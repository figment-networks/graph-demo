package ws

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/figment-networks/graph-demo/connectivity"
	wsapi "github.com/figment-networks/graph-demo/connectivity/ws"
	"github.com/figment-networks/graph-demo/manager/structs"
	"go.uber.org/zap"

	"github.com/gorilla/websocket"
)

type CosmosWSTransport struct {
	c    *websocket.Conn
	sess *wsapi.Session
	l    *zap.Logger
}

func NewCosmosWSTransport() *CosmosWSTransport {
	return &CosmosWSTransport{}
}

func (ng *CosmosWSTransport) Connect(ctx context.Context, address string, RH connectivity.FunctionCallHandler) (err error) {
	if ng.c, _, err = websocket.DefaultDialer.DialContext(ctx, address, nil); err != nil {
		return err
	}

	ng.sess = wsapi.NewSession(ctx, ng.c, ng.l, RH)
	go ng.sess.Recv()
	go ng.sess.Req()

	return nil
}

func (ng *CosmosWSTransport) GetBlockByHeight(ctx context.Context, height uint64) (bTx structs.BlockAndTx, err error) {
	resp, err := ng.sess.SendSync("get_by_height", []json.RawMessage{[]byte(strconv.FormatUint(height, 10))})
	if err != nil {
		return structs.BlockAndTx{}, err
	}

	if err = json.Unmarshal(resp.Result, &bTx); err != nil {
		return structs.BlockAndTx{}, err
	}

	return bTx, nil

}

func (ng *CosmosWSTransport) GetLatestByHeight(ctx context.Context) (bTx structs.BlockAndTx, err error) {
	resp, err := ng.sess.SendSync("get_latest", nil)
	if err != nil {
		return structs.BlockAndTx{}, err
	}

	if err = json.Unmarshal(resp.Result, &bTx); err != nil {
		return structs.BlockAndTx{}, err
	}

	return bTx, nil

}
