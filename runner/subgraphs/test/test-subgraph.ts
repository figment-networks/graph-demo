
import { callGQL, BlockResponse, NewBlockEvent, storeRecord, logInfo } from "../graph";

export class BlockEntity {
  id: string;
  height: number;
  time: Date;
  myNote: string;

  constructor(...args: any[]) {
    Object.assign(this, args);
  }
}

const GET_BLOCK = `query GetBlock($height: Int) {
    block( $height: Int = 0 ) {
      height
      time
      id
    }
  }`;

function handleNewBlock(newBlockEvent: NewBlockEvent) {
    const { data, error } = callGQL<BlockResponse>(newBlockEvent.network, GET_BLOCK, { height: newBlockEvent.height });

    if (error) {
      logInfo(JSON.stringify(error));
      return;
    }

    logInfo(JSON.stringify(data));

    const { height, id, time } = data;
    const entity = new BlockEntity(height, id, time, "ok");

    logInfo(JSON.stringify(entity));

    storeRecord("SubgraphStoreBlock", entity);
}
