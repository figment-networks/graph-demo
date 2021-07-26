package httpapi

import (
	"context"
	"net/http"

	"github.com/figment-networks/graph-demo/manager/structs"
)

type ManagerService interface {
	StoreBlock(ctx context.Context, block structs.Block) error
	StoreTransactions(ctx context.Context, txs []structs.Transaction) error
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
}

func (h *Handler) HandleStoreTransaction(w http.ResponseWriter, r *http.Request) {
}

func (h *Handler) HandleStore(w http.ResponseWriter, r *http.Request) {
	/*

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

		response, err := h.service.ProcessGraphqlQuery(ctx, req.Variables, req.Query)
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
			resp.Errors = []ErrorMessage{{Message: fmt.Sprintf("Error while encoding response: %w", err)}}
			enc.Encode(resp)
			return
		}

		return*/
}
