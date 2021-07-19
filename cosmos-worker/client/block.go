package client

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/figment-networks/graph-demo/cosmos-worker/client/mapper"
	"github.com/figment-networks/graph-demo/manager/structs"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/bytes"
	"go.uber.org/zap"
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
func (c *Client) GetBlock(ctx context.Context, height uint64) (blockAndTx structs.BlockAndTx, er error) {
	c.log.Debug("[COSMOS-WORKER] Getting block", zap.Uint64("height", height))

	b, err := c.tmServiceClient.GetBlockByHeight(ctx, &tmservice.GetBlockByHeightRequest{Height: int64(height)}, grpc.WaitForReady(true))
	if err != nil {
		c.log.Debug("[COSMOS-CLIENT] Error while getting block by height", zap.Uint64("height", height), zap.Error(err), zap.Int("txs", len(b.Block.Data.Txs)))
		return structs.BlockAndTx{}, err
	}

	bHash := bytes.HexBytes(b.BlockId.Hash).String()

	blockAndTx.Block = mapper.MapBlockResponseToStructs(b.Block, b.Block.Data, bHash)

	blockAndTx.Transactions = make([]structs.Transaction, 0)

	for _, tx := range b.Block.Data.GetTxs() {

		// tx

		decodedTx, err := decodeTx(tx)
		if err != nil {
			return structs.BlockAndTx{}, err
		}

		fmt.Println(decodedTx)
		// c.rawToTransaction(ctx, decodedTx, nil)

	}

	// blockID = structs.BlockID{
	// 	Hash: b.BlockId.Hash,
	// }

	c.log.Debug("[COSMOS-WORKER] Got block", zap.Uint64("height", height))
	return blockAndTx, nil
}

func decodeTx(txBytes []byte) (types.Tx, error) {
	if len(txBytes) == 0 {
		return nil, errors.New("tx bytes are empty")
	}
	var tx = legacytx.StdTx{}
	// UnmarshalBinaryBare

	// legacytx.Unmarshal(amino.U, &tx)

	err := amino.UnmarshalBinaryBare(txBytes, &tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// GetBlock fetches most recent block from chain
func (c *Client) GetLatest(ctx context.Context) (block structs.BlockAndTx, er error) {

	nctx, cancel := context.WithTimeout(ctx, c.cfg.TimeoutBlockCall)
	defer cancel()
	b, err := c.tmServiceClient.GetLatestBlock(nctx, &tmservice.GetLatestBlockRequest{}, grpc.WaitForReady(true))
	if err != nil {
		c.log.Debug("[COSMOS-CLIENT] Error getting latest block", zap.Error(err), zap.Int("txs", len(b.Block.Data.Txs)))
		return block, err
	}

	bHash := bytes.HexBytes(b.BlockId.Hash).String()

	c.log.Debug("[COSMOS-CLIENT] Got latest block", zap.Uint64("height", uint64(b.Block.Header.Height)), zap.Error(err))
	return structs.BlockAndTx{
		Block: mapper.MapBlockResponseToStructs(b.Block, b.Block.Data, bHash),
	}, nil

}
