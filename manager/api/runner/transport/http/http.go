package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type ManagerService interface {
	ProcessGraphqlQuery(ctx context.Context, q []byte, v map[string]interface{}) ([]byte, error)
}

type JSONGraphQLRequest struct {
	Query     []byte                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type JSONGraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []ErrorMessage  `json:"errors,omitempty"`
}

type ErrorMessage struct {
	Message string `json:"message,omitempty"`
}

type Handler struct {
	service ManagerService
}

func NewHandler(svc ManagerService) *Handler {
	return &Handler{
		service: svc,
	}
}

func (h *Handler) AttachMux(mux *http.ServeMux) {
	mux.HandleFunc("/graphQL", h.HandleRequest)
}

func (h *Handler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	resp := JSONGraphQLResponse{}

	ct := r.Header.Get("Content-Type")
	if ct != "" && !strings.Contains(ct, "json") {
		w.WriteHeader(http.StatusNotAcceptable)
		resp.Errors = []ErrorMessage{{Message: "wrong content type"}}
		enc.Encode(resp)
		return
	}

	dec := json.NewDecoder(r.Body)
	req := &JSONGraphQLRequest{}
	dec.Decode(req)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := h.service.ProcessGraphqlQuery(ctx, req.Query, req.Variables)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp.Errors = []ErrorMessage{{Message: err.Error()}}
		enc.Encode(resp)
		return
	}

	w.WriteHeader(http.StatusOK)

	resp.Data = response
	if err = enc.Encode(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp.Errors = []ErrorMessage{{Message: fmt.Errorf("Error while encoding response: %w", err).Error()}}
		enc.Encode(resp)
		return
	}

	return
}
