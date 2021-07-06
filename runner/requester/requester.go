package requester

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"sync"
)

type Rqstr struct {
	c *http.Client

	list  map[string]Destination
	llock sync.RWMutex
}

type Destination struct {
	Name    string
	Kind    string // http, ws etc
	Address string
}

func NewRqstr(c *http.Client) *Rqstr {
	return &Rqstr{
		c:    c,
		list: make(map[string]Destination),
	}
}
func (r *Rqstr) AddDestination(dest Destination) {
	r.llock.Lock()
	r.list[dest.Name] = dest
	r.llock.Unlock()
}

type GQLPayload struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
	//	OperationName string                 `json:"operationName"`
}

type GQLResponse struct {
	Data   interface{}   `json:"data"`
	Errors []interface{} `json:"errors"`
}

func (r *Rqstr) CallGQL(ctx context.Context, name string, query string, variables map[string]interface{}) ([]byte, error) {
	r.llock.RLock()
	d, ok := r.list[name]
	r.llock.RUnlock()
	if !ok {
		return nil, errors.New("graph not found")
	}

	buff := new(bytes.Buffer)
	defer buff.Reset()
	enc := json.NewEncoder(buff)
	if err := enc.Encode(GQLPayload{query, variables}); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.Address, buff)
	if err != nil {
		return nil, err
	}
	resp, err := r.c.Do(req)
	if err != nil {
		return nil, err
	}

	respD, err := ioutil.ReadAll(resp.Body)
	return respD, err
}
