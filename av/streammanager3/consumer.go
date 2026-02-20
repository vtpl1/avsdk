package streammanager3

import (
	"context"
	"sync"

	"github.com/vtpl1/avsdk/av"
)

const consumerChanSize = 64

type consumer struct {
	consumerID   string
	mux          av.MuxCloser
	muxerRemover av.MuxerRemover
	pktChan      chan av.Packet
	errChan      chan<- error
	// onDead is called in a new goroutine when run exits early due to a write
	// error, triggering automatic removal from the producer. It is never called
	// on a normal exit (i.e. when pktChan is closed by stop).
	onDead func()
	ctx    context.Context //nolint:containedctx
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func newConsumer(ctx context.Context, consumerID string, mux av.MuxCloser, muxerRemover av.MuxerRemover, errChan chan<- error) *consumer {
	cCtx, cancel := context.WithCancel(ctx)

	return &consumer{
		consumerID:   consumerID,
		mux:          mux,
		muxerRemover: muxerRemover,
		pktChan:      make(chan av.Packet, consumerChanSize),
		errChan:      errChan,
		ctx:          cCtx,
		cancel:       cancel,
	}
}

// start calls WriteHeader with the current stream list then spawns the write goroutine.
func (c *consumer) start(streams []av.Stream) error {
	if err := c.mux.WriteHeader(c.ctx, streams); err != nil {
		return err
	}
	c.wg.Add(1)
	go c.run()

	return nil
}

// run drains pktChan until it is closed, writing each packet to the muxer.
// It uses for-range so that all buffered packets are processed even after stop closes
// the channel — avoiding the race where ctx.Done() fires before the buffer is drained.
// On a write error, the error is forwarded to errChan and onDead is called in a new
// goroutine to trigger automatic removal of this consumer from its producer.
func (c *consumer) run() {
	defer c.wg.Done()
	for pkt := range c.pktChan {
		if pkt.NewCodecs != nil {
			if cc, ok := c.mux.(av.CodecChanger); ok {
				if err := cc.WriteCodecChange(c.ctx, pkt.NewCodecs); err != nil {
					select {
					case c.errChan <- err:
					default:
					}
					go c.onDead()

					return
				}
			}
		}
		if err := c.mux.WritePacket(c.ctx, pkt); err != nil {
			select {
			case c.errChan <- err:
			default:
			}
			go c.onDead()

			return
		}
	}
}

// stop closes pktChan (signalling the goroutine to drain and exit), waits for it,
// then finalises the muxer and calls muxerRemover.
// producerID is forwarded to muxerRemover as required by MuxerRemover's signature.
func (c *consumer) stop(ctx context.Context, producerID string) {
	close(c.pktChan)
	c.wg.Wait()
	c.cancel()
	_ = c.mux.WriteTrailer(ctx)
	_ = c.mux.Close()
	if c.muxerRemover != nil {
		_ = c.muxerRemover(ctx, producerID, c.consumerID)
	}
}
