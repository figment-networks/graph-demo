package mapper

import (
	"encoding/json"

	distribution "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func parseDistributionMessage(value []byte, msgStruct string) (raw json.RawMessage, err error) {
	switch msgStruct {
	case "MsgFundCommunityPool":
		return protoUnmarshal(value, &distribution.MsgFundCommunityPool{})

	case "MsgFundCommunityPoolResponse":
		return protoUnmarshal(value, &distribution.MsgFundCommunityPoolResponse{})

	case "MsgSetWithdrawAddress":
		return protoUnmarshal(value, &distribution.MsgSetWithdrawAddress{})

	case "MsgSetWithdrawAddressResponse":
		return protoUnmarshal(value, &distribution.MsgSetWithdrawAddressResponse{})

	case "MsgWithdrawDelegatorReward":
		return protoUnmarshal(value, &distribution.MsgWithdrawDelegatorReward{})

	case "MsgWithdrawDelegatorRewardResponse":
		return protoUnmarshal(value, &distribution.MsgWithdrawDelegatorRewardResponse{})

	case "MsgWithdrawValidatorCommission":
		return protoUnmarshal(value, &distribution.MsgWithdrawValidatorCommission{})

	case "MsgWithdrawValidatorCommissionResponse":
		return protoUnmarshal(value, &distribution.MsgWithdrawValidatorCommissionResponse{})

	default:
		return nil, errUnknownMessageStruct
	}
}
