package ratelimit

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

type rateLimitConfigJob struct{}

func NewRateLimitConfigJob() job.Job {
	return &rateLimitConfigJob{}
}

func (s *rateLimitConfigJob) Description() string {
	return "Copies the rate limit configurations from the database to Redis."
}

func (s *rateLimitConfigJob) Config() []env.Config {
	return nil
}

func (s *rateLimitConfigJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}
	rl, err := redispool.NewRateLimiter()
	if err != nil {
		return nil, err
	}
	hdlr := handler{
		codeHostStore: db.CodeHosts(),
		ratelimiter:   ratelimit.NewCodeHostRateLimiter(rl),
	}
	handlerFunc := func(ctx context.Context) error {
		return hdlr.Handle(ctx, observationCtx)
	}
	rlcWorker := makeRateLimitConfigWorker(handlerFunc)

	return []goroutine.BackgroundRoutine{rlcWorker}, nil
}

func makeRateLimitConfigWorker(handlerFunc func(ctx context.Context) error) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		goroutine.HandlerFunc(handlerFunc),
		goroutine.WithName("rate_limit_config_worker"),
		goroutine.WithDescription("copies the rate limit configurations from the database to Redis"),
		goroutine.WithInterval(30*time.Second),
	)
}
