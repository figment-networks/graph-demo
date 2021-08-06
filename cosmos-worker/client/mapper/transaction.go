package mapper

import (
	"context"
	"fmt"
	"time"

	"github.com/figment-networks/graph-demo/manager/structs"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/tendermint/tendermint/libs/bytes"
)

// TransactionMapper parse raw data from cosmos into transaction format with augmentation from blocks
func TransactionMapper(ctx context.Context, in *tx.Tx, resp *types.TxResponse, blockHash, chainID string) (tx structs.Transaction, err error) {
	tx = structs.Transaction{
		Height:     uint64(resp.Height),
		Hash:       resp.TxHash,
		BlockHash:  blockHash,
		ChainID:    chainID,
		CodeSpace:  resp.Codespace,
		Code:       uint64(resp.Code),
		Result:     resp.Data,
		Logs:       logs(resp.Logs),
		Info:       resp.Info,
		Signatures: signatures(in.Signatures),
		GasWanted:  uint64(resp.GasWanted),
		GasUsed:    uint64(resp.GasUsed),
	}

	if resp.RawLog != "" {
		tx.RawLog = []byte(resp.RawLog)
	} else {
		tx.RawLog = []byte(resp.Logs.String())
	}

	if txBytes := resp.Tx; txBytes != nil {
		txRaw, err := parseMessage(txBytes.Value, txBytes.TypeUrl)
		if err != nil {
			return structs.Transaction{}, err
		}
		tx.TxRaw = structs.TxRaw{
			TxRaw: txRaw,
			Raw: structs.Any{
				TypeURL: txBytes.TypeUrl,
				Value:   txBytes.Value,
			},
		}
	}

	if tx.AuthInfo, err = authInfo(in.AuthInfo); err != nil {
		return structs.Transaction{}, err
	}

	if body := in.Body; body != nil {
		tx.Memo = body.Memo
		tx.ExtensionOptions = extensionOptions(body.ExtensionOptions)
		tx.NonCriticalExtensionOptions = nonCriticalExtensionOptions(body.NonCriticalExtensionOptions)
		if tx.Messages, err = messages(body.Messages); err != nil {
			return structs.Transaction{}, err
		}
	}

	if tx.ExtensionOptions != nil {
		fmt.Println(tx.ExtensionOptions)
	}

	if tx.NonCriticalExtensionOptions != nil {
		fmt.Println(tx.NonCriticalExtensionOptions)
	}

	tx.Time, err = time.Parse(time.RFC3339, resp.Timestamp)
	if err != nil {
		return tx, err
	}

	return tx, nil
}

func authInfo(txAuthInfo *tx.AuthInfo) (authInfo *structs.AuthInfo, err error) {
	if txAuthInfo == nil {
		return nil, nil
	}

	authInfo = &structs.AuthInfo{
		Fee: fee(txAuthInfo.Fee),
	}
	authInfo.SignerInfos, err = signers(txAuthInfo.SignerInfos)

	return authInfo, err
}

func fee(fee *tx.Fee) *structs.Fee {
	if fee == nil {
		return nil
	}

	f := &structs.Fee{
		GasLimit:  fee.GasLimit,
		Sender:    fee.Payer,
		Recipient: fee.Granter,
	}

	if fee.Amount != nil {
		f.Amount = fee.Amount[0].Amount.BigInt()
		f.Currency = fee.Amount[0].Denom
	}

	return f
}

func signers(signerInfos []*tx.SignerInfo) ([]structs.SignerInfo, error) {
	signers := make([]structs.SignerInfo, len(signerInfos))
	for i, sInfo := range signerInfos {
		var pubKey string
		switch sInfo.PublicKey.TypeUrl {
		case "/cosmos.crypto.secp256k1.PubKey":
			pk := secp256k1.PubKey{}
			if err := pk.Unmarshal(sInfo.PublicKey.Value); err != nil {
				return nil, err
			}
			pubKey = bytes.HexBytes(pk.Key).String()
		case "/cosmos.crypto.ed25519.PubKey":
			pk := ed25519.PubKey{}
			if err := pk.Unmarshal(sInfo.PublicKey.Value); err != nil {
				return nil, err
			}
			pubKey = bytes.HexBytes(pk.Key).String()
		}

		signers[i] = structs.SignerInfo{
			PublicKey: &structs.PublicKey{
				Key: pubKey,
				Raw: structs.Any{
					TypeURL: sInfo.PublicKey.TypeUrl,
					Value:   sInfo.PublicKey.Value,
				},
			},
			ModeInfo: sInfo.ModeInfo.String(),
			Sequence: sInfo.Sequence,
		}
	}

	return signers, nil
}

func extensionOptions(extensionOptions []*codectypes.Any) []structs.ExtensionOptions {
	if extensionOptions == nil {
		return nil
	}

	eos := make([]structs.ExtensionOptions, len(extensionOptions))
	for i, eo := range extensionOptions {
		eos[i] = structs.ExtensionOptions{
			ExtensionOption: nil,
			Raw: structs.Any{
				TypeURL: eo.TypeUrl,
				Value:   eo.Value,
			},
		}
	}
	return eos
}

func messages(txMsgs []*codectypes.Any) ([]structs.Message, error) {
	if txMsgs == nil {
		return nil, nil
	}

	msgs := make([]structs.Message, len(txMsgs))
	for i, msg := range txMsgs {
		byteMsg, err := parseMessage(msg.Value, msg.TypeUrl)
		if err != nil {
			return nil, err
		}

		msgs[i] = structs.Message{
			Message: byteMsg,
			Raw: structs.Any{
				TypeURL: msg.TypeUrl,
				Value:   msg.Value,
			},
		}
	}

	return msgs, nil
}

func nonCriticalExtensionOptions(txNceos []*codectypes.Any) []structs.NonCriticalExtensionOptions {
	if txNceos == nil {
		return nil
	}

	nceos := make([]structs.NonCriticalExtensionOptions, len(txNceos))
	for i, nceo := range txNceos {
		nceos[i] = structs.NonCriticalExtensionOptions{
			NonCriticalExtensionOption: nil,
			Raw: structs.Any{
				TypeURL: nceo.TypeUrl,
				Value:   nceo.Value,
			},
		}
	}

	return nceos
}

func logs(respLogs types.ABCIMessageLogs) []structs.Log {
	logs := make([]structs.Log, len(respLogs))
	for i, log := range respLogs {
		events := make([]structs.Event, len(log.Events))
		for j, event := range log.Events {
			attributes := make(map[string]string, len(event.Attributes))
			for _, attr := range event.Attributes {
				attributes[attr.Key] = attr.Value
			}
			events[j] = structs.Event{
				Type:       event.Type,
				Attributes: attributes,
			}
		}
		logs[i] = structs.Log{
			MsgIndex: uint64(log.MsgIndex),
			Log:      log.Log,
			Events:   events,
		}
	}
	return logs
}

func signatures(sigs [][]byte) []string {
	signatures := make([]string, len(sigs))
	for i, sig := range sigs {
		signatures[i] = bytes.HexBytes(sig).String()
	}
	return signatures
}
