package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/figment-networks/graph-demo/graphcall"
	"github.com/figment-networks/graph-demo/manager/client"
	"github.com/figment-networks/graph-demo/manager/store"
	"github.com/figment-networks/graph-demo/manager/structs"
)

const NETWORK = "cosmos"

type Service struct {
	clients map[string]client.Client
	store   store.Store
}

func New(store store.Store, clients map[string]client.Client) *Service {
	return &Service{
		clients: clients,
		store:   store,
	}
}

func (s *Service) GetByHeight(ctx context.Context, height uint64, chainID string) (structs.All, error) {
	cli, ok := s.clients[chainID]
	if !ok {
		return structs.All{}, errors.New("Unknown chain id")
	}

	block, err := s.store.GetBlockByHeight(ctx, height, chainID, NETWORK)
	if err != nil && err == sql.ErrNoRows {
		return cli.GetByHeight(ctx, height)
	} else {
		return structs.All{}, err
	}

	if block.NumberOfTransactions == 0 {
		return structs.All{}, nil
	}

	txs, err := s.store.GetTransactions(ctx, height, chainID, NETWORK)
	if err != nil {
		return structs.All{}, err
	}

	return block, txs, nil

}

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

func (s *Service) ProcessGraphqlQuery(ctx context.Context, v map[string]interface{}, q string) ([]byte, error) {
	queries, err := graphcall.ParseQuery(q, v)
	if err != nil {
		return nil, fmt.Errorf("Error while parsing graphql query: %w", err)
	}

	blocks, err := s.getBlocks(ctx, &queries)
	if err != nil {
		return nil, fmt.Errorf("Error while fetching data: %w", err)
	}

	rawResp, err := graphcall.MapBlocksToResponse(queries.Queries, blocks)
	if err != nil {
		return nil, fmt.Errorf("Error while mapping response: %w", err)
	}

	return rawResp, nil
}

func (s *Service) getBlocks(ctx context.Context, query *graphcall.GraphQuery) (rStructs.QueriesResp, error) {
	qResp := make(map[string]map[uint64]rStructs.BlockAndTx)

	for _, q := range query.Queries {
		resp, err := s.getBlocksByHeight(ctx, &q)
		if err != nil {
			return nil, err
		}
		qResp[q.Name] = resp
	}

	return qResp, nil
}

func (s *Service) getBlocksByHeight(ctx context.Context, q *graphcall.Query) (map[uint64]rStructs.BlockAndTx, error) {
	heights, err := getHeightsToFetch(q.Params)
	if err != nil {
		return nil, err
	}

	blocks := make(map[uint64]rStructs.BlockAndTx)
	for _, h := range heights {
		b, txs, err := s.getBlockByHeight(ctx, h)
		if err != nil {
			return nil, err
		}
		blocks[h] = rStructs.BlockAndTx{
			Block: b,
			Txs:   txs,
		}
	}

	return blocks, nil
}

func getHeightsToFetch(params map[string]graphcall.Part) ([]uint64, error) {
	var i, startHeight, endHeight uint64
	var isStart, isEnd bool

	for key, v := range params {

		val := v.Params[key].Value
		if val == nil {
			return nil, errors.New("Empty parameter value")
		}

		switch key {
		case "height", "startHeight":
			startHeight = val.(uint64)
			isStart = true
		case "endHeight":
			endHeight = val.(uint64)
			isEnd = true
		}
	}

	if !isStart && isEnd || !isStart && !isEnd || (isEnd && (endHeight < startHeight)) {
		return nil, errors.New("Bad height parameters")
	}

	if !isEnd {
		return []uint64{startHeight}, nil
	}

	heights := make([]uint64, endHeight-startHeight+1)
	for i = 0; i < endHeight-startHeight+1; i++ {
		heights[i] = i
	}

	return heights, nil
}

func (s *Service) getBlockByHeight(ctx context.Context, height uint64) (structs.Block, []structs.Transaction, error) {
	var getBlockResp wStructs.GetBlockResp

	resp, err := http.Get(fmt.Sprintf("%s/getBlock/%d", s.url, height))
	if err != nil {
		return structs.Block{}, nil, err
	}
	defer resp.Body.Close()

	byteResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return structs.Block{}, nil, err
	}

	if err = json.Unmarshal(byteResp, &getBlockResp); err != nil {
		return structs.Block{}, nil, err
	}

	return getBlockResp.Block, getBlockResp.Txs, nil
}
