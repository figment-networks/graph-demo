package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/figment-networks/graph-demo/runner/api/mapper"
)

var (
	blockParams map[string]struct{} = map[string]struct{}{
		"height":      {},
		"startHeight": {},
		"endHeight":   {},
	}

	blockFields map[string]struct{} = map[string]struct{}{
		"height": {},
		"hash":   {},
		"time":   {},
		"txs":    {},
	}

	txsFields map[string]struct{} = map[string]struct{}{
		"height":    {},
		"hash":      {},
		"time":      {},
		"sender":    {},
		"recipient": {},
	}
)

type Service struct {
	url string
	cli *http.Client
}

func New(cli *http.Client, url string) *Service {
	return &Service{
		url: url,
		cli: cli,
	}
}

func (s *Service) ProcessGraphqlQuery(ctx context.Context, v map[string]interface{}, q string) ([]byte, error) {
	queries, err := mapper.ParseQuery(q, v)
	if err != nil {
		return nil, fmt.Errorf("Error while parsing graphql query: %w", err)
	}
	/*
		rawResp, err := mapper.MapBlocksToResponse(queries.Queries, blocks)
		if err != nil {
			return nil, fmt.Errorf("Error while mapping response: %w", err)
		}
	*/
	return nil, nil
}
