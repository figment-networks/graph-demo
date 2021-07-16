package scheduler

import (
	"context"

	"github.com/figment-networks/graph-demo/manager/client"
	"github.com/figment-networks/graph-demo/manager/store"
	"go.uber.org/zap"

	"github.com/robfig/cron"
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
		cron:   cron.New(),
		log:    log,
		store:  s,
	}
}

func (s *Scheduler) Start(height uint64) {
	s.height = height
	s.cron.AddFunc("@every 1m", s.fetchAndSaveBlockInDatbase)
}

func (s *Scheduler) fetchAndSaveBlockInDatbase() {

	block, txs, err := s.client.GetBlockByHeight(s.ctx, s.height)
	if err != nil {
		s.log.Error("[CRON] Error while getting block", zap.Uint64("height", s.height), zap.Error(err))
		return
	}

	if err := s.store.StoreBlock(s.ctx, block); err != nil {
		s.log.Error("[CRON] Error while saving block in database", zap.Uint64("height", s.height), zap.Error(err))
		return
	}

	if block.NumberOfTransactions > 0 {
		if err := s.store.StoreTransactions(s.ctx, txs); err != nil {
			s.log.Error("[CRON] Error while saving transactions in database", zap.Uint64("height", s.height), zap.Uint64("txs", block.NumberOfTransactions), zap.Error(err))
			return
		}
	}

}
