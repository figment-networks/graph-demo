"use strict";
// This test is used by jsRunner_test.go
exports.__esModule = true;
var graph_1 = require("../graph");
var BlockEntity = /** @class */ (function () {
    function BlockEntity(height, id, time, myNote) {
        this.id = id;
        this.height = height;
        this.time = time;
        this.myNote = myNote;
    }
    return BlockEntity;
}());
var GET_BLOCK = "query GetBlock($height: Int) {\n    block( $height: Int = 0 ) {\n      height\n      time\n      id\n    }\n  }";
function handleNewBlock(newBlockEvent) {
    var _a = graph_1.callGQL(newBlockEvent.network, GET_BLOCK, { height: newBlockEvent.height }), data = _a.data, error = _a.error;
    if (error) {
        graph_1.printA(JSON.stringify(error));
        return;
    }
    graph_1.printA(JSON.stringify(data));
    var height = data.height, id = data.id, time = data.time;
    var entity = new BlockEntity(height, id, time, "ok");
    graph_1.printA(JSON.stringify(entity));
    graph_1.storeRecord("StoreBlock", entity);
}
