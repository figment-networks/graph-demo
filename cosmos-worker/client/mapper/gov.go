package mapper

import (
	"encoding/json"

	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
)

func parseGovMessage(value []byte, msgStruct string) (raw json.RawMessage, err error) {
	switch msgStruct {
	case "MsgDeposit":
		return protoUnmarshal(value, &gov.MsgDeposit{})

	case "MsgDepositResponse":
		return protoUnmarshal(value, &gov.MsgDepositResponse{})

	case "MsgSubmitProposal":
		return protoUnmarshal(value, &gov.MsgSubmitProposal{})

	case "MsgSubmitProposalResponse":
		return protoUnmarshal(value, &gov.MsgSubmitProposalResponse{})

	case "MsgVote":
		return protoUnmarshal(value, &gov.MsgVote{})

	case "MsgVoteResponse":
		return protoUnmarshal(value, &gov.MsgVoteResponse{})

	default:
		return nil, errUnknownMessageStruct
	}
}
