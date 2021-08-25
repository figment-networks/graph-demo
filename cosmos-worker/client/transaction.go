package client

import (
	"context"
	"regexp"
	"strconv"
	"time"

	"github.com/figment-networks/graph-demo/cosmos-worker/client/mapper"
	"github.com/figment-networks/graph-demo/manager/structs"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var curencyRegex = regexp.MustCompile("([0-9\\.\\,\\-\\s]+)([^0-9\\s]+)$")

// SearchTx is making search api call
func (c *Client) SearchTx(ctx context.Context, block structs.Block) (txs []structs.Transaction, err error) {
	height := block.Header.Height
	c.log.Debug("[COSMOS-WORKER] Getting transactions", zap.Int64("height", height))

	pag := &query.PageRequest{
		CountTotal: true,
		Limit:      perPage,
	}

	var page = uint64(1)
	for {
		pag.Offset = (perPage * page) - perPage
		now := time.Now()

		nctx, cancel := context.WithTimeout(ctx, c.cfg.TimeoutSearchTxCall)
		grpcRes, err := c.txServiceClient.GetTxsEvent(nctx, &tx.GetTxsEventRequest{
			Events:     []string{"tx.height=" + strconv.FormatInt(height, 10)},
			Pagination: pag,
		}, grpc.WaitForReady(true))
		cancel()

		c.log.Debug("[COSMOS-API] Request Time (/tx_search)", zap.Duration("duration", time.Since(now)))
		if err != nil {
			return nil, err
		}

		pageTxs := make([]structs.Transaction, len(grpcRes.Txs))
		for i, trans := range grpcRes.Txs {
			if pageTxs[i], err = mapper.TransactionMapper(ctx, trans, grpcRes.TxResponses[i], block.Hash, block.Header.ChainID); err != nil {
				return nil, err
			}
		}
		txs = append(txs, pageTxs...)

		if grpcRes.Pagination.GetTotal() <= uint64(len(txs)) {
			break
		}

		page++

	}

	c.log.Debug("[COSMOS-WORKER] Got transactions", zap.Int64("height", height))
	return txs, nil
}
