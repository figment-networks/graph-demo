// This test is used by jsRunner_test.go

import { graphql, BlockEvent, store, log } from "../../graph";

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

function handleBlock(newBlockEvent: BlockEvent) {
    const { data, error } = graphql.call(newBlockEvent.network, GET_BLOCK, { height: newBlockEvent.height });

    if (error) {
      log.debug(JSON.stringify(error));
      return;
    }

    log.debug(JSON.stringify(data));

    const { height, id, time } = data;
    const entity = new BlockEntity(height, id, time, "ok");

    log.debug(JSON.stringify(entity));

    store.save("StoreBlock", entity);
}
