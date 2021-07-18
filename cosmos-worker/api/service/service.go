package service

import (
	"context"

	"github.com/figment-networks/graph-demo/cosmos-worker/api"
	"github.com/figment-networks/graph-demo/cosmos-worker/structs"
)

type ServiceClient interface {
	GetBlock(ctx context.Context, height uint64) (block structs.Block, er error)
}

type Service struct {
	client ServiceClient
}

func NewService(c *api.Client) *Service {
	return &Service{
		client: c,
	}
}

func (s *Service) GetAll(ctx context.Context, heightInt uint64) (*structs.All, error) {
	block, err := s.client.GetBlock(ctx, uint64(heightInt))
	if err != nil {
		return nil, err
	}

	txs, err := s.client.SearchTx(ctx, block, uint64(heightInt), 100)
	return &structs.All{block, txs}, nil
}

func (s *Service) GetLatest(ctx context.Context) (*structs.Latest, error) {
	return s.client.GetLatest(ctx)
}
