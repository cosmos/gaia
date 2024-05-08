# The Cosmos Hub Roadmap 2.0

This Cosmos Hub Product Roadmap incorporates input from product owners, key Cosmos stakeholders (ATOM holders, Interchain Foundation, and Cosmos Hub operations), the internal Gaia / ATOM call, the virtual Gaia Org Call, the [Cosmos Hub as a Port City](https://blog.cosmos.network/the-cosmos-hub-is-a-port-city-5b7f2d28debf) article, and the [ATOM2021](https://github.com/cosmosdevs/atom2021) presentation.

This roadmap gives a one-year guideline in which stakeholders can anticipate updated features on the Cosmos Hub, with the greatest degree of specificity available for the most immediate upgrades, and decreasing precision available the further out the timeline goes.

The upgrades aim to add features such as liquidity, economic security, usability, and participation. To highlight our focus on DeFi, we have chosen to use the [Greeks from Finance](https://en.wikipedia.org/wiki/Greeks_(finance)) in naming upcoming upgrades.

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

## v7-Theta Upgrade (expected Q1 2022)
- Gaia v7.0.x
- Cosmos SDK v0.45
  - Minimal update with small fixes
- IBC 3.0
  - Interchain Account Module
    - Allows the creation of accounts on a "Host" blockchain which are controlled by an authentication module on a "Controller" blockchain.
    - Arbitrary messages are able to be submitted from the "Controller" blockchain to the "Host" blockchain to be executed on behalf of the Interchain Account.
    - Uses ordered IBC channels, one per account.
- Interchain Account Message Auhothorization Module
    - Authentication module that authorizes any Account to create an Interchain Account on any IBC connected "Host" blockchain that has the Interchain Account IBC module.
    - Accounts can be private key controlled users, and eventually the Gov Module and any Groups Module.

## v8-Rho Upgrade (expected Q2 2022)
- Gaia v8.0.x
- Cosmos SDK v0.46
  - Groups module:
    - Enables higher-level multisig permissioned accounts, e.g., weight-based voting policies
  - Meta-Transactions
    - Allows transactions to be submitted by separate accounts that receive tips.
  - Gov Module Improvements
    - Execution of arbitraty transactions instead of just governance proposals.
    - Enables much more expressive governance module.
- NFT module
  - Enable simple management of NFT identifiers, their owners, and associated data, such as URIs, content, and provenance
  - An extensible base module for extensions including collectibles, custody, provenance, and marketplaces
- Tendermint v0.35
- Liquid Staking
  - Frees secure and low-risk delegations for use in other parts of the Cosmos ecosystem
  - Features include enabling transfer of rewards and voting rights
- Wasmd
  - Governance permissioned CosmWASM instance on the hub
- Budget Module (stretch-goal)
  - Inflation funding directed to arbitrary module and account addresses
- Global Fee Module (stretch-goal)
  - Allows denoms and min-fees to be governance parameters so gas can be paid in various denoms.
  - Visible on [tgrade](https://github.com/confio/tgrade/tree/main/x/globalfee) already and enabled in [ante.go](https://github.com/confio/tgrade/blob/main/app/ante.go#L72-L92)
- Bech32 Prefix forwarding (stretch-goal)
  - https://github.com/osmosis-labs/bech32-ibc

## v9-Lambda Upgrade (expected Q3 2022)
- Gaia v9.0.x
- Interchain Security v1 - Required Participation of Provider Chain Validators
  - The Cosmos solution to shared security that uses IBC Cross Chain Validation (CCV) to relay validator set composition from a Provider Chain (Cosmos Hub) to a Consumer Chain. This validator set is in charge of producing blocks on both networks using separate nodes. Misbehavior on the Consumer Chain results in slashing Provider Chain staking tokens (ATOM).
  - Allows independent modules like Gravity DEX or Bridge to live on separate chains with their own development cycles.
- Chain Name Service
  - Chain-ID registry
  - Node registry
  - IBC Path Resolution
  - Asset registry
  - Account registry
  - Bech32 registry

## v10-Epsilon (expected Q4 2022)
- Gaia v10.0.x
- Interchain Security v2 - Opt-In Participation of Provider Chain Validators
  - Where Provider Chain validators have the ability to opt-in to block production for various Consumer Chains.
- Cosmos SDK v0.47
  - Sparse Merkle Tree (SMT)
    - Various storage and performance optimizations 
  - Postgres indexing
  - Protobuf v2

## v11-Gamma (expected Q1 2023)
- Gaia v11.0.x
- Interchain Security v3 - Layered Security
  - Where Consumer Chains combine their own staking token validator set with Provider Chain validator set.

## Future Considerations
The Cosmos Hub is a decentralized network with many diverse contributors. As such there is no one authority of what is or can be part of the Cosmos Network. The Cosmos Hub team at Interchain does it's best to maintain the Gaia repository, which is the primary codebase that operates the Cosmos Network. The Interchain Foundation is one of the sources of funding for engineering work that may make its way onto the Cosmos Hub. We do our best to participate in ongoing conversations about the mission, vision and purpose of the Cosmos Hub, so that we can best support work to enabling it via funding, engineering, coordination and communication. Some of the topics which have been discussed by contributors inside and outside of Interchain are listed below, although have not been developed to the point of being included in the roadmap:
* Privacy
* Smart Contracts
* Rollups

The Cosmos Hub Roadmap is maintained by the Interchain Cosmos Hub team as a living document, and is updated in collaboration with key stakeholders from the multi-entity Cosmos community. 
