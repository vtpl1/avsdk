package streammanager4

import (
	"context"
	"sync"

	"github.com/vtpl1/avsdk/av"
)

type Producer struct {
	mu        sync.RWMutex
	consumers map[string]*consumer

	demuxer av.DemuxCloser
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
	panic("unimplemented")
}

// Stop implements [av.Stopper].
func (p *Producer) Stop() error {
	panic("unimplemented")
}

// WaitStop implements [av.Stopper].
func (p *Producer) WaitStop() error {
	panic("unimplemented")
}
