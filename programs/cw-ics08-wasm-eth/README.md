# `CosmWasm` ICS08 Wasm Ethereum Light Client

> [!CAUTION]
> ⚠ The Ethereum Light Client is currently under heavy development, and is not ready for use.

This is the `CosmWasm` implementation that can be used with ibc-go's `08-wasm` light client wrapper. 
It handles the client and consensus state, and calls into `packages/ethereum-light-client` for all the light client related logic.

## Acknowledgements

This work is based on the ethereum light client created by [Union](http://github.com/unionlabs/union/).
