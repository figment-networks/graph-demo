package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/figment-networks/graph-demo/manager/structs"
	"go.uber.org/zap"
)

type CosmosHTTPTransport struct {
	address string
	c       *http.Client
	log     *zap.Logger
}

func NewCosmosHTTPTransport(address string, c *http.Client, log *zap.Logger) *CosmosHTTPTransport {
	return &CosmosHTTPTransport{
		address: address,
		c:       c,
		log:     log,
	}
}

func (ng *CosmosHTTPTransport) GetAll(ctx context.Context, height uint64) (bTx structs.BlockAndTx, er error) {
	ng.log.Debug("[HTTP] Getting a block", zap.Uint64("height", height))

	resp, err := http.Get(fmt.Sprintf("%s/getAll/%d", ng.address, height))
	if err != nil {
		ng.log.Error("[HTTP] Error while getting a block from worker", zap.Uint64("height", height), zap.Error(err))
		return structs.BlockAndTx{}, err
	}
	defer resp.Body.Close()

	byteResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ng.log.Error("[HTTP] Error while reading block response body", zap.Uint64("height", height), zap.Error(err))
		return structs.BlockAndTx{}, err
	}

	if err = json.Unmarshal(byteResp, &bTx); err != nil {
		ng.log.Error("[HTTP] Error while decoding block with transactions", zap.Uint64("height", height), zap.Error(err))
		return structs.BlockAndTx{}, err
	}

	ng.log.Debug("[HTTP] Got a block", zap.Uint64("height", height), zap.Int("txs", len(bTx.Block.Data.Txs)))

	return structs.BlockAndTx{
		Block:        bTx.Block,
		Transactions: bTx.Transactions,
	}, nil
}

func (ng *CosmosHTTPTransport) GetLatest(ctx context.Context) (bTx structs.BlockAndTx, er error) {
	ng.log.Debug("[HTTP] Getting latest block")

	resp, err := http.Get(fmt.Sprintf("%s/getLatest", ng.address))
	if err != nil {
		ng.log.Error("[HTTP] Error while getting latest block from worker", zap.Error(err))
		return structs.BlockAndTx{}, err
	}
	defer resp.Body.Close()

	byteResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ng.log.Error("[HTTP] Error while reading latest block response body", zap.Error(err))
		return structs.BlockAndTx{}, err
	}

	if err = json.Unmarshal(byteResp, &bTx); err != nil {
		ng.log.Error("[HTTP] Error while decoding latest block with transactions", zap.Error(err))
		return structs.BlockAndTx{}, err
	}

	ng.log.Debug("[HTTP] Got latest block", zap.Int64("height", bTx.Block.Header.Height), zap.Int("txs", len(bTx.Block.Data.Txs)))

	return structs.BlockAndTx{
		Block:        bTx.Block,
		Transactions: bTx.Transactions,
	}, nil
}
