package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/figment-networks/graph-demo/connectivity"
	wsapi "github.com/figment-networks/graph-demo/connectivity/ws"
	"github.com/figment-networks/graph-demo/manager/structs"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type JSONGraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []ErrorMessage  `json:"errors,omitempty"`
}

type ErrorMessage struct {
	Message string `json:"message,omitempty"`
}

type CosmosClient interface {
	GetAll(ctx context.Context, height uint64) error
	GetLatest(ctx context.Context) (uint64, error)
}

type ProcessHandler struct {
	service CosmosClient
	log     *zap.Logger

	c    *websocket.Conn
	sess *wsapi.Session

	registry     map[string]connectivity.Handler
	registrySync sync.RWMutex
}

func NewProcessHandler(log *zap.Logger, svc CosmosClient) *ProcessHandler {
	ph := &ProcessHandler{
		log:      log,
		service:  svc,
		registry: make(map[string]connectivity.Handler),
	}

	ph.Add("get_all", ph.GetAll)
	ph.Add("get_latest", ph.GetLatest)

	return ph
}

func (ng *ProcessHandler) Connect(ctx context.Context, address string) (err error) {
	if ng.c, _, err = websocket.DefaultDialer.DialContext(ctx, address, nil); err != nil {
		return err
	}

	ng.sess = wsapi.NewSession(ctx, ng.c, ng.log, ng)
	go ng.sess.Recv()
	go ng.sess.Req()

	return nil
}

func (ng *ProcessHandler) Register(ctx context.Context, chainID string) (err error) {
	_, err = ng.sess.SendSync("register", []json.RawMessage{[]byte(`"` + chainID + `"`)})
	return err
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

func (ph *ProcessHandler) GetAll(ctx context.Context, req connectivity.Request, resp connectivity.Response) {
	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)
	r := JSONGraphQLResponse{}

	args := req.Arguments()
	if len(args) == 0 {
		r.Errors = append(r.Errors, ErrorMessage{
			Message: "Missing query (getAll)",
		})
		enc.Encode(r)
		resp.Send(json.RawMessage(b.Bytes()), nil)
		return
	}

	var height uint64
	if err := json.Unmarshal(args[0], &height); err != nil {
		r.Errors = append(r.Errors, ErrorMessage{
			Message: "Error unmarshaling query " + err.Error(),
		})
		enc.Encode(r)
		resp.Send(json.RawMessage(b.Bytes()), nil)
		return
	}

	if err := ph.service.GetAll(ctx, height); err != nil {
		r.Errors = append(r.Errors, ErrorMessage{
			Message: "Error unmarshaling query " + err.Error(),
		})
		enc.Encode(r)
		resp.Send(json.RawMessage(b.Bytes()), nil)
		return
	}

	resp.Send(json.RawMessage([]byte(`"ACK"`)), nil)
}

func (ph *ProcessHandler) GetLatest(ctx context.Context, req connectivity.Request, resp connectivity.Response) {
	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)
	r := JSONGraphQLResponse{}

	block, err := ph.service.GetLatest(ctx)
	if err != nil {
		r.Errors = append(r.Errors, ErrorMessage{
			Message: "Error getting latest",
		})
		enc.Encode(r)
		resp.Send(json.RawMessage(b.Bytes()), nil)
		return
	}

	encoded, err := json.Marshal(block)
	if err != nil {
		r.Errors = append(r.Errors, ErrorMessage{
			Message: "Error encoding respons",
		})
		enc.Encode(r)
		resp.Send(json.RawMessage(b.Bytes()), nil)
		return
	}
	resp.Send(encoded, nil)
}

func (ph *ProcessHandler) StoreTransactions(ctx context.Context, txs []structs.Transaction) error {
	txsM, err := json.Marshal(txs)
	if err != nil {
		return err
	}

	resp, err := ph.sess.SendSync("store_transactions", []json.RawMessage{txsM})
	if err != nil {
		return err
	}
	if resp.Error != nil && resp.Error.Message != "" {
		return errors.New(resp.Error.Message)
	}

	return nil
}

func (ph *ProcessHandler) StoreBlock(ctx context.Context, block structs.Block) error {
	blockM, err := json.Marshal(block)
	if err != nil {
		return err
	}

	resp, err := ph.sess.SendSync("store_block", []json.RawMessage{blockM})
	if err != nil {
		return err
	}
	if resp.Error != nil && resp.Error.Message != "" {
		return errors.New(resp.Error.Message)
	}

	return nil
}
