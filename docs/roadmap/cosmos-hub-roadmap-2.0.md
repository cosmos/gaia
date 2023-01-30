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
- Cosmos SDK v0.45
  - Minimal update with small fixes
- Interchain Security - Required Participation of Provider Chain Validators
  - The Cosmos solution to shared security that uses IBC Cross Chain Validation (CCV) to relay validator set composition from a Provider Chain (Cosmos Hub) to a Consumer Chain. This validator set is in charge of producing blocks on both networks using separate nodes. Misbehavior on the Consumer Chain results in slashing Provider Chain staking tokens (ATOM).
- Global Fee Module
  - Allows denoms and min-fees to be governance parameters so gas can be paid in various denoms.
  - Visible on [tgrade](https://github.com/confio/tgrade/tree/main/x/globalfee) already and enabled in [ante.go](https://github.com/confio/tgrade/blob/main/app/ante.go#L72-L92)

## v9-Lambda Upgrade (expected Q1 2023)

- Gaia v9.0.x
<<<<<<< HEAD
- Cosmos SDK v0.47
  - Groups module:
    - Enables higher-level multisig permissioned accounts, e.g., weight-based voting policies
  - Gov Module Improvements
    - Execution of arbitraty transactions instead of just governance proposals.
    - Enables much more expressive governance module.
  - Phasing out broadcast mode
  - Remove proposer based rewards
  - Consensus-param module
    - With the deprecation of the param module, this will allow governance or another account to modify these parameters.
  - Changes required for Interchain Security
  - Liquid Staking module
    - Free, secure and low-risk delegations for use in other parts of the Cosmos ecosystem
    - Features include enabling transfer of rewards and voting rights
- IBC 5.x
  - Relayer Incentivisation so that IBC packets contain fees to pay for relayer costs.
- Interchain Account Message Authorization Module
  - Authentication module that authorizes any Account to create an Interchain Account on any IBC connected "Host" blockchain that has the Interchain Account IBC module.
  - Accounts can be private key controlled users, and eventually the Gov Module and any Groups Module.
- IBC Msg Whitelist to skip MinFee in CheckTX
- Bech32 Prefix forwarding
  - <https://github.com/osmosis-labs/bech32-ibc>
- Liquidity Module Deprecation
  - Contains forced withdraw of liquidity

## v10-Epsilon (expected Q2 2023)

- Gaia v10.0.x
- IBC Queries
- Hub ATOM Liquidity (HAL)
  - Protocol Controlled Value application to acquire ATOM LP tokens with Interchain Security Tokens

## v11-Gamma (expected Q3 2023)

- Gaia v11.0.x
- Interchain Security v2 - Layered Security
  - Where Consumer Chains combine their own staking token validator set with Provider Chain validator set.

## Future Considerations

The Cosmos Hub is a decentralized network with many diverse contributors. As such there is no one authority of what is or can be part of the Cosmos Network. The Cosmos Hub team at Interchain does its best to maintain the Gaia repository, which is the primary codebase that operates the Cosmos Network. The Interchain Foundation is one of the sources of funding for engineering work that may make its way onto the Cosmos Hub. We do our best to participate in ongoing conversations about the mission, vision and purpose of the Cosmos Hub, so that we can best support work to enabling it via funding, engineering, coordination and communication. Some of the topics which have been discussed by contributors inside and outside of Interchain are listed below, although have not been developed to the point of being included in the roadmap:

- Multi-Hop Routing
  - Simplifies the topography of relayers such that packets from pairwise channels between chains can be routed through the hub while preserving the original channel and more importantly token denom path.
- Chain Name Service
  - Chain-ID registry
  - Node registry
  - IBC Path Resolution
  - Asset registry
  - Account registry
  - Bech32 registry
- IBC NFT
- NFT module
  - Enable simple management of NFT identifiers, their owners, and associated data, such as URIs, content, and provenance
  - An extensible base module for extensions including collectibles, custody, provenance, and marketplaces
  - Unless the Cosmos Hub plans to be a full blown platform for NFT publication, it should pair the inclusion of this module with the IBC NFT module similar to how the Cosmos Hub doesn't allow new Fungible Tokens to be published but does allow them to be transferred via IBC.
- Privacy
- Smart Contracts
- Rollups

The Cosmos Hub Roadmap is maintained by the Interchain Cosmos Hub team as a living document, and is updated in collaboration with key stakeholders from the multi-entity Cosmos community.
=======
- Cosmos SDK v0.45
- IBC 4.2
- Interchain Security - Replicated Security
  - The Cosmos solution to shared security that uses IBC Cross Chain Validation (CCV) to relay validator set composition from a Provider Chain (Cosmos Hub) to a Consumer Chain. This validator set is in charge of producing blocks on both networks using separate nodes. Misbehavior on the Consumer Chain results in slashing Provider Chain staking tokens (ATOM).
>>>>>>> 690b0c9 (Update cosmos-hub-roadmap-2.0.md (#2113))
