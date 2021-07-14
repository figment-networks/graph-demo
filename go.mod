module github.com/figment-networks/graph-demo

go 1.16

require (
	github.com/cosmos/cosmos-sdk v0.42.4
	github.com/figment-networks/indexer-manager v0.4.1
	github.com/figment-networks/indexing-engine v0.4.4
	github.com/gogo/protobuf v1.3.3
	github.com/golang/snappy v0.0.3 // indirect
	github.com/google/uuid v1.3.0
	github.com/gorilla/websocket v1.4.2
	github.com/graphql-go/graphql v0.7.9
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.11
	go.uber.org/zap v1.18.1
	golang.org/x/time v0.0.0-20210611083556-38a9dc6acbc6
	google.golang.org/grpc v1.37.0
	rogchap.com/v8go v0.6.0
)

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
