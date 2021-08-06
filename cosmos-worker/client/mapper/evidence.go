package mapper

import (
	"encoding/json"

	evidence "github.com/cosmos/cosmos-sdk/x/evidence/types"
)

func parseEvidenceMessage(value []byte, msgStruct string) (raw json.RawMessage, err error) {
	switch msgStruct {
	// checkEvidence type any
	case "MsgSubmitEvidence":
		return protoUnmarshal(value, &evidence.MsgSubmitEvidence{})

	case "MsgSubmitEvidenceResponse":
		return protoUnmarshal(value, &evidence.MsgSubmitEvidenceResponse{})

	default:
		return nil, errUnknownMessageStruct
	}
}
