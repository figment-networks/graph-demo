"use strict";
exports.__esModule = true;
exports.BlockEntity = void 0;
var graph_1 = require("../graph");
var BlockEntity = /** @class */ (function () {
    function BlockEntity() {
        var args = [];
        for (var _i = 0; _i < arguments.length; _i++) {
            args[_i] = arguments[_i];
        }
        Object.assign(this, args);
    }
    return BlockEntity;
}());
exports.BlockEntity = BlockEntity;
var GET_BLOCK = "query GetBlock($height: Int) {\n    block( $height: Int = 0 ) {\n      height\n      time\n      id\n    }\n  }";
function handleNewBlock(newBlockEvent) {
    var _a = graph_1.callGQL(newBlockEvent.network, GET_BLOCK, { height: newBlockEvent.height }), data = _a.data, error = _a.error;
    if (error) {
        graph_1.logInfo(JSON.stringify(error));
        return;
    }
    graph_1.logInfo(JSON.stringify(data));
    var height = data.height, id = data.id, time = data.time;
    var entity = new BlockEntity(height, id, time, "ok");
    graph_1.logInfo(JSON.stringify(entity));
    graph_1.storeRecord("SubgraphStoreBlock", entity);
}
