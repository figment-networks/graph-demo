package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/figment-networks/graph-demo/cosmos-worker/client"
	"github.com/figment-networks/graph-demo/manager/structs"
)

type GetBlockResp struct {
	Block structs.Block         `json:"block"`
	Txs   []structs.Transaction `json:"txs"`
}

type GetLastResp struct {
	LastHeight uint64 `json:"last_height"`
}

type Handler struct {
	client *client.Client
}

func NewHandler(c *client.Client) *Handler {
	return &Handler{
		client: c,
	}
}

func (h *Handler) AttachToMux(mux *http.ServeMux) {
	mux.HandleFunc("/getAll/", h.HandleGetAll)
	mux.HandleFunc("/getLast/", h.HandleGetLast)
}

func (h *Handler) HandleGetAll(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
<<<<<<< HEAD
	heightStr := strings.Trim(r.URL.Path, "/getAll/")
=======
	heightStr := strings.Trim(r.URL.Path, "/getBlock/")
>>>>>>> bca17a11c51b4e4f4f8d47ff80093a7fdd74ec7b

	heightInt, err := strconv.Atoi(heightStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
<<<<<<< HEAD
		w.Write([]byte(fmt.Sprintf("Error while parsing block height: %s", err.Error())))
		return
	}

	blockAndTx, err := h.client.GetBlock(ctx, int64(heightInt))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error while getting a block: %s", err.Error())))
=======
		w.Write([]byte(fmt.Sprintf("Error while parsing block height: %w", err)))
		return
	}

	blockAndTx, err := h.client.GetBlock(ctx, uint64(heightInt))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error while getting a block: %w", err)))
>>>>>>> bca17a11c51b4e4f4f8d47ff80093a7fdd74ec7b
		return
	}

	resp, err := json.Marshal(GetBlockResp{
		Block: blockAndTx.Block,
		Txs:   blockAndTx.Transactions,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
<<<<<<< HEAD
		w.Write([]byte(fmt.Sprintf("Error while marshalling response: %s", err.Error())))
=======
		w.Write([]byte(fmt.Sprintf("Error while marshalling response: %w", err)))
>>>>>>> bca17a11c51b4e4f4f8d47ff80093a7fdd74ec7b
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *Handler) HandleGetLast(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()

	blockAndTx, err := h.client.GetLatest(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
<<<<<<< HEAD
		w.Write([]byte(fmt.Sprintf("Error while getting a block: %s", err.Error())))
=======
		w.Write([]byte(fmt.Sprintf("Error while getting a block: %w", err)))
>>>>>>> bca17a11c51b4e4f4f8d47ff80093a7fdd74ec7b
		return
	}

	resp, err := json.Marshal(GetBlockResp{
		Block: blockAndTx.Block,
		Txs:   blockAndTx.Transactions,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
<<<<<<< HEAD
		w.Write([]byte(fmt.Sprintf("Error while marshalling response: %s", err.Error())))
=======
		w.Write([]byte(fmt.Sprintf("Error while marshalling response: %w", err)))
>>>>>>> bca17a11c51b4e4f4f8d47ff80093a7fdd74ec7b
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
