
import { callGQL, BlockResponse, NewBlockEvent, storeRecord, printA, Network, GraphQLResponse } from "../graph";

/***
 * Generated
 * Definitions here would be auto-generated by `graph codegen`
 */
export class BlockEntity {
  id: string;
  height: number;
  time: Date;
  myNote: string;

  constructor(...args: any[]) {
    Object.assign(this, args);
  }
}

function query<T>(network: Network, query: string, args: {}): GraphQLResponse<T> {
  const stringResponse = callGQL(network, query, args);
  return JSON.parse(stringResponse);
}

/**
 * Mapping
 */
const GET_BLOCK = `query GetBlock($height: Int) {
  block( $height: Int = 0 ) {
    height
    time
    id
  }
}`;

/**
 * This would be defined in the subgraph.yaml.
 * 
 * Ex:
 * ```yaml
 * # ...
 * mapping:
 *  kind: cosmos/events
 *  apiVersion: 0.0.1
 *  language: wasm/assemblyscript
 *  blockHandlers:
 *    - function: handleNewBlock
 * ```
 */
function handleNewBlock(newBlockEvent: NewBlockEvent) {
  printA('newEventData: ' + JSON.stringify(newBlockEvent));

  const response = query<BlockResponse>(newBlockEvent.network, GET_BLOCK, { height: newBlockEvent.height });

  printA('GQL call raw response: ' + JSON.stringify(response));

  const {error, data} = response;

  if (error) {
    printA('GQL call error: ' + JSON.stringify(error));
    return;
  }

  if (!data) {
    printA('GQL call returned no data');
    return;
  }

  printA('graphQL response: ' + JSON.stringify(data));

  const { height, id, time } = data;
  const entity = new BlockEntity(height, id, time, "ok");

  printA('entity: ' + JSON.stringify(entity));

  // replace with `entity.save()` for graph-ts
  storeRecord("SubgraphStoreBlock", entity);
}
