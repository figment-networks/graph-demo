package client

import (
	"context"
	"sync"

	"github.com/figment-networks/graph-demo/cosmos-worker/client/mapper"
	"github.com/figment-networks/graph-demo/manager/structs"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/tendermint/tendermint/libs/bytes"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const perPage = 100

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
func (c *Client) GetBlock(ctx context.Context, height int64) (blockAndTx structs.BlockAndTx, er error) {
	ctx, cancel := context.WithTimeout(ctx, c.cfg.TimeoutBlockCall)
	defer cancel()

	c.log.Debug("[COSMOS-WORKER] Getting block", zap.Int64("height", height))

	b, err := c.tmServiceClient.GetBlockByHeight(ctx, &tmservice.GetBlockByHeightRequest{Height: height}, grpc.WaitForReady(true))
	if err != nil {
		c.log.Debug("[COSMOS-CLIENT] Error while getting block by height", zap.Int64("height", height), zap.Error(err), zap.Int("txs", len(b.Block.Data.Txs)))
		return structs.BlockAndTx{}, err
	}

	bHash := bytes.HexBytes(b.BlockId.Hash).String()

	blockAndTx.Block = mapper.MapBlockResponseToStructs(b.Block, b.Block.Data, bHash)
	if blockAndTx.Transactions, err = c.SearchTx(ctx, blockAndTx.Block, perPage); err != nil {
		return structs.BlockAndTx{}, err
	}

	// blockID = structs.BlockID{
	// 	Hash: b.BlockId.Hash,
	// }

	c.log.Debug("[COSMOS-WORKER] Got block", zap.Int64("height", height))

	return blockAndTx, nil
}

// GetBlock fetches most recent block from chain
func (c *Client) GetLatest(ctx context.Context) (blockAndTx structs.BlockAndTx, er error) {
	ctx, cancel := context.WithTimeout(ctx, c.cfg.TimeoutBlockCall)
	defer cancel()

	c.log.Debug("[COSMOS-WORKER] Getting latest block")

	b, err := c.tmServiceClient.GetLatestBlock(ctx, &tmservice.GetLatestBlockRequest{}, grpc.WaitForReady(true))
	if err != nil {
		c.log.Debug("[COSMOS-CLIENT] Error getting latest block", zap.Error(err), zap.Int("txs", len(b.Block.Data.Txs)))
		return structs.BlockAndTx{}, err
	}

	bHash := bytes.HexBytes(b.BlockId.Hash).String()

	blockAndTx.Block = mapper.MapBlockResponseToStructs(b.Block, b.Block.Data, bHash)
	if blockAndTx.Transactions, err = c.SearchTx(ctx, blockAndTx.Block, perPage); err != nil {
		return structs.BlockAndTx{}, err
	}

	c.log.Debug("[COSMOS-CLIENT] Got latest block", zap.Uint64("height", uint64(b.Block.Header.Height)), zap.Error(err))

	return blockAndTx, nil

}
