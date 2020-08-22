package dispatcher

import (
	"context"
	"github.com/severgroup-tt/gopkg-app/client/sentry"
	"github.com/severgroup-tt/gopkg-app/middleware"
	database "github.com/severgroup-tt/gopkg-database"
	"github.com/severgroup-tt/gopkg-database/repository/dao"
	logger "github.com/severgroup-tt/gopkg-logger"
	"sync"
)

// Dispatcher implementation of facade
type Dispatcher struct {
	processors map[Event][]EventProcessor
	bgCtx      context.Context
	// syncWG     sync.WaitGroup
	asyncWG sync.WaitGroup
	rwMutex sync.RWMutex
}

// AddListener add listener processors
func (d *Dispatcher) AddListener(listener EventListener) {
	d.rwMutex.Lock()
	defer d.rwMutex.Unlock()
	for name, processor := range listener.EventProcessors() {
		d.processors[name] = append(d.processors[name], processor...)
	}
}

// AddProcessor add processor for specific event
func (d *Dispatcher) AddProcessor(name Event, processor EventProcessor) {
	d.rwMutex.Lock()
	defer d.rwMutex.Unlock()
	d.processors[name] = append(d.processors[name], processor)
}

// Dispatch process event
func (d *Dispatcher) Dispatch(ctx context.Context, name Event, msg interface{}) error {
	d.rwMutex.RLock()
	defer d.rwMutex.RUnlock()
	processorsCnt := len(d.processors[name])
	if processorsCnt == 0 {
		return nil
	}

	h := func(ctx context.Context) error {
		errs := make(chan error, len(d.processors[name]))
		for i := range d.processors[name] {
			go func(processor EventProcessor, errs chan<- error) {
				errs <- processor(ctx, msg)
			}(d.processors[name][i], errs)
		}
		for i := 0; i < len(d.processors[name]); i++ {
			if err := <-errs; err != nil {
				return err
			}
		}
		return nil
	}

	db := database.TryFromContext(ctx)
	if db != nil {
		repo := dao.New()
		return repo.WithTX(ctx, func(ctx context.Context) error {
			return h(ctx)
		})
	}

	return h(ctx)
}

// Stop wait until all async processors done
func (d *Dispatcher) Stop() {
	d.rwMutex.Lock()
	defer d.rwMutex.Unlock()
	d.asyncWG.Wait()
}

// withRecover wrap function with panic catch
func withRecover(ctx context.Context, fn func()) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(ctx, "dispatcher-panic error: %v", err)
			sentry.Panic(err, "request_id", middleware.GetRequestId(ctx))
		}
	}()

	fn()
}
