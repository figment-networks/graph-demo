# Manager - Worker Graph Demo

This codebase contains a proof-of-concept for integrating Figment's manager & worker pattern into The Graph ecosystem.

The flow covered in the demo presents the unoptimized way of subgraph data processing.

In this flow we can distinct two different data flows:

- Network data ingestion - a process between manager and worker where network data is synced.
- Subgraph generation - a process between runtime and manager that allows runtime to fetch the data.

## Major Components

Keep in mind that for the sake of the demo it presents only basic functions mostly for the "happy-path".

- connectivity - simple reusable connectivity protocols module (websocket / jsonrpc)
- cosmos-worker - worker process for fetching data from the Cosmos network
- graphcall - graphql query parsing package
- manager - orchestrates worker process(es), interface to data store, Network Graph API and subscription interface for subgraph runtime
- runner - a mock of `graph-node`'s WASM mapping runtime based on v8 engine
- subgraphs - contains sample subgraphs for this demo

## Setup

### Docker-compose

The whole stack can be run with Docker Compose. To get started:

The default config assumes a cosmoshub-4 node is running on `localhost:9090`
It is currently defined in docker-compose config as `host.docker.internal:9090` this parameter may not work on linux - in this case you may need to refer to the default gateway or the host address.

```sh
docker-compose build
docker-compose up
```

### The Debug (Visual Studio Code)

If you would like to test service step by step in the debug node. Repository includes `.vscode/launch.json` configuration.
The recommended way to run that would be to comment out service in the the docker-compose.yaml file, run the rest of the system from there and start service using debug mode in VSC UI.
Docker-compose configuration is just using driver overlay/attachable. so it should be able to connect to every service running inside the docker without any problems

### Make process

Project includes `Makefile` that builds almost everything. To do so you need to have the latest version of golang compiler installed.
Then just run `make build all`.

What `make` does not automatically do is generation of javascript files in subgraph, leaving that to the author of .
