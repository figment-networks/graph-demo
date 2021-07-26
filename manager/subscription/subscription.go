package subscription

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"go.uber.org/zap"
)

type Sub interface {
	Send(ctx context.Context, height uint64, resp json.RawMessage) error

	ID() string

	FromHeight() uint64
	CurrentHeight() uint64
}

type Evt struct {
	EvType string
	Height uint64
	Data   interface{}
}

type Handle struct {
	in        chan Evt
	l         sync.RWMutex
	log       *zap.Logger
	endpoints map[string]Sub
	finish    chan struct{}
}

func NewHandle() *Handle {
	return &Handle{
		endpoints: make(map[string]Sub),
		finish:    make(chan struct{}),
		in:        make(chan Evt),
	}
}

func (h *Handle) AddEndpoint(s Sub) {
	h.l.Lock()
	defer h.l.Unlock()
	h.endpoints[s.ID()] = s
}

func (h *Handle) RemoveEndpoint(id string) {
	h.l.Lock()
	defer h.l.Unlock()
	delete(h.endpoints, id)
}

func (h *Handle) Send(ctx context.Context, ev Evt) error {
	select {
	case <-ctx.Done():
		return errors.New("error sending event context done")
	case h.in <- ev:
	}
	return nil
}

// fan out event to all subscribers
func (h *Handle) Run(ctx context.Context) {
	for {
		select {
		case <-h.finish:
			return
		case <-ctx.Done():
			return
		case evt := <-h.in:
			mD, err := json.Marshal(evt.Data)
			if err != nil {
				h.log.Error("error marshaing response", zap.Any("data", evt.Data))
				continue
			}
			h.l.RLock()
			for _, sub := range h.endpoints {
				select {
				case <-ctx.Done():
					h.l.RUnlock()
					return
				default:
					sub.Send(ctx, evt.Height, mD)
				}
			}
			h.l.RUnlock()
		}
	}
}

type Subscriptions struct {
	types map[string]*Handle
	l     sync.RWMutex
}

func NewSubscriptions() *Subscriptions {
	return &Subscriptions{
		types: make(map[string]*Handle),
	}
}

// PopulateEvent - We populate events using heights, only to indicate a point time.
// It might be something else in different networks
func (s *Subscriptions) PopulateEvent(ctx context.Context, evType string, height uint64, data interface{}) error {
	s.l.RLock()
	defer s.l.RUnlock()

	t, ok := s.types[evType]
	if !ok { // noone is subscribed
		return nil
	}

	return t.Send(ctx, Evt{EvType: evType, Height: height, Data: data})
}

func (s *Subscriptions) Add(ev string, sub Sub) error {
	s.l.Lock()
	defer s.l.Unlock()
	t, ok := s.types[ev]
	if !ok {
		t = NewHandle()
	}
	t.AddEndpoint(sub)
	s.types[ev] = t

	return nil
}

func (s *Subscriptions) Remove(id string) error {
	s.l.Lock()
	defer s.l.Unlock()
	for _, t := range s.types {
		t.RemoveEndpoint(id)
	}
	return nil
}
