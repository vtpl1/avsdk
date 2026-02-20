package streammanager3_test

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"

	"github.com/vtpl1/avsdk/av"
	"github.com/vtpl1/avsdk/av/streammanager3"
)

// ---------------------------------------------------------------------------
// Helpers / mocks
// ---------------------------------------------------------------------------

// mockDemuxer emits a fixed sequence of packets then returns io.EOF.
type mockDemuxer struct {
	streams []av.Stream
	packets []av.Packet
	pos     int
	mu      sync.Mutex
	closed  bool
}

func (d *mockDemuxer) GetCodecs(_ context.Context) ([]av.Stream, error) { return d.streams, nil }

func (d *mockDemuxer) ReadPacket(_ context.Context) (av.Packet, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.pos >= len(d.packets) {
		return av.Packet{}, io.EOF
	}
	pkt := d.packets[d.pos]
	d.pos++
	return pkt, nil
}

func (d *mockDemuxer) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.closed = true
	return nil
}

// mockMuxer records every packet it receives.
type mockMuxer struct {
	mu       sync.Mutex
	received []av.Packet
	closed   bool
}

func (m *mockMuxer) WriteHeader(_ context.Context, _ []av.Stream) error { return nil }

func (m *mockMuxer) WritePacket(_ context.Context, pkt av.Packet) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.received = append(m.received, pkt)
	return nil
}

func (m *mockMuxer) WriteTrailer(_ context.Context) error { return nil }

func (m *mockMuxer) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *mockMuxer) Packets() []av.Packet {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]av.Packet, len(m.received))
	copy(cp, m.received)
	return cp
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
func TestInterfaceImplementations(t *testing.T) {
	var _ av.StreamManager = (*streammanager3.StreamManager)(nil)
	var _ av.Stopper = (*streammanager3.StreamManager)(nil)
	var _ av.Stopper = (*streammanager3.Producer)(nil)
}

func TestNew(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sm := streammanager3.New(ctx, func(_ context.Context, _ string) (av.DemuxCloser, error) {
		return nil, nil
	}, nil)

	if sm == nil {
		t.Fatal("expected non-nil StreamManager")
	}

	if sm.GetActiveProducersCount(ctx) != 0 {
		t.Fatalf("expected 0 active producers, got %d", sm.GetActiveProducersCount(ctx))
	}
}

func TestAddRemoveSingleConsumer(t *testing.T) {
	ctx := context.Background()

	dmx := &mockDemuxer{packets: []av.Packet{{FrameID: 1}}}
	sm := streammanager3.New(ctx, func(_ context.Context, _ string) (av.DemuxCloser, error) {
		return dmx, nil
	}, nil)

	mux := &mockMuxer{}
	errChan := make(chan error, 1)

	if err := sm.AddConsumer(ctx, "prod1", "cons1",
		func(_ context.Context, _, _ string) (av.MuxCloser, error) { return mux, nil },
		nil,
		errChan,
	); err != nil {
		t.Fatalf("AddConsumer: %v", err)
	}

	if sm.GetActiveProducersCount(ctx) != 1 {
		t.Fatalf("expected 1 active producer, got %d", sm.GetActiveProducersCount(ctx))
	}

	if err := sm.RemoveConsumer(ctx, "prod1", "cons1"); err != nil {
		t.Fatalf("RemoveConsumer: %v", err)
	}

	if sm.GetActiveProducersCount(ctx) != 0 {
		t.Fatalf("expected 0 active producers after removal, got %d", sm.GetActiveProducersCount(ctx))
	}
}

func TestMultipleConsumersSameProducer(t *testing.T) {
	ctx := context.Background()

	dmx := &mockDemuxer{}
	sm := streammanager3.New(ctx, func(_ context.Context, _ string) (av.DemuxCloser, error) {
		return dmx, nil
	}, nil)

	errChan := make(chan error, 10)
	consumerIDs := []string{"cons1", "cons2", "cons3"}

	for _, cID := range consumerIDs {
		cID := cID
		mux := &mockMuxer{}
		if err := sm.AddConsumer(ctx, "prod1", cID,
			func(_ context.Context, _, _ string) (av.MuxCloser, error) { return mux, nil },
			nil,
			errChan,
		); err != nil {
			t.Fatalf("AddConsumer %s: %v", cID, err)
		}
	}

	// All three consumers share one producer.
	if sm.GetActiveProducersCount(ctx) != 1 {
		t.Fatalf("expected 1 active producer, got %d", sm.GetActiveProducersCount(ctx))
	}

	// Remove consumers one by one; producer must stay until the last one leaves.
	for i, cID := range consumerIDs {
		if err := sm.RemoveConsumer(ctx, "prod1", cID); err != nil {
			t.Fatalf("RemoveConsumer %s: %v", cID, err)
		}
		remaining := len(consumerIDs) - i - 1
		expected := 0
		if remaining > 0 {
			expected = 1
		}
		if got := sm.GetActiveProducersCount(ctx); got != expected {
			t.Fatalf("after removing %s: expected %d producer(s), got %d", cID, expected, got)
		}
	}
}

func TestPacketFanout(t *testing.T) {
	ctx := context.Background()

	const numPackets = 5
	pkts := make([]av.Packet, numPackets)
	for i := range pkts {
		pkts[i] = av.Packet{FrameID: int64(i + 1)}
	}

	dmx := &mockDemuxer{packets: pkts}
	sm := streammanager3.New(ctx, func(_ context.Context, _ string) (av.DemuxCloser, error) {
		return dmx, nil
	}, nil)

	mux1 := &mockMuxer{}
	mux2 := &mockMuxer{}
	errChan := make(chan error, 10)

	for _, pair := range []struct {
		cID string
		mux *mockMuxer
	}{{"cons1", mux1}, {"cons2", mux2}} {
		p := pair
		if err := sm.AddConsumer(ctx, "prod1", p.cID,
			func(_ context.Context, _, _ string) (av.MuxCloser, error) { return p.mux, nil },
			nil,
			errChan,
		); err != nil {
			t.Fatalf("AddConsumer %s: %v", p.cID, err)
		}
	}

	// Wait for all packets (including io.EOF forwarded as error) to be processed.
	// The demuxer returns io.EOF after numPackets, which is sent to errChan.
	var eofCount int
	for err := range errChan {
		if errors.Is(err, io.EOF) {
			eofCount++
			if eofCount == 2 { // one per consumer
				break
			}
		}
	}

	if err := sm.RemoveConsumer(ctx, "prod1", "cons1"); err != nil {
		t.Fatalf("RemoveConsumer cons1: %v", err)
	}
	if err := sm.RemoveConsumer(ctx, "prod1", "cons2"); err != nil {
		t.Fatalf("RemoveConsumer cons2: %v", err)
	}

	for _, pair := range []struct {
		name string
		mux  *mockMuxer
	}{{"mux1", mux1}, {"mux2", mux2}} {
		got := pair.mux.Packets()
		if len(got) != numPackets {
			t.Errorf("%s: expected %d packets, got %d", pair.name, numPackets, len(got))
		}
		for i, p := range got {
			if p.FrameID != pkts[i].FrameID {
				t.Errorf("%s packet %d: FrameID want %d got %d", pair.name, i, pkts[i].FrameID, p.FrameID)
			}
		}
	}
}

func TestRemoveUnknownConsumer(t *testing.T) {
	ctx := context.Background()
	sm := streammanager3.New(ctx, nil, nil)
	err := sm.RemoveConsumer(ctx, "no-such-prod", "no-such-cons")
	if !errors.Is(err, streammanager3.ErrProducerNotFound) {
		t.Fatalf("expected ErrProducerNotFound, got %v", err)
	}
}

// errMuxer is a mockMuxer whose WritePacket always returns a configurable error.
type errMuxer struct {
	mockMuxer
	writeErr error
}

func (m *errMuxer) WritePacket(_ context.Context, _ av.Packet) error {
	return m.writeErr
}

// TestDeadConsumerAutoRemoval verifies that when a consumer's write goroutine exits due
// to a muxer error the consumer is automatically removed from its producer, and the
// producer itself is torn down once no consumers remain.
func TestDeadConsumerAutoRemoval(t *testing.T) {
	ctx := context.Background()

	writeErr := errors.New("write failed")
	dmx := &mockDemuxer{packets: []av.Packet{{FrameID: 1}, {FrameID: 2}}}
	sm := streammanager3.New(ctx, func(_ context.Context, _ string) (av.DemuxCloser, error) {
		return dmx, nil
	}, nil)

	errChan := make(chan error, 4)
	mux := &errMuxer{writeErr: writeErr}

	// muxRemoved is closed by the muxerRemover callback; it is the authoritative
	// signal that the full auto-removal teardown sequence has completed.
	muxRemoved := make(chan struct{})

	if err := sm.AddConsumer(ctx, "prod1", "cons1",
		func(_ context.Context, _, _ string) (av.MuxCloser, error) { return mux, nil },
		func(_ context.Context, _, _ string) error { close(muxRemoved); return nil },
		errChan,
	); err != nil {
		t.Fatalf("AddConsumer: %v", err)
	}

	// The demuxer EOF and the muxer writeErr both land in errChan; the EOF may
	// arrive first because runLoop can exhaust the demuxer before consumer.run()
	// processes the first packet. Drain until we see writeErr.
	for {
		if err := <-errChan; errors.Is(err, writeErr) {
			break
		}
	}

	// Block until muxerRemover fires, confirming the full auto-removal teardown ran.
	<-muxRemoved

	if got := sm.GetActiveProducersCount(ctx); got != 0 {
		t.Fatalf("expected 0 active producers after dead consumer auto-removal, got %d", got)
	}
}

func TestDemuxerFactoryError(t *testing.T) {
	ctx := context.Background()
	factoryErr := errors.New("source unavailable")
	sm := streammanager3.New(ctx, func(_ context.Context, _ string) (av.DemuxCloser, error) {
		return nil, factoryErr
	}, nil)

	errChan := make(chan error, 1)
	err := sm.AddConsumer(ctx, "prod1", "cons1",
		func(_ context.Context, _, _ string) (av.MuxCloser, error) { return &mockMuxer{}, nil },
		nil,
		errChan,
	)
	if !errors.Is(err, factoryErr) {
		t.Fatalf("expected factoryErr, got %v", err)
	}
	// Producer must not be registered when factory fails.
	if sm.GetActiveProducersCount(ctx) != 0 {
		t.Fatalf("expected 0 producers after factory error, got %d", sm.GetActiveProducersCount(ctx))
	}
}
