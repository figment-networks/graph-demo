package mapper

import (
	"encoding/json"
	"errors"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/tendermint/tendermint/libs/bytes"
)

func parseCryptoMessage(value []byte, msgStruct string) (raw json.RawMessage, err error) {
	switch msgStruct {
	case "ed25519.PrivKey":
		return ed25519PrivKey(value)
	case "ed25519.PubKey":
		return ed25519PubKey(value)
	case "multisig.LegacyAminoPubKey":
		return nil, errors.New("how")
	case "multisig.v1beta1.CompactBitArray":
		return nil, errors.New("how")
	case "multisig.v1beta1.MultiSignature":
		return nil, errors.New("how")
	case "secp256k1.PrivKey":
		return secp256k1PrivKey(value)
	case "secp256k1.PubKey":
		return secp256k1PubKey(value)

	default:
		return nil, errUnknownMessageStruct
	}
}

type key struct {
	Key string `json:"key,omitempty"`
}

func ed25519PrivKey(value []byte) (json.RawMessage, error) {
	pk := ed25519.PrivKey{}
	if err := pk.Unmarshal(value); err != nil {
		return nil, err
	}
	return json.Marshal(key{
		Key: bytes.HexBytes(pk.Key).String(),
	})
}

func ed25519PubKey(value []byte) (json.RawMessage, error) {
	pk := ed25519.PubKey{}
	if err := pk.Unmarshal(value); err != nil {
		return nil, err
	}
	return json.Marshal(key{
		Key: bytes.HexBytes(pk.Key).String(),
	})
}

func secp256k1PrivKey(value []byte) (json.RawMessage, error) {
	pk := secp256k1.PrivKey{}
	if err := pk.Unmarshal(value); err != nil {
		return nil, err
	}

	return json.Marshal(key{
		Key: bytes.HexBytes(pk.Key).String(),
	})
}

func secp256k1PubKey(value []byte) (json.RawMessage, error) {
	pk := secp256k1.PubKey{}
	if err := pk.Unmarshal(value); err != nil {
		return nil, err
	}

	return json.Marshal(key{
		Key: bytes.HexBytes(pk.Key).String(),
	})
}
