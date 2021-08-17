package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/figment-networks/graph-demo/manager/structs"
)

type CosmosClient interface {
	GetAll(ctx context.Context, height uint64) error
	GetLatest(ctx context.Context) (structs.Block, error)
}

type Handler struct {
	client CosmosClient
}

func NewHandler(c CosmosClient) *Handler {
	return &Handler{
		client: c,
	}
}

func (h *Handler) AttachToMux(mux *http.ServeMux) {
	mux.HandleFunc("/getAll/", h.HandleGetAll)
	mux.HandleFunc("/getLast/", h.HandleGetLast)
}

func (h *Handler) HandleGetAll(w http.ResponseWriter, r *http.Request) {
	heightStr := strings.Replace(r.URL.Path, "/getAll/", "", -1)
	heightInt, err := strconv.Atoi(heightStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error while parsing block height: %s", err.Error())))
		return
	}

	if err := h.client.GetAll(r.Context(), uint64(heightInt)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error while getting a block: %s", err.Error())))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ACK"))
}

func (h *Handler) HandleGetLast(w http.ResponseWriter, r *http.Request) {
	block, err := h.client.GetLatest(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error while getting a block: %s", err.Error())))
		return
	}

	resp, err := json.Marshal(block)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error while marshalling response: %s", err.Error())))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
