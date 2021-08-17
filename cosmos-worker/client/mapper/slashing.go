package mapper

import (
	"encoding/json"

	slashing "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

func parseSlashingMessage(value []byte, msgStruct string) (raw json.RawMessage, err error) {
	switch msgStruct {
	case "MsgUnjail":
		return protoUnmarshal(value, &slashing.MsgUnjail{})

	case "MsgUnjailResponse":
		return protoUnmarshal(value, &slashing.MsgUnjail{})

	default:
		return nil, errUnknownMessageStruct
	}
}
