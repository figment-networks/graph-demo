package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/figment-networks/graph-demo/cosmos-worker/api"
	"github.com/figment-networks/graph-demo/cosmos-worker/structs"
)

type Handler struct {
	client *api.Client
}

func New(c *api.Client) *Handler {
	return &Handler{
		client: c,
	}
}

func (h *Handler) AttachToMux(mux *http.ServeMux) {
	mux.HandleFunc("/get_block/", h.HandleGetBlock)
}

func (h *Handler) HandleGetBlock(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	heightStr := strings.Trim(r.URL.Path, "/get_block/")

	heightInt, err := strconv.Atoi(heightStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error while parsing block height: %w", err)))
		return
	}

	block, err := h.client.GetBlock(ctx, uint64(heightInt))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error while getting a block: %w", err)))
		return
	}

	txs, err := h.client.SearchTx(ctx, block, uint64(heightInt), 100)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error while getting transactions: %w", err)))
		return
	}

	resp, err := json.Marshal(structs.GetBlockResp{
		Block: block,
		Txs:   txs,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error while marshalling response: %w", err)))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
