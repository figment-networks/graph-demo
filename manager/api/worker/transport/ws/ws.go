package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"sync"

	"github.com/figment-networks/graph-demo/connectivity"
	"github.com/figment-networks/graph-demo/manager/scheduler"
	"github.com/figment-networks/graph-demo/manager/structs"

	wsConn "github.com/figment-networks/graph-demo/connectivity/ws"
	cliTr "github.com/figment-networks/graph-demo/manager/client/transport/ws"

	"go.uber.org/zap"
)

type JSONGraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []ErrorMessage  `json:"errors,omitempty"`
}

type ErrorMessage struct {
	Message string `json:"message,omitempty"`
}

type ManagerService interface {
	StoreBlock(ctx context.Context, block structs.Block) error
	StoreTransactions(ctx context.Context, txs []structs.Transaction) error
}

type ProcessHandler struct {
	service ManagerService
	log     *zap.Logger
	sched   *scheduler.Scheduler

	// session storagre
	reg *wsConn.Registry

	// function registry
	registry     map[string]connectivity.Handler
	registrySync sync.RWMutex
}

func NewProcessHandler(log *zap.Logger, svc ManagerService, sched *scheduler.Scheduler, reg *wsConn.Registry) *ProcessHandler {
	ph := &ProcessHandler{
		log:      log,
		service:  svc,
		reg:      reg,
		sched:    sched,
		registry: make(map[string]connectivity.Handler),
	}

	ph.Add("register", ph.Register)
	ph.Add("store_block", ph.StoreBlock)
	ph.Add("store_transactions", ph.StoreTransactions)

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

func (ph *ProcessHandler) StoreBlock(ctx context.Context, req connectivity.Request, resp connectivity.Response) {
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

	var block structs.Block
	if err := json.Unmarshal(args[0], &block); err != nil {
		r.Errors = append(r.Errors, ErrorMessage{
			Message: "Error unmarshaling query " + err.Error(),
		})
		enc.Encode(r)
		resp.Send(json.RawMessage(b.Bytes()), nil)
		return
	}

	err := ph.service.StoreBlock(ctx, block)
	resp.Send(json.RawMessage([]byte(`"ACK"`)), err)
}

func (ph *ProcessHandler) StoreTransactions(ctx context.Context, req connectivity.Request, resp connectivity.Response) {
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

	var txs []structs.Transaction
	if err := json.Unmarshal(args[0], &txs); err != nil {
		r.Errors = append(r.Errors, ErrorMessage{
			Message: "Error unmarshaling query " + err.Error(),
		})
		enc.Encode(r)
		resp.Send(json.RawMessage(b.Bytes()), nil)
		return
	}

	err := ph.service.StoreTransactions(ctx, txs)
	resp.Send(json.RawMessage([]byte(`"ACK"`)), err)
}

func (ph *ProcessHandler) Register(ctx context.Context, req connectivity.Request, resp connectivity.Response) {

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

	var chainID string
	if err := json.Unmarshal(args[0], &chainID); err != nil {
		r.Errors = append(r.Errors, ErrorMessage{
			Message: "Error unmarshaling query " + err.Error(),
		})
		enc.Encode(r)
		resp.Send(json.RawMessage(b.Bytes()), nil)
		return
	}

	ss, ok := ph.reg.Get(req.ConnID())
	if !ok {
		r.Errors = append(r.Errors, ErrorMessage{
			Message: "Session does not exists",
		})
		enc.Encode(r)
		resp.Send(json.RawMessage(b.Bytes()), nil)
		return
	}

	go ph.sched.Start(ctx, cliTr.NewCosmosWSTransport(ss), req.ConnID(), chainID)

	resp.Send(json.RawMessage([]byte(`"ACK"`)), nil)

}
