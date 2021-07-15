package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/figment-networks/graph-demo/manager/structs"
	"github.com/figment-networks/graph-demo/runner/api/graphql"
)

var (
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

func (s *Service) ProcessGraphqlQuery(v map[string]interface{}, q string) (map[string]interface{}, error) {
	queries, err := graphql.ParseQuery(q, v)
	if err != nil {
		return nil, fmt.Errorf("Error while parsing graphql query: %w", err)
	}

	if err := fetchData(&queries); err != nil {
		return nil, fmt.Errorf("Error while fetching data: %w", err)
	}

	rawResp, err := graphql.MapQueryToResponse(queries.Queries)
	if err != nil {
		return nil, fmt.Errorf("Error while mapping response: %w", err)
	}

	return rawResp, nil
}

func (s *Service) fetchData(q *graphql.GraphQuery) error {

	for _, q := range q.Queries {
		fmt.Println(q)
	}

	return nil
}

func (s *Service) getBlockByHeight(ctx context.Context, height uint64) (structs.Block, []structs.Transaction, error) {
	resp, err := http.Get(fmt.Sprintf("%s/get_block/%d", s.url, height))
	if err != nil {
		return structs.Block{}, nil, err
	}

}
