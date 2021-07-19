package service

import (
	"context"

	"github.com/figment-networks/graph-demo/cosmos-worker/api/structs"
	"github.com/figment-networks/graph-demo/cosmos-worker/client"
)

type Service struct {
	client client.GRPC
}

func NewService(c *client.Client) *Service {
	return &Service{
		client: c,
	}
}

func (s *Service) GetBlockAndTransactionsByHeight(ctx context.Context, height uint64) (*structs.BlockAndTx, error) {
	blockAndTx, err := s.client.GetBlock(ctx, height)
	if err != nil {
		return nil, err
	}

	return &structs.BlockAndTx{
		Block: blockAndTx.Block,
		Txs:   blockAndTx.Transactions,
	}, nil
}

func (s *Service) GetLatest(ctx context.Context) (*structs.BlockAndTx, error) {
	latestBlockAndTx, err := s.client.GetLatest(ctx)
	if err != nil {
		return nil, err
	}

	return &structs.BlockAndTx{
		Block: latestBlockAndTx.Block,
		Txs:   latestBlockAndTx.Transactions,
	}, nil
}
