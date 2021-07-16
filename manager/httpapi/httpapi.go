package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/figment-networks/graph-demo/runner/api/service"
)

type Handler struct {
	service *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{
		service: svc,
	}
}

func (h *Handler) AttachMux(mux *http.ServeMux) {
	mux.HandleFunc("/getBlock", h.HandleGetBlock)
}

func (h *Handler) HandleGetBlock(w http.ResponseWriter, r *http.Request) {
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

	response, err := h.service.ProcessGraphqlQuery(ctx, req.Variables, req.Query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp.Errors = []errorMessage{{Message: err.Error()}}
		enc.Encode(resp)
		return
	}

	w.WriteHeader(http.StatusOK)

	resp.Data = response
	if err = enc.Encode(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp.Errors = []errorMessage{{Message: fmt.Sprintf("Error while encoding response: %w", err)}}
		enc.Encode(resp)
		return
	}

	return
}