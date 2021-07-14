package api

import (
	"context"
	"sync"

	"github.com/figment-networks/indexing-engine/structs"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/tendermint/tendermint/libs/bytes"
	"google.golang.org/grpc"
)

const BLOCK_VERSION = "0.0.1"

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
	var ok bool
	if height != 0 {
		block, ok = c.Sbc.Get(height)
		if ok {
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

		return block, nil
	}

	bbh, err := c.tmServiceClient.GetBlockByHeight(nctx, &tmservice.GetBlockByHeightRequest{Height: int64(height)}, grpc.WaitForReady(true))
	if err != nil {
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

	return block, nil

}
