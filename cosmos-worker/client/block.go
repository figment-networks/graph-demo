package client

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/figment-networks/graph-demo/cosmos-worker/client/mapper"
	"github.com/figment-networks/graph-demo/manager/structs"

	// cryptoAmino "github.com/cosmos/amino-js/go/lib/tendermint/tendermint/crypto/encoding/amino"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/libs/bytes"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// BlocksMap map of blocks to control block map
// with extra summary of number of transactions
type BlocksMap struct {
	sync.Mutex
	Blocks map[uint64]structs.Block
	NumTxs uint64
}

// BlockErrorPair to wrap error response
type BlockErrorPair struct {
	Height uint64
	Block  structs.Block
	Err    error
}

// GetBlock fetches most recent block from chain
func (c *Client) GetBlock(ctx context.Context, height int64) (blockAndTx structs.BlockAndTx, er error) {
	c.log.Debug("[COSMOS-WORKER] Getting block", zap.Int64("height", height))

	b, err := c.tmServiceClient.GetBlockByHeight(ctx, &tmservice.GetBlockByHeightRequest{Height: height}, grpc.WaitForReady(true))
	if err != nil {
		c.log.Debug("[COSMOS-CLIENT] Error while getting block by height", zap.Int64("height", height), zap.Error(err))
		return structs.BlockAndTx{}, err
	}

	bHash := bytes.HexBytes(b.BlockId.Hash).String()

	blockAndTx.Block = mapper.MapBlockResponseToStructs(b.Block, b.Block.Data, bHash)

	txs := b.Block.Data.GetTxs()
	blockAndTx.Transactions = make([]structs.Transaction, len(txs))

	for i, t := range txs {
		decodedTx, err := c.txDecoder(t)
		if err != nil {
			return structs.BlockAndTx{}, err
		}

		// newTx. = decodedTx.GetMsgs()
		dt := decodedTx.(*wrapper).tx

		fmt.Println(dt)

		body := dt.Body

		events := make([]structs.TransactionEvent, len(body.Messages))
		for index, m := range body.Messages {
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
				return structs.BlockAndTx{}, err
			}

			events[index] = tev
		}

		// body.Memo
		blockAndTx.Transactions[i] = structs.Transaction{
			// Hash:      dt.GetH,
			BlockHash: bHash,
			Height:    uint64(height),
			ChainID:   c.chainID,
			// Epoch: c.,
			Time: b.Block.Header.Time,
			Memo: body.Memo,
		}

	}

	// blockID = structs.BlockID{
	// 	Hash: b.BlockId.Hash,
	// }

	c.log.Debug("[COSMOS-WORKER] Got block", zap.Int64("height", height))
	return blockAndTx, nil
}

func getLogsFromMsgs(ctx context.Context, msgs []*types.Msg) (*types.Result, error) {
	// func (app *BaseApp) runMsgs(ctx sdk.Context, msgs []sdk.Msg, mode runTxMode) (*sdk.Result, error) {
	msgLogs := make(types.ABCIMessageLogs, 0, len(msgs))
	events := types.EmptyEvents()
	txMsgData := &types.TxMsgData{
		Data: make([]*types.MsgData, 0, len(msgs)),
	}

	// NOTE: GasWanted is determined by the AnteHandler and GasUsed by the GasMeter.
	for i, msg := range msgs {
		// skip actual execution for (Re)CheckTx mode
		// if mode == runTxModeCheck || mode == runTxModeReCheck {
		// 	break
		// }

		var (
			msgResult    *types.Result
			eventMsgName string // name to use as value in event `message.action`
			err          error
		)

		if handler := app.msgServiceRouter.Handler(msg); handler != nil {
			// ADR 031 request type routing
			msgResult, err = handler(ctx, msg)
			eventMsgName = types.MsgTypeURL(msg)
		} else if legacyMsg, ok := msg.(legacytx.LegacyMsg); ok {
			// legacy sdk.Msg routing
			// Assuming that the app developer has migrated all their Msgs to
			// proto messages and has registered all `Msg services`, then this
			// path should never be called, because all those Msgs should be
			// registered within the `msgServiceRouter` already.
			msgRoute := legacyMsg.Route()
			eventMsgName = legacyMsg.Type()
			handler := app.router.Route(ctx, msgRoute)
			if handler == nil {
				return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized message route: %s; message index: %d", msgRoute, i)
			}

			msgResult, err = handler(ctx, msg)
		} else {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "can't route message %+v", msg)
		}

		if err != nil {
			return nil, sdkerrors.Wrapf(err, "failed to execute message; message index: %d", i)
		}

		msgEvents := types.Events{
			types.NewEvent(types.EventTypeMessage, types.NewAttribute(types.AttributeKeyAction, eventMsgName)),
		}
		msgEvents = msgEvents.AppendEvents(msgResult.GetEvents())

		// append message events, data and logs
		//
		// Note: Each message result's data must be length-prefixed in order to
		// separate each result.
		events = events.AppendEvents(msgEvents)

		txMsgData.Data = append(txMsgData.Data, &types.MsgData{MsgType: types.MsgTypeURL(msg), Data: msgResult.Data})
		msgLogs = append(msgLogs, types.NewABCIMessageLog(uint32(i), msgResult.Log, msgEvents))
	}

	data, err := proto.Marshal(txMsgData)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to marshal tx data")
	}

	return &types.Result{
		Data:   data,
		Log:    strings.TrimSpace(msgLogs.String()),
		Events: events.ToABCIEvents(),
	}, nil
}

// type Handler func(ctx Context, msg Msg) (*Result, error)

type handlerFunc func(ctx types.Context, req types.Msg) (*types.Result, error)

func (c *Client) handler(methodHandler) handlerFunc {
	return func(ctx types.Context, req types.Msg) (*types.Result, error) {
		ctx = ctx.WithEventManager(types.NewEventManager())
		interceptor := func(goCtx context.Context, _ interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			goCtx = context.WithValue(goCtx, types.SdkContextKey, ctx)
			return handler(goCtx, req)
		}
		// Call the method handler from the service description with the handler object.
		// We don't do any decoding here because the decoding was already done.
		res, err := methodHandler(handler, types.WrapSDKContext(ctx), noopDecoder, interceptor)
		if err != nil {
			return nil, err
		}

		resMsg, ok := res.(proto.Message)
		if !ok {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "Expecting proto.Message, got %T", resMsg)
		}

		return sdk.WrapServiceResult(ctx, resMsg, err)
	}
}

// func decodeTx(txBytes []byte) (types.Tx, error) {
// 	if len(txBytes) == 0 {
// 		return nil, errors.New("tx bytes are empty")
// 	}

// base64.NewEncoding(string(txBytes))
// hexBytes := []byte{}
// _, err := base64.StdEncoding.Decode(hexBytes, txBytes)
// if err != nil {
// 	return nil, err
// }

// var tx tx.Tx
// err := codec.UnmarshalBinaryBare(txBytes, &tx)

// if err != nil {
// 	return nil, err
// }

// fmt.Println(tx)

// bz2, err := codec.MarshalJSON(o)
// if err != nil {
// 	return nil, err
// }

// return nil, nil

// var stdTx = legacytx.StdTx{}
// // UnmarshalBinaryBare

// legacyAmino := codec.NewLegacyAmino()
// cCodec.RegisterCrypto(legacyAmino)

// legacyAmino.RegisterInterface((*types.Msg)(nil), nil)
// legacyAmino.RegisterInterface((*types.Tx)(nil), nil)
// legacyAmino.RegisterInterface((*types.Signature)(nil), nil)
// legacyAmino.RegisterInterface((*types.Fee)(nil), nil)
// legacyAmino.RegisterInterface((*types.TxWithTimeoutHeight)(nil), nil)
// legacyAmino.RegisterInterface((*types.TxWithMemo)(nil), nil)

// legacyAmino.

// legacyAmino.R
// cCodec.RegisterInterfaces()

// legacyAmino.RegisterConcrete(&ed25519.PubKey{}, ed25519.PubKeyName, nil)
// legacyAmino.RegisterConcrete(&ed25519.PrivKey{}, ed25519.PrivKeyName, nil)
// legacyAmino.RegisterInterface((*types.Sig)(nil), nil)
// legacyAmino.RegisterInterface((*types.Msg)(nil), "cosmos.base.v1beta1.Msg")
// legacyAmino.RegisterInterface((*tmtypes.Evidence)(nil), nil)
// legacyAmino.RegisterConcrete(&tmtypes.DuplicateVoteEvidence{}, "tendermint/DuplicateVoteEvidence", nil)
// err := legacyAmino.UnmarshalBinaryBare(txBytes, &stdTx)
// amino.Un
//
// legacytx.Unmarshal(amino.U, &tx)

// err := amino.UnmarshalBinaryBare(txBytes, &tx)
// if err != nil {
// 	return nil, err
// }
// return stdTx, nil
// }

// type wrapper struct {
// 	tx                           *types.Tx
// 	bodyBz                       []byte
// 	authInfoBz                   []byte
// 	txBodyHasUnknownNonCriticals bool
// }

// func TxDecoder(txBytes []byte) (types.Tx, error) {

// }

// GetBlock fetches most recent block from chain
func (c *Client) GetLatest(ctx context.Context) (block structs.BlockAndTx, er error) {

	nctx, cancel := context.WithTimeout(ctx, c.cfg.TimeoutBlockCall)
	defer cancel()
	b, err := c.tmServiceClient.GetLatestBlock(nctx, &tmservice.GetLatestBlockRequest{}, grpc.WaitForReady(true))
	if err != nil {
		c.log.Debug("[COSMOS-CLIENT] Error getting latest block", zap.Error(err), zap.Int("txs", len(b.Block.Data.Txs)))
		return block, err
	}

	bHash := bytes.HexBytes(b.BlockId.Hash).String()

	c.log.Debug("[COSMOS-CLIENT] Got latest block", zap.Uint64("height", uint64(b.Block.Header.Height)), zap.Error(err))
	return structs.BlockAndTx{
		Block: mapper.MapBlockResponseToStructs(b.Block, b.Block.Data, bHash),
	}, nil

}
