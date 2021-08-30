# Subgraphs

The subgraphs defined in this directory are examples and not compatible with Graph Node. These are meant to be run by the V8 runtime ([runner](../runner/README.md)) included in this demo.

# Building Subgraphs

- When making changes to a subgraph typescript file, you need to rebuild (e.g. `npm run build:simple-example`) to javascript to reflect changes in the V8 runtime.
- The subgraph needs to be compiled into a single js file. The V8 runtime does not handle multiple files (e.g. using `import`)
