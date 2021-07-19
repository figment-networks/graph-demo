package ws

import (
	"context"
	"errors"
	"net/http"

	"github.com/figment-networks/graph-demo/connectivity"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var ErrConnectionClosed = errors.New("connection closed")

type Conn struct {
	RH connectivity.FunctionCallHandler
	l  *zap.Logger
}

func (c *Conn) AttachToMux(ctx context.Context, mux *http.ServeMux) {
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		uConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			c.l.Warn("Error upgrading connection", zap.Error(err))
			return
		}

		sess := NewSession(ctx, uConn, c.l, c.RH)
		go sess.Recv()
		go sess.Req()
	})
}
