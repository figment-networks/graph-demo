package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/figment-networks/graph-demo/graphcall"
	"github.com/figment-networks/graph-demo/manager/client"
	"github.com/figment-networks/graph-demo/manager/store"
	"github.com/figment-networks/graph-demo/manager/structs"
	rStructs "github.com/figment-networks/graph-demo/runner/api/structs"
)

type Service struct {
	clients map[string]client.Client
	store   store.Store

	blocksByChainID map[string]map[uint64]rStructs.BlockAndTx
	lock            sync.RWMutex
}

func New(store store.Store, clients map[string]client.Client) *Service {
	return &Service{
		clients:         clients,
		store:           store,
		blocksByChainID: make(map[string]map[uint64]rStructs.BlockAndTx),
	}
}

func (s *Service) Close() error {
	return s.store.Close()
}

func (s *Service) GetByHeight(ctx context.Context, height uint64, chainID string) (structs.BlockAndTx, error) {
	cli, ok := s.clients[chainID]
	if !ok {
		return structs.BlockAndTx{}, errors.New("Unknown chain id")
	}

	block, err := s.store.GetBlockByHeight(ctx, height, chainID)
	if err != nil && err == sql.ErrNoRows {
		return cli.GetByHeight(ctx, height)
	} else {
		return structs.BlockAndTx{}, err
	}

	if block.NumberOfTransactions == 0 {
		return structs.BlockAndTx{}, nil
	}

	txs, err := s.store.GetTransactionsByHeight(ctx, height, chainID)
	if err != nil {
		return structs.BlockAndTx{}, err
	}

	return structs.BlockAndTx{
		Block:        block,
		Transactions: txs,
	}, nil

}

func (s *Service) ProcessGraphqlQuery(ctx context.Context, v map[string]interface{}, q string) ([]byte, error) {
	queries, err := graphcall.ParseQuery(q, v)
	if err != nil {
		return nil, fmt.Errorf("Error while parsing graphql query: %w", err)
	}

	blocks, err := s.getBlocks(ctx, &queries)
	if err != nil {
		return nil, fmt.Errorf("Error while fetching data: %w", err)
	}

	rawResp, err := mapBlocksToResponse(queries.Queries, blocks)
	if err != nil {
		return nil, fmt.Errorf("Error while mapping response: %w", err)
	}

	return rawResp, nil
}

func (s *Service) getBlocks(ctx context.Context, query *graphcall.GraphQuery) (rStructs.QueriesResp, error) {
	qResp := make(map[string]map[uint64]rStructs.BlockAndTx)

	for _, query := range query.Queries {
		resp, err := s.getQueryBlocksByHeights(ctx, query.Params)
		if err != nil {
			return nil, err
		}

		qResp[query.Name] = resp
	}
	return qResp, nil
}

func (s *Service) getBlockFromCache(chainID string, height uint64) (rStructs.BlockAndTx, bool) {
	blocks, ok := s.blocksByChainID[chainID]
	if !ok {
		return rStructs.BlockAndTx{}, false
	}

	if r, ok := blocks[height]; ok {
		return r, true
	}

	return rStructs.BlockAndTx{}, false
}

type blocksByChain map[string]map[uint64]rStructs.BlockAndTx

func (s *Service) getQueryBlocksByHeights(ctx context.Context, params map[string]graphcall.Part) (resp map[uint64]rStructs.BlockAndTx, err error) {
	chainID, heights, err := getHeightsToFetchByChain(params)
	if err != nil {
		return nil, err
	}

	cli, ok := s.clients[chainID]
	if !ok {
		return nil, errors.New("Unknown chain id")
	}

	resp = make(map[uint64]rStructs.BlockAndTx)
	for _, h := range heights {
		if r, ok := s.getBlockFromCache(chainID, h); ok {
			resp[h] = r
			continue
		}

		bTx, err := cli.GetByHeight(ctx, h)
		if err != nil {
			return nil, err
		}

		resp[h] = rStructs.BlockAndTx{
			Block: bTx.Block,
			Txs:   bTx.Transactions,
		}
	}

	return nil, err
}

func getHeightsToFetchByChain(params map[string]graphcall.Part) (chainID string, heights []uint64, err error) {
	var i, startHeight, endHeight uint64
	var isStart, isEnd bool

	for key, v := range params {

		val := v.Params[key].Value
		if val == nil {
			return "", nil, errors.New("Empty parameter value")
		}

		switch key {
		case "height", "startHeight":
			startHeight = val.(uint64)
			isStart = true
		case "endHeight":
			endHeight = val.(uint64)
			isEnd = true
		case "chain_id":
			chainID = val.(string)
		}
	}

	if !isStart && isEnd || !isStart && !isEnd || (isEnd && (endHeight < startHeight)) {
		return "", nil, errors.New("Bad height parameters")
	}

	if !isEnd {
		return chainID, []uint64{startHeight}, nil
	}

	heights = make([]uint64, endHeight-startHeight+1)
	for i = 0; i < endHeight-startHeight+1; i++ {
		heights[i] = startHeight + i
	}

	return chainID, heights, nil
}
