package mapper

import (
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/figment-networks/graph-demo/manager/structs"
	"github.com/tendermint/tendermint/libs/bytes"
)

func BlockMapper(b *tmservice.GetBlockByHeightResponse) structs.Block {
	bHeader := b.Block.Header
	bLastCommit := b.Block.LastCommit
	return structs.Block{
		Hash:                 bytes.HexBytes(b.BlockId.Hash).String(),
		Height:               uint64(bHeader.Height),
		Time:                 bHeader.Time,
		ChainID:              bHeader.ChainID,
		NumberOfTransactions: uint64(len(b.Block.Data.Txs)),

		Data: structs.BlockData{
			Txs: b.Block.Data.Txs,
		},

		Header: structs.BlockHeader{
			Version: structs.Consensus(bHeader.Version),
			ChainID: bHeader.ChainID,
			Time:    bHeader.Time,
			Height:  bHeader.Height,
			LastBlockId: structs.BlockID{
				Hash: bytes.HexBytes(bHeader.LastBlockId.Hash).String(),
			},
			LastCommitHash:     bytes.HexBytes(bHeader.LastCommitHash).String(),
			DataHash:           bytes.HexBytes(bHeader.DataHash).String(),
			ValidatorsHash:     bytes.HexBytes(bHeader.ValidatorsHash).String(),
			NextValidatorsHash: bytes.HexBytes(bHeader.NextValidatorsHash).String(),
			ConsensusHash:      bytes.HexBytes(bHeader.ConsensusHash).String(),
			AppHash:            bytes.HexBytes(bHeader.AppHash).String(),
			LastResultsHash:    bytes.HexBytes(bHeader.LastResultsHash).String(),
			EvidenceHash:       bytes.HexBytes(bHeader.EvidenceHash).String(),
			ProposerAddress:    bytes.HexBytes(bHeader.ProposerAddress).String(),
		},

		LastCommit: &structs.Commit{
			Height: bLastCommit.Height,
			Round:  bLastCommit.Round,
			BlockID: structs.BlockID{
				Hash: bytes.HexBytes(bLastCommit.BlockID.Hash).String(),
				PartSetHeader: structs.PartSetHeader{
					Total: bLastCommit.BlockID.PartSetHeader.Total,
					Hash:  bytes.HexBytes(bLastCommit.BlockID.PartSetHeader.Hash).String(),
				}},
		},
	}
}
