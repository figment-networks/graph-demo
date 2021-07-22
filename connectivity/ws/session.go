package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/figment-networks/graph-demo/connectivity"
	"github.com/figment-networks/graph-demo/connectivity/jsonrpc"
	"github.com/google/uuid"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	pingTime = 50 * time.Second
)

// Session represents websocket connection during it's livetime
type Session struct {
	ID        string
	c         *websocket.Conn
	reg       connectivity.FunctionCallHandler
	ctx       context.Context
	ctxCancel context.CancelFunc
	l         *zap.Logger

	// Buffered channel of outbound messages.
	send     chan jsonrpc.Request
	response chan jsonrpc.Response

	routing     map[uint64]*Waiting
	routingLock sync.RWMutex
	newID       *uint64
}

type Waiting struct {
	returnCh chan jsonrpc.Response
}

func NewWaiting() *Waiting {
	return &Waiting{returnCh: make(chan jsonrpc.Response, 1)}
}

func NewSession(ctx context.Context, c *websocket.Conn, l *zap.Logger, reg connectivity.FunctionCallHandler) *Session {
	nCtx, cancel := context.WithCancel(ctx)

	firstCall := uint64(0)
	return &Session{
		ID:        uuid.NewString(),
		reg:       reg,
		c:         c,
		ctx:       nCtx,
		ctxCancel: cancel,
		l:         l,
		send:      make(chan jsonrpc.Request, 10),
		response:  make(chan jsonrpc.Response, 10),
		newID:     &firstCall,
		routing:   make(map[uint64]*Waiting),
	}
}

func (s *Session) Send(req jsonrpc.Request) {
	s.send <- req
}

func (s *Session) SendSync(method string, params []json.RawMessage) (jsonrpc.Response, error) {

	w := NewWaiting()
	defer close(w.returnCh)
	id := atomic.AddUint64(s.newID, 1)

	s.routingLock.Lock()
	s.routing[id] = w
	s.routingLock.Unlock()

	s.send <- jsonrpc.Request{
		ID:      id,
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	return <-w.returnCh, nil
}

func (s *Session) Recv() {
	defer func() {
		s.c.Close()
	}()

	err := s.c.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		s.l.Error("error setting read deadline", zap.Error(err))
		return
	}

	s.c.SetPongHandler(func(string) error {
		return s.c.SetReadDeadline(time.Now().Add(pongWait))
	})

	readr := bytes.NewReader(nil)
	dec := json.NewDecoder(readr)
	for {
		_, message, err := s.c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.l.Error(" websocket unexpected close error", zap.Error(err))
			}
			break
		}

		readr.Reset(message)
		req := &jsonrpc.Hybrid{}
		if err = dec.Decode(req); err != nil {
			s.l.Error("error unmarshaling jsonrpc response", zap.Error(err))
			continue
		}

		if req.JSONRPC != "2.0" {
			s.response <- jsonrpc.Response{JSONRPC: "2.0", Error: &jsonrpc.Error{Code: -32700, Message: "Parse error"}}
		}

		if req.Result != nil || req.Error != nil {
			s.routingLock.RLock()
			waitO, ok := s.routing[req.ID]
			s.routingLock.RUnlock()
			if !ok {
				s.l.Error("unexpected message", zap.Any("message", req))
			}
			waitO.returnCh <- jsonrpc.Response{ID: req.ID, JSONRPC: "2.0", Result: req.Result, Error: req.Error}
			delete(s.routing, req.ID)
			s.routingLock.RUnlock()

			continue
		}

		h, ok := s.reg.Get(req.Method)
		if !ok {
			s.response <- jsonrpc.Response{ID: req.ID, JSONRPC: "2.0", Error: &jsonrpc.Error{Code: -32601, Message: "Method not found"}}
		}

		go h(s.ctx, &SessionRequest{args: req.Params, connID: s.ID}, &SessionResponse{
			ID:             req.ID,
			SessionContext: s.ctx,
			RespCh:         s.response,
		})
	}
}

func (s *Session) Req() {

	tckr := time.NewTicker(pingTime)
	defer tckr.Stop()

	buff := new(bytes.Buffer)
	enc := json.NewEncoder(buff)
WSLOOP:
	for {
		select {
		case <-s.ctx.Done():
			s.l.Info("closing connection on context done")
			if err := s.c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
				s.l.Error("error closing websocket ", zap.Error(err))
				break WSLOOP
			}

			<-time.After(time.Second)

			break WSLOOP

		case message, ok := <-s.send:
			if !ok {
				s.l.Info("send is closed")
				if s.c != nil {
					s.c.WriteMessage(websocket.CloseMessage, []byte{})
				}
				return
			}

			buff.Reset()
			if err := enc.Encode(message); err != nil {
				s.l.Info("error in encode", zap.Error(err))
				continue WSLOOP
			}

			if err := s.c.WriteMessage(websocket.TextMessage, buff.Bytes()); err != nil {
				s.l.Error("error sending data websocket ", zap.Error(err))
				break WSLOOP
			}

		case message, ok := <-s.response:
			if !ok {
				s.l.Info("send is closed")
				if s.c != nil {
					s.c.WriteMessage(websocket.CloseMessage, []byte{})
				}
				return
			}

			buff.Reset()
			if err := enc.Encode(message); err != nil {
				s.l.Info("error in encode", zap.Error(err))
				continue WSLOOP
			}

			if err := s.c.WriteMessage(websocket.TextMessage, buff.Bytes()); err != nil {
				s.l.Error("error sending data websocket ", zap.Error(err))
				break WSLOOP
			}

		case message, ok := <-s.response:
			if !ok {
				s.l.Info("send is closed")
				if s.c != nil {
					s.c.WriteMessage(websocket.CloseMessage, []byte{})
				}
				return
			}

			buff.Reset()
			if err := enc.Encode(message); err != nil {
				s.l.Info("error in encode", zap.Error(err))
				/*	req.RespCH <- Response{
					ID:    originalID,
					Type:  req.Method,
					Error: fmt.Errorf("error encoding message: %w ", err),
				}*/
				continue WSLOOP
			}

			if err := s.c.WriteMessage(websocket.TextMessage, buff.Bytes()); err != nil {
				s.l.Error("error sending data websocket ", zap.Error(err))
				break WSLOOP
			}

			/*
				if err := w.Close(); err != nil {
					return
				}
			*/
		case <-tckr.C:
			// conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := s.c.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		}
	}
}

type SessionResponse struct {
	ID uint64

	SessionContext context.Context
	RespCh         chan jsonrpc.Response
}

func (s *SessionResponse) Send(result json.RawMessage, er error) error {

	resp := jsonrpc.Response{
		ID:      s.ID,
		JSONRPC: "2.0",
		Result:  result,
	}

	if er != nil {
		resp.Error = &jsonrpc.Error{Message: er.Error()}
	}

	s.RespCh <- resp
	return nil
}

type SessionRequest struct {
	connID string
	args   []json.RawMessage
}

func (sR *SessionRequest) Arguments() []json.RawMessage {
	return sR.args
}

func (sR *SessionRequest) ConnID() string {
	return sR.connID
}
