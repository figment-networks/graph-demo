package client

import (
	"context"
	"errors"

	"github.com/figment-networks/graph-demo/manager/store"
	"github.com/figment-networks/graph-demo/manager/structs"
	"go.uber.org/zap"
)

type NetworkClient interface {
	GetAll(ctx context.Context, height uint64) error
	GetLatest(ctx context.Context) (structs.Block, error)
}


type SubscriptionClient interface {
	PopulateEvent(ctx context.Context, event string, height uint64, data interface{}) error
}

type Client struct {
	sc SubscriptionClient
	l  *zap.Logger
	st store.Storager
}

func NewClient(l *zap.Logger, st store.Storager, sc SubscriptionClient) *Client {
	return &Client{
		l:  l,
		st: st,
		sc: sc,
	}
}

func (c *Client) ProcessHeight(ctx context.Context, nc NetworkClient, height uint64) (err error) {
	if err = c.getByHeight(ctx, nc, height); err != nil {
		return err
	}

	// We can populate some errors from here
	if err := c.PopulateEvent(ctx, structs.EVENT_NEW_BLOCK, height, structs.EventNewBlock{Height: height}); err != nil {
		return err
	}

	if err := c.PopulateEvent(ctx, structs.EVENT_NEW_TRANSACTION, height, structs.EventNewTransaction{Height: height}); err != nil {
		return err
	}

	return nil
}

func (c *Client) PopulateEvent(ctx context.Context, event string, height uint64, data interface{}) error {
	if c.sc == nil {
		return errors.New("there is now subscription client linked")
	}
	return c.sc.PopulateEvent(ctx, event, height, data)
}

func (c *Client) getByHeight(ctx context.Context, nc NetworkClient, height uint64) (err error) {
	return nc.GetAll(ctx, height)
}

func (c *Client) GetLatest(ctx context.Context, nc NetworkClient) (uint64, error) {
	lst, err := nc.GetLatest(ctx)
	if err != nil {
		return 0, err
	}
	return lst.Height, err
}

func (c *Client) GetLatestFromStorage(ctx context.Context, chainID string) (height uint64, err error) {
	return c.st.GetLatestHeight(ctx, chainID)
}

func (c *Client) SetLatestFromStorage(ctx context.Context, chainID string, height uint64) (err error) {
	return c.st.SetLatestHeight(ctx, chainID, height)
}
