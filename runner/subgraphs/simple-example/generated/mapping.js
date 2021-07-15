"use strict";
exports.__esModule = true;
exports.BlockEntity = void 0;
var graph_1 = require("../../graph");
/***
 * Generated
 * Definitions here would be auto-generated by `graph codegen`
 */
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
/**
 * Mapping
 */
var GET_BLOCK = "query GetBlock($height: Int) {\n  block( $height: Int = 0 ) {\n    height\n    time\n    id\n  }\n}";
/**
 * This function is defined in the subgraph.yaml.
 *
 * Ex:
 * ```yaml
 * # ...
 * mapping:
 *  kind: cosmos/blocks
 *  apiVersion: 0.0.1
 *  language: wasm/assemblyscript
 *  blockHandlers:
 *    - function: handleNewBlock
 * ```
 */
function handleBlock(newBlockEvent) {
    graph_1.log.debug('newBlockEvent: ' + JSON.stringify(newBlockEvent));
    var _a = graph_1.graphql.call("cosmos", GET_BLOCK, { height: newBlockEvent.height }, "0.0.1"), error = _a.error, data = _a.data;
    if (error) {
        graph_1.log.debug('GQL call error: ' + JSON.stringify(error));
        return;
    }
    if (!data) {
        graph_1.log.debug('GQL call returned no data');
        return;
    }
    graph_1.log.debug('GQL call data: ' + JSON.stringify(data));
    var height = data.height, id = data.id, time = data.time;
    var entity = new BlockEntity(height, id, time, "ok");
    graph_1.log.debug('Entity: ' + JSON.stringify(entity));
    var storeErr = graph_1.store.save("SubgraphStoreBlock", entity).storeErr;
    if (storeErr) {
        graph_1.log.debug('Error storing entity: ' + JSON.stringify(storeErr));
    }
}