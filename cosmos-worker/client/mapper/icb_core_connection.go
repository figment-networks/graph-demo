package mapper

import (
	"encoding/json"

	connection "github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
)

func parseICBCoreConnectionMessage(value []byte, msgStruct string) (raw json.RawMessage, err error) {
	switch msgStruct {
	case "MsgConnectionOpenAck":
		return protoUnmarshal(value, &connection.MsgConnectionOpenAck{})

	case "MsgConnectionOpenAckResponse":
		return protoUnmarshal(value, &connection.MsgConnectionOpenAckResponse{})

	case "MsgConnectionOpenConfirm":
		return protoUnmarshal(value, &connection.MsgConnectionOpenConfirm{})

	case "MsgConnectionOpenConfirmResponse":
		return protoUnmarshal(value, &connection.MsgConnectionOpenConfirmResponse{})

	case "MsgConnectionOpenInit":
		return protoUnmarshal(value, &connection.MsgConnectionOpenInit{})

	case "MsgConnectionOpenInitResponse":
		return protoUnmarshal(value, &connection.MsgConnectionOpenInitResponse{})

	case "MsgConnectionOpenTry":
		return protoUnmarshal(value, &connection.MsgConnectionOpenTry{})

	case "MsgConnectionOpenTryResponse":
		return protoUnmarshal(value, &connection.MsgConnectionOpenTryResponse{})

	default:
		return nil, errUnknownMessageStruct
	}
}
