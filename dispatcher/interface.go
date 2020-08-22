package dispatcher

import "context"

// Event ...
type Event string

// EventProcessor ...
type EventProcessor func(ctx context.Context, msg interface{}) error

// EventListener ...
type EventListener interface {
	EventProcessors() map[Event][]EventProcessor
}

// EventDispatcher ...
type EventDispatcher interface {
	Dispatch(ctx context.Context, name Event, msg interface{})
	AddListener(listener EventListener)
	AddProcessor(name Event, processor EventProcessor)
	Stop()
}
