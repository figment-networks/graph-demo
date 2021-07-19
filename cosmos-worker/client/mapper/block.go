package mapper

import (
	"github.com/figment-networks/graph-demo/manager/structs"
	"github.com/tendermint/tendermint/proto/tendermint/types"
)

func MapBlockResponseToStructs(b *types.Block, txs types.Data, bHash string) structs.Block {
	return structs.Block{
		Hash:    bHash,
		Height:  uint64(b.Header.Height),
		Time:    b.Header.Time,
		ChainID: b.Header.ChainID,

		Header: structs.BlockHeader{
			ChainID: b.Header.ChainID,
			Time:    b.Header.Time,
			Height:  b.Header.Height,
			LastBlockId: structs.BlockID{
				Hash: b.Header.LastBlockId.Hash,
			},
			LastCommitHash:     b.Header.LastCommitHash,
			DataHash:           b.Header.DataHash,
			ValidatorsHash:     b.Header.ValidatorsHash,
			NextValidatorsHash: b.Header.NextValidatorsHash,
			ConsensusHash:      b.Header.ConsensusHash,
			AppHash:            b.Header.AppHash,
			LastResultsHash:    b.Header.LastResultsHash,
			EvidenceHash:       b.Header.EvidenceHash,
			ProposerAddress:    b.Header.ProposerAddress,
		},
		Data: structs.BlockData{
			Txs: txs.Txs,
		},
		LastCommit: &structs.Commit{
			Height: b.LastCommit.Height,
			Round:  b.LastCommit.Round,
			BlockID: structs.BlockID{
				Hash:          b.LastCommit.BlockID.Hash,
				PartSetHeader: structs.PartSetHeader(b.LastCommit.BlockID.PartSetHeader),
			},
		},
	}
}
