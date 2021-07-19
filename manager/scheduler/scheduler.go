package scheduler

import (
	"context"
	"time"

	"github.com/figment-networks/graph-demo/manager/client"
	"github.com/figment-networks/graph-demo/manager/store"
	"github.com/robfig/cron"
	"go.uber.org/zap"
)

type Scheduler struct {
	height uint64
	ctx    context.Context
	cron   *cron.Cron
	client client.Client
	log    *zap.Logger
	store  store.Store
}

func New(ctx context.Context, c client.Client, s store.Store, log *zap.Logger) *Scheduler {
	return &Scheduler{
		ctx:    ctx,
		client: c,
		log:    log,
		store:  s,
	}
}

func (s *Scheduler) Start(ctx context.Context, height uint64) {
	s.height = height

	tckr := time.NewTicker(time.Minute)
	defer tckr.Stop()
	for {
		select {
		case <-tckr.C:
			s.fetchAndSaveBlockInDatbase()
		case <-ctx.Done():
			break
		}
	}

}

func (s *Scheduler) fetchAndSaveBlockInDatbase() {

	all, err := s.client.GetByHeight(s.ctx, s.height)
	if err != nil {
		s.log.Error("[CRON] Error while getting block", zap.Uint64("height", s.height), zap.Error(err))
		return
	}

	if err := s.store.StoreBlock(s.ctx, all.Block); err != nil {
		s.log.Error("[CRON] Error while saving block in database", zap.Uint64("height", s.height), zap.Error(err))
		return
	}

	if all.Block.NumberOfTransactions > 0 {
		if err := s.store.StoreTransactions(s.ctx, all.Transactions); err != nil {
			s.log.Error("[CRON] Error while saving transactions in database", zap.Uint64("height", s.height), zap.Uint64("txs", all.Block.NumberOfTransactions), zap.Error(err))
			return
		}
	}

}
