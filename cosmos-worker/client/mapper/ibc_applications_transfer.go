package mapper

import (
	"encoding/json"

	transfer "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
)

func parseICBApplicationsTransferMessage(value []byte, msgStruct string) (raw json.RawMessage, err error) {
	switch msgStruct {
	case "MsgTransfer":
		return protoUnmarshal(value, &transfer.MsgTransfer{})

	case "MsgTransferResponse":
		return protoUnmarshal(value, &transfer.MsgTransferResponse{})

	default:
		return nil, errUnknownMessageStruct
	}
}
