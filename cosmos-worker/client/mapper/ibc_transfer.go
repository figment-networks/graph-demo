package mapper

import (
	"fmt"

	"github.com/figment-networks/graph-demo/manager/structs"

	transfer "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	"github.com/gogo/protobuf/proto"
)

// IBCTransferToSub transforms ibc.MsgTransfer sdk messages to SubsetEvent
func IBCTransferToSub(msg []byte) (se structs.SubsetEvent, err error) {
	m := &transfer.MsgTransfer{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a transfer type: %w", err)
	}

	return structs.SubsetEvent{
		Type:   []string{"transfer"},
		Module: "ibc",
	}, nil
}
