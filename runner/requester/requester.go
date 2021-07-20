package requester

import (
	"context"
	"errors"
	"sync"
)

type Caller interface {
	CallGQL(ctx context.Context, name string, query string, variables map[string]interface{}, version string) ([]byte, error)
}

type Rqstr struct {
	list  map[string]Caller
	llock sync.RWMutex
}

func NewRqstr() *Rqstr {
	return &Rqstr{
		list: make(map[string]Caller),
	}
}

func (r *Rqstr) AddDestination(name string, dest Caller) {
	r.llock.Lock()
	r.list[name] = dest
	r.llock.Unlock()
}

func (r *Rqstr) CallGQL(ctx context.Context, name string, query string, variables map[string]interface{}, version string) ([]byte, error) {
	r.llock.RLock()
	d, ok := r.list[name]
	r.llock.RUnlock()
	if !ok {
		return nil, errors.New("graph not found: " + name)
	}

	return d.CallGQL(ctx, name, query, variables, version)
}
