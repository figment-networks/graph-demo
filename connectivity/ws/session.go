package ws

import (
	"context"

	"github.com/figment-networks/graph-demo/connectivity"
)

type FunctionCallHandler interface {
	Get(name string) (h connectivity.Handler, ok bool)
}

// Session represents websocket connection during it's livetime
type Session struct {
	h         FunctionCallHandler
	ctx       context.Context
	ctxCancel context.CancelFunc

	// Buffered channel of outbound messages.
	//	send chan jsonrpc.JsonRPCResponse
}

func NewSession(ctx context.Context, h FunctionCallHandler) *Session {
	nCtx, cancel := context.WithCancel(ctx)
	return &Session{
		h:         h,
		ctx:       nCtx,
		ctxCancel: cancel,
		//	send: make(chan jsonrpc.JsonRPCResponse, 10)
	}
}
