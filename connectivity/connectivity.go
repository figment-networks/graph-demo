package connectivity

import (
	"context"
	"encoding/json"
)

type Response interface {
	Send(result json.RawMessage, er error) error
}

type Request interface {
	ConnID() string
	Arguments() []json.RawMessage
}

type FunctionCallHandler interface {
	Get(name string) (h Handler, ok bool)
	Add(name string, h Handler)
}

type Handler func(ctx context.Context, req Request, resp Response)

type WsConn interface {
	Attach(string, Handler)
}
