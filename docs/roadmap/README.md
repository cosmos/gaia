<!--
order: false
parent:
  title: Previous Releases
  order: 7
-->

# Previous Releases

Please see the table below for library versions and other dependencies.  
  
  
## Cosmos Hub Release Details

### Delta Upgrade (Completed July 12, 2021)

- Gaia v5.0.x
- Gravity DEX:
  - A scalable AMM model for token swaps
  - Drives liquidity for tokens on the Cosmos Hub
  - Delivers price consistency and order execution

### Vega Upgrade (Completed December 14, 2021)

- Gaia v6.0.x
- Cosmos SDK v0.44
  - Fee grant module:
    - Allows paying fees on behalf of another account
  - Authz module:
    - Provide governance functions to execute transactions on behalf of another account
- Liquidity Module v1.4.2
  - The Gravity DEX with updates for dependencies
- IBC v2.0.0
- Tendermint v0.34.14
- Cosmosvisor v0.1.0
- IBC packet forward middleware v1.0.1
  - Cosmos Hub as a router

- External chain launch: Gravity Bridge
  - Transfer ATOM, ETH, ERC-20, and other Cosmos tokens between Ethereum and the Gravity Bridge Chain and by extension all IBC connected chains.
  - Fee and reward model hosted across Cosmos and Ethereum

### v7-Theta Upgrade (Completed March 25, 2022)

- Gaia v7.0.x
- Cosmos SDK v0.45
  - Minimal update with small fixes
- Gravity DEX: Liquidity v1.4.5
  - Adds a circuit breaker governance proposal type to disable adding new liquidity in order to make a migration possible.
- IBC 3.0.0
  - Interchain Account Module
    - Allows the creation of accounts on a "Host" blockchain which are controlled by an authentication module on a "Controller" blockchain.
    - Arbitrary messages are able to be submitted from the "Controller" blockchain to the "Host" blockchain to be executed on behalf of the Interchain Account.
    - Uses ordered IBC channels, one per account.

### v8-Rho Upgrade (expected Q1 2023)

- Gaia v8.0.x
- Cosmos SDK v0.45.12
  - Version bump with a number of fixes
  - See [changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.45.12/CHANGELOG.md) for details
- IBC v3.4
  - See [changelog](https://github.com/cosmos/ibc-go/blob/v3.4.0/CHANGELOG.md) for details
- IBC Packet Forward Middleware v3.1.1
- IBC Msg Whitelist to skip MinFee in CheckTX
- Global Fee Module
  - Allows denoms and min-fees to be governance parameters so gas can be paid in various denoms.
  - Visible on [tgrade](https://github.com/confio/tgrade/tree/main/x/globalfee) already and enabled in [ante.go](https://github.com/confio/tgrade/blob/main/app/ante.go#L72-L92)

### v9-Lambda Upgrade (expected Q1 2023)

- Gaia v9.0.x
- Cosmos SDK v0.45-ics
- IBC 4.2
- Interchain Security - Replicated Security
  - The Cosmos solution to shared security that uses IBC Cross Chain Validation (CCV) to relay validator set composition from a Provider Chain (Cosmos Hub) to a Consumer Chain. This validator set is in charge of producing blocks on both networks using separate nodes. Misbehavior on the Consumer Chain results in slashing Provider Chain staking tokens (ATOM).  
  
## Cosmos Hub Summary

| Upgrade Name        | Date          | Height    | Chain Identifier | Tm      | Cosmos SDK | Gaia                     | IBC                      |
|---------------------|---------------|-----------|---------------|------------|------------|--------------------------|--------------------------|
| Mainnet Launch      | 13/03/19    | 0         | `cosmoshub-1` | [v0.31.x](https://github.com/tendermint/tendermint/releases/tag/v0.31.11)         | [v0.33.x](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.33.2)      |  _Included in Cosmos SDK_ | n/a                      |
| [Security Hard Fork](https://forum.cosmos.network/t/critical-cosmossdk-security-advisory-updated/2211)  | 21/04/19    | 482,100   | `cosmoshub-1` | [v0.31.x](https://github.com/tendermint/tendermint/releases/tag/v0.31.11)          | [v0.34.x](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.34.6)    |   _Included in Cosmos SDK_)                  | n/a                      |
| Upgrade #1          | 21/01/20    |   500043 | `cosmoshub-2` | [v0.31.x](https://github.com/tendermint/tendermint/releases/tag/v0.31.11)         | [v0.34.x](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.34.10)     |   _Included in Cosmos SDK_)                  | n/a                      |
| Upgrade #2          | 07/08/20    |  2902000 | `cosmoshub-3` | [v0.32.x](https://github.com/tendermint/tendermint/releases/tag/v0.32.14)         | [v0.37.x](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.37.15)     | [v2.0.x](https://github.com/cosmos/gaia/releases/tag/v2.0.14)                   | n/a                      |
| Stargate            | 18/02/21    |  5200791 | `cosmoshub-4` | [v0.34.x](https://github.com/tendermint/tendermint/releases/tag/v0.34.3)          | [v0.40.x](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.40.1)      | [v4.0.x](https://github.com/cosmos/gaia/releases/tag/v4.0.6)                   | _Included in Cosmos SDK_ |
| Security Hard Fork  | ?             | ?         | `cosmoshub-4` | [v0.34.x](https://github.com/tendermint/tendermint/releases/tag/v0.34.8)       | [v0.41.x](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.41.4)      | [v4.2.x](https://github.com/cosmos/gaia/releases/tag/v4.2.1)                   | _Included in Cosmos SDK_ |
| Delta (Gravity DEX) | 13/07/21    |  6910000 | `cosmoshub-4` | [v0.34.x](https://github.com/tendermint/tendermint/releases/tag/v0.34.13)         | [v0.42.x](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.10)     | [v5.0.x](https://github.com/cosmos/gaia/releases/tag/v5.0.8)                   | _Included in Cosmos SDK_ |
| Vega    v6          | 13/12/21    |  8695000 | `cosmoshub-4` | [v0.34.x](https://github.com/tendermint/tendermint/releases/tag/v0.34.14)         | [v0.44.x](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.44.5)      | [v6.0.x](https://github.com/cosmos/gaia/releases/tag/v6.0.4)                   | [v2.0.x](https://github.com/cosmos/ibc-go/releases/tag/v2.0.3)                   |
| Theta   v7          | 12/04/22    | 10085397 | `cosmoshub-4` | [v0.34.x](https://github.com/tendermint/tendermint/releases/tag/v0.34.14)         | [v0.45.x](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.1)      | [v7.0.x](https://github.com/cosmos/gaia/releases/tag/v7.0.0)                   | [v3.0.x](https://github.com/cosmos/ibc-go/releases/tag/v3.0.0)                   |
| Rho     v8          | 16/02/23    | 14099412 | `cosmoshub-4` | [v0.34.x](https://github.com/informalsystems/tendermint/releases/tag/v0.34.24)    | [v0.45.x](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.12)     | [v8.0.x](https://github.com/cosmos/gaia/releases/tag/v8.0.0)                   | [v3.4.x](https://github.com/cosmos/ibc-go/releases/tag/v3.4.0)                   |
| Lambda  v9          | 15/03/23    | 14470501 | `cosmoshub-4` | [v0.34.x](https://github.com/informalsystems/tendermint/releases/tag/v0.34.25)    | [v0.45.x](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.13-ics) | [v9.0.x](https://github.com/cosmos/gaia/releases/tag/v9.0.0)                   | [v4.2.x](https://github.com/cosmos/ibc-go/releases/tag/v4.2.0)                   |
| v10                 | 21/06/23    | 15816200 | `cosmoshub-4` | [v0.34.x](https://github.com/cometbft/cometbft/releases/tag/v0.34.28)             | [v0.45.x](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.16-ics) | [v10.0.x](https://github.com/cosmos/gaia/releases/tag/v10.0.0)                 | [v4.4.x](https://github.com/cosmos/ibc-go/releases/tag/v4.4.0)                   |
| v11                 | 16/08/23    | 16596000 | `cosmoshub-4` | [v0.34.x](https://github.com/cometbft/cometbft/releases/tag/v0.34.29)             | [v0.45.x](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.16-ics) | [v11.x](https://github.com/cosmos/gaia/releases/tag/v11.0.0)                 | [v4.4.x](https://github.com/cosmos/ibc-go/releases/tag/v4.4.2)                   |


