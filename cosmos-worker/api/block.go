package api

import (
	"context"
	"sync"

	"github.com/figment-networks/graph-demo/manager/structs"
	"go.uber.org/zap"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/tendermint/tendermint/libs/bytes"
	"google.golang.org/grpc"
)

// BlocksMap map of blocks to control block map
// with extra summary of number of transactions
type BlocksMap struct {
	sync.Mutex
	Blocks map[uint64]structs.Block
	NumTxs uint64
}

// BlockErrorPair to wrap error response
type BlockErrorPair struct {
	Height uint64
	Block  structs.Block
	Err    error
}

// GetBlock fetches most recent block from chain
func (c *Client) GetBlock(ctx context.Context, height uint64) (block structs.Block, er error) {
	c.log.Debug("[COSMOS-WORKER] Getting block", zap.Uint64("height", height))

	var ok bool
	if height != 0 {
		block, ok = c.Sbc.Get(height)
		if ok {
			c.log.Debug("[COSMOS-CLIENT] Got block from cache", zap.Uint64("height", height), zap.Uint64("txs", block.NumberOfTransactions))
			return block, nil
		}
	}

	if err := c.rateLimiterGRPC.Wait(ctx); err != nil {
		return block, err
	}

	nctx, cancel := context.WithTimeout(ctx, c.cfg.TimeoutBlockCall)
	defer cancel()
	if height == 0 {
		lb, err := c.tmServiceClient.GetLatestBlock(nctx, &tmservice.GetLatestBlockRequest{})
		if err != nil {
			c.log.Debug("[COSMOS-CLIENT] Error getting latest block", zap.Uint64("height", height), zap.Error(err), zap.Uint64("txs", block.NumberOfTransactions))
			return block, err
		}

		bh := bytes.HexBytes(lb.BlockId.Hash)

		block = structs.Block{
			Hash:                 bh.String(),
			Height:               uint64(lb.Block.Header.Height),
			Time:                 lb.Block.Header.Time,
			ChainID:              lb.Block.Header.ChainID,
			NumberOfTransactions: uint64(len(lb.Block.Data.Txs)),
		}
		c.Sbc.Add(block)

		c.log.Debug("[COSMOS-CLIENT] Got latest block", zap.Uint64("height", height), zap.Error(err), zap.Uint64("txs", block.NumberOfTransactions))
		return block, nil
	}

	bbh, err := c.tmServiceClient.GetBlockByHeight(nctx, &tmservice.GetBlockByHeightRequest{Height: int64(height)}, grpc.WaitForReady(true))
	if err != nil {
		c.log.Debug("[COSMOS-CLIENT] Error while getting block by height", zap.Uint64("height", height), zap.Error(err), zap.Uint64("txs", block.NumberOfTransactions))
		return block, err
	}

	hb := bytes.HexBytes(bbh.BlockId.Hash)

	block = structs.Block{
		Hash:                 hb.String(),
		Height:               uint64(bbh.Block.Header.Height),
		Time:                 bbh.Block.Header.Time,
		ChainID:              bbh.Block.Header.ChainID,
		NumberOfTransactions: uint64(len(bbh.Block.Data.Txs)),
	}

	c.Sbc.Add(block)

	c.log.Debug("[COSMOS-WORKER] Got block", zap.Uint64("height", height), zap.Uint64("txs", block.NumberOfTransactions))
	return block, nil
}
