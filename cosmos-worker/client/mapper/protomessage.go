package mapper

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	distribution "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidence "github.com/cosmos/cosmos-sdk/x/evidence/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	transfer "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	client "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	slashing "github.com/cosmos/cosmos-sdk/x/slashing/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gogo/protobuf/proto"
)

func parseKey(value []byte, typeURL string) (result json.RawMessage, err error) {
	switch typeURL {
	case "/cosmos.crypto.ed25519.PrivKey":

	case "/cosmos.crypto.ed25519.PubKey":

	case "/cosmos.crypto.multisig.LegacyAminoPubKey":

	case "/cosmos.crypto.multisig.v1beta1/CompactBitArray":

	case "/cosmos.crypto.multisig.v1beta1/MultiSignature":

	case "/cosmos.crypto.secp256k1.PrivKey":

	case "/cosmos.crypto.secp256k1.PubKey":

	default:
		return nil, fmt.Errorf("unknown proto key type %q", typeURL)
	}

	return result, nil
}

func parseMessage(value []byte, typeURL string) (result json.RawMessage, err error) {
	switch typeURL {
	case "/cosmos.bank.v1beta1.MsgMultiSend":
		return protoUnmarshal(value, &bank.MsgMultiSend{})

	case "/cosmos.bank.v1beta1.MsgMultiSendResponse":
		return protoUnmarshal(value, &bank.MsgMultiSendResponse{})

	case "/cosmos.bank.v1beta1.MsgSend":
		return protoUnmarshal(value, &bank.MsgSend{})

	case "/cosmos.bank.v1beta1.MsgSendResponse":
		return protoUnmarshal(value, &bank.MsgSendResponse{})

	case "/cosmos.distribution.v1beta1.MsgFundCommunityPool":
		return protoUnmarshal(value, &distribution.MsgFundCommunityPool{})

	case "/cosmos.distribution.v1beta1.MsgFundCommunityPoolResponse":
		return protoUnmarshal(value, &distribution.MsgFundCommunityPoolResponse{})

	case "/cosmos.distribution.v1beta1.MsgSetWithdrawAddress":
		return protoUnmarshal(value, &distribution.MsgSetWithdrawAddress{})

	case "/cosmos.distribution.v1beta1.MsgSetWithdrawAddressResponse":
		return protoUnmarshal(value, &distribution.MsgSetWithdrawAddressResponse{})

	case "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward":
		return protoUnmarshal(value, &distribution.MsgWithdrawDelegatorReward{})

	case "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorRewardResponse":
		return protoUnmarshal(value, &distribution.MsgWithdrawDelegatorRewardResponse{})

	case "/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission":
		return protoUnmarshal(value, &distribution.MsgWithdrawValidatorCommission{})

	case "/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommissionResponse":
		return protoUnmarshal(value, &distribution.MsgWithdrawValidatorCommissionResponse{})

	// checkEvidence type any
	case "/cosmos.evidence.v1beta1.MsgSubmitEvidence":
		return protoUnmarshal(value, &evidence.MsgSubmitEvidence{})

	case "/cosmos.evidence.v1beta1.MsgSubmitEvidenceResponse":
		return protoUnmarshal(value, &evidence.MsgSubmitEvidenceResponse{})

	case "/cosmos.gov.v1beta1.MsgDeposit":
		return protoUnmarshal(value, &gov.MsgDeposit{})

	case "/cosmos.gov.v1beta1.MsgDepositResponse":
		return protoUnmarshal(value, &gov.MsgDepositResponse{})

	case "/cosmos.gov.v1beta1.MsgSubmitProposal":
		return protoUnmarshal(value, &gov.MsgSubmitProposal{})

	case "/cosmos.gov.v1beta1.MsgSubmitProposalResponse":
		return protoUnmarshal(value, &gov.MsgSubmitProposalResponse{})

	case "/cosmos.gov.v1beta1.MsgVote":
		return protoUnmarshal(value, &gov.MsgVote{})

	case "/cosmos.gov.v1beta1.MsgVoteResponse":
		return protoUnmarshal(value, &gov.MsgVoteResponse{})

	case "/cosmos.slashing.v1beta1.MsgUnjail":
		return protoUnmarshal(value, &slashing.MsgUnjail{})

	case "/cosmos.slashing.v1beta1.MsgUnjailResponse":
		return protoUnmarshal(value, &slashing.MsgUnjail{})

	case "/cosmos.staking.v1beta1.MsgBeginRedelegate":
		return protoUnmarshal(value, &staking.MsgBeginRedelegate{})

	case "/cosmos.staking.v1beta1.MsgBeginRedelegateResponse":
		return protoUnmarshal(value, &staking.MsgBeginRedelegateResponse{})

	case "/cosmos.staking.v1beta1.MsgCreateValidator":
		return protoUnmarshal(value, &staking.MsgCreateValidator{})

	case "/cosmos.staking.v1beta1.MsgCreateValidatorResponse":
		return protoUnmarshal(value, &staking.MsgCreateValidatorResponse{})

	case "/cosmos.staking.v1beta1.MsgDelegate":
		return protoUnmarshal(value, &staking.MsgDelegate{})

	case "/cosmos.staking.v1beta1.MsgDelegateResponse":
		return protoUnmarshal(value, &staking.MsgDelegateResponse{})

	case "/cosmos.staking.v1beta1.MsgEditValidator":
		return protoUnmarshal(value, &staking.MsgEditValidator{})

	case "/cosmos.staking.v1beta1.MsgEditValidatorResponse":
		return protoUnmarshal(value, &staking.MsgEditValidatorResponse{})

	case "/cosmos.staking.v1beta1.MsgUndelegate":
		return protoUnmarshal(value, &staking.MsgUndelegate{})

	case "/cosmos.staking.v1beta1.MsgUndelegateResponse":
		return protoUnmarshal(value, &staking.MsgUndelegateResponse{})

	case "/cosmos.tx.v1beta1.AuthInfo":
		return protoUnmarshal(value, &tx.AuthInfo{})

	case "/cosmos.tx.v1beta1.Fee":
		return protoUnmarshal(value, &tx.Fee{})

	case "/cosmos.tx.v1beta1.ModeInfo":
		return protoUnmarshal(value, &tx.ModeInfo{})

	case "/cosmos.tx.v1beta1.ModeInfo.Multi":
		return protoUnmarshal(value, &tx.ModeInfo_Multi{})

	case "/cosmos.tx.v1beta1.ModeInfo.Single":
		return protoUnmarshal(value, &tx.ModeInfo_Single{})

	case "/cosmos.tx.v1beta1.SignDoc":
		return protoUnmarshal(value, &tx.SignDoc{})

	case "/cosmos.tx.v1beta1.SignerInfo":
		return protoUnmarshal(value, &tx.SignerInfo{})

	case "/cosmos.tx.v1beta1.Tx":
		return protoUnmarshal(value, &tx.Tx{})

	case "/cosmos.tx.v1beta1.TxBody":
		return protoUnmarshal(value, &tx.TxBody{})

	case "/cosmos.tx.v1beta1.TxRaw":
		return protoUnmarshal(value, &tx.TxRaw{})

	case "/cosmos.vesting.v1beta1.MsgCreateVestingAccount":
		return protoUnmarshal(value, &vesting.MsgCreateVestingAccount{})

	case "/cosmos.vesting.v1beta1.MsgCreateVestingAccountResponse":
		return protoUnmarshal(value, &vesting.MsgCreateVestingAccountResponse{})

	case "/ibc.applications.transfer.v1.MsgTransfer":
		return protoUnmarshal(value, &transfer.MsgTransfer{})

	case "/ibc.applications.transfer.v1.MsgTransferResponse":
		return protoUnmarshal(value, &transfer.MsgTransferResponse{})

	case "/ibc.core.channel.v1.MsgAcknowledgement":
		return protoUnmarshal(value, &channel.MsgAcknowledgement{})

	case "/ibc.core.channel.v1.MsgAcknowledgementResponse":
		return protoUnmarshal(value, &channel.MsgAcknowledgementResponse{})

	case "/ibc.core.channel.v1.MsgChannelCloseConfirm":
		return protoUnmarshal(value, &channel.MsgChannelCloseConfirm{})

	case "/ibc.core.channel.v1.MsgChannelCloseConfirmResponse":
		return protoUnmarshal(value, &channel.MsgChannelCloseConfirmResponse{})

	case "/ibc.core.channel.v1.MsgChannelCloseInit":
		return protoUnmarshal(value, &channel.MsgChannelCloseInit{})

	case "/ibc.core.channel.v1.MsgChannelCloseInitResponse":
		return protoUnmarshal(value, &channel.MsgChannelCloseInitResponse{})

	case "/ibc.core.channel.v1.MsgChannelOpenAck":
		return protoUnmarshal(value, &channel.MsgChannelOpenAck{})

	case "/ibc.core.channel.v1.MsgChannelOpenAckResponse":
		return protoUnmarshal(value, &channel.MsgChannelOpenAckResponse{})

	case "/ibc.core.channel.v1.MsgChannelOpenConfirm":
		return protoUnmarshal(value, &channel.MsgChannelOpenConfirm{})

	case "/ibc.core.channel.v1.MsgChannelOpenConfirmResponse":
		return protoUnmarshal(value, &channel.MsgChannelOpenConfirmResponse{})

	case "/ibc.core.channel.v1.MsgChannelOpenInit":
		return protoUnmarshal(value, &channel.MsgChannelOpenInit{})

	case "/ibc.core.channel.v1.MsgChannelOpenInitResponse":
		return protoUnmarshal(value, &channel.MsgChannelOpenInitResponse{})

	case "/ibc.core.channel.v1.MsgChannelOpenTry":
		return protoUnmarshal(value, &channel.MsgChannelOpenTry{})

	case "/ibc.core.channel.v1.MsgChannelOpenTryResponse":
		return protoUnmarshal(value, &channel.MsgChannelOpenTryResponse{})

	case "/ibc.core.channel.v1.MsgRecvPacket":
		return protoUnmarshal(value, &channel.MsgRecvPacket{})

	case "/ibc.core.channel.v1.MsgRecvPacketResponse":
		return protoUnmarshal(value, &channel.MsgRecvPacketResponse{})

	case "/ibc.core.channel.v1.MsgTimeout":
		return protoUnmarshal(value, &channel.MsgTimeout{})

	case "/ibc.core.channel.v1.MsgTimeoutOnClose":
		return protoUnmarshal(value, &channel.MsgTimeoutOnClose{})

	case "/ibc.core.channel.v1.MsgTimeoutOnCloseResponse":
		return protoUnmarshal(value, &channel.MsgTimeoutOnCloseResponse{})

	case "/ibc.core.channel.v1.MsgTimeoutResponse":
		return protoUnmarshal(value, &channel.MsgTimeoutResponse{})

	case "/ibc.core.client.v1.MsgCreateClient":
		return protoUnmarshal(value, &client.MsgCreateClient{})

	case "/ibc.core.client.v1.MsgCreateClientResponse":
		return protoUnmarshal(value, &client.MsgCreateClientResponse{})

	case "/ibc.core.client.v1.MsgSubmitMisbehaviour":
		return protoUnmarshal(value, &client.MsgSubmitMisbehaviour{})

	case "/ibc.core.client.v1.MsgSubmitMisbehaviourResponse":
		return protoUnmarshal(value, &client.MsgSubmitMisbehaviourResponse{})

	case "/ibc.core.client.v1.MsgUpdateClient":
		return protoUnmarshal(value, &client.MsgUpdateClient{})

	case "/ibc.core.client.v1.MsgUpdateClientResponse":
		return protoUnmarshal(value, &client.MsgUpdateClientResponse{})

	case "/ibc.core.client.v1.MsgUpgradeClient":
		return protoUnmarshal(value, &client.MsgUpgradeClient{})

	case "/ibc.core.client.v1.MsgUpgradeClientResponse":
		return protoUnmarshal(value, &client.MsgUpgradeClientResponse{})

	case "/ibc.core.connection.v1.MsgConnectionOpenAck":
		return protoUnmarshal(value, &connection.MsgConnectionOpenAck{})

	case "/ibc.core.connection.v1.MsgConnectionOpenAckResponse":
		return protoUnmarshal(value, &connection.MsgConnectionOpenAckResponse{})

	case "/ibc.core.connection.v1.MsgConnectionOpenConfirm":
		return protoUnmarshal(value, &connection.MsgConnectionOpenConfirm{})

	case "/ibc.core.connection.v1.MsgConnectionOpenConfirmResponse":
		return protoUnmarshal(value, &connection.MsgConnectionOpenConfirmResponse{})

	case "/ibc.core.connection.v1.MsgConnectionOpenInit":
		return protoUnmarshal(value, &connection.MsgConnectionOpenInit{})

	case "/ibc.core.connection.v1.MsgConnectionOpenInitResponse":
		return protoUnmarshal(value, &connection.MsgConnectionOpenInitResponse{})

	case "/ibc.core.connection.v1.MsgConnectionOpenTry":
		return protoUnmarshal(value, &connection.MsgConnectionOpenTry{})

	case "/ibc.core.connection.v1.MsgConnectionOpenTryResponse":
		return protoUnmarshal(value, &connection.MsgConnectionOpenTryResponse{})

	case "/ibc.lightclients.tendermint.v1.ClientState":
		return protoUnmarshal(value, &tendermint.ClientState{})

	case "/ibc.lightclients.tendermint.v1.ConsensusState":
		return protoUnmarshal(value, &tendermint.ConsensusState{})

	case "/ibc.lightclients.tendermint.v1.Fraction":
		return protoUnmarshal(value, &tendermint.Fraction{})

	case "/ibc.lightclients.tendermint.v1.Header":
		return protoUnmarshal(value, &tendermint.Header{})

	case "/ibc.lightclients.tendermint.v1.Misbehaviour":
		return protoUnmarshal(value, &tendermint.Misbehaviour{})

	default:
		return nil, fmt.Errorf("unknown proto type url %q", typeURL)
	}
}

func protoUnmarshal(value []byte, pb proto.Message) (raw json.RawMessage, err error) {
	if err := proto.Unmarshal(value, pb); err != nil {
		return nil, err
	}

	v := reflect.Indirect(reflect.ValueOf(pb))
	result := make(map[string]json.RawMessage)

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i).Interface()
		fieldName := v.Type().Field(i).Name
		switch reflect.TypeOf(fieldValue) {
		case reflect.TypeOf(&types.Any{}):
			any := fieldValue.(*types.Any)
			if result[fieldName], err = parseMessage(any.Value, any.TypeUrl); err != nil {
				return nil, err
			}
		default:
			if result[fieldName], err = json.Marshal(fieldValue); err != nil {
				return nil, err
			}
		}
	}

	return json.Marshal(result)
}
