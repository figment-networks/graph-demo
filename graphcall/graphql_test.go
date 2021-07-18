package graphcall_test

import (
	"reflect"
	"testing"

	"github.com/figment-networks/graph-demo/graphcall"
)

const t1 = `query GetBlock($height: Int = 6940033) {
	block(height: 6940034) {
	  height
	  time
	  id
	  unknwonField
	}
	getTransactions(height: $height) {
	  hash
	  time
	  id
	  unknwonField
	}
	getBlockHashAndTransactionHashes(height: $height) {
	  block {
		height
		time
		hash
		id
	  }
	  unknownField
	  txs {
		hash
		time
	  }
	}
  }`

func TestParseQuery(t *testing.T) {
	type args struct {
		query     string
		variables map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    graphcall.GraphQuery
		wantErr bool
	}{
		{name: "simple", args: args{
			query:     t1,
			variables: map[string]interface{}{"height": float64(120)},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := graphcall.ParseQuery(tt.args.query, tt.args.variables)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}
