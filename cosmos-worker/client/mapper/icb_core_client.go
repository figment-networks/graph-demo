package mapper

import (
	"encoding/json"

	client "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
)

func parseICBCoreClientMessage(value []byte, msgStruct string) (raw json.RawMessage, err error) {
	switch msgStruct {
	case "MsgCreateClient":
		return protoUnmarshal(value, &client.MsgCreateClient{})

	case "MsgCreateClientResponse":
		return protoUnmarshal(value, &client.MsgCreateClientResponse{})

	case "MsgSubmitMisbehaviour":
		return protoUnmarshal(value, &client.MsgSubmitMisbehaviour{})

	case "MsgSubmitMisbehaviourResponse":
		return protoUnmarshal(value, &client.MsgSubmitMisbehaviourResponse{})

	case "MsgUpdateClient":
		return protoUnmarshal(value, &client.MsgUpdateClient{})

	case "MsgUpdateClientResponse":
		return protoUnmarshal(value, &client.MsgUpdateClientResponse{})

	case "MsgUpgradeClient":
		return protoUnmarshal(value, &client.MsgUpgradeClient{})

	case "MsgUpgradeClientResponse":
		return protoUnmarshal(value, &client.MsgUpgradeClientResponse{})

	default:
		return nil, errUnknownMessageStruct
	}
}
