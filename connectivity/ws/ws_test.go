package ws

import (
	"context"
	"net/http"
	"testing"

	"github.com/figment-networks/graph-demo/connectivity"
	"github.com/figment-networks/graph-demo/connectivity/jsonrpc"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestConn(t *testing.T) {
	type fields struct {
		RH connectivity.FunctionCallHandler
		l  *zap.Logger
	}
	type args struct {
		ctx context.Context
		mux *http.ServeMux
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			logr := zaptest.NewLogger(t)

			AReg := jsonrpc.NewRegistryHandler()
			cA := &Conn{
				RH: AReg,
				l:  logr,
			}
			cA.AttachToMux(tt.args.ctx, tt.args.mux)

			BReg := jsonrpc.NewRegistryHandler()
			cB := &Conn{
				RH: BReg,
				l:  logr,
			}
			cB.AttachToMux(tt.args.ctx, tt.args.mux)

		})
	}
}
