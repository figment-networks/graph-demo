package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/figment-networks/graph-demo/manager/client"
	"github.com/figment-networks/graph-demo/manager/store"
	"github.com/figment-networks/graph-demo/manager/structs"
)

const NETWORK = "cosmos"

type Service struct {
	clients map[string]client.Client
	store   store.Store
}

func New(store store.Store, clients map[string]client.Client) *Service {
	return &Service{
		clients: clients,
		store:   store,
	}
}

func (s *Service) GetBlockByHeight(ctx context.Context, height uint64, chainID string) (structs.Block, []structs.Transaction, error) {
	cli, ok := s.clients[chainID]
	if !ok {
		return structs.Block{}, nil, errors.New("Unknown chain id")
	}

	block, err := s.store.GetBlockByHeight(ctx, height, chainID, NETWORK)
	if err != nil && err == sql.ErrNoRows {
		return cli.GetBlockByHeight(ctx, height)
	} else {
		return structs.Block{}, nil, err
	}

	if block.NumberOfTransactions == 0 {
		return block, nil, nil
	}

	txs, err := s.store.GetTransactions(ctx, height, chainID, NETWORK)
	if err != nil {
		return structs.Block{}, nil, err
	}

	return block, txs, nil

}
