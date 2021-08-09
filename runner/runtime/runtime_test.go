package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/figment-networks/graph-demo/cmd/common/logger"
	"github.com/figment-networks/graph-demo/connectivity"
	runnerClient "github.com/figment-networks/graph-demo/runner/client"
	"github.com/figment-networks/graph-demo/runner/requester"
	"github.com/figment-networks/graph-demo/runner/schema"
	"github.com/figment-networks/graph-demo/runner/store/memap"
	"github.com/figment-networks/graph-demo/runner/structs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLoader_LoadJS(t *testing.T) {
	type connectArgs struct {
		managerURL string
		subs       []structs.Subs
	}
	type storeArgs struct {
		str   string
		data  map[string]interface{}
		key   string
		value string
	}
	type args struct {
		subgraphName string
		schema       string
		connect      connectArgs
		store        storeArgs
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "a",
			args: args{
				"test",
				"../../subgraphs/test",
				connectArgs{
					managerURL: "ws://localhost:8085",
					subs: []structs.Subs{
						{
							Name:           "newTransaction",
							StartingHeight: 5200244,
						},
						{
							Name:           "newBlock",
							StartingHeight: 5200244,
						},
					},
				},
				storeArgs{
					"Block",
					map[string]interface{}{
						"id":      "qazxsw23edcvfr45tgbnhyujm",
						"network": "testNetwork",
						"height":  1234,
						"time":    time.Now().String(),
						"mynote":  "sugar",
					},
					"id",
					"qazxsw23edcvfr45tgbnhyujm",
				},
			},
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			logger.Init("console", "debug", []string{"stderr"})
			defer logger.Sync()
			log := logger.GetLogger()

			rqstr := requester.NewRqstr()
			wst := &networkGraphWSTransportMock{}
			wst.On("Connect", mock.AnythingOfType("*context.emptyCtx"), tt.args.connect.managerURL, mock.AnythingOfType("*client.NetworkGraphClient")).Return(nil)
			wst.On("Subscribe", mock.AnythingOfType("*context.emptyCtx"), tt.args.connect.subs).Return(nil)

			rqstr.AddDestination("testNetwork", wst)

			sStore := memap.NewSubgraphStore()
			loader := NewLoader(log, rqstr, sStore)

			ngc := runnerClient.NewNetworkGraphClient(log, loader)

			require.Nil(t, wst.Connect(ctx, tt.args.connect.managerURL, ngc))

			schemas := schema.NewSchemas(sStore, loader, rqstr)
			require.Nil(t, schemas.LoadFromSubgraphYaml(tt.args.schema))

			require.Nil(t, sStore.Store(ctx, tt.args.subgraphName, tt.args.store.str, tt.args.store.data))

			st, err := sStore.Get(ctx, tt.args.subgraphName, tt.args.store.str, tt.args.store.key, tt.args.store.value)
			require.Nil(t, err)
			require.NotNil(t, st)

			expected := []map[string]interface{}{tt.args.store.data}
			assert.Equal(t, expected, st)

		})
	}
}

type networkGraphWSTransportMock struct {
	mock.Mock
}

func (ng *networkGraphWSTransportMock) Connect(ctx context.Context, address string, RH connectivity.FunctionCallHandler) (err error) {
	args := ng.Called(ctx, address, RH)
	return args.Error(0)
}

func (ng *networkGraphWSTransportMock) CallGQL(ctx context.Context, name string, query string, variables map[string]interface{}, version string) ([]byte, error) {
	args := ng.Called(ctx, name, query, variables, version)
	return args.Get(0).([]byte), args.Error(1)
}

func (ng *networkGraphWSTransportMock) Subscribe(ctx context.Context, events []structs.Subs) error {
	return ng.Called(ctx, events).Error(0)
}

func (ng *networkGraphWSTransportMock) Unsubscribe(ctx context.Context, events []string) error {
	return ng.Called(ctx, events).Error(0)
}
