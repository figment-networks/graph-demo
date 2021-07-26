package requester

import (
	"context"
	"errors"
	"sync"

	"github.com/figment-networks/graph-demo/runner/structs"
)

type Caller interface {
	CallGQL(ctx context.Context, name string, query string, variables map[string]interface{}, version string) ([]byte, error)

	Subscribe(ctx context.Context, events []structs.Subs) error
	Unsubscribe(ctx context.Context, events []string) error
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

func (r *Rqstr) Subscribe(ctx context.Context, name string, events []structs.Subs) error {
	r.llock.RLock()
	d, ok := r.list[name]
	r.llock.RUnlock()
	if !ok {
		return errors.New("graph not found: " + name)
	}

	return d.Subscribe(ctx, events)
}

func (r *Rqstr) Unsubscribe(ctx context.Context, name string, events []string) error {
	r.llock.RLock()
	d, ok := r.list[name]
	r.llock.RUnlock()
	if !ok {
		return errors.New("graph not found: " + name)
	}

	return d.Unsubscribe(ctx, events)
}
