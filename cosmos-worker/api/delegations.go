package api

import (
	"context"
	"fmt"
	"strconv"

	"github.com/figment-networks/indexing-engine/structs"

	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"google.golang.org/grpc/metadata"
)

// GetAccountDelegations fetches account delegations
func (c *Client) GetAccountDelegations(ctx context.Context, params structs.HeightAccount) (resp structs.GetAccountDelegationsResponse, err error) {
	resp.Height = params.Height

	delResp, err := c.stakingClient.DelegatorDelegations(metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, strconv.FormatUint(params.Height, 10)),
		&types.QueryDelegatorDelegationsRequest{DelegatorAddr: params.Account})
	if err != nil {
		return resp, fmt.Errorf("[COSMOS-API] Error fetching delegations: %w", err)
	}

	for _, dr := range delResp.DelegationResponses {
		resp.Delegations = append(resp.Delegations,
			structs.Delegation{
				Delegator: dr.Delegation.DelegatorAddress,
				Validator: structs.Validator(dr.Delegation.ValidatorAddress),
				Shares: structs.RewardAmount{
					Numeric: dr.Delegation.Shares.BigInt(),
					Exp:     sdk.Precision,
				},
				Balance: structs.RewardAmount{
					Text:     dr.Balance.Amount.String(),
					Numeric:  dr.Balance.Amount.BigInt(),
					Currency: dr.Balance.Denom,
				},
			},
		)
	}

	return resp, err
}
