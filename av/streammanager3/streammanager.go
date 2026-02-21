package streammanager3

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vtpl1/avsdk/av"
)

var (
	ErrProducerNotFound           = errors.New("producer not found")
	ErrProducerDemuxFactory       = errors.New("producer demux factory")
	ErrConsumerNotFound           = errors.New("consumer not found")
	ErrConsumerMuxFactory         = errors.New("consumer mux factory")
	ErrStreamManagerClosing       = errors.New("stream manager closing")
	ErrProducerClosing            = errors.New("producer closing")
	ErrProducerLastError          = errors.New("producer last error")
	ErrConsumerClosing            = errors.New("consumer closing")
	ErrConsumerAlreadyExists      = errors.New("consumer already exists")
	ErrCodecsNotAvailable         = errors.New("codecs not available")
	ErrStreamManagerNotStartedYet = errors.New("stream manager not started yet")
	ErrProducerNotStartedYet      = errors.New("producer not started yet")
	ErrConsumerNotStartedYet      = errors.New("consumer not started yet")
	ErrMuxerWritePacket           = errors.New("muxer write packet")
	ErrMuxerWriteHeader           = errors.New("muxer write header")
)

type Option func(*StreamManager)

type StreamManager struct {
	demuxerFactory av.DemuxerFactory
	demuxerRemover av.DemuxerRemover

	cancel         context.CancelFunc
	wg             sync.WaitGroup
	mu             sync.RWMutex
	alreadyClosing atomic.Bool
	started        atomic.Bool
	producers      map[string]*Producer

	producersToStart chan *Producer
}

func New(demuxerFactory av.DemuxerFactory, demuxerRemover av.DemuxerRemover, opts ...Option) *StreamManager {
	m := &StreamManager{
		demuxerFactory:   demuxerFactory,
		demuxerRemover:   demuxerRemover,
		producers:        make(map[string]*Producer),
		producersToStart: make(chan *Producer),
	}
	for _, o := range opts {
		o(m)
	}

	return m
}

func (m *StreamManager) Start(ctx context.Context) error {
	m.wg.Add(1)
	sctx, cancel := context.WithCancel(ctx)
	m.cancel = cancel

	go func(sctx context.Context, cancel context.CancelFunc) {
		defer m.wg.Done()
		defer cancel()
		defer func() {
			m.mu.RLock()
			inactive := make(map[string]*Producer, len(m.producers))
			for producerID, p := range m.producers {
				inactive[producerID] = p
			}
			m.mu.RUnlock()

			for _, p := range inactive {
				_ = p.Close()
			}

			m.mu.Lock()
			for producerID := range m.producers {
				delete(m.producers, producerID)
			}
			m.mu.Unlock()
		}()
		m.started.Store(true)
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				m.mu.RLock()
				inactive := make(map[string]*Producer, len(m.producers))
				for producerID, p := range m.producers {
					if p.ConsumerCount() == 0 {
						inactive[producerID] = p
					}
				}
				m.mu.RUnlock()
				for _, p := range inactive {
					_ = p.Close()
				}

				m.mu.Lock()
				for producerID := range inactive {
					delete(m.producers, producerID)
				}
				m.mu.Unlock()
			case <-sctx.Done():
				return
			case p, ok := <-m.producersToStart:
				if ok {
					err := p.Start(sctx)
					if err != nil {
						return
					}
				}
			}
		}
	}(sctx, cancel)

	return nil
}

func (m *StreamManager) AddConsumer(ctx context.Context, producerID string, consumerID string, muxerFactory av.MuxerFactory, muxerRemover av.MuxerRemover, errChan chan<- error) error {
	if m.alreadyClosing.Load() {
		return ErrStreamManagerClosing
	}
	if !m.started.Load() {
		return ErrStreamManagerNotStartedYet
	}
	for {
		m.mu.Lock()
		p, existed := m.producers[producerID]
		if !existed {
			p = NewProducer(producerID, m.demuxerFactory, m.demuxerRemover)
			m.producers[producerID] = p
		}
		m.mu.Unlock()
		if !existed {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case m.producersToStart <- p:
			}
		}

		if p.lastError() != nil {
			return fmt.Errorf("%s: %w", producerID, errors.Join(ErrProducerLastError, p.lastError()))
		}

		if err := p.AddConsumer(ctx, consumerID, muxerFactory, muxerRemover, errChan); err != nil {
			if errors.Is(err, ErrProducerClosing) {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			if errors.Is(err, ErrProducerNotStartedYet) {
				time.Sleep(10 * time.Millisecond)
				continue
			}

			return err
		}

		return nil
	}
}

func (m *StreamManager) RemoveConsumer(ctx context.Context, producerID string, consumerID string) error {
	m.mu.RLock()
	p, ok := m.producers[producerID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("%s: %w", producerID, ErrProducerNotFound)
	}

	return p.RemoveConsumer(ctx, consumerID)
}

func (m *StreamManager) GetActiveProducersCount(ctx context.Context) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.producers)
}

func (m *StreamManager) PauseProducer(ctx context.Context, producerID string) error {
	m.mu.RLock()
	p, ok := m.producers[producerID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("%s: %w", producerID, ErrProducerNotFound)
	}

	return p.Pause(ctx)
}

func (m *StreamManager) ResumeProducer(ctx context.Context, producerID string) error {
	m.mu.RLock()
	p, ok := m.producers[producerID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("%s: %w", producerID, ErrProducerNotFound)
	}

	return p.Resume(ctx)
}

func (m *StreamManager) SignalStop() bool {
	if !m.alreadyClosing.CompareAndSwap(false, true) {
		return false
	}

	if m.cancel != nil {
		m.cancel()
	}

	return true
}

func (m *StreamManager) WaitStop() error {
	m.wg.Wait()

	return nil
}

func (m *StreamManager) Stop() error {
	if !m.SignalStop() {
		return nil
	}

	return m.WaitStop()
}
