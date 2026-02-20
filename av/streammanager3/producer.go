package streammanager3

import (
	"context"
	"fmt"
	"sync"

	"github.com/vtpl1/avsdk/av"
)

// Producer manages a single demuxer and fans its packets out to one or more consumers.
// The demuxer is created lazily when the first consumer is added, and torn down when the
// last consumer is removed.
type Producer struct {
	producerID     string
	demuxerFactory av.DemuxerFactory
	removeMe       av.ProducerRemover

	mu        sync.RWMutex
	consumers map[string]*consumer
	// closing is set to true (under mu) when the last consumer is removed. Any
	// concurrent AddConsumer call will see this flag and return errProducerClosing
	// instead of re-using a producer that is already mid-teardown.
	closing bool

	demuxer av.DemuxCloser
	streams []av.Stream

	ctx    context.Context //nolint:containedctx
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// SignalStop implements [av.Stopper].
func (p *Producer) SignalStop() bool {
	if p.cancel != nil {
		p.cancel()
	}

	return true
}

// Stop implements [av.Stopper].
func (p *Producer) Stop() error {
	p.SignalStop()

	return p.WaitStop()
}

// WaitStop implements [av.Stopper].
func (p *Producer) WaitStop() error {
	p.wg.Wait()

	return nil
}

// NewProducer creates a Producer that will use demuxerFactory to open the source on demand.
// removeMe is called (with producerID) after the last consumer is removed and the demuxer
// has been closed, allowing the caller to deregister the producer from any registry.
// Pass nil for removeMe if no deregistration callback is needed.
func NewProducer(ctx context.Context, producerID string, demuxerFactory av.DemuxerFactory, removeMe av.ProducerRemover) *Producer {
	cCtx, cancel := context.WithCancel(ctx)

	return &Producer{
		producerID:     producerID,
		demuxerFactory: demuxerFactory,
		removeMe:       removeMe,
		consumers:      make(map[string]*consumer),
		ctx:            cCtx,
		cancel:         cancel,
	}
}

// AddConsumer opens a muxer via muxerFactory and attaches it to this producer.
// If this is the first consumer, the demuxer is created and the read loop is started.
// Errors from the muxer write loop are sent to errChan.
// Returns errProducerClosing if the producer has already started tearing down.
func (p *Producer) AddConsumer(ctx context.Context, producerID, consumerID string,
	muxerFactory av.MuxerFactory,
	muxerRemover av.MuxerRemover,
	errChan chan<- error,
) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Reject new consumers if teardown has already begun.
	if p.closing {
		return errProducerClosing
	}

	// Lazy demuxer init on first consumer.
	isFirst := len(p.consumers) == 0
	if isFirst {
		dmx, err := p.demuxerFactory(ctx, producerID)
		if err != nil {
			return err
		}
		streams, err := dmx.GetCodecs(ctx)
		if err != nil {
			_ = dmx.Close()

			return err
		}
		p.demuxer = dmx
		p.streams = streams
	}

	mux, err := muxerFactory(ctx, producerID, consumerID)
	if err != nil {
		if isFirst {
			_ = p.demuxer.Close()
			p.demuxer = nil
		}

		return err
	}

	//nolint:contextcheck // p.ctx is intentional: the consumer must outlive the AddConsumer call.
	c := newConsumer(p.ctx, consumerID, mux, muxerRemover, errChan)
	// Wire auto-removal: when the write goroutine exits due to a muxer error it
	// calls RemoveConsumer so the producer stops fanning packets to a dead sink.
	// context.Background() is used so cleanup runs even if the caller's ctx is gone.
	// A concurrent explicit RemoveConsumer for the same ID returns ErrConsumerNotFound,
	// which is silently ignored here — both paths call the same stop() sequence.
	c.onDead = func() { _ = p.RemoveConsumer(context.Background(), p.producerID, consumerID) }
	if err = c.start(p.streams); err != nil {
		_ = mux.Close()
		if isFirst {
			_ = p.demuxer.Close()
			p.demuxer = nil
		}

		return err
	}

	p.consumers[consumerID] = c

	// Start the read loop only after the first consumer is confirmed.
	if isFirst {
		p.wg.Add(1)
		go p.runLoop()
	}

	return nil
}

// RemoveConsumer stops the consumer identified by consumerID and tears down its muxer.
// When the last consumer is removed the demuxer is also closed and removeMe is called.
func (p *Producer) RemoveConsumer(ctx context.Context, producerID, consumerID string) error {
	p.mu.Lock()
	c, ok := p.consumers[consumerID]
	if !ok {
		p.mu.Unlock()

		return fmt.Errorf("%s: %w", consumerID, ErrConsumerNotFound)
	}
	delete(p.consumers, consumerID)
	empty := len(p.consumers) == 0
	if empty {
		// Signal closing before releasing the lock so that any concurrent AddConsumer
		// that acquires the lock next sees the flag and returns errProducerClosing.
		p.closing = true
	}
	p.mu.Unlock()

	c.stop(ctx, producerID)

	if empty {
		p.cancel()
		p.wg.Wait()
		_ = p.demuxer.Close()
		if p.removeMe != nil {
			return p.removeMe(ctx, producerID)
		}
	}

	return nil
}

// Pause pauses packet delivery if the underlying demuxer supports av.Pauser.
func (p *Producer) Pause(ctx context.Context) error {
	p.mu.RLock()
	dmx := p.demuxer
	p.mu.RUnlock()
	if pauser, ok := dmx.(av.Pauser); ok {
		return pauser.Pause(ctx)
	}

	return nil
}

// Resume resumes packet delivery if the underlying demuxer supports av.Pauser.
func (p *Producer) Resume(ctx context.Context) error {
	p.mu.RLock()
	dmx := p.demuxer
	p.mu.RUnlock()
	if pauser, ok := dmx.(av.Pauser); ok {
		return pauser.Resume(ctx)
	}

	return nil
}

// runLoop reads packets from the demuxer and fans each one out to all active consumers.
// It exits when the producer context is cancelled or the demuxer returns an error.
func (p *Producer) runLoop() {
	defer p.wg.Done()

	// Capture the demuxer under the read lock. p.demuxer is set before runLoop is
	// started (inside AddConsumer while p.mu is held) and is never reassigned while
	// the loop runs. Reading it under the lock satisfies the race detector, which
	// cannot infer the happens-before relationship from goroutine creation alone.
	p.mu.RLock()
	dmx := p.demuxer
	p.mu.RUnlock()

	for {
		pkt, err := dmx.ReadPacket(p.ctx)
		if err != nil {
			p.mu.RLock()
			for _, c := range p.consumers {
				select {
				case c.errChan <- err:
				default:
				}
			}
			p.mu.RUnlock()

			return
		}

		// Keep p.streams current so that consumers added after a mid-stream codec
		// change receive the correct stream list in their WriteHeader call.
		if pkt.NewCodecs != nil {
			p.mu.Lock()
			p.streams = pkt.NewCodecs
			p.mu.Unlock()
		}

		p.mu.RLock()
		for _, c := range p.consumers {
			select {
			case c.pktChan <- pkt:
			default:
				// Slow consumer: drop packet rather than blocking the read loop.
			}
		}
		p.mu.RUnlock()
	}
}
