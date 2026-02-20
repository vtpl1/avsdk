package streammanager4

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
}

// AddConsumer implements [av.StreamManager].
func (m *StreamManager) AddConsumer(ctx context.Context, producerID string, consumerID string,
	muxerFactory av.MuxerFactory,
	muxerRemover av.MuxerRemover,
	errChan chan<- error,
) error {
	panic("unimplemented")
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

// Stop implements [av.StreamManager].
func (m *StreamManager) Stop() error {
	panic("unimplemented")
}

// WaitStop implements [av.StreamManager].
func (m *StreamManager) WaitStop() error {
	panic("unimplemented")
}
