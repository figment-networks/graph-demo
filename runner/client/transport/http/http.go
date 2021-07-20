package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type GQLPayload struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
	//	OperationName string                 `json:"operationName"`
}

type GQLResponse struct {
	Data   interface{}   `json:"data"`
	Errors []interface{} `json:"errors"`
}

type NetworkGraphHTTPTransport struct {
	c       *http.Client
	address string
}

func NewNetworkGraphHTTPTransport(address string, c *http.Client) *NetworkGraphHTTPTransport {
	return &NetworkGraphHTTPTransport{
		address: address,
		c:       c,
	}
}

func (ng *NetworkGraphHTTPTransport) CallGQL(ctx context.Context, name string, query string, variables map[string]interface{}, version string) ([]byte, error) {
	buff := new(bytes.Buffer)
	defer buff.Reset()
	enc := json.NewEncoder(buff)
	if err := enc.Encode(GQLPayload{query, variables}); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ng.address, buff)
	if err != nil {
		return nil, err
	}
	resp, err := ng.c.Do(req)
	if err != nil {
		return nil, err
	}

	respD, err := ioutil.ReadAll(resp.Body)
	return respD, err
}
