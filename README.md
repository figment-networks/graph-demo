# Manager - Worker Graph Demo

This codebase contains a proof-of-concept for integrating Figment's manager & worker pattern into The Graph ecosystem.

## Major Components

- connectivity - reusable connectivity protocols module (websocket / jsonrpc)
- cosmos-worker - worker process for fetching data from the Cosmos network
- graphcall
- manager - orchestrates worker process(es), interface to data store, and exposes API for querying Network Graph
- runner - a mock of `graph-node`'s WASM mapping runtime
- subgraphs - contains sample subgraphs for this demo

## Setup

The whole stack can be run with Docker Compose. To get started:

```sh
# This assumes a cosmoshub node is running on host.docker.internal:9090
docker-compose up
```
