package mapper

import (
	"fmt"
	"strconv"

	"github.com/figment-networks/graph-demo/manager/structs"

	"github.com/cosmos/cosmos-sdk/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/gogo/protobuf/proto"
)

// VestingMsgCreateVestingAccountToSub transforms vesting.MsgCreateVestingAccount sdk messages to SubsetEvent
func VestingMsgCreateVestingAccountToSub(msg []byte, lg types.ABCIMessageLog) (se structs.SubsetEvent, err error) {

	cva := &vesting.MsgCreateVestingAccount{}
	if err := proto.Unmarshal(msg, cva); err != nil {
		return se, fmt.Errorf("not a msg_create_vesting_account type: %w", err)
	}

	se = structs.SubsetEvent{
		Type:       []string{"msg_create_vesting_account"},
		Module:     "vesting",
		Additional: make(map[string][]string),
	}

	evt, _ := vestingProduceEvTx(cva.FromAddress, cva.Amount)
	se.Sender = append(se.Sender, evt)

	evtV, _ := vestingProduceEvTx(cva.ToAddress, cva.Amount)
	se.Recipient = append(se.Recipient, evtV)

	se.Additional["end_time"] = []string{strconv.Itoa(int(cva.EndTime))}
	del := "false"
	if cva.Delayed {
		del = "true"
	}
	se.Additional["delayed"] = []string{del}

	return se, err
}

func vestingProduceEvTx(account string, coins types.Coins) (evt structs.EventTransfer, err error) {
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
