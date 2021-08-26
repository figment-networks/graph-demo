# Manager

The manager orchestrates the workers and also handles requests from runner (i.e. a mock `graph-node`).

## Directory Structure

- `api` - API interface for manager  and runner to communicate to manager
- `client` - client interface to communicate with workers
- `scheduler` - triggers internal events to start the process of orchestrating workers to consume new data from the networks
- `store` - the data store interface
- `structs` - contains the data structures for the data stored in the store
- `subscription` - subscription abstraction on runner connection
