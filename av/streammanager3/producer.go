package streammanager3

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vtpl1/avsdk/av"
)

type Producer struct {
	producerID     string
	demuxerFactory av.DemuxerFactory
	demuxerRemover av.DemuxerRemover

	cancel           context.CancelFunc
	wg               sync.WaitGroup
	mu               sync.RWMutex
	alreadyClosing   atomic.Bool
	started          atomic.Bool
	consumers        map[string]*Consumer
	consumersToStart chan *Consumer

	demuxer   av.DemuxCloser
	streams   []av.Stream
	codecsErr error
	codecsCh  chan struct{}
}

func (m *Producer) Close() error {
	if !m.alreadyClosing.CompareAndSwap(false, true) {
		return nil
	}

	if m.cancel != nil {
		m.cancel()
	}
	m.wg.Wait()

	return nil
}

func (m *Producer) GetCodecs(ctx context.Context) ([]av.Stream, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-m.codecsCh:
		return m.streams, m.codecsErr
	}
}

func (m *Producer) ReadPacket(ctx context.Context) (av.Packet, error) {
	return m.demuxer.ReadPacket(ctx)
}

func NewProducer(producerID string,
	demuxerFactory av.DemuxerFactory,
	demuxerRemover av.DemuxerRemover,
) *Producer {
	m := &Producer{
		producerID:       producerID,
		demuxerFactory:   demuxerFactory,
		demuxerRemover:   demuxerRemover,
		consumersToStart: make(chan *Consumer),
		codecsCh:         make(chan struct{}),
		consumers:        make(map[string]*Consumer),
	}

	return m
}

func (m *Producer) ConsumerCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.consumers)
}

func (m *Producer) Start(ctx context.Context) error {
	m.wg.Add(1)
	sctx, cancel := context.WithCancel(ctx)
	m.cancel = cancel
	go func(ctx context.Context, cancel context.CancelFunc) {
		defer m.wg.Done()
		defer cancel()
		m.started.Store(true)
		demuxer, err := m.demuxerFactory(ctx, m.producerID)
		if err != nil {
			m.setLastCodecError(err)

			return
		}
		m.demuxer = demuxer
		defer m.demuxer.Close()
		defer func(ctx context.Context) {
			if m.demuxerRemover != nil {
				ctxDetached := context.WithoutCancel(ctx)
				ctxTimeout, cancel := context.WithTimeout(ctxDetached, 5*time.Second)
				defer cancel()
				_ = m.demuxerRemover(ctxTimeout, m.producerID)
			}
		}(ctx)
		defer func() {
			m.mu.RLock()
			inactive := make(map[string]*Consumer, len(m.consumers))
			for consumerID, c := range m.consumers {
				inactive[consumerID] = c
			}
			m.mu.RUnlock()

			for _, c := range inactive {
				_ = c.Close()
			}

			m.mu.Lock()
			for consumerID := range m.consumers {
				delete(m.consumers, consumerID)
			}
			m.mu.Unlock()
		}()
		streams, err := m.demuxer.GetCodecs(ctx)
		if err != nil {
			m.setLastCodecError(err)

			return
		}
		m.mu.Lock()
		m.streams = streams
		close(m.codecsCh)
		m.mu.Unlock()

		m.wg.Add(1)
		go func() {
			defer m.wg.Done()
			m.readWriteLoop(ctx)
		}()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case c, ok := <-m.consumersToStart:
				if !ok {
					continue
				}
				_ = c.Start(ctx)
			case <-ticker.C:
				m.mu.RLock()
				inactive := make(map[string]*Consumer, len(m.consumers))
				for consumerID, c := range m.consumers {
					if !c.inactive.Load() {
						continue
					}
					inactive[consumerID] = c
				}
				m.mu.RUnlock()
				for _, c := range inactive {
					_ = c.Close()
				}

				m.mu.Lock()
				for consumerID := range inactive {
					delete(m.consumers, consumerID)
				}
				m.mu.Unlock()

			case <-ctx.Done():
				return
			}
		}
	}(sctx, cancel)

	return nil
}

func (m *Producer) readWriteLoop(ctx context.Context) {
	FPS := 2500
	fpsLimitTicker := time.NewTicker(time.Second / time.Duration(FPS))
	defer fpsLimitTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-fpsLimitTicker.C:
			pkt, err := m.ReadPacket(ctx)
			if err != nil {
				return
			}
			m.mu.RLock()
			active := make(map[string]*Consumer, len(m.consumers))
			for consumerID, c := range m.consumers {
				if c.LastError() != nil {
					continue
				}
				if c.inactive.Load() {
					continue
				}
				active[consumerID] = c
			}
			m.mu.RUnlock()
			for _, c := range active {
				_ = c.WritePacket(ctx, pkt)
			}
		}
	}
}

func (m *Producer) setLastCodecError(err error) {
	if err == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.codecsErr = err
	close(m.codecsCh)
}

func (m *Producer) lastError() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.codecsErr
}

func (m *Producer) AddConsumer(ctx context.Context, consumerID string, muxerFactory av.MuxerFactory, muxerRemover av.MuxerRemover, errChan chan<- error) error {
	if m.alreadyClosing.Load() {
		return ErrProducerClosing
	}
	if !m.started.Load() {
		return ErrProducerNotStartedYet
	}
	if m.lastError() != nil {
		return m.lastError()
	}
	m.mu.Lock()
	_, existed := m.consumers[consumerID]
	if existed {
		m.mu.Unlock()
		return ErrConsumerAlreadyExists
	}
	c := NewConsumer(consumerID, muxerFactory, muxerRemover, errChan)
	m.consumers[consumerID] = c
	m.mu.Unlock()
	streams, err := m.GetCodecs(ctx)
	if err != nil {
		c.setLastError(errors.Join(ErrCodecsNotAvailable, err))

		return err
	}
	select {
	case <-ctx.Done():
		c.inactive.Store(true)
		return ErrProducerClosing
	case m.consumersToStart <- c:
	}
	return c.WriteHeader(ctx, streams)
}

func (m *Producer) RemoveConsumer(ctx context.Context, consumerID string) error {
	m.mu.RLock()
	consumer, exists := m.consumers[consumerID]
	m.mu.RUnlock()
	if exists {
		_ = consumer.Close()
	}

	return nil
}

func (m *Producer) Pause(ctx context.Context) error {
	if m.alreadyClosing.Load() {
		return ErrProducerClosing
	}
	if !m.started.Load() {
		return ErrProducerNotStartedYet
	}
	m.mu.RLock()
	dmx := m.demuxer
	m.mu.RUnlock()
	if pauser, ok := dmx.(av.Pauser); ok {
		return pauser.Pause(ctx)
	}

	return nil
}

func (m *Producer) Resume(ctx context.Context) error {
	if m.alreadyClosing.Load() {
		return ErrProducerClosing
	}
	if !m.started.Load() {
		return ErrProducerNotStartedYet
	}
	m.mu.RLock()
	dmx := m.demuxer
	m.mu.RUnlock()
	if pauser, ok := dmx.(av.Pauser); ok {
		return pauser.Resume(ctx)
	}

	return nil
}
