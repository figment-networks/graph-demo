
import { callGQL, Block, SubgraphNewBlock, storeRecord, printA } from "../graph";

class SubgraphStoreBlock {
  id: string;
  height: number;
  time: Date;
  myNote: string;

  constructor( height: number, id: string, time: Date, myNote: string ) {
      this.height = height;
      this.time = time;
      this.id = id;
      this.myNote = myNote;
  }
}

const QUERY_1 = `query GetBlock($height: Int) {
    block( $height: Int = 0 ) {
      height
      time
      id
    }
  }`;


function handleNewBlock(newBlock: SubgraphNewBlock) {
    const resp = callGQL<Block>(newBlock.network, QUERY_1, { height: newBlock.height });
    printA(JSON.stringify(resp));

    const { height, id, time } = resp.data;
    printA(JSON.stringify(new SubgraphStoreBlock(height, id, time, "ok" )));
    storeRecord("SubgraphStoreBlock", new SubgraphStoreBlock(height, id, time, "ok" ))
}
