package conn

import (
	"context"
	"io"
)

type Response interface {
	Send(io.ReadCloser, error) error

	Write(p []byte) (n int, err error)
}

type Request interface {
	Arguments() []interface{}
}

type Handler func(ctx context.Context, req Request, resp Response)

type WsConn interface {
	Attach(string, Handler)
}
