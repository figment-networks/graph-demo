"use strict";
exports.__esModule = true;
var graph_1 = require("../graph");
var entities_1 = require("./entities");
var GET_BLOCK = "query GetBlock($height: Int) {\n    block( $height: Int = 0 ) {\n      height\n      time\n      id\n    }\n  }";
/**
 * This would be defined in the subgraph.yaml.
 *
 * Ex:
 * ```yaml
 * # ...
 * mapping:
 *  kind: cosmos/events
 *  apiVersion: 0.0.1
 *  language: wasm/assemblyscript
 *  blockHandlers:
 *    - function: handleNewBlock
 * ```
 */
function handleNewBlock(newBlockEvent) {
    var _a = graph_1.callGQL(newBlockEvent.network, GET_BLOCK, { height: newBlockEvent.height }), data = _a.data, error = _a.error;
    if (error) {
        graph_1.logInfo(JSON.stringify(error));
        return;
    }
    graph_1.logInfo(JSON.stringify(data));
    var height = data.height, id = data.id, time = data.time;
    var entity = new entities_1.BlockEntity(height, id, time, "ok");
    // log.info
    graph_1.logInfo(JSON.stringify(entity));
    // replace with `entity.save()` for graph-ts
    graph_1.storeRecord("SubgraphStoreBlock", entity);
}
