package mapper

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"
)

var (
	errUnknownMessageStruct = errors.New("unknown message struct")
	errUnexpectedMsgVersion = errors.New("unexpected message version")
	errUnknownMsgType       = errors.New("unknown message type")
)

func parseMessage(value []byte, typeURL string) (result json.RawMessage, err error) {
	parts := strings.Split(typeURL[1:], ".")
	partsLen := len(parts)
	msgType := parts[1]

	switch parts[0] {

	// example cosmos message type url:  "/cosmos.bank.v1beta1.MsgSend"
	case "cosmos":
		isCryptoMsg := msgType == "crypto"
		if partsLen != 4 && (isCryptoMsg && partsLen != 5) {
			return nil, fmt.Errorf("unexpected url format")
		}
		msgStruct := parts[3]
		version := parts[2]

		if isCryptoMsg {
			msgStruct = strings.Join(parts[2:], ".")
		}

		result, err = parseCosmosMessage(value, msgStruct, msgType, version)

	// example ibc message type url: "/ibc.applications.transfer.v1.MsgTransfer"
	case "ibc":
		if partsLen != 5 {
			return nil, fmt.Errorf("unexpected url format")
		}
		msgType += fmt.Sprintf(".%s", parts[2])
		msgStruct := parts[4]
		version := parts[3]

		result, err = parseIBCMessage(value, msgStruct, msgType, version)

	default:
		return nil, fmt.Errorf("unknown url value %q", typeURL)
	}

	if err != nil {
		return nil, fmt.Errorf("%s: typeURL: %q", err.Error(), typeURL)
	}

	return
}

func parseCosmosMessage(value []byte, msgStruct, msgType, version string) (raw json.RawMessage, err error) {
	if version != "v1beta1" && msgType != "crypto" {
		return nil, errUnexpectedMsgVersion
	}

	switch msgType {
	case "bank":
		return parseBankMessage(value, msgStruct)
	case "crypto":
		return parseCryptoMessage(value, msgStruct)
	case "distribution":
		return parseDistributionMessage(value, msgStruct)
	case "evidence":
		return parseEvidenceMessage(value, msgStruct)
	case "gov":
		return parseGovMessage(value, msgStruct)
	case "slashing":
		return parseSlashingMessage(value, msgStruct)
	case "staking":
		return parseStakingMessage(value, msgStruct)
	case "tx":
		return parseTxMessage(value, msgStruct)
	case "vesting":
		return parseVestingMessage(value, msgStruct)

	default:
		return nil, errUnknownMsgType
	}
}

func parseIBCMessage(value []byte, msgStruct, msgType, version string) (raw json.RawMessage, err error) {
	if version != "v1" {
		return nil, errUnexpectedMsgVersion
	}

	switch msgType {
	case "applications.transfer":
		return parseICBApplicationsTransferMessage(value, msgStruct)
	case "core.channel":
		return parseICBCoreChannelMessage(value, msgStruct)
	case "core.client":
		return parseICBCoreClientMessage(value, msgStruct)
	case "lightclients.tendermint":
		return parseICBLightclientsTendermintMessage(value, msgStruct)

	default:
		return nil, errUnknownMsgType
	}
}

func protoUnmarshal(value []byte, pb proto.Message) (raw json.RawMessage, err error) {
	if err := proto.Unmarshal(value, pb); err != nil {
		return nil, err
	}

	return proto.Marshal(pb)
}
