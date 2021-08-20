# Runner

This runtime is meant to mimic the behaviors of [graph-node](https://github.com/graphprotocol/graph-node) which executes wasm subgraph mapping code. Ultimately, this code may be adapted into graph-node.

Our runner uses the Go V8 runtime to run javascript.

## Directory Structure

- `api` - APIs to interact with the runner
- `runtime` - creates the V8 runtime and executes the javascript
- `requester` - layer to make GraphQL queries to the manager
- `schema` - model for loading graphQL schemas from subgraphs
- `store` - in-memory store for subgraph data
