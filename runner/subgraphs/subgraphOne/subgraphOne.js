"use strict";
var __extends = (this && this.__extends) || (function () {
    var extendStatics = function (d, b) {
        extendStatics = Object.setPrototypeOf ||
            ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
            function (d, b) { for (var p in b) if (Object.prototype.hasOwnProperty.call(b, p)) d[p] = b[p]; };
        return extendStatics(d, b);
    };
    return function (d, b) {
        if (typeof b !== "function" && b !== null)
            throw new TypeError("Class extends value " + String(b) + " is not a constructor or null");
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
exports.__esModule = true;
var graph_1 = require("../graph");
var BlockEntity = /** @class */ (function (_super) {
    __extends(BlockEntity, _super);
    function BlockEntity() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    return BlockEntity;
}(graph_1.Entity));
var GET_BLOCK = "query GetBlock($height: Int) {\n    block( $height: Int = 0 ) {\n      height\n      time\n      id\n    }\n  }";
// This is called by the mapping runtime when a new block is available to be processed
function handleNewBlock(newBlock) {
    var _a = graph_1.callGQL(newBlock.network, GET_BLOCK, { height: newBlock.height }), data = _a.data, error = _a.error;
    if (error) {
        graph_1.printA(JSON.stringify(error));
        return;
    }
    graph_1.printA(JSON.stringify(data));
    var height = data.height, id = data.id, time = data.time;
    var entity = new BlockEntity(height, id, time, "ok");
    graph_1.printA(JSON.stringify(entity));
    graph_1.storeRecord("SubgraphStoreBlock", entity);
}
