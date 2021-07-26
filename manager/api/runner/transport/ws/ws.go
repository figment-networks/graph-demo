package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/figment-networks/graph-demo/connectivity"
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
	Add(ev string, sub subscription.Sub) error
	Remove(id string) error
}

type ErrorMessage struct {
	Message string `json:"message,omitempty"`
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

	registry     map[string]connectivity.Handler
	registrySync sync.RWMutex

	subscriptions Subscriber
}

func NewProcessHandler(log *zap.Logger, svc ManagerService, subscriptions Subscriber) *ProcessHandler {
	ph := &ProcessHandler{
		log:           log,
		service:       svc,
		subscriptions: subscriptions,
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
			Message: "Missing query",
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

	var vars map[string]interface{}
	if len(args) == 2 {
		if err := json.Unmarshal(args[1], &vars); err != nil {
			r.Errors = append(r.Errors, ErrorMessage{
				Message: "Error unmarshaling query variables " + err.Error(),
			})
			enc.Encode(r)
			resp.Send(json.RawMessage(b.Bytes()), nil)
			return
		}
	}

	response, err := ph.service.ProcessGraphqlQuery(ctx, vars, query)
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

	if err := json.Unmarshal(args[1], &events); err != nil {
		r.Errors = append(r.Errors, ErrorMessage{
			Message: "Missing subscription",
		})
		enc.Encode(r)
		// TODO(lukanus): error
		resp.Send(json.RawMessage(b.Bytes()), nil)
		return
	}

	for _, ev := range events {
		ph.subscriptions.Add(ev.Name, NewSubscriptionInstance(req.ConnID(), resp, ev.StartingHeight))
		ph.log.Debug("added subscription for event", zap.String("id", req.ConnID()), zap.String("event", ev.Name), zap.Uint64("from", ev.StartingHeight))
	}
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

	if err := json.Unmarshal(args[1], &events); err != nil {
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
}

func NewSubscriptionInstance(id string, resp connectivity.Response, from uint64) subscription.Sub {
	return &SubscriptionInstance{
		id:   id,
		from: from,
		resp: resp,
	}
}

type SubscriptionInstance struct {
	id string

	resp    connectivity.Response
	from    uint64
	current uint64
}

func (si *SubscriptionInstance) Send(ctx context.Context, height uint64, resp json.RawMessage) error {

	// TODO(l): send only if not initial

	return si.resp.Send(resp, nil)

}

func (si *SubscriptionInstance) ID() string {
	return si.id
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
