package mapper

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/types/tx"
)

func parseTxMessage(value []byte, msgStruct string) (raw json.RawMessage, err error) {
	switch msgStruct {
	case "AuthInfo":
		return protoUnmarshal(value, &tx.AuthInfo{})

	case "Fee":
		return protoUnmarshal(value, &tx.Fee{})

	case "ModeInfo":
		return protoUnmarshal(value, &tx.ModeInfo{})

	case "ModeInfo.Multi":
		return protoUnmarshal(value, &tx.ModeInfo_Multi{})

	case "ModeInfo.Single":
		return protoUnmarshal(value, &tx.ModeInfo_Single{})

	case "SignDoc":
		return protoUnmarshal(value, &tx.SignDoc{})

	case "SignerInfo":
		return protoUnmarshal(value, &tx.SignerInfo{})

	case "Tx":
		return protoUnmarshal(value, &tx.Tx{})

	case "TxBody":
		return protoUnmarshal(value, &tx.TxBody{})

	case "TxRaw":
		return protoUnmarshal(value, &tx.TxRaw{})

	default:
		return nil, errUnknownMessageStruct
	}
}
