package ws

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/figment-networks/graph-demo/connectivity"
	wsapi "github.com/figment-networks/graph-demo/connectivity/ws"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type GQLPayload struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
	//	OperationName string                 `json:"operationName"`
}

type GQLResponse struct {
	Data   interface{}   `json:"data"`
	Errors []interface{} `json:"errors"`
}

type NetworkGraphWSTransport struct {
	c    *websocket.Conn
	sess *wsapi.Session
	l    *zap.Logger
}

func NewNetworkGraphWSTransport(l *zap.Logger) *NetworkGraphWSTransport {
	return &NetworkGraphWSTransport{l: l}
}

func (ng *NetworkGraphWSTransport) Connect(ctx context.Context, address string, RH connectivity.FunctionCallHandler) (err error) {

	ng.c, _, err = websocket.DefaultDialer.DialContext(ctx, address, nil)

	ng.sess = wsapi.NewSession(ctx, ng.c, ng.l, RH)
	go ng.sess.Recv()
	go ng.sess.Req()

	return err
}

func (ng *NetworkGraphWSTransport) CallGQL(ctx context.Context, name string, query string, variables map[string]interface{}) ([]byte, error) {
	buff := new(bytes.Buffer)
	defer buff.Reset()
	enc := json.NewEncoder(buff)
	if err := enc.Encode(GQLPayload{query, variables}); err != nil {
		return nil, err
	}

	resp, err := ng.sess.SendSync(name, []json.RawMessage{[]byte(query), buff.Bytes()})
	buff.Reset()

	return resp.Result, err
}
