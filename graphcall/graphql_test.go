package graphcall_test

import (
	"testing"

	"github.com/figment-networks/graph-demo/graphcall"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	t1 = []byte(`query GetBlock($height: Int = 6000000) {
		block(height: 7000000) {
		  height
		  time
		  id
		  unknownField
		}
		getTransactions(height: $height) {
		  hash
		  time
		  id
		}
		getBlockHashAndTransactionHashes(height: $height) {
		  block {
			height
			time
			hash
			id
		  }
		  transactions {
			hash
			time
		  }
		}
	  }`)

	t2 = []byte(`query GetBlock2($height: Int = 6000000) {
		block(height: 7000000) {
		  hash
		  time
		  id
		}
	}`)
)

func TestParseQuery(t *testing.T) {
	type args struct {
		query     []byte
		variables map[string]interface{}
	}
	tests := []struct {
		name       string
		args       args
		graphQuery graphcall.GraphQuery
		err        error
	}{
		{
			name: "simple",
			args: args{
				query:     t1,
				variables: map[string]interface{}{"height": float64(120)},
			},
			graphQuery: graphcall.GraphQuery{
				Q: graphcall.Part{
					Name: "GetBlock",
					Params: map[string]graphcall.Param{
						"height": {
							Field:    "height",
							Type:     "Int",
							Variable: "uint64",
							Value:    uint64(120),
						},
					},
				},
				Queries: []graphcall.Query{
					{
						Name:  "block",
						Order: 0,
						Params: map[string]graphcall.Part{
							"height": {
								Name: "height",
								Params: map[string]graphcall.Param{
									"height": {
										Field:    "height",
										Type:     "Int",
										Variable: "uint64",
										Value:    uint64(7000000),
									},
								},
							},
						},
						Fields: map[string]graphcall.Field{
							"height": {
								Name:  "height",
								Order: 0,
							},
							"time": {
								Name:  "time",
								Order: 1,
							},
							"id": {
								Name:  "id",
								Order: 2,
							},
							"unknownfield": {
								Name:  "unknownField",
								Order: 3,
							},
						},
					},
					{
						Name:  "getTransactions",
						Order: 1,
						Params: map[string]graphcall.Part{
							"height": {
								Name: "height",
								Params: map[string]graphcall.Param{
									"height": {
										Field:    "height",
										Type:     "Int",
										Variable: "uint64",
										Value:    uint64(120),
									},
								},
							},
						},
						Fields: map[string]graphcall.Field{
							"hash": {
								Name:  "hash",
								Order: 0,
							},
							"time": {
								Name:  "time",
								Order: 1,
							},
							"id": {
								Name:  "id",
								Order: 2,
							},
						},
					},
					{
						Name:  "getBlockHashAndTransactionHashes",
						Order: 2,
						Params: map[string]graphcall.Part{
							"height": {
								Name: "height",
								Params: map[string]graphcall.Param{
									"height": {
										Field:    "height",
										Type:     "Int",
										Variable: "uint64",
										Value:    uint64(120),
									},
								},
							},
						},
						Fields: map[string]graphcall.Field{
							"block": {
								Name:  "block",
								Order: 0,
								Fields: map[string]graphcall.Field{
									"height": {
										Name:  "height",
										Order: 0,
									},
									"time": {
										Name:  "time",
										Order: 1,
									},
									"hash": {
										Name:  "hash",
										Order: 2,
									},
									"id": {
										Name:  "id",
										Order: 3,
									},
								},
							},
							"transactions": {
								Name:  "transactions",
								Order: 1,
								Fields: map[string]graphcall.Field{
									"hash": {
										Name:  "hash",
										Order: 0,
									},
									"time": {
										Name:  "time",
										Order: 1,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "empty variables",
			args: args{
				query:     t2,
				variables: nil,
			},
			graphQuery: graphcall.GraphQuery{
				Q: graphcall.Part{
					Name: "GetBlock2",
					Params: map[string]graphcall.Param{
						"height": {
							Field:    "height",
							Type:     "Int",
							Variable: "uint64",
							Value:    uint64(6000000),
						},
					},
				},
				Queries: []graphcall.Query{
					{
						Name:  "block",
						Order: 0,
						Params: map[string]graphcall.Part{
							"height": {
								Name: "height",
								Params: map[string]graphcall.Param{
									"height": {
										Field:    "height",
										Type:     "Int",
										Variable: "uint64",
										Value:    uint64(7000000),
									},
								},
							},
						},
						Fields: map[string]graphcall.Field{
							"hash": {
								Name:  "hash",
								Order: 0,
							},
							"time": {
								Name:  "time",
								Order: 1,
							},
							"id": {
								Name:  "id",
								Order: 2,
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graphQuery, err := graphcall.ParseQuery(tt.args.query, tt.args.variables)

			require.Equal(t, tt.err, err)
			assert.Equal(t, tt.graphQuery, graphQuery)
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
		name      string
		args      args
		subgrapgh *graphcall.Subgraph
		err       error
	}{
		{
			name: "simple",
			args: args{
				query: schemaT1,
			},
			subgrapgh: &graphcall.Subgraph{
				Name: "simple",
				Entities: map[string]*graphcall.Entity{
					"Block": {
						Name: "Block",
						Fields: map[string]graphcall.Fields{
							"height": {
								Name:    "height",
								Type:    "Int",
								IsArray: false,
								NotNull: true,
							},
							"id": {
								Name:    "id",
								Type:    "ID",
								IsArray: false,
								NotNull: true,
							},
							"mynote": {
								Name:    "myNote",
								Type:    "String",
								IsArray: false,
								NotNull: false,
							},
							"time": {
								Name:    "time",
								Type:    "String",
								IsArray: false,
								NotNull: true,
							},
							"transactions": {
								Name:    "transactions",
								Type:    "Transaction",
								IsArray: true,
								NotNull: true,
							},
						},
					},
					"Transaction": {
						Name: "Transaction",
						Fields: map[string]graphcall.Fields{
							"blockid": {
								Name:    "blockID",
								Type:    "Int",
								IsArray: false,
								NotNull: true,
							},
							"height": {
								Name:    "height",
								Type:    "Int",
								IsArray: false,
								NotNull: true,
							},
							"id": {
								Name:    "id",
								Type:    "ID",
								IsArray: false,
								NotNull: true,
							},
							"mynote": {
								Name:    "myNote",
								Type:    "String",
								IsArray: false,
								NotNull: true,
							},
							"time": {
								Name:    "time",
								Type:    "String",
								IsArray: false,
								NotNull: true,
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subgrapgh, err := graphcall.ParseSchema(tt.name, tt.args.query)
			require.Equal(t, tt.err, err)
			assert.Equal(t, tt.subgrapgh, subgrapgh)
		})
	}
}
