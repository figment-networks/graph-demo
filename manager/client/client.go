package client

import (
	"context"

	"github.com/figment-networks/graph-demo/manager/structs"
)

type NetworkClient interface {
	GetByHeight(ctx context.Context, height uint64) (structs.BlockAndTx, error)
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

func (c *Client) GetByHeight(ctx context.Context, height uint64) (structs.BlockAndTx, error) {
	return c.nc.GetByHeight(ctx, height)
}

func (c *Client) PopulateEvent(ctx context.Context, event uint64) error {
	return c.PopulateEvent(ctx, event)
}
