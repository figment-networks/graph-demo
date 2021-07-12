package api

import (
	"context"
	"fmt"
	"strconv"

	"github.com/figment-networks/indexing-engine/structs"

	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"google.golang.org/grpc/metadata"
)

type responseWithHeight struct {
	Height string                                   `json:"height"`
	Result types.QueryDelegatorTotalRewardsResponse `json:"result"`
}

const maxRetries = 3

// GetReward fetches total rewards for delegator account
func (c *Client) GetReward(ctx context.Context, params structs.HeightAccount) (resp structs.GetRewardResponse, err error) {
	resp.Height = params.Height
	resp.Rewards = make(map[structs.Validator][]structs.RewardAmount, 0)

	valResp, err := c.distributionClient.DelegatorValidators(metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, strconv.FormatUint(params.Height, 10)),
		&types.QueryDelegatorValidatorsRequest{DelegatorAddress: params.Account})
	if err != nil {
		return resp, fmt.Errorf("[COSMOS-API] Error fetching validators: %w", err)
	}

	for _, val := range valResp.Validators {
		delResp, err := c.distributionClient.DelegationRewards(metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, strconv.FormatUint(params.Height, 10)),
			&types.QueryDelegationRewardsRequest{DelegatorAddress: params.Account, ValidatorAddress: val})
		if err != nil {
			return resp, fmt.Errorf("[COSMOS-API] Error fetching delegation rewards: %w", err)
		}

		valRewards := make([]structs.RewardAmount, 0, len(delResp.GetRewards()))
		for _, reward := range delResp.GetRewards() {
			valRewards = append(valRewards,
				structs.RewardAmount{
					Text:     reward.Amount.String(),
					Numeric:  reward.Amount.BigInt(),
					Currency: reward.Denom,
					Exp:      sdk.Precision,
				},
			)
		}

		resp.Rewards[structs.Validator(val)] = valRewards
	}

	return resp, err
}
