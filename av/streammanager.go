package av

import "context"

type StreamManager interface {
	GetActiveProducersCount(ctx context.Context) int
	AddConsumer(ctx context.Context, producerID, consumerID string,
		muxerFactory MuxerFactory,
		muxerRemover MuxerRemover,
		errChan chan<- error) error
	RemoveConsumer(ctx context.Context, producerID, consumerID string) error
	PauseProducer(ctx context.Context, producerID string) error
	ResumeProducer(ctx context.Context, producerID string) error
	Stopper
}
