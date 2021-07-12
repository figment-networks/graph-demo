package mapper

import (
	"fmt"
	"strconv"

	"github.com/figment-networks/indexing-engine/structs"

	"github.com/cosmos/cosmos-sdk/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/gogo/protobuf/proto"
)

// GovDepositToSub transforms gov.MsgDeposit sdk messages to SubsetEvent
func GovDepositToSub(msg []byte, lg types.ABCIMessageLog) (se structs.SubsetEvent, err error) {
	dep := &gov.MsgDeposit{}
	if err := proto.Unmarshal(msg, dep); err != nil {
		return se, fmt.Errorf("Not a deposit type: %w", err)
	}

	se = structs.SubsetEvent{
		Type:       []string{"deposit"},
		Module:     "gov",
		Node:       map[string][]structs.Account{"depositor": {{ID: dep.Depositor}}},
		Additional: map[string][]string{"proposalID": {strconv.FormatUint(dep.ProposalId, 10)}},
	}

	sender := structs.EventTransfer{Account: structs.Account{ID: dep.Depositor}}
	txAmount := map[string]structs.TransactionAmount{}

	for i, coin := range dep.Amount {
		am := structs.TransactionAmount{
			Currency: coin.Denom,
			Numeric:  coin.Amount.BigInt(),
			Text:     coin.Amount.String(),
		}

		sender.Amounts = append(sender.Amounts, am)
		key := "deposit"
		if i > 0 {
			key += "_" + strconv.Itoa(i)
		}

		txAmount[key] = am
	}

	se.Sender = []structs.EventTransfer{sender}
	se.Amount = txAmount

	err = produceTransfers(&se, "send", "", lg)
	return se, err
}

// GovVoteToSub transforms gov.MsgVote sdk messages to SubsetEvent
func GovVoteToSub(msg []byte) (se structs.SubsetEvent, err error) {
	vote := &gov.MsgVote{}
	if err := proto.Unmarshal(msg, vote); err != nil {
		return se, fmt.Errorf("Not a vote type: %w", err)
	}

	return structs.SubsetEvent{
		Type:   []string{"vote"},
		Module: "gov",
		Node:   map[string][]structs.Account{"voter": {{ID: vote.Voter}}},
		Additional: map[string][]string{
			"proposalID": {strconv.FormatUint(vote.ProposalId, 10)},
			"option":     {vote.Option.String()},
		},
	}, nil
}

// GovSubmitProposalToSub transforms gov.MsgSubmitProposal sdk messages to SubsetEvent
func GovSubmitProposalToSub(msg []byte, lg types.ABCIMessageLog) (se structs.SubsetEvent, err error) {
	sp := &gov.MsgSubmitProposal{}
	if err := proto.Unmarshal(msg, sp); err != nil {
		return se, fmt.Errorf("Not a submit_proposal type: %w", err)
	}

	se = structs.SubsetEvent{
		Type:   []string{"submit_proposal"},
		Module: "gov",
		Node:   map[string][]structs.Account{"proposer": {{ID: sp.Proposer}}},
	}

	sender := structs.EventTransfer{Account: structs.Account{ID: sp.Proposer}}
	txAmount := map[string]structs.TransactionAmount{}

	for i, coin := range sp.InitialDeposit {
		am := structs.TransactionAmount{
			Currency: coin.Denom,
			Numeric:  coin.Amount.BigInt(),
			Text:     coin.Amount.String(),
		}

		sender.Amounts = append(sender.Amounts, am)
		key := "initial_deposit"
		if i > 0 {
			key += "_" + strconv.Itoa(i)
		}

		txAmount[key] = am
	}
	se.Sender = []structs.EventTransfer{sender}
	se.Amount = txAmount

	err = produceTransfers(&se, "send", "", lg)
	if err != nil {
		return se, err
	}

	content := sp.GetContent()
	if content == nil {
		return se, nil
	}
	se.Additional = map[string][]string{}

	if content.ProposalRoute() != "" {
		se.Additional["proposal_route"] = []string{content.ProposalRoute()}
	}
	if content.ProposalType() != "" {
		se.Additional["proposal_type"] = []string{content.ProposalType()}
	}
	if content.GetDescription() != "" {
		se.Additional["descritpion"] = []string{content.GetDescription()}
	}
	if content.GetTitle() != "" {
		se.Additional["title"] = []string{content.GetTitle()}
	}
	if content.String() != "" {
		se.Additional["content"] = []string{content.String()}
	}

	return se, nil
}
