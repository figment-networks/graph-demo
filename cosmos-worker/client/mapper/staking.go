package mapper

import (
	"encoding/json"

	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func parseStakingMessage(value []byte, msgStruct string) (raw json.RawMessage, err error) {
	switch msgStruct {
	case "MsgBeginRedelegate":
		return protoUnmarshal(value, &staking.MsgBeginRedelegate{})

	case "MsgBeginRedelegateResponse":
		return protoUnmarshal(value, &staking.MsgBeginRedelegateResponse{})

	case "MsgCreateValidator":
		return protoUnmarshal(value, &staking.MsgCreateValidator{})

	case "MsgCreateValidatorResponse":
		return protoUnmarshal(value, &staking.MsgCreateValidatorResponse{})

	case "MsgDelegate":
		return protoUnmarshal(value, &staking.MsgDelegate{})

	case "MsgDelegateResponse":
		return protoUnmarshal(value, &staking.MsgDelegateResponse{})

	case "MsgEditValidator":
		return protoUnmarshal(value, &staking.MsgEditValidator{})

	case "MsgEditValidatorResponse":
		return protoUnmarshal(value, &staking.MsgEditValidatorResponse{})

	case "MsgUndelegate":
		return protoUnmarshal(value, &staking.MsgUndelegate{})

	case "MsgUndelegateResponse":
		return protoUnmarshal(value, &staking.MsgUndelegateResponse{})

	default:
		return nil, errUnknownMessageStruct
	}
}
