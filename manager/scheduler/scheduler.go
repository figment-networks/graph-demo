package scheduler

import (
	"context"
	"time"

	"github.com/figment-networks/graph-demo/manager/client"
	"go.uber.org/zap"
)

type Scheduler struct {
	client client.Client
	log    *zap.Logger
}

func New(log *zap.Logger, c client.Client) *Scheduler {
	return &Scheduler{
		client: c,
		log:    log,
	}
}

func (s *Scheduler) Start(ctx context.Context, height uint64) {
	h := height

	tckr := time.NewTicker(10 * time.Second)
	defer tckr.Stop()
	for {
		select {
		case <-tckr.C:
			_, err := s.client.GetByHeight(ctx, h)
			if err != nil {
				s.log.Error("error getting height", zap.Uint64("height", h), zap.Error(err))
				continue
			}
			h++
		case <-ctx.Done():
			return
		}
	}
}

/*
	s.fetchAndSaveBlockInDatbase()
			s.height++
		case <-ctx.Done():
			break
		}
	}

}

func (s *Scheduler) fetchAndSaveBlockInDatbase() {
	bTx, err := s.client.GetBlockByHeight(s.ctx, s.height)
	if err != nil {
		s.log.Error("[CRON] Error while getting block", zap.Uint64("height", s.height), zap.Error(err))
		return
	}

	s.storeBlockAndTxs(bTx)
}

func (s *Scheduler) storeBlockAndTxs(bTx structs.BlockAndTx) {
	if err := s.store.StoreBlock(s.ctx, bTx.Block); err != nil {
		s.log.Error("[CRON] Error while saving block in database", zap.Uint64("height", s.height), zap.Error(err))
		return
	}

	if bTx.Block.NumberOfTransactions > 0 {
		if err := s.store.StoreTransactions(s.ctx, bTx.Transactions); err != nil {
			s.log.Error("[CRON] Error while saving transactions in database", zap.Uint64("height", s.height), zap.Uint64("txs", bTx.Block.NumberOfTransactions), zap.Error(err))
*/
