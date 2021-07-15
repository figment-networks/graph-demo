package mapper

import (
	"fmt"

	"github.com/figment-networks/graph-demo/manager/structs"

	crisis "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/gogo/protobuf/proto"
)

// CrisisVerifyInvariantToSub transforms crisis.MsgVerifyInvariant sdk messages to SubsetEvent
func CrisisVerifyInvariantToSub(msg []byte) (se structs.SubsetEvent, er error) {
	mvi := &crisis.MsgVerifyInvariant{}
	if err := proto.Unmarshal(msg, mvi); err != nil {
		return se, fmt.Errorf("Not a crisis type: %w", err)
	}

	return structs.SubsetEvent{
		Type:   []string{"verify_invariant"},
		Module: "crisis",
		Sender: []structs.EventTransfer{{
			Account: structs.Account{ID: mvi.Sender},
		}},
		Additional: map[string][]string{
			"invariant_route":       {mvi.InvariantRoute},
			"invariant_module_name": {mvi.InvariantModuleName},
		},
	}, nil
}
