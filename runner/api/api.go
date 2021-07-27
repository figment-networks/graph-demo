package api

import (
	"context"
	"fmt"
	"log"

	"github.com/figment-networks/graph-demo/graphcall"
	"github.com/figment-networks/graph-demo/runner/store"
)

type Service struct {
	store store.Storage
}

func NewService(store store.Storage) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) ProcessGraphqlQuery(ctx context.Context, v map[string]interface{}, q string) ([]byte, error) {
	queries, err := graphcall.ParseQuery(q, v)
	if err != nil {
		return nil, fmt.Errorf("error while parsing graphql query: %w", err)
	}

	log.Println("queries", queries)
	/*
		blocks, err := s.getBlocks(ctx, &queries)
		if err != nil {
			return nil, fmt.Errorf("error while fetching data: %w", err)
		}
		rawResp, err := mapBlocksToResponse(queries.Queries, blocks)
		if err != nil {
			return nil, fmt.Errorf("error while mapping response: %w", err)
		}
	*/
	return []byte{}, nil

}

/*
func (s *Service) getBlocks(ctx context.Context, query *graphcall.GraphQuery) (structs.QueriesResp, error) {
	qResp := make(map[string]map[uint64]structs.BlockAndTx)

	for _, query := range query.Queries {
		resp, err := s.getQueryBlocksByHeights(ctx, query.Params)
		if err != nil {
			return nil, err
		}

		qResp[query.Name] = resp
	}
	return qResp, nil
}

func (s *Service) getQueryBlocksByHeights(ctx context.Context, params map[string]graphcall.Part) (resp map[uint64]structs.BlockAndTx, err error) {
	chainID, heights, err := getHeightsToFetchByChain(params)
	if err != nil {
		return nil, err
	}

	resp = make(map[uint64]structs.BlockAndTx)
	for _, h := range heights {
		bTx, err := s.GetByHeight(ctx, h, chainID)
		if err != nil {
			return nil, err
		}

		resp[h] = structs.BlockAndTx{
			Block:        bTx.Block,
			Transactions: bTx.Transactions,
		}
	}

	return resp, err
}
*/
