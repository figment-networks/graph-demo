package service

import (
	"context"

	"github.com/figment-networks/graph-demo/manager/store"
)

type Service struct {
	store store.Store
}

func New(store store.Store) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) GetBlockByHeight(ctx context.Context, height uint64) {

}
