package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

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
	Data   json.RawMessage   `json:"data"`
	Errors []json.RawMessage `json:"errors,omitempty"`
}

func AttachMux(mux http.ServeMux) {
	mux.HandleFunc("/subgraph", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		resp := JSONGraphQLResponse{}

		ct := r.Header.Get("Content-Type")
		if ct != "" && !strings.Contains(ct, "json") {
			w.WriteHeader(http.StatusNotAcceptable)
			resp.Errors = append(resp.Errors, json.RawMessage([]byte("wrong contenttype")))
			enc.Encode(resp)
			return
		}

		dec := json.NewDecoder(r.Body)
		req := &JSONGraphQLRequest{}
		dec.Decode(req)

		w.WriteHeader(http.StatusOK)
		resp.Data = json.RawMessage([]byte("ok"))
		enc.Encode(resp)
	})
}
