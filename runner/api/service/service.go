package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	wStructs "github.com/figment-networks/graph-demo/cosmos-worker/structs"
	"github.com/figment-networks/graph-demo/manager/structs"
	"github.com/figment-networks/graph-demo/runner/api/mapper"
	rStructs "github.com/figment-networks/graph-demo/runner/api/structs"
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

	blocks, err := s.getBlocks(ctx, &queries)
	if err != nil {
		return nil, fmt.Errorf("Error while fetching data: %w", err)
	}

	rawResp, err := mapper.MapBlocksToResponse(queries.Queries, blocks)
	if err != nil {
		return nil, fmt.Errorf("Error while mapping response: %w", err)
	}

	return rawResp, nil
}

func (s *Service) getBlocks(ctx context.Context, query *rStructs.GraphQuery) (rStructs.QueriesResp, error) {
	qResp := make(map[string]map[uint64]rStructs.BlockAndTx)

	for _, q := range query.Queries {
		fmt.Println(q)
		resp, err := s.getBlocksByHeight(ctx, &q)
		if err != nil {
			return nil, err
		}
		qResp[q.Name] = resp
	}

	return qResp, nil
}

func (s *Service) getBlocksByHeight(ctx context.Context, q *rStructs.Query) (map[uint64]rStructs.BlockAndTx, error) {
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

func getHeightsToFetch(params map[string]rStructs.Part) ([]uint64, error) {
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

	resp, err := http.Get(fmt.Sprintf("%s/get_block/%d", s.url, height))
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
