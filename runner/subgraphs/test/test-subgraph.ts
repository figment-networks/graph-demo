
import { callGQL, BlockResponse, NewBlockEvent, storeRecord, logInfo } from "../graph";

class BlockEntity {
  id: string;
  height: number;
  time: Date;
  myNote: string;

  constructor(height: number, id: string, time: Date, myNote: string) {
    this.id = id;
    this.height = height;
    this.time = time;
    this.myNote = myNote;
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

    storeRecord("StoreBlock", entity);
}
