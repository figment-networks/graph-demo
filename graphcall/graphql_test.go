package graphcall_test

import (
	"reflect"
	"testing"

	"github.com/figment-networks/graph-demo/graphcall"
)

var t1 = []byte(`query GetBlock($height: Int = 6940033) {
	block(height: 6940034) {
	  height
	  time
	  id
	  unknownField
	}
	getTransactions(height: $height) {
	  hash
	  time
	  id
	  unknownField
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
  }`)

func TestParseQuery(t *testing.T) {
	type args struct {
		query     []byte
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
		{name: "empty variables", args: args{
			query:     t1,
			variables: nil,
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

var schemaT1 = []byte(`type Block @entity {
	id: ID!
	height: Int!
	time: String!
	myNote: String
	transactions: [Transaction]
  }

  type Transaction @entity {
	id: ID!
	blockID: Int!
	height: Int!
	time: String!
	myNote: String!
  }`)

func TestParseSchema(t *testing.T) {
	type args struct {
		query []byte
	}
	tests := []struct {
		name    string
		args    args
		want    graphcall.GraphQuery
		wantErr bool
	}{
		{name: "simple", args: args{
			query: schemaT1,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := graphcall.ParseSchema(tt.name, tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
