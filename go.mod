module github.com/figment-networks/graph-demo

go 1.16

require (
	github.com/bearcherian/rollzap v1.0.2
	github.com/cosmos/cosmos-sdk v0.42.4
	github.com/figment-networks/indexer-manager v0.4.1
	github.com/figment-networks/indexing-engine v0.4.4
	github.com/gogo/protobuf v1.3.3
	github.com/golang-migrate/migrate/v4 v4.14.1 // indirect
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/graphql-go/graphql v0.7.9 // indirect
	github.com/hasura/go-graphql-client v0.2.0 // indirect
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/klauspost/compress v1.13.1 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.9
	go.uber.org/zap v1.18.1
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e // indirect
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba
	google.golang.org/grpc v1.36.0
	nhooyr.io/websocket v1.8.7 // indirect
	rogchap.com/v8go v0.6.0 // indirect
)

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
