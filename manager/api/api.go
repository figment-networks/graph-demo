package api

import (
	"context"
	"errors"
	"fmt"

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

func (s *Service) ProcessGraphqlQuery(ctx context.Context, q []byte, v map[string]interface{}) ([]byte, error) {
	queries, err := graphcall.ParseQuery(q, v)
	if err != nil {
		return nil, fmt.Errorf("error while parsing graphql query: %w", err)
	}

	d, err := s.getData(ctx, &queries)
	if err != nil {
		return nil, fmt.Errorf("error while fetching data: %w", err)
	}

	rawResp, err := mapBlocksToResponse(queries.Queries, d)
	if err != nil {
		return nil, fmt.Errorf("error while mapping response: %w", err)
	}

	return rawResp, nil
}

func (s *Service) getData(ctx context.Context, query *graphcall.GraphQuery) (qresp structs.QueriesResp, err error) {
	qResp := make(map[string]map[uint64]structs.BlockAndTx)

	for _, query := range query.Queries {

		cID := query.Params["chain_id"]
		cIDv := cID.Params["chain_id"].Value
		if cIDv == nil {
			return nil, errors.New("empty parameter chain_id")
		}
		chainID, ok := cIDv.(string)
		if !ok {
			return nil, errors.New("chain_id is not a string")
		}

		resp := make(map[uint64]structs.BlockAndTx)
		if query.Name == "block" {

			cID := query.Params["height"]
			cIDv := cID.Params["height"].Value
			if cIDv == nil {
				return nil, errors.New("empty parameter height")
			}
			height, ok := cIDv.(uint64)
			if !ok {
				return nil, errors.New("height is not a string")
			}

			btx := structs.BlockAndTx{}
			btx.Block, err = s.store.GetBlockByHeight(ctx, height, chainID)
			if err != nil {
				return nil, err
			}

			if len(btx.Block.Data.Txs) > 0 {
				if btx.Transactions, err = s.store.GetTransactionsByParam(ctx, chainID, "height", height); err != nil {
					return nil, err
				}
			}

			resp[height] = btx
		} else if query.Name == "transaction" {

			for k, h := range query.Params {
				if k == "chain_id" {
					continue
				}

				btx := structs.BlockAndTx{}
				if btx.Transactions, err = s.store.GetTransactionsByParam(ctx, chainID, k, h.Params[k].Value); err != nil {
					return nil, err
				}

				resp[0] = btx
				break
			}
		}

		qResp[query.Name] = resp
	}
	return qResp, nil
}
