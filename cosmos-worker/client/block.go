package client

import (
	"context"

	"github.com/figment-networks/graph-demo/cosmos-worker/client/mapper"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const perPage = 100

// GetAll fetches all data for given height
func (c *Client) GetAll(ctx context.Context, height uint64) (blockID string, txIDs []string, er error) {
	ctx, cancel := context.WithTimeout(ctx, c.cfg.TimeoutBlockCall)
	defer cancel()

	c.log.Debug("[COSMOS-WORKER] Getting block", zap.Uint64("height", height))

	b, err := c.tmServiceClient.GetBlockByHeight(ctx, &tmservice.GetBlockByHeightRequest{Height: int64(height)}, grpc.WaitForReady(true))
	if err != nil {
		c.log.Debug("[COSMOS-CLIENT] Error getting block by height", zap.Uint64("height", height), zap.Error(err))
		return "", nil, err
	}

	c.log.Debug("[COSMOS-WORKER] Got block", zap.Uint64("height", height))

	block := mapper.BlockMapper(b)

	if c.persistor != nil {
		if blockID, err = c.persistor.StoreBlock(ctx, block); err != nil {
			c.log.Debug("[COSMOS-CLIENT] Error storing block at height", zap.Uint64("height", height), zap.Error(err))
			return "", nil, err
		}
	}

	txs, err := c.SearchTx(ctx, block)
	if err != nil {
		c.log.Debug("[COSMOS-CLIENT] Error getting transactions by height", zap.Uint64("height", height), zap.Error(err))
		return "", nil, err
	}

	if c.persistor != nil {
		if txIDs, err = c.persistor.StoreTransactions(ctx, txs); err != nil {
			c.log.Debug("[COSMOS-CLIENT] Error storing transaction at height", zap.Uint64("height", height), zap.Error(err))
			return "", nil, err
		}
	}

	return blockID, txIDs, nil
}

// GetBlock fetches most recent block from chain
func (c *Client) GetLatest(ctx context.Context) (uint64, error) {
	ctx, cancel := context.WithTimeout(ctx, c.cfg.TimeoutBlockCall)
	defer cancel()

	c.log.Debug("[COSMOS-WORKER] Getting latest block")

	b, err := c.tmServiceClient.GetLatestBlock(ctx, &tmservice.GetLatestBlockRequest{}, grpc.WaitForReady(true))
	if err != nil {
		c.log.Debug("[COSMOS-CLIENT] Error getting latest block", zap.Error(err))
		return 0, err
	}

	c.log.Debug("[COSMOS-CLIENT] Got latest block", zap.Uint64("height", uint64(b.Block.Header.Height)))

	return uint64(b.Block.Header.Height), nil

}
