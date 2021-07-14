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

		queries, err := graphql.ParseQuery(req.Query, req.Variables)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			resp.Errors = []errorMessage{{Message: fmt.Sprintf("Error while parsing graphql query: %s", err)}}
			enc.Encode(resp)
			return
		}

		if err := fetchData(&queries); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			resp.Errors = []errorMessage{{Message: fmt.Sprintf("Error while fetching data: %w", err)}}
			enc.Encode(resp)
			return
		}

		rawResp, err := graphql.MapQueryToResponse(queries.Queries)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			resp.Errors = []errorMessage{{Message: err.Error()}}
			enc.Encode(resp)
			return
		}

		w.WriteHeader(http.StatusOK)

		raw, _ := json.Marshal(rawResp)
		resp.Data = raw
		enc.Encode(resp)
		return
	})
}

func fetchData(q *graphql.GraphQuery) error {
	for _, q := range q.Queries {
		fmt.Println(q)
	}

	return nil
}
