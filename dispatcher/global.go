package dispatcher

import "context"

var instance *Dispatcher

func init() {
	instance = &Dispatcher{
		processors: make(map[Event][]EventProcessor),
		bgCtx:      context.Background(),
	}
}

// SetBackgroundContext ...
func SetBackgroundContext(ctx context.Context) {
	instance.bgCtx = ctx
}

// Dispatch dispatch message using global dispatcher
func Dispatch(ctx context.Context, name Event, msg interface{}) error {
	return instance.Dispatch(ctx, name, msg)
}

// AddListener add listener processors using global dispatcher
func AddListener(listener EventListener) {
	instance.AddListener(listener)
}
