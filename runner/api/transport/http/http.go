package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type API interface {
	ProcessGraphqlQuery(ctx context.Context, v map[string]interface{}, q string) ([]byte, error)
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
	api API
}

func NewHandler(api API) *Handler {
	return &Handler{
		api: api,
	}
}

func (h *Handler) AttachMux(mux *http.ServeMux) {
	mux.HandleFunc("/subgraph", h.HandleGraphql)
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := h.api.ProcessGraphqlQuery(ctx, req.Variables, req.Query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp.Errors = []errorMessage{{Message: err.Error()}}
		enc.Encode(resp)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp.Errors = []errorMessage{{Message: fmt.Sprintf("Error while writing response: %w", err)}}
		enc.Encode(resp)
		return
	}

	return
}
