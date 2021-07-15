package mapper

import (
	"fmt"

	"github.com/figment-networks/graph-demo/manager/structs"

	connection "github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	"github.com/gogo/protobuf/proto"
)

// IBCConnectionOpenInitToSub transforms ibc.MsgConnectionOpenInit sdk messages to SubsetEvent
func IBCConnectionOpenInitToSub(msg []byte) (se structs.SubsetEvent, err error) {
	m := &connection.MsgConnectionOpenInit{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a connection_open_init type: %w", err)
	}

	return structs.SubsetEvent{
		Type:   []string{"connection_open_init"},
		Module: "ibc",
	}, nil
}

// IBCConnectionOpenConfirmToSub transforms ibc.MsgConnectionOpenConfirm sdk messages to SubsetEvent
func IBCConnectionOpenConfirmToSub(msg []byte) (se structs.SubsetEvent, err error) {
	m := &connection.MsgConnectionOpenConfirm{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a connection_open_confirm type: %w", err)
	}

	return structs.SubsetEvent{
		Type:   []string{"connection_open_confirm"},
		Module: "ibc",
	}, nil
}

// IBCConnectionOpenAckToSub transforms ibc.MsgConnectionOpenAck sdk messages to SubsetEvent
func IBCConnectionOpenAckToSub(msg []byte) (se structs.SubsetEvent, err error) {
	m := &connection.MsgConnectionOpenAck{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a connection_open_ack type: %w", err)
	}

	return structs.SubsetEvent{
		Type:   []string{"connection_open_ack"},
		Module: "ibc",
	}, nil
}

// IBCConnectionOpenTryToSub transforms ibc.MsgConnectionOpenTry sdk messages to SubsetEvent
func IBCConnectionOpenTryToSub(msg []byte) (se structs.SubsetEvent, err error) {
	m := &connection.MsgConnectionOpenTry{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a connection_open_try type: %w", err)
	}

	return structs.SubsetEvent{
		Type:   []string{"connection_open_try"},
		Module: "ibc",
	}, nil
}
