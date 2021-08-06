package mapper

import (
	"encoding/json"

	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func parseBankMessage(value []byte, msgStruct string) (raw json.RawMessage, err error) {
	switch msgStruct {
	case "MsgMultiSend":
		return protoUnmarshal(value, &bank.MsgMultiSend{})

	case "MsgMultiSendResponse":
		return protoUnmarshal(value, &bank.MsgMultiSendResponse{})

	case "MsgSend":
		return protoUnmarshal(value, &bank.MsgSend{})

	case "MsgSendResponse":
		return protoUnmarshal(value, &bank.MsgSendResponse{})

	default:
		return nil, errUnknownMessageStruct
	}
}
