package client

import (
	"context"
	"sync"

	"github.com/figment-networks/graph-demo/manager/structs"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
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
func (c *Client) GetBlock(ctx context.Context, height uint64) (block structs.Block, blockID structs.BlockID, er error) {
	c.log.Debug("[COSMOS-WORKER] Getting block", zap.Uint64("height", height))

	b, err := c.tmServiceClient.GetBlockByHeight(ctx, &tmservice.GetBlockByHeightRequest{Height: int64(height)}, grpc.WaitForReady(true))
	if err != nil {
		c.log.Debug("[COSMOS-CLIENT] Error while getting block by height", zap.Uint64("height", height), zap.Error(err), zap.Int("txs", len(b.Block.Data.Txs)))
		return block, blockID, err
	}

	bhash := bytes.HexBytes(b.BlockId.Hash)

	block = structs.Block{
		Hash:    bhash.String(),
		Height:  uint64(b.Block.Header.Height),
		Time:    b.Block.Header.Time,
		ChainID: b.Block.Header.ChainID,

		Header: structs.BlockHeader{
			ChainID: b.Block.Header.ChainID,
			Time:    b.Block.Header.Time,
			Height:  b.Block.Header.Height,
			LastBlockId: structs.BlockID{
				Hash: b.Block.Header.LastBlockId.Hash,
			},
			LastCommitHash:     b.Block.Header.LastCommitHash,
			DataHash:           b.Block.Header.DataHash,
			ValidatorsHash:     b.Block.Header.ValidatorsHash,
			NextValidatorsHash: b.Block.Header.NextValidatorsHash,
			ConsensusHash:      b.Block.Header.ConsensusHash,
			AppHash:            b.Block.Header.AppHash,
			LastResultsHash:    b.Block.Header.LastResultsHash,
			EvidenceHash:       b.Block.Header.EvidenceHash,
			ProposerAddress:    b.Block.Header.ProposerAddress,
		},
		Data: structs.BlockData{
			Txs: b.Block.Data.Txs,
		},
		LastCommit: &structs.Commit{
			Height: b.Block.LastCommit.Height,
			Round:  b.Block.LastCommit.Round,
			BlockID: structs.BlockID{
				Hash:          b.Block.LastCommit.BlockID.Hash,
				PartSetHeader: structs.PartSetHeader(b.Block.LastCommit.BlockID.PartSetHeader),
			},
		},
	}

	blockID = structs.BlockID{
		Hash: b.BlockId.Hash,
	}

	c.log.Debug("[COSMOS-WORKER] Got block", zap.Uint64("height", height))
	return block, blockID, nil
}

// GetBlock fetches most recent block from chain
func (c *Client) GetLatest(ctx context.Context) (block structs.Block, er error) {

	nctx, cancel := context.WithTimeout(ctx, c.cfg.TimeoutBlockCall)
	defer cancel()
	b, err := c.tmServiceClient.GetLatestBlock(nctx, &tmservice.GetLatestBlockRequest{}, grpc.WaitForReady(true))
	if err != nil {
		c.log.Debug("[COSMOS-CLIENT] Error getting latest block", zap.Error(err), zap.Int("txs", len(b.Block.Data.Txs)))
		return block, err
	}

	c.log.Debug("[COSMOS-CLIENT] Got latest block", zap.Uint64("height", uint64(b.Block.Header.Height)), zap.Error(err))
	return structs.Block{
		Hash:   bytes.HexBytes(b.BlockId.Hash).String(),
		Height: uint64(b.Block.Header.Height),
	}, nil

}