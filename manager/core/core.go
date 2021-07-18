package core

import (
	"context"
	"sync"

	"github.com/figment-networks/graph-demo/connectivity"
	"github.com/figment-networks/graph-demo/manager/structs"
	"github.com/graphql-go/graphql"
)

var schema, _ = graphql.NewSchema(
	graphql.SchemaConfig{
		Query: queryType,
	},
)

type ManagerClient interface {
	GetByHeight(ctx context.Context, height uint64) (structs.All, error)
}

type ProcessHandler struct {
	registry     map[string]connectivity.Handler
	c            ManagerClient
	registrySync sync.RWMutex
}

func NewProcessHandler() *ProcessHandler {
	ph := &ProcessHandler{registry: make(map[string]connectivity.Handler)}
	ph.Add("query", ph.GraphQLRequest)
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
	args := req.Arguments()
	query := args[0]

	resp.Send()
}
