package mapper

import (
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/figment-networks/graph-demo/manager/structs"
	"github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/proto/tendermint/types"
)

func BlockMapper(b *tmservice.GetBlockByHeightResponse) structs.Block {
	bHeader := b.Block.Header

	return structs.Block{
		Hash:    bytes.HexBytes(b.BlockId.Hash).String(),
		Height:  uint64(bHeader.Height),
		Time:    bHeader.Time,
		ChainID: bHeader.ChainID,
		Data: structs.BlockData{
			Txs: b.Block.Data.Txs,
		},
		Evidence: getEvidence(b.Block.Evidence.GetEvidence()),
		Header: structs.BlockHeader{
			Version: structs.Consensus(bHeader.Version),
			ChainID: bHeader.ChainID,
			Time:    bHeader.Time,
			Height:  bHeader.Height,
			LastBlockId: structs.BlockID{
				Hash: bytes.HexBytes(bHeader.LastBlockId.Hash).String(),
				PartSetHeader: structs.PartSetHeader{
					Total: bHeader.LastBlockId.PartSetHeader.Total,
					Hash:  bytes.HexBytes(bHeader.LastBlockId.PartSetHeader.Hash).String(),
				},
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
		LastCommit: getLastCommit(b.Block.LastCommit),
	}
}

func getEvidence(evidences []types.Evidence) []structs.BlockEvidence {
	evidenceList := make([]structs.BlockEvidence, len(evidences))
	for i, evidence := range evidences {
		evidenceList[i] = structs.BlockEvidence(evidence.String())
	}
	return evidenceList
}

func getLastCommit(c *types.Commit) *structs.Commit {
	if c == nil {
		return nil
	}

	commitSigs := make([]structs.CommitSig, len(c.Signatures))
	for i, sig := range c.Signatures {
		commitSigs[i] = structs.CommitSig{
			BlockIdFlag:      int32(sig.BlockIdFlag),
			ValidatorAddress: bytes.HexBytes(sig.ValidatorAddress).String(),
			Timestamp:        sig.Timestamp,
			Signature:        bytes.HexBytes(sig.Signature).String(),
		}
	}

	return &structs.Commit{
		Height: c.Height,
		Round:  c.Round,
		BlockID: structs.BlockID{
			Hash: bytes.HexBytes(c.BlockID.Hash).String(),
			PartSetHeader: structs.PartSetHeader{
				Total: c.BlockID.PartSetHeader.Total,
				Hash:  bytes.HexBytes(c.BlockID.PartSetHeader.Hash).String(),
			}},
		Signatures: commitSigs,
	}
}
