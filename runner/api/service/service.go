package service

import (
	"context"

	"github.com/figment-networks/graph-demo/runner/api/client"
	"github.com/figment-networks/graph-demo/runner/store/memap"

	"go.uber.org/zap"
)

type Service struct {
	client *client.Client
	store  *memap.SubgraphStore
	log    *zap.Logger
}

func New(client *client.Client, store *memap.SubgraphStore) *Service {
	return &Service{
		client: client,
		store:  store,
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

	return s.client.ProcessGraphqlQuery(ctx, q, v)
}
