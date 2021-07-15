package client

import (
	"context"

	"github.com/figment-networks/graph-demo/manager/store"
	"github.com/figment-networks/graph-demo/manager/store/params"
	"github.com/figment-networks/graph-demo/manager/structs"

	"go.uber.org/zap"
)

type Client struct {
	storeEng store.DataStore
	logger   *zap.Logger
}

func NewClient(storeEng store.DataStore, logger *zap.Logger) *Client {
	c := &Client{
		storeEng: storeEng,
		logger:   logger,
	}
	return c
}

// SearchTransactions is the search
func (hc *Client) SearchTransactions(ctx context.Context, ts structs.TransactionSearch) ([]structs.Transaction, error) {
	return hc.storeEng.GetTransactions(ctx, params.TransactionSearch{
		Network:      ts.Network,
		ChainIDs:     ts.ChainIDs,
		Epoch:        ts.Epoch,
		Hash:         ts.Hash,
		Height:       ts.Height,
		Type:         params.SearchArr{Value: ts.Type.Value},
		BlockHash:    ts.BlockHash,
		Account:      ts.Account,
		Sender:       ts.Sender,
		Receiver:     ts.Receiver,
		Memo:         ts.Memo,
		AfterTime:    ts.AfterTime,
		BeforeTime:   ts.BeforeTime,
		AfterHeight:  ts.AfterHeight,
		BeforeHeight: ts.BeforeHeight,
		Limit:        ts.Limit,
		Offset:       ts.Offset,
		WithRaw:      ts.WithRaw,
		WithRawLog:   ts.WithRawLog,
	})
}
