package jsonrpc

import (
	"sync"

	"github.com/figment-networks/graph-demo/connectivity"
)

// RegistryHandler takes care of functions attached to interface calls
type RegistryHandler struct {
	Registry     map[string]connectivity.Handler
	registrySync sync.RWMutex
}

func NewRegistryHandler() *RegistryHandler {
	return &RegistryHandler{Registry: make(map[string]connectivity.Handler)}
}

func (rh *RegistryHandler) Add(name string, handler connectivity.Handler) {
	rh.registrySync.Lock()
	defer rh.registrySync.Unlock()
	rh.Registry[name] = handler
}

func (rh *RegistryHandler) Get(name string) (h connectivity.Handler, ok bool) {
	rh.registrySync.RLock()
	defer rh.registrySync.RUnlock()

	h, ok = rh.Registry[name]
	return h, ok
}
