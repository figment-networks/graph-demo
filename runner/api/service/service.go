package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/figment-networks/graph-demo/graphcall"
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
	queries, err := graphcall.ParseQuery(q, v)
	if err != nil {
		return nil, fmt.Errorf("error while parsing graphql query: %w", err)
	}

	for _, query := range queries.Queries {
		id, ok := query.Params["id"]
		if !ok {
			return nil, errors.New("missing required parameter: id")
		}

		idParam, ok := id.Params["id"]
		if !ok {
			return nil, errors.New("missing required parameter value: id")
		}

		for name, fields := range query.Fields {

			records, err := s.store.Get(ctx, "name", name, id.Name, idParam.Value.(string))

		}
		id, ok := query.Params["id"]
		if !ok {
			return nil, errors.New("missing required field: id")
		}

		id.Params[id]

		s.store.Get(ctx, query.Name, id.Name, id.Params)
	}

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
