
import { callGQL, GetBlockResponse, GrapQLResponse, SubgraphNewBlock, storeRecord, printA } from "../graph";

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


function handleNewBlock(nb: SubgraphNewBlock) {
    const resp = callGQL("networkOne", QUERY_1, {height: nb.height}) as GrapQLResponse;
    const data = resp.data as  GetBlockResponse;
    printA(JSON.stringify(resp));
    printA(JSON.stringify(new SubgraphStoreBlock(data.height, data.id, data.time, "ok" )));
    storeRecord("SubgraphStoreBlock", new SubgraphStoreBlock(data.height, data.id, data.time, "ok" ))
}
