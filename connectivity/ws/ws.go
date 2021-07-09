package ws

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/figment-networks/graph-demo/connectivity"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var ErrConnectionClosed = errors.New("connection closed")

type Conn struct {
	l  *log.Logger
	RH *RegistryHandler
}

func (c *Conn) Attach(name string, handler connectivity.Handler) {
	c.RH.Add(name, handler)
}

func (c *Conn) AttachToMux(ctx context.Context, mux *http.ServeMux) {
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		uConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {

			//conn.l.Warn("Error upgrading connection", zap.Error(err))
			return
		}

		sess := NewSession(ctx, c.RH)
		go sess.recv(uConn)
		go sess.resp(uConn)
	})
}
