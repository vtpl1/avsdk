package av

import "context"

type SignalStopper interface {
	SignalStop() bool
}

type Stopper interface {
	SignalStopper
	WaitStop() error
	Stop() error
}

type StartStopper interface {
	Start(ctx context.Context) error
	Stopper
}
