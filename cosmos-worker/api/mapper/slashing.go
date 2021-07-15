package mapper

import (
	"fmt"

	"github.com/figment-networks/graph-demo/manager/structs"

	slashing "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/gogo/protobuf/proto"
)

// SlashingUnjailToSub transforms slashing.MsgUnjail sdk messages to SubsetEvent
func SlashingUnjailToSub(msg []byte) (se structs.SubsetEvent, er error) {
	unjail := &slashing.MsgUnjail{}
	if err := proto.Unmarshal(msg, unjail); err != nil {
		return se, fmt.Errorf("Not a unjail type: %w", err)
	}

	return structs.SubsetEvent{
		Type:   []string{"unjail"},
		Module: "slashing",
		Node:   map[string][]structs.Account{"validator": {{ID: unjail.ValidatorAddr}}},
	}, nil
}
