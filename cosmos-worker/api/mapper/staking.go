package mapper

import (
	"fmt"

	"github.com/figment-networks/indexing-engine/structs"

	"github.com/cosmos/cosmos-sdk/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gogo/protobuf/proto"
)

const unbondedAddr = "cosmos1tygms3xhhs3yv487phx3dw4a95jn7t7lpm470r"

// StakingUndelegateToSub transforms staking.MsgUndelegate sdk messages to SubsetEvent
func StakingUndelegateToSub(msg []byte, lg types.ABCIMessageLog) (se structs.SubsetEvent, err error) {
	u := &staking.MsgUndelegate{}
	if err := proto.Unmarshal(msg, u); err != nil {
		return se, fmt.Errorf("Not a undelegate type: %w", err)
	}
	se = structs.SubsetEvent{
		Type:   []string{"undelegate"},
		Module: "staking",
		Node: map[string][]structs.Account{
			"delegator": {{ID: u.DelegatorAddress}},
			"validator": {{ID: u.ValidatorAddress}},
		},
		Amount: map[string]structs.TransactionAmount{
			"undelegate": {
				Currency: u.Amount.Denom,
				Numeric:  u.Amount.Amount.BigInt(),
				Text:     u.Amount.String(),
			},
		},
	}

	err = produceTransfers(&se, "reward", unbondedAddr, lg)
	return se, err
}

// StakingDelegateToSub transforms staking.MsgDelegate sdk messages to SubsetEvent
func StakingDelegateToSub(msg []byte, lg types.ABCIMessageLog) (se structs.SubsetEvent, err error) {
	d := &staking.MsgDelegate{}
	if err := proto.Unmarshal(msg, d); err != nil {
		return se, fmt.Errorf("Not a delegate type: %w", err)
	}

	se = structs.SubsetEvent{
		Type:   []string{"delegate"},
		Module: "staking",
		Node: map[string][]structs.Account{
			"delegator": {{ID: d.DelegatorAddress}},
			"validator": {{ID: d.ValidatorAddress}},
		},
		Amount: map[string]structs.TransactionAmount{
			"delegate": {
				Currency: d.Amount.Denom,
				Numeric:  d.Amount.Amount.BigInt(),
				Text:     d.Amount.String(),
			},
		},
	}

	err = produceTransfers(&se, "reward", "", lg)
	return se, err
}

// StakingBeginRedelegateToSub transforms staking.MsgBeginRedelegate sdk messages to SubsetEvent
func StakingBeginRedelegateToSub(msg []byte, lg types.ABCIMessageLog) (se structs.SubsetEvent, err error) {
	br := &staking.MsgBeginRedelegate{}
	if err := proto.Unmarshal(msg, br); err != nil {
		return se, fmt.Errorf("Not a begin_redelegate type: %w", err)
	}

	se = structs.SubsetEvent{
		Type:   []string{"begin_redelegate"},
		Module: "staking",
		Node: map[string][]structs.Account{
			"delegator":             {{ID: br.DelegatorAddress}},
			"validator_destination": {{ID: br.ValidatorDstAddress}},
			"validator_source":      {{ID: br.ValidatorSrcAddress}},
		},
		Amount: map[string]structs.TransactionAmount{
			"delegate": {
				Currency: br.Amount.Denom,
				Numeric:  br.Amount.Amount.BigInt(),
				Text:     br.Amount.String(),
			},
		},
	}

	err = produceTransfers(&se, "reward", "", lg)
	return se, err
}

// StakingCreateValidatorToSub transforms staking.MsgCreateValidator sdk messages to SubsetEvent
func StakingCreateValidatorToSub(msg []byte) (se structs.SubsetEvent, err error) {
	ev := &staking.MsgCreateValidator{}
	if err := proto.Unmarshal(msg, ev); err != nil {
		return se, fmt.Errorf("Not a create_validator type: %w", err)
	}
	return structs.SubsetEvent{
		Type:   []string{"create_validator"},
		Module: "distribution",
		Node: map[string][]structs.Account{
			"delegator": {{ID: ev.DelegatorAddress}},
			"validator": {
				{
					ID: ev.ValidatorAddress,
					Details: &structs.AccountDetails{
						Name:        ev.Description.Moniker,
						Description: ev.Description.Details,
						Contact:     ev.Description.SecurityContact,
						Website:     ev.Description.Website,
					},
				},
			},
		},
		Amount: map[string]structs.TransactionAmount{
			"self_delegation": {
				Currency: ev.Value.Denom,
				Numeric:  ev.Value.Amount.BigInt(),
				Text:     ev.Value.String(),
			},
			"self_delegation_min": {
				Text:    ev.MinSelfDelegation.String(),
				Numeric: ev.MinSelfDelegation.BigInt(),
			},
			"commission_rate": {
				Text:    ev.Commission.Rate.String(),
				Numeric: ev.Commission.Rate.BigInt(),
			},
			"commission_max_rate": {
				Text:    ev.Commission.MaxRate.String(),
				Numeric: ev.Commission.MaxRate.BigInt(),
			},
			"commission_max_change_rate": {
				Text:    ev.Commission.MaxChangeRate.String(),
				Numeric: ev.Commission.MaxChangeRate.BigInt(),
			}},
	}, err
}

// StakingEditValidatorToSub transforms staking.MsgEditValidator sdk messages to SubsetEvent
func StakingEditValidatorToSub(msg []byte) (se structs.SubsetEvent, err error) {
	ev := &staking.MsgEditValidator{}
	if err := proto.Unmarshal(msg, ev); err != nil {
		return se, fmt.Errorf("Not a edit_validator type: %w", err)
	}
	sev := structs.SubsetEvent{
		Type:   []string{"edit_validator"},
		Module: "distribution",
		Node: map[string][]structs.Account{
			"validator": {
				{
					ID: ev.ValidatorAddress,
					Details: &structs.AccountDetails{
						Name:        ev.Description.Moniker,
						Description: ev.Description.Details,
						Contact:     ev.Description.SecurityContact,
						Website:     ev.Description.Website,
					},
				},
			},
		},
	}

	if ev.MinSelfDelegation != nil || ev.CommissionRate != nil {
		sev.Amount = map[string]structs.TransactionAmount{}
		if ev.MinSelfDelegation != nil {
			sev.Amount["self_delegation_min"] = structs.TransactionAmount{
				Text:    ev.MinSelfDelegation.String(),
				Numeric: ev.MinSelfDelegation.BigInt(),
			}
		}

		if ev.CommissionRate != nil {
			sev.Amount["commission_rate"] = structs.TransactionAmount{
				Text:    ev.CommissionRate.String(),
				Numeric: ev.CommissionRate.BigInt(),
			}
		}
	}
	return sev, err
}
