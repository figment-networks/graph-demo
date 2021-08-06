package mapper

import (
	"encoding/json"

	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

func parseVestingMessage(value []byte, msgStruct string) (raw json.RawMessage, err error) {
	switch msgStruct {
	case "MsgCreateVestingAccount":
		return protoUnmarshal(value, &vesting.MsgCreateVestingAccount{})

	case "MsgCreateVestingAccountResponse":
		return protoUnmarshal(value, &vesting.MsgCreateVestingAccountResponse{})

	default:
		return nil, errUnknownMessageStruct
	}
}
