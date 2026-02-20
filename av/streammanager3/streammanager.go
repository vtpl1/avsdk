package streammanager3

import (
	"context"
	"errors"
	"fmt"
	"sync"

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

// Option is a functional option for configuring a StreamManager.
type Option func(*StreamManager)

// StreamManager implements [av.StreamManager]. It lazily creates a [Producer] (and its
// underlying demuxer) when the first consumer for a given producerID is added, and
// tears it down when the last consumer for that producerID is removed.
type StreamManager struct {
	demuxerFactory av.DemuxerFactory
	demuxerRemover av.DemuxerRemover

	mu        sync.RWMutex
	producers map[string]*Producer

	ctx    context.Context //nolint:containedctx
	cancel context.CancelFunc
}

// New returns a StreamManager that uses demuxerFactory to open sources and
// demuxerRemover to clean them up. opts are applied in order after construction.
func New(ctx context.Context, demuxerFactory av.DemuxerFactory, demuxerRemover av.DemuxerRemover, opts ...Option) *StreamManager {
	cCtx, cancel := context.WithCancel(ctx)
	m := &StreamManager{
		demuxerFactory: demuxerFactory,
		demuxerRemover: demuxerRemover,
		producers:      make(map[string]*Producer),
		ctx:            cCtx,
		cancel:         cancel,
	}
	for _, o := range opts {
		o(m)
	}

	return m
}

// AddConsumer implements [av.StreamManager].
// If no live producer exists for producerID, one is created. If the producer found in
// the map is concurrently shutting down (errProducerClosing), its stale map entry is
// removed and the call retries with a fresh producer.
// If creating the first consumer fails, the freshly created producer is removed from
// the map.
func (m *StreamManager) AddConsumer(ctx context.Context, producerID, consumerID string,
	muxerFactory av.MuxerFactory,
	muxerRemover av.MuxerRemover,
	errChan chan<- error,
) error {
	for {
		m.mu.Lock()
		p, existed := m.producers[producerID]
		if !existed {
			p = NewProducer(m.ctx, producerID, m.demuxerFactory, nil) //nolint:contextcheck
			p.removeMe = m.makeProducerRemover(p, producerID)
			m.producers[producerID] = p
		}
		m.mu.Unlock()

		err := p.AddConsumer(ctx, producerID, consumerID, muxerFactory, muxerRemover, errChan)
		if err != nil {
			if errors.Is(err, errProducerClosing) {
				// Producer is shutting down concurrently. Remove its stale map entry
				// (if it is still there) and loop to get or create a fresh one.
				m.mu.Lock()
				if cur, ok := m.producers[producerID]; ok && cur == p {
					delete(m.producers, producerID)
				}
				m.mu.Unlock()

				continue
			}

			if !existed {
				// The producer was just created but its first consumer failed; remove it.
				m.mu.Lock()
				if cur, ok := m.producers[producerID]; ok && cur == p {
					delete(m.producers, producerID)
				}
				m.mu.Unlock()
			}

			return err
		}

		return nil
	}
}

// RemoveConsumer implements [av.StreamManager].
func (m *StreamManager) RemoveConsumer(ctx context.Context, producerID, consumerID string) error {
	m.mu.RLock()
	p, ok := m.producers[producerID]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("%s: %w", producerID, ErrProducerNotFound)
	}

	return p.RemoveConsumer(ctx, producerID, consumerID)
}

// SignalStop implements [av.StreamManager].
func (m *StreamManager) SignalStop() bool {
	if m.cancel != nil {
		m.cancel()
	}

	return true
}

// WaitStop implements [av.StreamManager]. It waits for all active producer goroutines
// to finish. Call SignalStop (or Stop) first to trigger shutdown of the read loops.
func (m *StreamManager) WaitStop() error {
	// Snapshot the producer list so we can wait without holding the lock.
	m.mu.RLock()
	ps := make([]*Producer, 0, len(m.producers))
	for _, p := range m.producers {
		ps = append(ps, p)
	}
	m.mu.RUnlock()

	for _, p := range ps {
		_ = p.WaitStop()
	}

	return nil
}

// Stop implements [av.StreamManager].
func (m *StreamManager) Stop() error {
	m.SignalStop()

	return m.WaitStop()
}

// GetActiveProducersCount implements [av.StreamManager].
func (m *StreamManager) GetActiveProducersCount(_ context.Context) int {
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

// makeProducerRemover returns a ProducerRemover that only removes p from the map when
// it is still the current entry for producerID. The identity check prevents a closing
// producer from inadvertently evicting a replacement producer that was registered while
// the closing producer's teardown was still in progress.
func (m *StreamManager) makeProducerRemover(p *Producer, producerID string) av.ProducerRemover {
	return func(ctx context.Context, _ string) error {
		m.mu.Lock()
		if cur, ok := m.producers[producerID]; ok && cur == p {
			delete(m.producers, producerID)
		}
		m.mu.Unlock()

		if m.demuxerRemover != nil {
			return m.demuxerRemover(ctx, producerID)
		}

		return nil
	}
}
