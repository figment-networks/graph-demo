
import { callGQL, BlockResponse, SubgraphNewBlock, storeRecord, printA } from "../graph";

class BlockEntity {
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

// This is called by the mapping runtime when a new block is available to be processed
function handleNewBlock(newBlock: SubgraphNewBlock) {
    const { data, error } = callGQL<BlockResponse>(newBlock.network, GET_BLOCK, { height: newBlock.height });
    if (error) {
      printA(JSON.stringify(error));
      return;
    }
    printA(JSON.stringify(data));

    const { height, id, time } = data;
    const entity = new BlockEntity(height, id, time, "ok");

    printA(JSON.stringify(entity));
    storeRecord("SubgraphStoreBlock", entity);
}
