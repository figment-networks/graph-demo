package scheduler

import (
	"context"
	"time"

	"github.com/figment-networks/graph-demo/manager/structs"
	"go.uber.org/zap"
)

type NetworkClient interface {
	GetAll(ctx context.Context, height uint64) error
	GetLatest(ctx context.Context) (structs.Block, error)
}

type Clienter interface {
	ProcessHeight(ctx context.Context, nc NetworkClient, height uint64) (err error)
	GetLatest(ctx context.Context, nc NetworkClient) (b structs.Block, err error)

	GetLatestFromStorage(ctx context.Context, chainID string) (height uint64, err error)
	SetLatestFromStorage(ctx context.Context, chainID string, height uint64) (err error)
}

type Scheduler struct {
	log *zap.Logger
	c   Clienter
}

func NewScheduler(log *zap.Logger, c Clienter) *Scheduler {
	return &Scheduler{log: log, c: c}
}

func (s *Scheduler) Start(ctx context.Context, nc NetworkClient, connID, chainID string) {

	h, err := s.c.GetLatestFromStorage(ctx, chainID)
	if err != nil {
		s.log.Error("error getting height", zap.Uint64("height", h), zap.Error(err))
	}
	h++

	tckr := time.NewTicker(10 * time.Second)
	defer tckr.Stop()
	for {
		select {
		case <-tckr.C:
			lb, err := s.c.GetLatest(ctx, nc)
			if err != nil {
				s.log.Error("error getting latest height", zap.Error(err))
				continue
			}

			if lb.Height-h > 2 {
				tckr.Reset(time.Millisecond)
			} else {
				tckr.Reset(10 * time.Second)
			}

			// Here we can create multiple queries depending on number of currently connected workers
			err = s.c.ProcessHeight(ctx, nc, h)
			if err != nil {
				s.log.Error("error getting height", zap.Uint64("height", h), zap.Error(err))
				continue
			}

			if err := s.c.SetLatestFromStorage(ctx, chainID, h); err != nil {
				s.log.Error("error setting latest height", zap.Uint64("height", h), zap.Error(err))
				continue
			}
			h++
		case <-ctx.Done():
			return
		}
	}
}
