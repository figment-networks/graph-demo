package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/figment-networks/graph-demo/runner/api/service"
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

type Handler struct {
	service service.Service
}

func (h *Handler) HandleGraphql(w http.ResponseWriter, r *http.Request) {
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

	response, err := h.service.ProcessGraphqlQuery(req.Variables, req.Query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp.Errors = []errorMessage{{Message: err.Error()}}
		enc.Encode(resp)
		return
	}

	w.WriteHeader(http.StatusOK)

	rawData, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp.Errors = []errorMessage{{Message: fmt.Sprintf("Error while marshalling response: %w", err)}}
		enc.Encode(resp)
		return
	}

	resp.Data = rawData
	enc.Encode(resp)
	return
}

func (h *Handler) AttachMux(mux *http.ServeMux) {
	mux.HandleFunc("/subgraph", h.HandleGraphql)
}
