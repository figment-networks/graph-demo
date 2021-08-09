package mapper

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

	return jsonMarshalWithTypeAny(pb)
}

func jsonMarshalWithTypeAny(str interface{}) (raw json.RawMessage, err error) {
	strValue := reflect.ValueOf(str)
	kind := strValue.Kind()
	switch kind {
	case reflect.Slice:
		sliceLen := strValue.Len()
		byteResp := []byte("[")
		for i := 0; i < sliceLen; i++ {
			resp, err := jsonMarshalStructWithTypeAny(strValue.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			byteResp = append(byteResp, resp...)
			if sliceLen > 1 && i+1 < sliceLen {
				byteResp = append(byteResp, []byte(",")...)
			}
		}
		return append(byteResp, []byte("]")...), nil

	case reflect.Struct:
		return jsonMarshalStructWithTypeAny(str)
	case reflect.Ptr:
		return jsonMarshalWithTypeAny(strValue.Elem().Interface())
	default:
		return nil, errors.New("internal server error, should never happen")
	}
}

func jsonMarshalStructWithTypeAny(str interface{}) (raw json.RawMessage, err error) {
	v := reflect.Indirect(reflect.ValueOf(str))
	result := make(map[string]json.RawMessage)

	for i := 0; i < v.NumField(); i++ {
		fieldName := v.Type().Field(i).Name
		fieldValue := v.Field(i).Interface()
		fieldType := reflect.TypeOf(fieldValue)

		switch fieldType {
		case reflect.TypeOf([]*types.Any{}):
			anySlice := fieldValue.([]*types.Any)
			resultBytes := []byte("[")
			resultLen := len(anySlice)
			for i, any := range anySlice {
				msgRaw, err := parseMessage(any.Value, any.TypeUrl)
				if err != nil {
					return nil, err
				}

				resultBytes = append(resultBytes, msgRaw...)

				if resultLen > 1 && i+1 < resultLen {
					resultBytes = append(resultBytes, []byte(",")...)
				}

			}
			result[fieldName] = append(resultBytes, []byte("]")...)

		case reflect.TypeOf(&types.Any{}):
			any := fieldValue.(*types.Any)
			if result[fieldName], err = parseMessage(any.Value, any.TypeUrl); err != nil {
				return nil, err
			}

		default:
			if isSliceOrStructOrStructPtr(fieldType) {
				if anyTypeFounded := checkForTypeAnyRecursively(fieldValue); anyTypeFounded {
					result[fieldName], err = jsonMarshalWithTypeAny(fieldValue)
				} else {
					result[fieldName], err = json.Marshal(fieldValue)
				}
			} else {
				result[fieldName], err = json.Marshal(fieldValue)
			}

			if err != nil {
				return nil, err
			}
		}
	}

	return json.Marshal(result)
}

func isSliceOrStructOrStructPtr(fieldType reflect.Type) bool {
	fieldKind := fieldType.Kind()
	isArray := fieldKind == reflect.Array && elemIsStructOrStructPtr(fieldType.Elem())
	isSlice := fieldKind == reflect.Slice && elemIsStructOrStructPtr(fieldType.Elem())
	isStructOrStructPtr := elemIsStructOrStructPtr(fieldType)
	return isArray || isSlice || isStructOrStructPtr
}

func elemIsStructOrStructPtr(elemType reflect.Type) bool {
	elemKind := elemType.Kind()
	return elemKind == reflect.Struct || elemKind == reflect.Ptr && elemType.Elem().Kind() == reflect.Struct
}

func checkForTypeAnyRecursively(str interface{}) bool {
	strValue := reflect.ValueOf(str)
	v := reflect.Indirect(strValue)

	switch strValue.Kind() {
	case reflect.Slice:
		return checkForTypeAnyInSlice(v)
	case reflect.Struct:
		return checkForTypeAnyInStruct(v)
	case reflect.Ptr:
		return checkForTypeAnyRecursively(strValue.Elem().Interface())
	default:
		return false
	}
}

func checkForTypeAnyInSlice(v reflect.Value) bool {
	for i := 0; i < v.Len(); i++ {
		fieldValue := v.Index(i).Interface()
		fieldType := reflect.TypeOf(fieldValue)
		switch fieldType {
		case reflect.TypeOf(&types.Any{}):
			return true
		default:
			if isSliceOrStructOrStructPtr(fieldType) {
				if anyTypeFounded := checkForTypeAnyRecursively(fieldValue); anyTypeFounded {
					return true
				}
			}
		}
	}

	return false
}

func checkForTypeAnyInStruct(v reflect.Value) bool {
	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i).Interface()
		fieldType := reflect.TypeOf(fieldValue)
		switch fieldType {
		case reflect.TypeOf(&types.Any{}), reflect.TypeOf([]*types.Any{}):
			return true
		case reflect.TypeOf(sdk.Int{}):
			continue
		default:
			if isSliceOrStructOrStructPtr(fieldType) {
				if anyTypeFounded := checkForTypeAnyRecursively(fieldValue); anyTypeFounded {
					return true
				}
			}
		}
	}

	return false
}
