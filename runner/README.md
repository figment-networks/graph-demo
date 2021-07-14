# Runner

This runtime is meant to mimic the behaviors of graph-node which executes wasm subgraph mapping code. Instead, we use the Go V8 runtime to run javascript.

## Directories

- `api` - APIs to interact with the runner
- `jsRuntime` - creates the V8 runtime and executes the javascript
- `requester` - layer to make GraphQL queries to the manager
- `schema` - model for loading graphQL schemas from subgraphs
- `store` - in-memory store for subgraph data
- `subgraphs` - sample subgraphs (typescript + graphQL schema)

## Subgraphs

When making changes to a subgraph typescript file, you need to rebuild (e.g. `npm run build:simple-example`) to javascript to reflect changes in the V8 runtime.
