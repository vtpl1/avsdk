package streammanager4

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/vtpl1/avsdk/av"
)

// Sentinel errors returned by StreamManager and Producer methods.
var (
	ErrProducerNotFound = errors.New("producer not found")
	ErrConsumerNotFound = errors.New("consumer not found")
)

// errProducerClosing is returned by Producer.AddConsumer when the producer has already
// started tearing down (its last consumer was concurrently removed). AddConsumer detects
// this and retries with a fresh producer rather than propagating the error to the caller.
var errProducerClosing = errors.New("producer closing")
var errConsumerClosing = errors.New("consumer closing")

// Option is a functional option for configuring a StreamManager.
type Option func(*StreamManager)

// StreamManager implements [av.StreamManager]. It lazily creates a [Producer] (and its
// underlying demuxer) when the first consumer for a given producerID is added, and
// tears it down when the last consumer for that producerID is removed.
type StreamManager struct {
	demuxerFactory av.DemuxerFactory
	demuxerRemover av.DemuxerRemover

	wg        sync.WaitGroup
	mu        sync.RWMutex
	producers map[string]*Producer
}

func New(demuxerFactory av.DemuxerFactory, demuxerRemover av.DemuxerRemover, opts ...Option) *StreamManager {
	m := &StreamManager{
		demuxerFactory: demuxerFactory,
		demuxerRemover: demuxerRemover,
		producers:      make(map[string]*Producer),
	}
	for _, o := range opts {
		o(m)
	}

	return m
}

// Start implements [av.StreamManager].
func (m *StreamManager) Start(ctx context.Context) error {
	m.wg.Add(1)
	sctx, cancel := context.WithCancel(ctx)

	go func(ctx context.Context, cancel context.CancelFunc) {
		defer m.wg.Done()
		defer cancel()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		defer func() {
			m.mu.RLock()
			ps := make(map[string]*Producer, len(m.producers))
			for producerID, p := range m.producers {
				ps[producerID] = p
			}
			m.mu.RUnlock()

			for _, p := range ps {
				_ = p.Stop()
			}
			m.mu.Lock()
			for producerID := range m.producers {
				delete(m.producers, producerID)
			}
			m.mu.Unlock()
		}()
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
					_ = p.Stop()
				}
				m.mu.Lock()
				for producerID := range inactive {
					delete(m.producers, producerID)
				}
				m.mu.Unlock()
			case <-sctx.Done():
				return
			}
		}
	}(sctx, cancel)

	return nil
}

// AddConsumer implements [av.StreamManager].
func (m *StreamManager) AddConsumer(ctx context.Context, producerID string, consumerID string,
	muxerFactory av.MuxerFactory,
	muxerRemover av.MuxerRemover,
	errChan chan<- error,
) error {
	var p *Producer
	if err := p.AddConsumer(ctx, producerID, consumerID, muxerFactory, muxerRemover, errChan); err != nil {
		return err
	}

	return nil
}

// RemoveConsumer implements [av.StreamManager].
func (m *StreamManager) RemoveConsumer(ctx context.Context, producerID string, consumerID string) error {
	panic("unimplemented")
}

// GetActiveProducersCount implements [av.StreamManager].
func (m *StreamManager) GetActiveProducersCount(ctx context.Context) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.producers)
}

// PauseProducer implements [av.StreamManager].
func (m *StreamManager) PauseProducer(ctx context.Context, producerID string) error {
	m.mu.RLock()
	p, ok := m.producers[producerID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("%s: %w", producerID, ErrProducerNotFound)
	}

	return p.Pause(ctx)
}

// ResumeProducer implements [av.StreamManager].
func (m *StreamManager) ResumeProducer(ctx context.Context, producerID string) error {
	m.mu.RLock()
	p, ok := m.producers[producerID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("%s: %w", producerID, ErrProducerNotFound)
	}

	return p.Resume(ctx)
}

// SignalStop implements [av.StreamManager].
func (m *StreamManager) SignalStop() bool {
	panic("unimplemented")
}

// WaitStop implements [av.StreamManager].
func (m *StreamManager) WaitStop() error {
	panic("unimplemented")
}

// Stop implements [av.StreamManager].
func (m *StreamManager) Stop() error {
	m.SignalStop()

	return m.WaitStop()
}
