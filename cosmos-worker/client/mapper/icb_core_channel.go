package mapper

import (
	"encoding/json"

	channel "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
)

func parseICBCoreChannelMessage(value []byte, msgStruct string) (raw json.RawMessage, err error) {
	switch msgStruct {
	case "MsgAcknowledgement":
		return protoUnmarshal(value, &channel.MsgAcknowledgement{})

	case "MsgAcknowledgementResponse":
		return protoUnmarshal(value, &channel.MsgAcknowledgementResponse{})

	case "MsgChannelCloseConfirm":
		return protoUnmarshal(value, &channel.MsgChannelCloseConfirm{})

	case "MsgChannelCloseConfirmResponse":
		return protoUnmarshal(value, &channel.MsgChannelCloseConfirmResponse{})

	case "MsgChannelCloseInit":
		return protoUnmarshal(value, &channel.MsgChannelCloseInit{})

	case "MsgChannelCloseInitResponse":
		return protoUnmarshal(value, &channel.MsgChannelCloseInitResponse{})

	case "MsgChannelOpenAck":
		return protoUnmarshal(value, &channel.MsgChannelOpenAck{})

	case "MsgChannelOpenAckResponse":
		return protoUnmarshal(value, &channel.MsgChannelOpenAckResponse{})

	case "MsgChannelOpenConfirm":
		return protoUnmarshal(value, &channel.MsgChannelOpenConfirm{})

	case "MsgChannelOpenConfirmResponse":
		return protoUnmarshal(value, &channel.MsgChannelOpenConfirmResponse{})

	case "MsgChannelOpenInit":
		return protoUnmarshal(value, &channel.MsgChannelOpenInit{})

	case "MsgChannelOpenInitResponse":
		return protoUnmarshal(value, &channel.MsgChannelOpenInitResponse{})

	case "MsgChannelOpenTry":
		return protoUnmarshal(value, &channel.MsgChannelOpenTry{})

	case "MsgChannelOpenTryResponse":
		return protoUnmarshal(value, &channel.MsgChannelOpenTryResponse{})

	case "MsgRecvPacket":
		return protoUnmarshal(value, &channel.MsgRecvPacket{})

	case "MsgRecvPacketResponse":
		return protoUnmarshal(value, &channel.MsgRecvPacketResponse{})

	case "MsgTimeout":
		return protoUnmarshal(value, &channel.MsgTimeout{})

	case "MsgTimeoutOnClose":
		return protoUnmarshal(value, &channel.MsgTimeoutOnClose{})

	case "MsgTimeoutOnCloseResponse":
		return protoUnmarshal(value, &channel.MsgTimeoutOnCloseResponse{})

	case "MsgTimeoutResponse":
		return protoUnmarshal(value, &channel.MsgTimeoutResponse{})

	default:
		return nil, errUnknownMessageStruct
	}
}
