package api

import (
	"context"
	"fmt"
	"strconv"

	"github.com/figment-networks/indexing-engine/structs"

	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	"google.golang.org/grpc/metadata"
)

// GetAccountBalance fetches account balance
func (c *Client) GetAccountBalance(ctx context.Context, params structs.HeightAccount) (resp structs.GetAccountBalanceResponse, err error) {
	resp.Height = params.Height

	balResp, err := c.bankClient.AllBalances(metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, strconv.FormatUint(params.Height, 10)),
		&types.QueryAllBalancesRequest{Address: params.Account})
	if err != nil {
		return resp, fmt.Errorf("[COSMOS-API] Error fetching balances: %w", err)
	}

	for _, blnc := range balResp.Balances {
		resp.Balances = append(resp.Balances,
			structs.TransactionAmount{
				Text:     blnc.Amount.String(),
				Numeric:  blnc.Amount.BigInt(),
				Currency: blnc.Denom,
			},
		)
	}

	return resp, err
}
