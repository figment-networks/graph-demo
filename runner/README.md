# Runner

This runtime is meant to mimic the behaviors of [graph-node](https://github.com/graphprotocol/graph-node) which executes wasm subgraph mapping code.
Our runner uses the Go V8 runtime to run javascript code.

## Directory Structure

- `api` - APIs to interact with the runner (`http` graphql requests for subgraph data )
- `client` - API to interact with graphql subscription for data
- `requester` - Runtime queries abstraction layer
- `runtime` - creates the V8 runtime and executes the javascript
- `schema` - model for loading graphQL schemas from subgraphs
- `store` - simple unefficient in-memory store for subgraph data (dynamic postgres table generation for graphql entities far exceed the scope of that POC )

## Flow

After start, runner reads given directory to read subgraph configuration.
Afterwards runner starts ws connection and subscribe to data.
When event reaches runner, it triggers respective handler of the subgraph.
Subgraph may trigger additional functions, including call() which triggers graphql query using ws connection to manager.
After processing, subgraph may persist structure into the memory database.
Afterwards, it's possible to use graphql interface to fetch the data, the same way as from real graph-node.
