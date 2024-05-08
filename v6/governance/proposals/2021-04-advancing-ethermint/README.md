# Advancing Ethermint - Governance Proposal: GTM and Engineering Plan for the Ethermint Chain 

> **NOTE**: this is a short version of the full proposal. To read the full document click [here](https://forum.cosmos.network/t/advancing-ethermint-governance-proposal-gtm-and-engineering-plan-for-the-ethermint-chain/4554).

### Author

Federico Kunze Küllmer: [@fedekunze](https://github.com/fedekunze)

## Summary

Tharsis is requesting 100,000 ATOMs from the Cosmos Hub community pool to fund, develop and advance the Ethermint project and launch an Ethermint Chain. The team will use these resources to grow our dedicated Cosmos-EVM team, so we can take on the commitments and responsibilities necessary to maintain and support the Cosmos-EVM chain and codebase.

An Ethermint environment has enormous potential to unlock new use cases within the Cosmos ecosystem that are not possible on Ethereum today. Use cases that require scalability and composability are some of the endless possibilities for Ethermint.

Ethermint is a way of vertically and horizontally scaling the projects that deploy on Ethereum, by allowing:

- Access to greater liquidity through IBC
- Faster transaction times through Tendermint BFT's instant finality
- Less strain on the Ethereum chain to process transactions (i.e. through gravity bridge)
- Seamless portability and composability with Ethereum Virtual Machine support

The commitments stated in this proposal will ensure a successful launch for the Ethermint chain together with its existing contributors (ChainSafe, OKEx, Iris, Injective, etc). Launching collaboratively with Ethermint stakeholders will result in a robust and independent community within Cosmos that will enable greater developer adoption of Cosmos technology.

The Ethermint stakeholders are partnering to execute on the long-awaited Ethermint Chain by joining forces to develop and deploy a new Cosmos EVM chain that will be used by thousands of crypto users at launch.

## Governance Votes

The following items summarize the voting options and what it means for this proposal. All addresses that vote on the proposal might be eligible for a future airdrop.

- **YES**: You approve the proposal statements and distribute the amount of 100,000 ATOMs to the multisig address. The treasury will allocate the funds to the Tharsis team, leading Ethermint's core development efforts and supporting the chain's go-to-market strategy. It will also lead to core technology maintenance and continuous discussion to ensure the project's longevity.
- **NO**: The NO vote is a request for improvements or adjustments. You agree that this proposal's motivation is valuable and that the team should create a follow-up proposal once the amendments are included.
- **NO (VETO)**: You veto the entire motivation for the proposal and expect the ICF and current maintainers to make the determination and continue the stewardship of the project. The proposers will not create a follow-up proposal.
- **ABSTAIN**: You are impartial to the outcome of the proposal.

## Multisig and release of funds

Upon the approval of the proposal, the treasury will distribute the funds to a ⅔ multi-signature account managed by the following individuals/partners:

- Federico Kunze Küllmer - Tharsis (proposer)
- Zaki Manian - Iqlusion
- Marko Baricevic - Interchain GmbH

The account address is: `cosmos124ezy53svellxqs075g69n4f5c0yzcy5slw7xz`

If the proposal passes, the team will immediately receive 40% of the funds to expand its engineering team and other business development efforts to support GTM for the chain. The remaining 60% will be released in an equal proportion to the number of milestones upon the completion of each milestone. For any reason, if the proposer has not completed the next milestone within a year of the last payment, the remaining funds held in the multisig account will be returned to the community pool.

## Product commitment

The current proposal aims to develop all the necessary components for a successful Ethermint chain. Our team will lead the core development efforts to execute the points below.

### Hard Commitments

These are the items that are mandatory for the release of funds. The items will be split into four milestones.

> NOTE: Some of the items below are currently stated under ChainSafe's service agreement with the ICF for Ethermint. Our team will collaborate with them on these items so that they are included by the time the EVM chain is launched. These items are marked below as [CS]

#### Milestone 1: Developer Usability and Testing 

This milestone aims to reach a stage where developers can begin deployments of Ethermint with the latest Cosmos SDK version and test their smart contracts in what will feel like a seamless experience.

- **Starport support**: Collaborate to ensure compatibility with Starport for developers that wish to use the EVM module with the latest SDK version on their sovereign chains.
- **Rosetta API support**: Support Ethermint transactions and queries on Coinbase’s Rosetta API that has been integrated into the SDK.
- **EVM Consistency**: Ensure that Ethermint can operate the same state as Ethereum and deterministically runs smart contract executions, exactly how Geth does (for example, checking the gas used between Ethermint and Geth)
- **Replay attack protection**: Register Ethermint permanent testnet and mainnet chain-ids to [ChainID Registry](https://chainid.network/) according to [EIP 155](https://eips.ethereum.org/EIPS/eip-155).
- **Documentation**: Ensure the documentation for both Ethermint and the EVM module are up to date with the implementation. JSON-RPC and OpenAPI (Swagger) docs for gRPC gateway and Rosetta will also be available for client developers. The team will create relevant sections to compare and distinguish key components of Ethermint and their corresponding ones on Ethereum. [CS]
- **Metrics**: We plan to list relevant metrics available through the SDK telemetry system for user engagement information such as the number of contracts deployed, amount transacted, gas usage per block, number of accounts created, number and amount IBC transfers to and from Ethermint, etc. These metrics will be displayed in a Dashboard UI in the form of charts. [CS]
- **Ensure compatibility with Ethereum tooling**: Test and coordinate with dev teams to test compatibility with (Truffle, Ganache, Metamask, web3.js, ethers.js, etc) and ensure the same dev UX as with Ethereum. The compatibility will then be ensured through end-to-end and integration tests. [CS]
- **User Guides**: Relevant guides will be added to connect Ethermint with the tools mentioned above.
- **Cosmjs Library support**: Make Ethermint keys, signing, queries, and txs compatible with the [cosmjs](https://github.com/cosmos/cosmjs) library.
- [**EIP 3085**](https://eips.ethereum.org/EIPS/eip-3085) **support**: add `wallet_addEthereumChain` JSON-RPC endpoint for Ethermint.

#### Milestone 2: Maximizing Performance and Compatibility

This milestone aims to enhance and benchmark the Ethermint chain's performance so developers can experience its superior benefits over existing solutions in the market.

- **EVM module readiness**: The current x/evm module from Ethermint suffers from technical debt regarding its architecture. The current proposal will do a bottleneck analysis of the EVM state transitions to redesign the EVM module to boost performance.
- **Benchmarks**: As a final step, we will be performing benchmarks for Ethereum transactions before and after the EVM refactor has been completed. [CS]
- **Maintain a permanent testnet**: Ethermint will have a permanent testnet to ease the development process for Ethereum developers and clients that wish to connect to Ethermint. The team will create a dedicated website, infrastructure, and faucet UI for users to request funds.
- **Faucet support**: The team will ensure an Ethermint-compatible faucet implementation is supported to ensure the sustainability of the permanent testnet. This will be also integrated into the existing faucet library of cosmjs. [CS]
- **Ethereum Bridge**: Integrate a combination of the following bridges in order to make Ethermint interoperable with Ethereum ERC20s: Cosmos Gravity bridge, IBC solo machine bridge, Chainbridge [CS].

#### Milestone 3: Mainnet readiness

This milestone's objective is to enhance security and users' accessibility to Ethermint, and stress-test the network before the mainnet launch.

- **Relayer Integration**: While the Ethermint migration to the SDK Stargate version supports IBC fungible token transfers on the app level, additional setup and integration is required to the IBC relayers to enable compatibility with Ethermint fully. The team will integrate the Ethermint keys and the remaining pieces to the relayer for full IBC support.
- **Ledger Support**: The team will perform an assessment of the current Cosmos and Ethereum ledger device applications to test their compatibility with Ethermint. If the keys or signing is not supported, the team will coordinate with ZondaX, the Ledger team, and other key partners to integrate the patches to the corresponding apps.
- **Simulations**: fuzz transaction testing for Ethermint and the EVM module. This will be done through the implementation of simulations and the [manticore](https://github.com/trailofbits/manticore) smart contract execution analysis tool.

#### Milestone 4: Mainnet launch

This milestone aims to provide support and coordination across the Cosmos community to ensure a safe and successful launch of the Ethermint mainnet.

- **Incentivized Testnet:** Planification, coordination and launch of the upcoming Ethermint’s incentivized testnet: Game of Ethermint.
- **Support Mainnet launch**: The team will support Ethermint’s mainnet launch by coordinating with key stakeholders, ecosystem partners, validators, community, etc. [CS]

#### Ongoing tasks

Below are hard commitment items that are required for a successful launch but don’t fit into any particular milestone as they are recurring over the whole development period.

- **Core Ethermint repository maintenance**: The team will commit to review community contributions and engage with issues and discussions regarding bugs and feature requests in the core codebase.
- **Coordination with Cosmos SDK core team**: Since the Ethermint codebase uses a lot of custom functionality (keys, `AnteHandler`, modular servers, etc) some changes/patches will need to be upstreamed to the Cosmos SDK to ensure modularity and non-breakingness.
- **Client support**: Develop partnerships with exchanges and wallets to support Ethermint through the Ethereum-compatible JSON-RPC or the gRPC services from the SDK since day one.
- **Community support**: Respond and support the community inquiries on Discord and other relevant channels.
- **Security Audit**: perform an internal and a third-party security audit prior to launch.
- **Bug bounty**: Coordinate a bug bounty program for the EVM module and the JSON-RPC server prior to launch.

## Soft Commitments

See the [full version](https://forum.cosmos.network/t/advancing-ethermint-governance-proposal-gtm-and-engineering-plan-for-the-ethermint-chain/4554) of this document.

## Conclusion

With this proposal, Tharsis plans to expedite the Ethermint chain's development and launch the network by Q4 2021. Ethermint will be the first EVM-compatible chain on Cosmos that will be fully interoperable with other BFT and EVM chains via IBC and the Gravity bridge. 

By creating and envisioning this long-term roadmap, we believe Ethermint can act as the vital component of the Interchain and serve as the gateway between the Ethereum and Cosmos ecosystems: The Ethermint launch will combine the Cosmos and Ethereum communities and provide new economic opportunities for millions of users.
