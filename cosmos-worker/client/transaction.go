package client

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/figment-networks/graph-demo/cosmos-worker/client/mapper"
	"github.com/figment-networks/graph-demo/manager/structs"

	codec_types "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	errUnknownMessageType = fmt.Errorf("unknown message type")
)

var curencyRegex = regexp.MustCompile("([0-9\\.\\,\\-\\s]+)([^0-9\\s]+)$")

// SearchTx is making search api call
func (c *Client) SearchTx(ctx context.Context, block structs.Block) (txs []structs.Transaction, err error) {
	height := block.Height
	c.log.Debug("[COSMOS-WORKER] Getting transactions", zap.Uint64("height", height))

	pag := &query.PageRequest{
		CountTotal: true,
		Limit:      perPage,
	}

	var page = uint64(1)
	for {
		pag.Offset = (perPage * page) - perPage
		now := time.Now()

		nctx, cancel := context.WithTimeout(ctx, c.cfg.TimeoutSearchTxCall)
		grpcRes, err := c.txServiceClient.GetTxsEvent(nctx, &tx.GetTxsEventRequest{
			Events:     []string{"tx.height=" + strconv.FormatUint(height, 10)},
			Pagination: pag,
		}, grpc.WaitForReady(true))
		cancel()

		c.log.Debug("[COSMOS-API] Request Time (/tx_search)", zap.Duration("duration", time.Now().Sub(now)))
		if err != nil {
			return nil, err
		}

		for i, trans := range grpcRes.Txs {
			resp := grpcRes.TxResponses[i]
			tx, err := c.rawToTransaction(ctx, trans, resp)
			if err != nil {
				return nil, err
			}
			tx.BlockHash = block.Hash
			tx.ChainID = block.ChainID
			tx.Time = block.Time
			txs = append(txs, tx)
		}

		if grpcRes.Pagination.GetTotal() <= uint64(len(txs)) {
			break
		}

		page++

	}

	c.log.Debug("[COSMOS-WORKER] Got transactions", zap.Uint64("height", height))
	return txs, nil
}

// transform raw data from cosmos into transaction format with augmentation from blocks
func (c *Client) rawToTransaction(ctx context.Context, in *tx.Tx, resp *types.TxResponse) (tx structs.Transaction, err error) {

	tx = structs.Transaction{
		Height:    uint64(resp.Height),
		Hash:      resp.TxHash,
		GasWanted: uint64(resp.GasWanted),
		GasUsed:   uint64(resp.GasUsed),
	}

	if resp.RawLog != "" {
		tx.RawLog = []byte(resp.RawLog)
	} else {
		tx.RawLog = []byte(resp.Logs.String())
	}

	tx.Raw, err = in.Marshal()
	if err != nil {
		return structs.Transaction{}, errors.New("error marshaling tx to raw")
	}

	if in.Body != nil {
		tx.Memo = in.Body.Memo

		for index, m := range in.Body.Messages {
			tev := structs.TransactionEvent{
				ID: strconv.Itoa(index),
			}
			lg := findLog(resp.Logs, index)

			// tPath is "/cosmos.bank.v1beta1.MsgSend" or "/ibc.core.client.v1.MsgCreateClient"
			tPath := strings.Split(m.TypeUrl, ".")

			var err error
			var msgType string
			if len(tPath) == 5 && tPath[0] == "/ibc" {
				msgType = tPath[4]
				err = addIBCSubEvent(tPath[2], msgType, &tev, m, lg)
			} else if len(tPath) == 4 && tPath[0] == "/cosmos" {
				msgType = tPath[3]
				err = addSubEvent(tPath[1], msgType, &tev, m, lg)
			} else {
				err = fmt.Errorf("TypeURL is in wrong format: %v", m.TypeUrl)
			}

			if err != nil {
				c.log.Error("[COSMOS-API] Problem decoding transaction ", zap.Error(err), zap.String("type", msgType), zap.String("route", m.TypeUrl), zap.Int64("height", resp.Height))
				return structs.Transaction{}, err
			}

			tx.Events = append(tx.Events, tev)
		}
	}

	if in.AuthInfo != nil {
		for _, coin := range in.AuthInfo.Fee.Amount {
			tx.Fee = append(tx.Fee, structs.TransactionAmount{
				Text:     coin.Amount.String(),
				Numeric:  coin.Amount.BigInt(),
				Currency: coin.Denom,
			})
		}
	}

	if resp.Code > 0 {
		tx.Events = append(tx.Events, structs.TransactionEvent{
			Kind: "error",
			Sub: []structs.SubsetEvent{{
				Type:   []string{"error"},
				Module: resp.Codespace,
				Error: &structs.SubsetEventError{
					Message: resp.RawLog,
				},
			}},
		})
	}

	return tx, nil
}

func findLog(logs types.ABCIMessageLogs, index int) types.ABCIMessageLog {
	if len(logs) <= index {
		return types.ABCIMessageLog{}
	}
	if lg := logs[index]; lg.GetMsgIndex() == uint32(index) {
		return lg
	}
	for _, lg := range logs {
		if lg.GetMsgIndex() == uint32(index) {
			return lg
		}
	}
	return types.ABCIMessageLog{}
}

func addSubEvent(msgRoute, msgType string, tev *structs.TransactionEvent, m *codec_types.Any, lg types.ABCIMessageLog) (err error) {
	var ev structs.SubsetEvent
	switch msgRoute {
	case "bank":
		switch msgType {
		case "MsgSend":
			ev, err = mapper.BankSendToSub(m.Value, lg)
		case "MsgMultiSend":
			ev, err = mapper.BankMultisendToSub(m.Value, lg)
		default:
			err = fmt.Errorf("problem with %s - %s: %w", msgRoute, msgType, errUnknownMessageType)
		}
	case "crisis":
		switch msgType {
		case "MsgVerifyInvariant":
			ev, err = mapper.CrisisVerifyInvariantToSub(m.Value)
		default:
			err = fmt.Errorf("problem with %s - %s: %w", msgRoute, msgType, errUnknownMessageType)
		}
	case "distribution":
		switch msgType {
		case "MsgWithdrawValidatorCommission":
			ev, err = mapper.DistributionWithdrawValidatorCommissionToSub(m.Value, lg)
		case "MsgSetWithdrawAddress":
			ev, err = mapper.DistributionSetWithdrawAddressToSub(m.Value)
		case "MsgWithdrawDelegatorReward":
			ev, err = mapper.DistributionWithdrawDelegatorRewardToSub(m.Value, lg)
		case "MsgFundCommunityPool":
			ev, err = mapper.DistributionFundCommunityPoolToSub(m.Value)
		default:
			err = fmt.Errorf("problem with %s - %s: %w", msgRoute, msgType, errUnknownMessageType)
		}
	case "evidence":
		switch msgType {
		case "MsgSubmitEvidence":
			ev, err = mapper.EvidenceSubmitEvidenceToSub(m.Value)
		default:
			err = fmt.Errorf("problem with %s - %s: %w", msgRoute, msgType, errUnknownMessageType)
		}
	case "gov":
		switch msgType {
		case "MsgDeposit":
			ev, err = mapper.GovDepositToSub(m.Value, lg)
		case "MsgVote":
			ev, err = mapper.GovVoteToSub(m.Value)
		case "MsgSubmitProposal":
			ev, err = mapper.GovSubmitProposalToSub(m.Value, lg)
		default:
			err = fmt.Errorf("problem with %s - %s: %w", msgRoute, msgType, errUnknownMessageType)
		}
	case "slashing":
		switch msgType {
		case "MsgUnjail":
			ev, err = mapper.SlashingUnjailToSub(m.Value)
		default:
			err = fmt.Errorf("problem with %s - %s: %w", msgRoute, msgType, errUnknownMessageType)
		}
	case "vesting":
		switch msgType {
		case "MsgCreateVestingAccount":
			ev, err = mapper.VestingMsgCreateVestingAccountToSub(m.Value, lg)
		default:
			err = fmt.Errorf("problem with %s - %s: %w", msgRoute, msgType, errUnknownMessageType)
		}
	case "staking":
		switch msgType {
		case "MsgUndelegate":
			ev, err = mapper.StakingUndelegateToSub(m.Value, lg)
		case "MsgEditValidator":
			ev, err = mapper.StakingEditValidatorToSub(m.Value)
		case "MsgCreateValidator":
			ev, err = mapper.StakingCreateValidatorToSub(m.Value)
		case "MsgDelegate":
			ev, err = mapper.StakingDelegateToSub(m.Value, lg)
		case "MsgBeginRedelegate":
			ev, err = mapper.StakingBeginRedelegateToSub(m.Value, lg)
		default:
			err = fmt.Errorf("problem with %s - %s: %w", msgRoute, msgType, errUnknownMessageType)
		}
	default:
		err = fmt.Errorf("problem with %s - %s: %w", msgRoute, msgType, errUnknownMessageType)
	}

	if len(ev.Type) > 0 {
		tev.Sub = append(tev.Sub, ev)
		tev.Kind = ev.Type[0]
	}
	return err
}

func addIBCSubEvent(msgRoute, msgType string, tev *structs.TransactionEvent, m *codec_types.Any, lg types.ABCIMessageLog) (err error) {
	var ev structs.SubsetEvent

	switch msgRoute {
	case "client":
		switch msgType {
		case "MsgCreateClient":
			ev, err = mapper.IBCCreateClientToSub(m.Value)
		case "MsgUpdateClient":
			ev, err = mapper.IBCUpdateClientToSub(m.Value)
		case "MsgUpgradeClient":
			ev, err = mapper.IBCUpgradeClientToSub(m.Value)
		case "MsgSubmitMisbehaviour":
			ev, err = mapper.IBCSubmitMisbehaviourToSub(m.Value)
		default:
			err = fmt.Errorf("problem with %s - %s: %w", msgRoute, msgType, errUnknownMessageType)
		}
	case "connection":
		switch msgType {
		case "MsgConnectionOpenInit":
			ev, err = mapper.IBCConnectionOpenInitToSub(m.Value)
		case "MsgConnectionOpenConfirm":
			ev, err = mapper.IBCConnectionOpenConfirmToSub(m.Value)
		case "MsgConnectionOpenAck":
			ev, err = mapper.IBCConnectionOpenAckToSub(m.Value)
		case "MsgConnectionOpenTry":
			ev, err = mapper.IBCConnectionOpenTryToSub(m.Value)
		default:
			err = fmt.Errorf("problem with %s - %s:  %w", msgRoute, msgType, errUnknownMessageType)
		}
	case "channel":
		switch msgType {
		case "MsgChannelOpenInit":
			ev, err = mapper.IBCChannelOpenInitToSub(m.Value)
		case "MsgChannelOpenTry":
			ev, err = mapper.IBCChannelOpenTryToSub(m.Value)
		case "MsgChannelOpenConfirm":
			ev, err = mapper.IBCChannelOpenConfirmToSub(m.Value)
		case "MsgChannelOpenAck":
			ev, err = mapper.IBCChannelOpenAckToSub(m.Value)
		case "MsgChannelCloseInit":
			ev, err = mapper.IBCChannelCloseInitToSub(m.Value)
		case "MsgChannelCloseConfirm":
			ev, err = mapper.IBCChannelCloseConfirmToSub(m.Value)
		case "MsgRecvPacket":
			ev, err = mapper.IBCChannelRecvPacketToSub(m.Value)
		case "MsgTimeout":
			ev, err = mapper.IBCChannelTimeoutToSub(m.Value)
		case "MsgAcknowledgement":
			ev, err = mapper.IBCChannelAcknowledgementToSub(m.Value)

		default:
			err = fmt.Errorf("problem with %s - %s:  %w", msgRoute, msgType, errUnknownMessageType)
		}
	case "transfer":
		switch msgType {
		case "MsgTransfer":
			ev, err = mapper.IBCTransferToSub(m.Value)
		default:
			err = fmt.Errorf("problem with %s - %s:  %w", msgRoute, msgType, errUnknownMessageType)
		}
	default:
		err = fmt.Errorf("problem with %s - %s:  %w", msgRoute, msgType, errUnknownMessageType)
	}

	if len(ev.Type) > 0 {
		tev.Sub = append(tev.Sub, ev)
		tev.Kind = ev.Type[0]
	}

	return err
}
