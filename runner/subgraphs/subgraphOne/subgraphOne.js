"use strict";
exports.__esModule = true;
var graph_1 = require("../graph");
var SubgraphStoreBlock = /** @class */ (function () {
    function SubgraphStoreBlock(height, id, time, myNote) {
        this.height = height;
        this.time = time;
        this.id = id;
        this.myNote = myNote;
    }
    return SubgraphStoreBlock;
}());
var QUERY_1 = "query GetBlock($height: Int) {\n    block( $height: Int = 0 ) {\n      height\n      time\n      id\n    }\n  }";
function handleNewBlock(nb) {
    var resp = graph_1.callGQL("networkOne", QUERY_1, { height: nb.height });
    var data = resp.data;
    graph_1.printA(JSON.stringify(resp));
    graph_1.printA(JSON.stringify(new SubgraphStoreBlock(data.height, data.id, data.time, "ok")));
    graph_1.storeRecord("StoreBlock", new SubgraphStoreBlock(data.height, data.id, data.time, "ok"));
}
