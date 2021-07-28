package service

import (
	"context"

	"github.com/figment-networks/graph-demo/runner/store/memap"

	"go.uber.org/zap"
)

type Service struct {
	store *memap.SubgraphStore
	log   *zap.Logger
}

func New(store *memap.SubgraphStore) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) ProcessGraphqlQuery(ctx context.Context, q []byte, v map[string]interface{}) ([]byte, error) {
	/*
		queries, err := graphcall.ParseQuery(q, v)
		if err != nil {
			return nil, fmt.Errorf("error while parsing graphql query: %w", err)
		}

			for _, query := range queries.Queries {
				s.store.Get(ctx, query.Name, query.)
			}
	*/

	return nil, nil // s.client.ProcessGraphqlQuery(ctx, q, v)
}
