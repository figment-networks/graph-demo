package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"

	"github.com/figment-networks/graph-demo/connectivity"
	wsConn "github.com/figment-networks/graph-demo/connectivity/ws"
	"github.com/figment-networks/graph-demo/manager/structs"
	"github.com/figment-networks/graph-demo/manager/subscription"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var ErrConnectionClosed = errors.New("connection closed")

type Subscriber interface {
	Add(ctx context.Context, ev string, sub subscription.Sub) error
	Remove(id string) error
}

type ErrorMessage struct {
	Message string `json:"message,omitempty"`
}

type JSONGraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type JSONGraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []ErrorMessage  `json:"errors,omitempty"`
}

type ManagerService interface {
	ProcessGraphqlQuery(ctx context.Context, v map[string]interface{}, q string) ([]byte, error)
}

type ProcessHandler struct {
	service ManagerService
	log     *zap.Logger

	reg *wsConn.Registry

	registry     map[string]connectivity.Handler
	registrySync sync.RWMutex

	subscriptions Subscriber
}

func NewProcessHandler(log *zap.Logger, svc ManagerService, reg *wsConn.Registry, subscriptions Subscriber) *ProcessHandler {
	ph := &ProcessHandler{
		log:           log,
		service:       svc,
		subscriptions: subscriptions,
		reg:           reg,
		registry:      make(map[string]connectivity.Handler),
	}
	ph.Add("query", ph.GraphQLRequest)
	ph.Add("subscribe", ph.Subscribe)
	ph.Add("unsubscribe", ph.Unsubscribe)
	return ph
}

func (ph *ProcessHandler) Add(name string, handler connectivity.Handler) {
	ph.registrySync.Lock()
	defer ph.registrySync.Unlock()
	ph.registry[name] = handler
}

func (ph *ProcessHandler) Get(name string) (h connectivity.Handler, ok bool) {
	ph.registrySync.RLock()
	defer ph.registrySync.RUnlock()

	h, ok = ph.registry[name]
	return h, ok
}

func (ph *ProcessHandler) GraphQLRequest(ctx context.Context, req connectivity.Request, resp connectivity.Response) {
	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)
	r := JSONGraphQLResponse{}

	args := req.Arguments()
	if len(args) == 0 {
		r.Errors = append(r.Errors, ErrorMessage{
			Message: "Missing query (GraphQLRequest)",
		})
		enc.Encode(r)
		resp.Send(json.RawMessage(b.Bytes()), nil)
		return
	}

	var query string
	if err := json.Unmarshal(args[0], &query); err != nil {
		r.Errors = append(r.Errors, ErrorMessage{
			Message: "Error unmarshaling query " + err.Error(),
		})
		enc.Encode(r)
		resp.Send(json.RawMessage(b.Bytes()), nil)
		return
	}

	var gQLReq JSONGraphQLRequest
	if len(args) > 1 {
		if err := json.Unmarshal(args[1], &gQLReq); err != nil {
			r.Errors = append(r.Errors, ErrorMessage{
				Message: "Error unmarshaling query variables " + err.Error(),
			})
			enc.Encode(r)
			resp.Send(json.RawMessage(b.Bytes()), nil)
			return
		}
	}

	response, err := ph.service.ProcessGraphqlQuery(ctx, gQLReq.Variables, gQLReq.Query)

	log.Println("response 111", string(response))
	resp.Send(response, err)
}

func (ph *ProcessHandler) Subscribe(ctx context.Context, req connectivity.Request, resp connectivity.Response) {
	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)
	r := JSONGraphQLResponse{}

	args := req.Arguments()
	if len(args) == 0 {
		r.Errors = append(r.Errors, ErrorMessage{
			Message: "Missing subscription",
		})
		enc.Encode(r)
		resp.Send(json.RawMessage(b.Bytes()), nil)
		return
	}

	var events []structs.Subs

	if err := json.Unmarshal(args[0], &events); err != nil {
		r.Errors = append(r.Errors, ErrorMessage{
			Message: "Missing subscription",
		})
		enc.Encode(r)
		// TODO(lukanus): error
		resp.Send(json.RawMessage(b.Bytes()), nil)
		return
	}

	for _, ev := range events {
		ph.subscriptions.Add(ctx, ev.Name, NewSubscriptionInstance(req.ConnID(), ph.reg, ev.StartingHeight))
		ph.log.Debug("added subscription for event", zap.String("id", req.ConnID()), zap.String("event", ev.Name), zap.Uint64("from", ev.StartingHeight))
	}

	resp.Send(json.RawMessage([]byte(`"ACK"`)), nil)
}

func (ph *ProcessHandler) Unsubscribe(ctx context.Context, req connectivity.Request, resp connectivity.Response) {
	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)
	r := JSONGraphQLResponse{}

	args := req.Arguments()
	if len(args) == 0 {
		r.Errors = append(r.Errors, ErrorMessage{
			Message: "Missing subscription",
		})
		enc.Encode(r)
		resp.Send(json.RawMessage(b.Bytes()), nil)
		return
	}

	var events []string

	if err := json.Unmarshal(args[0], &events); err != nil {
		r.Errors = append(r.Errors, ErrorMessage{
			Message: "Missing subscription",
		})
		enc.Encode(r)
		// TODO(lukanus): error
		resp.Send(json.RawMessage(b.Bytes()), nil)

		return
	}

	for _, ev := range events {
		err := ph.subscriptions.Remove(req.ConnID())
		if err != nil {
			r.Errors = append(r.Errors, ErrorMessage{
				Message: "Missing subscription",
			})
			enc.Encode(r)
			return
		}
		// TODO(lukanus): error
		ph.log.Debug("removed subscription for event", zap.String("id", req.ConnID()), zap.String("event", ev))
	}

	resp.Send(json.RawMessage([]byte(`"ACK"`)), nil)
}

func NewSubscriptionInstance(connID string, reg *wsConn.Registry, from uint64) subscription.Sub {
	return &SubscriptionInstance{
		connID: connID,
		reg:    reg,
		from:   from,
	}
}

type SubscriptionInstance struct {
	connID string

	reg     *wsConn.Registry
	from    uint64
	current uint64
}

func (si *SubscriptionInstance) Send(ctx context.Context, height uint64, name string, resp json.RawMessage) error {

	// TODO(l): send only if not initial
	log.Println("log event", height, string(resp))

	ss, ok := si.reg.Get(si.connID)
	if !ok || ss == nil {
		return errors.New("connection does not exists")
	}

	_, err := ss.SendSync("event", []json.RawMessage{[]byte(`"` + name + `"`), resp})

	return err

}

func (si *SubscriptionInstance) ID() string {
	return si.connID
}

func (si *SubscriptionInstance) FromHeight() uint64 {
	return si.from
}

func (si *SubscriptionInstance) CurrentHeight() uint64 {
	return si.current
}

func (si *SubscriptionInstance) SetCurrentHeight(c uint64) {
	si.current = c
}
