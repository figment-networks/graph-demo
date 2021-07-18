package mapper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/figment-networks/graph-demo/manager/structs"
	"github.com/gogo/protobuf/proto"
)

// BankMultisendToSub transforms bank.MsgMultiSend sdk messages to SubsetEvent
func BankMultisendToSub(msg []byte, lg types.ABCIMessageLog) (se structs.SubsetEvent, err error) {
	multisend := &bank.MsgMultiSend{}
	if err := proto.Unmarshal(msg, multisend); err != nil {
		return se, fmt.Errorf("Not a multisend type: %w", err)
	}

	se = structs.SubsetEvent{
		Type:   []string{"multisend"},
		Module: "bank",
	}
	for _, i := range multisend.Inputs {
		evt, err := bankProduceEvTx(i.Address, i.Coins)
		if err != nil {
			continue
		}
		se.Sender = append(se.Sender, evt)
	}

	for _, o := range multisend.Outputs {
		evt, err := bankProduceEvTx(o.Address, o.Coins)
		if err != nil {
			continue
		}
		se.Recipient = append(se.Recipient, evt)
	}

	err = produceTransfers(&se, "send", "", lg)
	return se, err
}

// BankSendToSub transforms bank.MsgSend sdk messages to SubsetEvent
func BankSendToSub(msg []byte, lg types.ABCIMessageLog) (se structs.SubsetEvent, err error) {
	send := &bank.MsgSend{}
	if err := proto.Unmarshal(msg, send); err != nil {
		return se, fmt.Errorf("Not a send type: %w", err)
	}

	se = structs.SubsetEvent{
		Type:   []string{"send"},
		Module: "bank",
	}

	evt, _ := bankProduceEvTx(send.FromAddress, send.Amount)
	se.Sender = append(se.Sender, evt)

	evt, _ = bankProduceEvTx(send.ToAddress, send.Amount)
	se.Recipient = append(se.Recipient, evt)

	err = produceTransfers(&se, "send", "", lg)
	return se, err
}

func bankProduceEvTx(account string, coins types.Coins) (evt structs.EventTransfer, err error) {
	evt = structs.EventTransfer{
		Account: structs.Account{ID: account},
	}
	if len(coins) > 0 {
		evt.Amounts = []structs.TransactionAmount{}
		for _, coin := range coins {
			evt.Amounts = append(evt.Amounts, structs.TransactionAmount{
				Currency: coin.Denom,
				Numeric:  coin.Amount.BigInt(),
				Text:     coin.Amount.String(),
			})
		}
	}

	return evt, nil
}
