package client

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/figment-networks/graph-demo/cosmos-worker/client/mapper"
	"github.com/figment-networks/graph-demo/manager/structs"

	// cryptoAmino "github.com/cosmos/amino-js/go/lib/tendermint/tendermint/crypto/encoding/amino"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"github.com/gogo/protobuf/proto"
	amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
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
func (c *Client) GetBlock(ctx context.Context, height uint64) (blockAndTx structs.BlockAndTx, er error) {
	c.log.Debug("[COSMOS-WORKER] Getting block", zap.Uint64("height", height))

	b, err := c.tmServiceClient.GetBlockByHeight(ctx, &tmservice.GetBlockByHeightRequest{Height: int64(height)}, grpc.WaitForReady(true))
	if err != nil {
		c.log.Debug("[COSMOS-CLIENT] Error while getting block by height", zap.Uint64("height", height), zap.Error(err), zap.Int("txs", len(b.Block.Data.Txs)))
		return structs.BlockAndTx{}, err
	}

	bHash := bytes.HexBytes(b.BlockId.Hash).String()

	blockAndTx.Block = mapper.MapBlockResponseToStructs(b.Block, b.Block.Data, bHash)

	blockAndTx.Transactions = make([]structs.Transaction, 0)

	for _, tx := range b.Block.Data.GetTxs() {

		// tx

		decodedTx, err := decodeTx(tx)
		if err != nil {
			return structs.BlockAndTx{}, err
		}

		fmt.Println(decodedTx)
		// c.rawToTransaction(ctx, decodedTx, nil)

	}

	// blockID = structs.BlockID{
	// 	Hash: b.BlockId.Hash,
	// }

	c.log.Debug("[COSMOS-WORKER] Got block", zap.Int64("height", height))
	return blockAndTx, nil
}

func (c *Client) getContextForTx(txBytes []byte, header tmproto.Header) (ctx types.Context, traceCtx types.TraceContext) {
	ms := new(types.CacheMultiStore)
	fmt.Println("TracingEnabled ", (*ms).TracingEnabled())
	logger := log.NewNopLogger()
	ctx = types.NewContext(*ms, header, false, logger)

	traceCtx = types.TraceContext(
		map[string]interface{}{"blockHeight": header.Height},
	)

	return ctx, traceCtx
}

func (c *Client) getLogsFromMsg(ctx types.Context, msg types.Msg, i int) (*types.Result, error) {
	events := types.EmptyEvents()
	// NOTE: GasWanted is determined by the AnteHandler and GasUsed by the GasMeter.

	var (
		msgResult    *types.Result
		eventMsgName string // name to use as value in event `message.action`
		err          error
	)

	// grpcServer := grpc.NewServer()
	// tmservice.RegisterServiceServer(grpcServer, c.tmServiceServer)

	// MsgServer
	// // c.tmServiceServer.
	// h := c.handler(mh, h)

	// msgResult, err = h(ctx, *msg)

	if svcMsg, ok := (*msg).(sdk.ServiceMsg); ok {
		msgFqName = svcMsg.MethodName
		handler := app.msgServiceRouter.Handler(msgFqName)
		if handler == nil {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized message service method: %s; message index: %d", msgFqName, i)
		}
		msgResult, err = handler(ctx, svcMsg.Request)
	} else {
		// legacy sdk.Msg routing
		msgRoute := msg.Route()
		msgFqName = msg.Type()
		handler := app.router.Route(ctx, msgRoute)
		if handler == nil {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized message route: %s; message index: %d", msgRoute, i)
		}

		msgResult, err = handler(ctx, msg)
	}

	if err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to execute message; message index: %d", i)
	}

	msgEvents := types.Events{
		types.NewEvent(types.EventTypeMessage, types.NewAttribute(types.AttributeKeyAction, eventMsgName)),
	}
	msgEvents = msgEvents.AppendEvents(msgResult.GetEvents())
	events = events.AppendEvents(msgEvents)
	txMsgData := &types.TxMsgData{
		Data: []*types.MsgData{&types.MsgData{MsgType: msg.Type(), Data: msgResult.Data}},
	}
	msgLog := types.NewABCIMessageLog(uint32(i), msgResult.Log, msgEvents)

	data, err := proto.Marshal(txMsgData)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to marshal tx data")
	}

	return &types.Result{
		Data:   data,
		Log:    msgLog.String(),
		Events: events.ToABCIEvents(),
	}, nil
}

// type Handler func(ctx Context, msg Msg) (*Result, error)

type handlerFunc func(ctx types.Context, req types.Msg) (*types.Result, error)

type methodHandler func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error)

func (c *Client) handler(methodHandler methodHandler, handler interface{}) handlerFunc {
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

		return types.WrapServiceResult(ctx, resMsg, err)
	}
}

func noopDecoder(_ interface{}) error { return nil }

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
