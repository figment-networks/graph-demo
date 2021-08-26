package httpapi

import (
	"context"
	"net/http"

	"github.com/figment-networks/graph-demo/manager/structs"
)

type ManagerService interface {
	StoreBlock(ctx context.Context, block structs.Block) (string, error)
	StoreTransactions(ctx context.Context, txs []structs.Transaction) ([]string, error)
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
	mux.HandleFunc("/storeBlock", h.HandleStoreBlock)
	mux.HandleFunc("/storeTransaction", h.HandleStoreTransaction)
}

func (h *Handler) HandleStoreBlock(w http.ResponseWriter, r *http.Request) {
	// (lukanus): might be also implemented for http
}

func (h *Handler) HandleStoreTransaction(w http.ResponseWriter, r *http.Request) {
	// (lukanus): might be also implemented for http
}
