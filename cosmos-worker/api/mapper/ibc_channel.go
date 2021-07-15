package mapper

import (
	"fmt"

	"github.com/figment-networks/graph-demo/manager/structs"

	channel "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	"github.com/gogo/protobuf/proto"
)

// IBCChannelOpenInitToSub transforms ibc.MsgChannelOpenInit sdk messages to SubsetEvent
func IBCChannelOpenInitToSub(msg []byte) (se structs.SubsetEvent, err error) {
	m := &channel.MsgChannelOpenInit{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a channel_open_init type: %w", err)
	}

	return structs.SubsetEvent{
		Type:   []string{"channel_open_init"},
		Module: "ibc",
	}, nil
}

// IBCChannelOpenConfirmToSub transforms ibc.MsgChannelOpenConfirm sdk messages to SubsetEvent
func IBCChannelOpenConfirmToSub(msg []byte) (se structs.SubsetEvent, err error) {
	m := &channel.MsgChannelOpenConfirm{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a channel_open_confirm type: %w", err)
	}

	return structs.SubsetEvent{
		Type:   []string{"channel_open_confirm"},
		Module: "ibc",
	}, nil
}

// IBCChannelOpenAckToSub transforms ibc.MsgChannelOpenAck sdk messages to SubsetEvent
func IBCChannelOpenAckToSub(msg []byte) (se structs.SubsetEvent, err error) {
	m := &channel.MsgChannelOpenAck{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a channel_open_ack type: %w", err)
	}

	return structs.SubsetEvent{
		Type:   []string{"channel_open_ack"},
		Module: "ibc",
	}, nil
}

// IBCChannelOpenTryToSub transforms ibc.MsgChannelOpenTry sdk messages to SubsetEvent
func IBCChannelOpenTryToSub(msg []byte) (se structs.SubsetEvent, err error) {
	m := &channel.MsgChannelOpenTry{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a channel_open_try type: %w", err)
	}

	return structs.SubsetEvent{
		Type:   []string{"channel_open_try"},
		Module: "ibc",
	}, nil
}

// IBCChannelCloseInitToSub transforms ibc.MsgChannelCloseInit sdk messages to SubsetEvent
func IBCChannelCloseInitToSub(msg []byte) (se structs.SubsetEvent, err error) {
	m := &channel.MsgChannelCloseInit{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a channel_close_init type: %w", err)
	}

	return structs.SubsetEvent{
		Type:   []string{"channel_close_init"},
		Module: "ibc",
	}, nil
}

// IBCChannelCloseConfirmToSub transforms ibc.MsgChannelCloseConfirm sdk messages to SubsetEvent
func IBCChannelCloseConfirmToSub(msg []byte) (se structs.SubsetEvent, err error) {
	m := &channel.MsgChannelCloseConfirm{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a channel_close_confirm type: %w", err)
	}

	return structs.SubsetEvent{
		Type:   []string{"channel_close_confirm"},
		Module: "ibc",
	}, nil
}

// IBCChannelRecvPacketToSub transforms ibc.MsgRecvPacket sdk messages to SubsetEvent
func IBCChannelRecvPacketToSub(msg []byte) (se structs.SubsetEvent, err error) {
	m := &channel.MsgRecvPacket{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a recv_packet type: %w", err)
	}

	return structs.SubsetEvent{
		Type:   []string{"recv_packet"},
		Module: "ibc",
	}, nil
}

// IBCChannelTimeoutToSub transforms ibc.MsgTimeout sdk messages to SubsetEvent
func IBCChannelTimeoutToSub(msg []byte) (se structs.SubsetEvent, err error) {
	m := &channel.MsgTimeout{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a timeout type: %w", err)
	}

	return structs.SubsetEvent{
		Type:   []string{"timeout"},
		Module: "ibc",
	}, nil
}

// IBCChannelAcknowledgementToSub transforms ibc.MsgAcknowledgement sdk messages to SubsetEvent
func IBCChannelAcknowledgementToSub(msg []byte) (se structs.SubsetEvent, err error) {
	m := &channel.MsgAcknowledgement{}
	if err := proto.Unmarshal(msg, m); err != nil {
		return se, fmt.Errorf("Not a channel_acknowledgement type: %w", err)
	}

	return structs.SubsetEvent{
		Type:   []string{"channel_acknowledgement"},
		Module: "ibc",
	}, nil
}
