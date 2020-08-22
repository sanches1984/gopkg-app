package job

import (
	"context"

	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
)

func NewPool(ctx context.Context, concurrency uint, redisNS string, redisPool *redis.Pool, jobList []WorkRecord, jobCtx interface{}, middlewareList []interface{}) *work.WorkerPool {
	pool := work.NewWorkerPool(jobCtx, uint(concurrency), redisNS, redisPool)
	for _, mw := range middlewareList {
		pool.Middleware(mw)
	}
	for _, item := range jobList {
		if item.Schedule != "" {
			if err := registerPeriodicalJob(ctx, pool, item.Job, item.Fn, item.Schedule); err != nil {
				panic("Register periodical Job error: " + err.Error())
			}
		} else {
			if err := registerJob(ctx, pool, item.Job, item.Fn); err != nil {
				panic("Register Job error: " + err.Error())
			}
		}
	}
	return pool
}
