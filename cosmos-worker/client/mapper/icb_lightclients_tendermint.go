package mapper

import (
	"encoding/json"

	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
)

func parseICBLightclientsTendermintMessage(value []byte, msgStruct string) (raw json.RawMessage, err error) {
	switch msgStruct {
	case "ClientState":
		return protoUnmarshal(value, &tendermint.ClientState{})

	case "ConsensusState":
		return protoUnmarshal(value, &tendermint.ConsensusState{})

	case "Fraction":
		return protoUnmarshal(value, &tendermint.Fraction{})

	case "Header":
		return protoUnmarshal(value, &tendermint.Header{})

	case "Misbehaviour":
		return protoUnmarshal(value, &tendermint.Misbehaviour{})

	default:
		return nil, errUnknownMessageStruct
	}
}
