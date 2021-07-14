package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/figment-networks/graph-demo/runner/api/graphql"
	"github.com/figment-networks/graph-demo/runner/store"
)

type API struct {
	s store.Storage
}

type JSONGraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type JSONGraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []errorMessage  `json:"errors,omitempty"`
}

type errorMessage struct {
	Message string `json:"message",omitempty`
}

func AttachMux(mux *http.ServeMux) {
	mux.HandleFunc("/subgraph", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		resp := JSONGraphQLResponse{}

		ct := r.Header.Get("Content-Type")
		if ct != "" && !strings.Contains(ct, "json") {
			w.WriteHeader(http.StatusNotAcceptable)
			resp.Errors = []errorMessage{{Message: "wrong content type"}}
			enc.Encode(resp)
			return
		}

		dec := json.NewDecoder(r.Body)
		req := &JSONGraphQLRequest{}
		dec.Decode(req)
	})
}

func fetchData(q *graphql.GraphQuery) error {
	for _, q := range q.Queries {
		fmt.Println(q)
	}

	return nil
}
