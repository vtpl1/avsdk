package streammanager3

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vtpl1/avsdk/av"
)

type Consumer struct {
	consumerID   string
	muxerFactory av.MuxerFactory
	muxerRemover av.MuxerRemover
	errCh        chan<- error

	cancel         context.CancelFunc
	wg             sync.WaitGroup
	alreadyClosing atomic.Bool
	inactive       atomic.Bool
	writeOnce      sync.Once

	mu      sync.RWMutex
	streams []av.Stream

	lastError error
	streamsCh chan []av.Stream
	queue     chan av.Packet
}

func NewConsumer(consumerID string,
	muxerFactory av.MuxerFactory,
	muxerRemover av.MuxerRemover, errCh chan<- error,
) *Consumer {
	m := &Consumer{
		consumerID:   consumerID,
		muxerFactory: muxerFactory,
		muxerRemover: muxerRemover,
		errCh:        errCh,
		streamsCh:    make(chan []av.Stream),
		queue:        make(chan av.Packet, 50),
	}

	return m
}

func (m *Consumer) Start(ctx context.Context) error {
	sctx, cancel := context.WithCancel(ctx)
	m.cancel = cancel
	m.wg.Add(1)
	go func(ctx context.Context, cancel context.CancelFunc) {
		defer m.wg.Done()
		defer cancel()
		defer func(_ context.Context) {
			if m.muxerRemover != nil {
				ctxDetached := context.WithoutCancel(ctx)
				ctxTimeout, cancel := context.WithTimeout(ctxDetached, 5*time.Second)
				defer cancel()
				_ = m.muxerRemover(ctxTimeout, m.consumerID)
			}
		}(ctx)
		defer m.inactive.Store(true)

		select {
		case <-ctx.Done():
			m.setLastError(ctx.Err())

			return
		case _, ok := <-m.streamsCh:
			if !ok {
				return
			}
			muxer, err := m.muxerFactory(ctx, m.consumerID)
			if err != nil {
				m.setLastError(errors.Join(ErrConsumerMuxFactory, err))

				return
			}
			defer muxer.Close()
			m.mu.RLock()
			streams := m.streams
			m.mu.RUnlock()
			if err := muxer.WriteHeader(ctx, streams); err != nil {
				m.setLastError(errors.Join(ErrMuxerWriteHeader, err))

				return
			}
			for {
				select {
				case <-ctx.Done():
					return
				case pkt, ok := <-m.queue:
					if !ok {
						return
					}

					if err := muxer.WritePacket(ctx, pkt); err != nil {
						m.setLastError(errors.Join(ErrMuxerWritePacket, err))

						return
					}
				}
			}
		}
	}(sctx, cancel)

	return nil
}

func (m *Consumer) WriteHeader(ctx context.Context, streams []av.Stream) error {
	m.writeOnce.Do(func() {
		defer close(m.streamsCh)
		if len(streams) == 0 {
			m.setLastError(ErrCodecsNotAvailable)

			return
		}
		_ = m.WriteCodecChange(ctx, streams)
		select {
		case <-ctx.Done():
		case m.streamsCh <- streams:
		}
	})

	return nil
}

func (m *Consumer) WritePacket(ctx context.Context, pkt av.Packet) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case m.queue <- pkt:
	}
	return nil
}

func (m *Consumer) WriteCodecChange(ctx context.Context, changed []av.Stream) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.streams = changed

	return nil
}

func (m *Consumer) Close() error {
	if !m.alreadyClosing.CompareAndSwap(false, true) {
		return nil
	}
	m.inactive.Store(true)
	if m.cancel != nil {
		m.cancel()
	}
	m.wg.Wait()

	return nil
}

func (m *Consumer) WriteTrailer(ctx context.Context) error {
	return nil
}

func (m *Consumer) setLastError(err error) {
	if err == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	m.lastError = err
	if m.errCh == nil {
		return
	}
	select {
	case m.errCh <- err:
	default:
	}
	m.inactive.Store(true)
}

func (m *Consumer) LastError() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.lastError
}
