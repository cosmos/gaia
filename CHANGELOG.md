# CHANGELOG

## 10.0.2

*July 03, 2023*

This release bumps several dependencies and enables extra queries. 

### DEPENDENCIES

- Bump [ibc-go](https://github.com/cosmos/ibc-go) to
  [v4.4.2](https://github.com/cosmos/ibc-go/releases/tag/v4.4.2)
  ([\#2554](https://github.com/cosmos/gaia/pull/2554))
- Bump [CometBFT](https://github.com/cometbft/cometbft) to
  [v0.34.29](https://github.com/cometbft/cometbft/releases/tag/v0.34.29)
  ([\#2594](https://github.com/cosmos/gaia/pull/2594))

### FEATURES

- Register NodeService to enable query `/cosmos/base/node/v1beta1/config`
  gRPC query to disclose node operator's configured minimum-gas-price.
  ([\#2629](https://github.com/cosmos/gaia/issues/2629))

## [v10.0.1] 2023-05-25

* (deps) [#2543](https://github.com/cosmos/gaia/pull/2543) Bump [ibc-go](https://github.com/cosmos/ibc-go) to [v4.4.1](https://github.com/cosmos/ibc-go/releases/tag/v4.4.1).

## [v10.0.0] 2023-05-19

* (deps) [#2498](https://github.com/cosmos/gaia/pull/2498) Bump multiple dependencies. 
  * Bump [cosmos-sdk](https://github.com/cosmos/cosmos-sdk) to [v0.45.16-ics](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.16-ics). See the [v0.45.16 release notes](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.16) for details. 
  * Bump [ibc-go](https://github.com/cosmos/ibc-go) to [v4.4.0](https://github.com/cosmos/ibc-go/releases/tag/v4.4.0).
  * Bump [CometBFT](https://github.com/cometbft/cometbft) to [v0.34.28](https://github.com/cometbft/cometbft/releases/tag/v0.34.28).
* (gaia) Bump Golang prerequisite from 1.18 to 1.20. See (https://go.dev/blog/go1.20) for details.

## Previous Versions

[CHANGELOG of previous versions](https://github.com/cosmos/gaia/blob/main/CHANGELOG.md)

<!-- Release links -->
[v10.0.1]: https://github.com/cosmos/gaia/releases/tag/v10.0.1
[v10.0.0]: https://github.com/cosmos/gaia/releases/tag/v10.0.0

