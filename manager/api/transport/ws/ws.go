package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/figment-networks/graph-demo/connectivity"
	"github.com/figment-networks/graph-demo/connectivity/jsonrpc"
	"github.com/figment-networks/graph-demo/manager/api/service"
	"github.com/figment-networks/graph-demo/manager/subscription"
)

type Sub interface {
	Send(jsonrpc.Response)

	ID() string

	FromHash() uint64
	CurrentHash() uint64
}

type Subscriber interface {
	Add(ev string, sub Sub) error
	Remove(id string) error
}

type ErrorMessage struct {
	Message string `json:"message",omitempty`
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

	registry     map[string]connectivity.Handler
	registrySync sync.RWMutex

	subscriptions Subscriber
}

func NewProcessHandler(svc *service.Service) *ProcessHandler {
	ph := &ProcessHandler{
		service:  svc,
		registry: make(map[string]connectivity.Handler),
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
				Message: "Error unmarshaling quevatiablesry " + err.Error(),
			})
			enc.Encode(r)
			resp.Send(json.RawMessage(b.Bytes()), nil)
			return
		}
	}

	log.Println("ProcessGraphqlQuery", vars, query)
	//	b, err := ph.service.ProcessGraphqlQuery(ctx, vars, query)

	resp.Send(json.RawMessage(b.Bytes()), nil)

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
		subscription.NewSubI()
		err = ph.subscriptions.Add(ev)
	}

}

func (ph *ProcessHandler) Unsubscribe(ctx context.Context, req connectivity.Request, resp connectivity.Response) {

}
