package api

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/figment-networks/graph-demo/graphcall"
	"github.com/figment-networks/graph-demo/manager/store"
	"github.com/figment-networks/graph-demo/manager/structs"
)

type Service struct {
	store store.Storager
}

func NewService(store store.Storager) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) StoreBlock(ctx context.Context, block structs.Block) error {
	return s.store.StoreBlock(ctx, block)
}

func (s *Service) StoreTransactions(ctx context.Context, txs []structs.Transaction) error {
	if len(txs) > 0 {
		return s.store.StoreTransactions(ctx, txs)
	}
	return nil
}

func (s *Service) GetByHeight(ctx context.Context, height uint64, chainID string) (bTx structs.BlockAndTx, err error) {
	if chainID == "" {
		return bTx, errors.New("ChainID is empty")
	}

	bTx.Block, err = s.store.GetBlockByHeight(ctx, height, chainID)
	if err != nil {
		return bTx, err
	}

	if len(bTx.Block.Data.Txs) == 0 {
		return bTx, nil
	}

	bTx.Transactions, err = s.store.GetTransactionsByHeight(ctx, height, chainID)
	if err != nil {
		return bTx, err
	}

	return bTx, nil
}

func (s *Service) ProcessGraphqlQuery(ctx context.Context, q []byte, v map[string]interface{}) ([]byte, error) {
	queries, err := graphcall.ParseQuery(q, v)
	if err != nil {
		return nil, fmt.Errorf("error while parsing graphql query: %w", err)
	}

	blocks, err := s.getBlocks(ctx, &queries)
	if err != nil {
		return nil, fmt.Errorf("error while fetching data: %w", err)
	}

	rawResp, err := mapBlocksToResponse(queries.Queries, blocks)
	if err != nil {
		return nil, fmt.Errorf("error while mapping response: %w", err)
	}

	return rawResp, nil
}

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

/*
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
*/

func (s *Service) getQueryBlocksByHeights(ctx context.Context, params map[string]graphcall.Part) (resp map[uint64]structs.BlockAndTx, err error) {
	chainID, heights, err := getHeightsToFetchByChain(params)
	if err != nil {
		return nil, err
	}

	resp = make(map[uint64]structs.BlockAndTx)
	for _, h := range heights {
		/*	if r, ok := s.getBlockFromCache(chainID, h); ok {
				resp[h] = r
				continue
			}
		*/
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

func getHeightsToFetchByChain(params map[string]graphcall.Part) (chainID string, heights []uint64, err error) {
	var i, startHeight, endHeight uint64
	var isStart, isEnd bool

	for key, v := range params {

		val := v.Params[key].Value
		if val == nil {
			return "", nil, errors.New("empty parameter value")
		}

		switch strings.ToLower(key) {
		case "height", "startheight", "start_height":
			startHeight = val.(uint64)
			isStart = true
		case "endheight", "end_height":
			endHeight = val.(uint64)
			isEnd = true
		case "chain_id", "chainid":
			chainID = val.(string)
		}
	}

	if !isStart && isEnd || !isStart && !isEnd || (isEnd && (endHeight < startHeight)) {
		return "", nil, errors.New("query parameters ar wrong")
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
