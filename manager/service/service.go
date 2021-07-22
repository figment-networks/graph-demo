package service

import (
	"context"
	"database/sql"

	"github.com/figment-networks/graph-demo/manager/store"
	"github.com/figment-networks/graph-demo/manager/structs"
)

const NETWORK = "cosmos"

type Service struct {
	store store.Store
}

func New(store store.Store) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) GetBlockByHeight(ctx context.Context, height uint64, chainID string) (structs.Block, []structs.Transaction, error) {
	block, err := s.store.GetBlockByHeight(ctx, height, chainID)
	if err != nil && err == sql.ErrNoRows {
		return structs.Block{}, nil, err
	}

	if block.NumberOfTransactions == 0 {
		return block, nil, nil
	}

	txs, err := s.store.GetTransactionsByHeight(ctx, height, chainID)
	if err != nil {
		return structs.Block{}, nil, err
	}

	return block, txs, nil

}
