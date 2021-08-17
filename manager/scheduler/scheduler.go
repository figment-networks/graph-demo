package scheduler

import (
	"context"
	"time"

	"github.com/figment-networks/graph-demo/manager/client"
	"go.uber.org/zap"
)

type Clienter interface {
	ProcessHeight(ctx context.Context, nc client.NetworkClient, height uint64) (err error)
	GetLatest(ctx context.Context, nc client.NetworkClient) (height uint64, err error)

	GetLatestFromStorage(ctx context.Context, chainID string) (height uint64, err error)
	SetLatestFromStorage(ctx context.Context, chainID string, height uint64) (err error)
}

type Scheduler struct {
	log           *zap.Logger
	c             Clienter
	lowestHeights map[string]uint64
}

func NewScheduler(log *zap.Logger, c Clienter, lowestHeights map[string]uint64) *Scheduler {
	return &Scheduler{log: log, c: c, lowestHeights: lowestHeights}
}

func (s *Scheduler) Start(ctx context.Context, nc client.NetworkClient, connID, chainID string) {

	h, err := s.c.GetLatestFromStorage(ctx, chainID)
	if err != nil {
		s.log.Error("error getting height", zap.Uint64("height", h), zap.Error(err))
	}

	if lh, ok := s.lowestHeights[chainID]; ok && h < lh {
		h = lh
	} else {
		h++
	}

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

			if lb-h > 2 {
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
