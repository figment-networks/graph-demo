package client

import (
	"context"

	"github.com/figment-networks/graph-demo/manager/store"
	"github.com/figment-networks/graph-demo/manager/structs"
	"go.uber.org/zap"
)

type NetworkClient interface {
	GetBlock(ctx context.Context, height uint64) (structs.BlockAndTx, error)
	GetLatest(ctx context.Context) (structs.BlockAndTx, error)
}

type RunnerClient interface {
	PopulateEvent(ctx context.Context, event string, data interface{}) error
}

type Client struct {
	nc    NetworkClient
	rc    RunnerClient
	log   *zap.Logger
	store store.Store
}

func NewClient(nc NetworkClient) *Client {
	return &Client{nc: nc}
}

func (c *Client) ProcessHeight(ctx context.Context, height uint64) (bTx structs.BlockAndTx, err error) {
	btx, err := c.GetByHeight(ctx, height)

	// We can populate some errors from here
	if err := c.PopulateEvent(ctx, structs.EVENT_NEW_BLOCK, structs.EventNewBlock{btx.Block.Height}); err != nil {
		return bTx, err
	}

	for _, tx := range btx.Transactions {
		c.PopulateEvent(ctx, structs.EVENT_NEW_TRANSACTION, structs.EventNewTransaction{tx.Height})
	}

}

func (c *Client) GetByHeight(ctx context.Context, height uint64) (bTx structs.BlockAndTx, err error) {

	bTx, err = c.nc.GetByHeight(ctx, height)
	if err != nil {
		c.log.Error("[CRON] Error while getting block", zap.Uint64("height", height), zap.Error(err))
		return bTx, err
	}

	if err = c.store.StoreBlock(ctx, bTx.Block); err != nil {
		c.log.Error("[CRON] Error while saving block in database", zap.Uint64("height", height), zap.Error(err))
		return bTx, err
	}

	if bTx.Block.NumberOfTransactions > 0 {
		if err = c.store.StoreTransactions(ctx, bTx.Transactions); err != nil {
			c.log.Error("[CRON] Error while saving transactions in database", zap.Uint64("height", height), zap.Uint64("txs", bTx.Block.NumberOfTransactions), zap.Error(err))
			return bTx, err
		}
	}

	return bTx, err
}

func (c *Client) PopulateEvent(ctx context.Context, event string, data interface{}) error {
	return c.rc.PopulateEvent(ctx, event, data)
}


/*
func (c *Client) GetBlockByHeight(ctx context.Context, height uint64) (structs.BlockAndTx, error) {
	return c.nc.GetBlock(ctx, height)
}

func (c *Client) GetLatestBlock(ctx context.Context) (structs.BlockAndTx, error) {
	return c.nc.GetLatest(ctx)
	*/
