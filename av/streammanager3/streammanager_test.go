package streammanager3_test

import (
	"context"
	"testing"

	"github.com/vtpl1/avsdk/av"
	"github.com/vtpl1/avsdk/av/streammanager3"
)

func TestInterfaceImplementations(t *testing.T) {
	var _ av.StreamManager = (*streammanager3.StreamManager)(nil)
	var _ av.StartStopper = (*streammanager3.StreamManager)(nil)
	var _ av.DemuxCloser = (*streammanager3.Producer)(nil)
	var _ av.MuxCloser = (*streammanager3.Consumer)(nil)
	var _ av.CodecChanger = (*streammanager3.Consumer)(nil)
}

func TestNew(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sm := streammanager3.New(func(_ context.Context, _ string) (av.DemuxCloser, error) {
		return nil, nil
	}, nil)
	sm.Start(ctx)

	if sm == nil {
		t.Fatal("expected non-nil StreamManager")
	}

	if sm.GetActiveProducersCount(ctx) != 0 {
		t.Fatalf("expected 0 active producers, got %d", sm.GetActiveProducersCount(ctx))
	}
}
