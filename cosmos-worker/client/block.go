package client

import (
	"context"

	"github.com/figment-networks/graph-demo/manager/structs"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/tendermint/tendermint/libs/bytes"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const perPage = 100

// GetAll fetches all data for given height
func (c *Client) GetAll(ctx context.Context, height uint64) (er error) {
	ctx, cancel := context.WithTimeout(ctx, c.cfg.TimeoutBlockCall)
	defer cancel()

	c.log.Debug("[COSMOS-WORKER] Getting block", zap.Uint64("height", height))

	b, err := c.tmServiceClient.GetBlockByHeight(ctx, &tmservice.GetBlockByHeightRequest{Height: int64(height)}, grpc.WaitForReady(true))
	if err != nil {
		c.log.Debug("[COSMOS-CLIENT] Error getting block by height", zap.Uint64("height", height), zap.Error(err))
		return err
	}

	c.log.Debug("[COSMOS-WORKER] Got block", zap.Uint64("height", height))

	block := structs.Block{
		Hash:    bytes.HexBytes(b.BlockId.Hash).String(),
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

	if c.persistor != nil {
		if err := c.persistor.StoreBlock(ctx, block); err != nil {
			c.log.Debug("[COSMOS-CLIENT] Error storing block at height", zap.Uint64("height", height), zap.Error(err))
			return err
		}
	}

	txs, err := c.SearchTx(ctx, block)
	if err != nil {
		c.log.Debug("[COSMOS-CLIENT] Error getting transactions by height", zap.Uint64("height", height), zap.Error(err))
		return err
	}

	if c.persistor != nil {
		if err := c.persistor.StoreTransactions(ctx, txs); err != nil {
			c.log.Debug("[COSMOS-CLIENT] Error storing transaction at height", zap.Uint64("height", height), zap.Error(err))
			return err
		}
	}

	return nil
}

// GetBlock fetches most recent block from chain
func (c *Client) GetLatest(ctx context.Context) (bl structs.Block, er error) {
	ctx, cancel := context.WithTimeout(ctx, c.cfg.TimeoutBlockCall)
	defer cancel()

	c.log.Debug("[COSMOS-WORKER] Getting latest block")

	b, err := c.tmServiceClient.GetLatestBlock(ctx, &tmservice.GetLatestBlockRequest{}, grpc.WaitForReady(true))
	if err != nil {
		c.log.Debug("[COSMOS-CLIENT] Error getting latest block", zap.Error(err), zap.Int("txs", len(b.Block.Data.Txs)))
		return bl, err
	}

	c.log.Debug("[COSMOS-CLIENT] Got latest block", zap.Uint64("height", uint64(b.Block.Header.Height)))

	return structs.Block{Height: uint64(b.Block.Header.Height)}, nil

}
