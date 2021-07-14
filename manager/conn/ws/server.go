package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/figment-networks/graph-demo/manager/conn"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var ErrConnectionClosed = errors.New("connection closed")

type JsonRPCRequest struct {
	ID      uint64        `json:"id"`
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type JsonRPCSend struct {
	JsonRPCRequest
	RespCH chan Response
}

type JsonRPCError struct {
	Code    int64         `json:"code"`
	Message string        `json:"message"`
	Data    []interface{} `json:"data"`
}

type JsonRPCResponse struct {
	ID      uint64          `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Error   *JsonRPCError   `json:"error,omitempty"`
	Result  json.RawMessage `json:"result"`
}

type ResponseStore struct {
	ID     uint64 `json:"id"` // originalID
	Type   string
	RespCH chan Response
}

type Response struct {
	ID     uint64
	Error  error
	Type   string
	Result json.RawMessage
}

type RegistryHandler struct {
	Registry     map[string]conn.Handler
	registrySync sync.RWMutex
}

func (rh *RegistryHandler) Add(name string, handler conn.Handler) {
	rh.registrySync.Lock()
	defer rh.registrySync.Unlock()
	rh.Registry[name] = handler
}

func (rh *RegistryHandler) Get(name string) (h conn.Handler, ok bool) {
	rh.registrySync.RLock()
	defer rh.registrySync.RUnlock()

	h, ok = rh.Registry[name]
	return h, ok
}

type Conn struct {
	l  *zap.Logger
	RH *RegistryHandler
}

type LockedResponseMap struct {
	Map map[uint64]ResponseStore
	L   sync.RWMutex
}

func NewConn(l *zap.Logger) *Conn {
	return &Conn{
		l:  l,
		RH: &RegistryHandler{Registry: make(map[string]conn.Handler)},
	}
}

func (conn *Conn) Attach(name string, handler conn.Handler) {
	conn.RH.Add(name, handler)
}

func (conn *Conn) AttachToMux(mux *http.ServeMux) {
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		uConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			conn.l.Warn("Error upgrading connection", zap.Error(err))
			return
		}

		sess := NewSession()
		ctx, cancel := context.WithCancel(context.Background())
		go sess.recv(ctx, cancel, conn.l, uConn, conn.RH)
		go sess.resp(ctx, cancel, conn.l, uConn, conn.RH)
	})
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type Session struct {
	// Buffered channel of outbound messages.
	send chan JsonRPCResponse
}

type SessionRequest struct {
	args []interface{}
}

func (sR *SessionRequest) Arguments() []interface{} {
	return sR.args
}

type SessionResponse struct {
	ID uint64

	SessionContext context.Context
	RespCh         chan JsonRPCResponse
}

func (sR *SessionResponse) Send(data io.ReadCloser, err error) error {
	select {
	case <-sR.SessionContext.Done():
		return errors.New("session context closed")
	default:
		resp := JsonRPCResponse{
			ID:      sR.ID,
			JSONRPC: "2.0",
		}
		if data != nil {
			defer data.Close()
			var encodingError error
			resp.Result, encodingError = io.ReadAll(data)
			if encodingError != nil {
				resp.Error = &JsonRPCError{Message: err.Error()}
				sR.RespCh <- resp
				return encodingError
			}
		}
		if err != nil {
			resp.Error = &JsonRPCError{Message: err.Error()}
		}
		sR.RespCh <- resp
	}
	return nil
}

func (sR *SessionResponse) Write(p []byte) (n int, err error) {
	return len(p), sR.Send(io.NopCloser(bytes.NewReader(p)), nil)
}

func NewSession() *Session {
	return &Session{send: make(chan JsonRPCResponse, 10)}
}

func (s *Session) recv(ctx context.Context, close context.CancelFunc, l *zap.Logger, conn *websocket.Conn, registry *RegistryHandler) {
	defer func() {
		conn.Close()
	}()

	// TODO check this
	err := conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		l.Error("error setting read deadline", zap.Error(err))
		return
	}

	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				l.Error("[CONN] websocket unexpected close error", zap.Error(err))
			}
			break
		}
		req := &JsonRPCRequest{}
		err = json.Unmarshal(message, req)
		if err != nil {
			l.Error("error unmarshaling jsonrpc response", zap.Error(err))
			continue
		}

		if req.JSONRPC != "2.0" {
			s.send <- JsonRPCResponse{JSONRPC: "2.0", Error: &JsonRPCError{Code: -32700, Message: "Parse error"}}
		}

		h, ok := registry.Get(req.Method)
		if !ok {
			s.send <- JsonRPCResponse{ID: req.ID, JSONRPC: "2.0", Error: &JsonRPCError{Code: -32601, Message: "Method not found"}}
		}

		go h(ctx, &SessionRequest{args: req.Params}, &SessionResponse{
			ID:             req.ID,
			SessionContext: ctx,
			RespCh:         s.send,
		})
	}
}

func (s *Session) resp(ctx context.Context, close context.CancelFunc, l *zap.Logger, conn *websocket.Conn, registry *RegistryHandler) {
	defer l.Sync()

	tckr := time.NewTicker(pingPeriod)
	defer tckr.Stop()

	buff := new(bytes.Buffer)
	enc := json.NewEncoder(buff)

WSLOOP:
	for {
		select {
		case <-ctx.Done():
			l.Info("[API] closing connection on context done")
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				l.Error("[API] Error closing websocket ", zap.Error(err))
				break WSLOOP
			}

			<-time.After(time.Second)

			break WSLOOP
		case message, ok := <-s.send:
			if !ok {
				l.Info("[API] send is closed")
				if conn != nil {
					conn.WriteMessage(websocket.CloseMessage, []byte{})
				}
				return
			}

			buff.Reset()
			if err := enc.Encode(message); err != nil {
				l.Info("[API] error in encode")
				/*	req.RespCH <- Response{
					ID:    originalID,
					Type:  req.Method,
					Error: fmt.Errorf("error encoding message: %w ", err),
				}*/
				continue WSLOOP
			}
			/*
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if !ok {
					// The hub closed the channel.
					conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
			*/

			err := conn.WriteMessage(websocket.TextMessage, buff.Bytes())
			if err != nil {
				l.Error("[API] Error sending data websocket ", zap.Error(err))
				break WSLOOP
			}

			/*
				if err := w.Close(); err != nil {
					return
				}
			*/
		case <-tckr.C:
			// conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		}
	}
}
