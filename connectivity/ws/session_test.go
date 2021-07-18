package ws

import (
	"context"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestSession(t *testing.T) {
	type args struct {
		ctx context.Context
		l   *zap.Logger
	}
	tests := []struct {
		name string
		args args
		want *Session
	}{

		{
			name: "test",
			args: args{
				ctx: context.Background(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logr := zaptest.NewLogger(t)

			/*	AReg := jsonrpc.NewRegistryHandler()
				A := NewSession(tt.args.ctx, tt.args.c, logr, AReg)

				BReg := jsonrpc.NewRegistryHandler()
				B := NewSession(tt.args.ctx, tt.args.c, logr, BReg)
			*/
		})
	}
}
