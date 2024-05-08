# The Cosmos Hub Roadmap

This Cosmos Hub roadmap serves as a reference for the current planned features of upcoming releases, as well as providing a record of past releases.

## Delta Upgrade (Completed July 12, 2021)

- Gaia v5.0.x
- Gravity DEX:
  - A scalable AMM model for token swaps
  - Drives liquidity for tokens on the Cosmos Hub
  - Delivers price consistency and order execution

## Vega Upgrade (Completed December 14, 2021)

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

## v7-Theta Upgrade (Completed March 25, 2022)

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

## v8-Rho Upgrade (expected Q1 2023)

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



## v9-Lambda Upgrade (expected Q1 2023)

- Gaia v9.0.x
- Cosmos SDK v0.45
- IBC 4.2
- IBC Packet Forward Middleware v4.0.1
- Interchain Security - Replicated Security
  - The Cosmos solution to shared security that uses IBC Cross Chain Validation (CCV) to relay validator set composition from a Provider Chain (Cosmos Hub) to a Consumer Chain. This validator set is in charge of producing blocks on both networks using separate nodes. Misbehavior on the Consumer Chain results in slashing Provider Chain staking tokens (ATOM).

