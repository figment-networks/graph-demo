
import { graphql, BlockEvent, store, log, TransactionEvent } from "../../graph";

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
 * This function is defined in the subgraph.yaml.
 * 
 * Ex:
 * ```yaml
 * # ...
 * mapping:
 *  kind: cosmos/blocks
 *  apiVersion: 0.0.1
 *  language: wasm/assemblyscript
 *  transactionHandlers:
 *    - function: handleTransaction
 *  blockHandlers:
 *    - function: handleBlock
 * ```
 */
function handleBlock(newBlockEvent: BlockEvent) {
  log.debug('newBlockEvent: ' + JSON.stringify(newBlockEvent));

  const {error, data} = graphql.call("cosmos", GET_BLOCK, { height: newBlockEvent.height }, "0.0.1");

  if (error) {
     log.debug('GQL call error: ' + JSON.stringify(error));
    return;
  }

  if (!data) {
     log.debug('GQL call returned no data');
    return;
  }

   log.debug('GQL call data: ' + JSON.stringify(data));

  const { height, id, time } = data;
  const entity = new BlockEntity(height, id, time, "ok");

  log.debug('Entity: ' + JSON.stringify(entity));

  const {storeErr} = store.save("SubgraphStoreBlock", entity);
  if (storeErr) {
     log.debug('Error storing entity: ' + JSON.stringify(storeErr));
  }
}

function handleTransaction(newTxnEvent: TransactionEvent) {
  log.debug('newTxnEvent: ' + JSON.stringify(newTxnEvent));
}
