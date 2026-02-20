package streammanager4

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/vtpl1/avsdk/av"
)

type Producer struct {
	mu        sync.RWMutex
	consumers map[string]*consumer

	alreadyClosing atomic.Bool
	demuxer        av.DemuxCloser
}

func (p *Producer) ConsumerCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.consumers)
}

// Start implements [av.StartStopper].
func (p *Producer) Start(ctx context.Context) error {
	panic("unimplemented")
}

func (p *Producer) AddConsumer(ctx context.Context, producerID string, consumerID string, muxerFactory av.MuxerFactory, muxerRemover av.MuxerRemover, errChan chan<- error) error {
	panic("unimplemented")
}

func (p *Producer) Resume(ctx context.Context) error {
	p.mu.RLock()
	dmx := p.demuxer
	p.mu.RUnlock()
	if pauser, ok := dmx.(av.Pauser); ok {
		return pauser.Resume(ctx)
	}

	return nil
}

func (p *Producer) Pause(ctx context.Context) error {
	p.mu.RLock()
	dmx := p.demuxer
	p.mu.RUnlock()
	if pauser, ok := dmx.(av.Pauser); ok {
		return pauser.Pause(ctx)
	}

	return nil
}

// SignalStop implements [av.Stopper].
func (p *Producer) SignalStop() bool {
	p.alreadyClosing.Store(true)
	return true
}

// Stop implements [av.Stopper].
func (p *Producer) Stop() error {
	p.SignalStop()
	return p.WaitStop()
}

// WaitStop implements [av.Stopper].
func (p *Producer) WaitStop() error {
	return nil
}
