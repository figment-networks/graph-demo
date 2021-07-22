package client

import (
	"context"

	"github.com/figment-networks/graph-demo/manager/structs"
)

type NetworkClient interface {
	GetBlock(ctx context.Context, height uint64) (structs.BlockAndTx, error)
	GetLatest(ctx context.Context) (structs.BlockAndTx, error)
}

type RunnerClient interface {
	PopulateEvent(ctx context.Context, event uint64)
}

type Client struct {
	nc NetworkClient
}

func NewClient(nc NetworkClient) *Client {
	return &Client{nc: nc}
}

func (c *Client) GetBlockByHeight(ctx context.Context, height uint64) (structs.BlockAndTx, error) {
	return c.nc.GetBlock(ctx, height)
}

func (c *Client) GetLatestBlock(ctx context.Context) (structs.BlockAndTx, error) {
	return c.nc.GetLatest(ctx)
}

func (c *Client) PopulateEvent(ctx context.Context, event uint64) error {
	return c.PopulateEvent(ctx, event)
}
