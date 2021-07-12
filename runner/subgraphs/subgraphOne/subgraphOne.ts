
import { callGQL, BlockResponse, NewBlockEvent, storeRecord, logInfo } from "../graph";
import { BlockEntity } from './entities';

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
    const { data, error } = callGQL<BlockResponse>(newBlockEvent.network, GET_BLOCK, { height: newBlockEvent.height });

    if (error) {
      logInfo(JSON.stringify(error));
      return;
    }

    logInfo(JSON.stringify(data));

    const { height, id, time } = data;
    const entity = new BlockEntity(height, id, time, "ok");

    // log.info
    logInfo(JSON.stringify(entity));

    // replace with `entity.save()` for graph-ts
    storeRecord("SubgraphStoreBlock", entity);
}
