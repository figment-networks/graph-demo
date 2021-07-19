package subscription

import (
	"context"
	"sync"

	"github.com/figment-networks/graph-demo/connectivity/jsonrpc"
)

type Sub interface {
	Send(jsonrpc.Response) error

	ID() string

	FromHeight() uint64
	CurrentHeight() uint64
}

type Subscriptions struct {
	types map[string]map[string]Sub
	l     sync.RWMutex
}

func NewSubscriptions() *Subscriptions {
	return &Subscriptions{
		types: make(map[string]map[string]Sub),
	}
}

func (s *Subscriptions) Add(ev string, sub Sub) error {
	s.l.Lock()
	defer s.l.Unlock()
	t, ok := s.types[ev]
	if !ok {
		t = make(map[string]Sub)
	}
	t[sub.ID()] = sub
	s.types[ev] = t

	return nil
}

func (s *Subscriptions) Remove(id string) error {
	s.l.Lock()
	defer s.l.Unlock()

	for _, t := range s.types {
		delete(t, id)
	}
	return nil
}

func (s *Subscriptions) Send(ctx context.Context, event string, resp jsonrpc.Response) error {
	s.l.RLock()
	defer s.l.RUnlock()
	for _, h := range s.types[event] {
		select {
		case <-ctx.Done():
			return nil
		default:
			h.Send(resp)
		}
	}
	return nil
}

func NewSubI(jsonrpc.Response) Sub {
	return &SubscriptionInstance{}
}

type SubscriptionInstance struct {
	id string

	from    uint64
	current uint64
}

func (si *SubscriptionInstance) Send(jsonrpc.Response) error {

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
