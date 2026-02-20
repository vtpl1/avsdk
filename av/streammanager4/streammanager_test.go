package streammanager4_test

import (
	"context"
	"testing"

	"github.com/vtpl1/avsdk/av"
	"github.com/vtpl1/avsdk/av/streammanager4"
)

func TestInterfaceImplementations(t *testing.T) {
	var _ av.StreamManager = (*streammanager4.StreamManager)(nil)
	var _ av.StartStopper = (*streammanager4.StreamManager)(nil)
	var _ av.StartStopper = (*streammanager4.Producer)(nil)
}

func TestNew(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sm := streammanager4.New(func(_ context.Context, _ string) (av.DemuxCloser, error) {
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
