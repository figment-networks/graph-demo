package ws

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/figment-networks/graph-demo/connectivity"
	wsapi "github.com/figment-networks/graph-demo/connectivity/ws"
	"github.com/figment-networks/graph-demo/runner/structs"
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
	ph := &NetworkGraphWSTransport{
		l: l,
	}
	return ph
}

func (ng *NetworkGraphWSTransport) Connect(ctx context.Context, address string, RH connectivity.FunctionCallHandler) (err error) {
	ng.c, _, err = websocket.DefaultDialer.DialContext(ctx, address, nil)
	if err != nil {
		return err
	}
	ng.sess = wsapi.NewSession(ctx, ng.c, ng.l, RH)
	go ng.sess.Recv()
	go ng.sess.Req()

	return nil
}

func (ng *NetworkGraphWSTransport) CallGQL(ctx context.Context, name string, query string, variables map[string]interface{}, version string) ([]byte, error) {
	buff := new(bytes.Buffer)
	defer buff.Reset()
	enc := json.NewEncoder(buff)
	if err := enc.Encode(GQLPayload{query, variables}); err != nil {
		return nil, err
	}

	q, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	v, err := json.Marshal(version)
	if err != nil {
		return nil, err
	}

	resp, err := ng.sess.SendSync("query", []json.RawMessage{q, buff.Bytes(), v})
	buff.Reset()
	return resp.Result, err
}

func (ng *NetworkGraphWSTransport) Subscribe(ctx context.Context, events []structs.Subs) error {
	buff := new(bytes.Buffer)
	defer buff.Reset()
	enc := json.NewEncoder(buff)
	if err := enc.Encode(events); err != nil {
		return err
	}

	_, err := ng.sess.SendSync("subscribe", []json.RawMessage{buff.Bytes()})
	buff.Reset()

	return err
}

func (ng *NetworkGraphWSTransport) Unsubscribe(ctx context.Context, events []string) error {
	buff := new(bytes.Buffer)
	defer buff.Reset()
	enc := json.NewEncoder(buff)
	if err := enc.Encode(events); err != nil {
		return err
	}

	_, err := ng.sess.SendSync("unsubscribe", []json.RawMessage{buff.Bytes()})
	buff.Reset()

	return err
}
