package jsRuntime

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/figment-networks/graph-demo/runner/requester"
	"github.com/figment-networks/graph-demo/runner/schema"
	"github.com/figment-networks/graph-demo/runner/store"
	"github.com/figment-networks/graph-demo/runner/store/memap"
	"github.com/stretchr/testify/require"
)

func TestLoader_LoadJS(t *testing.T) {
	type args struct {
		name   string
		path   string
		schema string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "a",
			args: args{
				"one",
				"../subgraphs/subgraphOne/subgraphOne.js",
				"../subgraphs/subgraphOne/schema.graphql"},
		},
	}
	for _, tt := range tests {

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, `{"data": {"time": "12345677", "height":2345, "id": "qazxsw23edcvfr45tgbnhyujm"} }`)
		}))
		defer ts.Close()

		t.Run(tt.name, func(t *testing.T) {
			cli := &http.Client{}
			rqstr := requester.NewRqstr(cli)
			rqstr.AddDestination(requester.Destination{
				Name:    "networkOne",
				Kind:    "http",
				Address: ts.URL,
			})

			schemas := schema.NewSchemas()
			schemas.LoadFromFile(tt.args.name, tt.args.schema)

			sStore := memap.NewSubgraphStore()

			for _, sg := range schemas.Subgraphs {
				for _, ent := range sg.Entities {
					indexed := []store.NT{}
					for k, v := range ent.Params {
						indexed = append(indexed, store.NT{Name: k, Type: v.Type})
					}
					sStore.NewStore(tt.args.name, ent.Name, indexed)
				}
			}

			l := NewLoader(rqstr, sStore)
			if err := l.LoadJS(tt.args.name, tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("Loader.LoadJS() error = %v, wantErr %v", err, tt.wantErr)
			}

			m := map[string]interface{}{
				"height": 1234,
			}
			if err := l.CallSubgraphHandler(tt.args.name,
				&SubgraphHandler{
					name:   "handleNewBlock",
					values: []interface{}{m},
				}); err != nil {
				t.Errorf("Loader.CallSubgraphHandler() error = %v", err)
			}

			st, err := sStore.Get(context.Background(), tt.args.name, "StoreBlock", "id", "qazxsw23edcvfr45tgbnhyujm")
			if err != nil {
				t.Errorf("mStore.Get error = %v, wantErr %v", err, tt.wantErr)
			}

			require.NotNil(t, st)

		})
	}
}
