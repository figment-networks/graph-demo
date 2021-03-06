package client

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/figment-networks/graph-demo/connectivity"
	"go.uber.org/zap"
)

type EventClient interface {
	NewEvent(typ string, data map[string]interface{}) error
}

type NetworkGraphClient struct {
	ec EventClient

	registry     map[string]connectivity.Handler
	registrySync sync.RWMutex
	l            *zap.Logger
}

func NewNetworkGraphClient(l *zap.Logger, ec EventClient) *NetworkGraphClient {
	ph := &NetworkGraphClient{
		registry: make(map[string]connectivity.Handler),
		l:        l,
		ec:       ec,
	}
	ph.Add("event", ph.EventHandler)
	return ph
}

func (ng *NetworkGraphClient) EventHandler(ctx context.Context, req connectivity.Request, resp connectivity.Response) {
	args := req.Arguments()
	data := make(map[string]interface{})

	if err := json.Unmarshal(args[1], &data); err != nil {
		ng.l.Error("unmarshal error", zap.Error(err))
	}

	if err := ng.ec.NewEvent(strings.Replace(string(args[0]), `"`, "", -1), data); err != nil {
		ng.l.Error("new event error", zap.Error(err))
	}

	resp.Send(json.RawMessage([]byte(`"ACK"`)), nil)
}

func (ph *NetworkGraphClient) Add(name string, handler connectivity.Handler) {
	ph.registrySync.Lock()
	defer ph.registrySync.Unlock()
	ph.registry[name] = handler
}

func (ph *NetworkGraphClient) Get(name string) (h connectivity.Handler, ok bool) {
	ph.registrySync.RLock()
	defer ph.registrySync.RUnlock()

	h, ok = ph.registry[name]
	return h, ok
}
