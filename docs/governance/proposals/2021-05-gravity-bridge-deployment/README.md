# Signaling Proposal - Deployment of Gravity Bridge on the Cosmos Hub

![Gravity Bridge](https://raw.githubusercontent.com/althea-net/cosmos-gravity-bridge/main/gravity-bridge.svg)

## Summary

This proposal is a Request For Comment from the ATOM community regarding the activation of the Gravity Bridge module onto the Cosmos Hub.

By voting YES to this proposal, you will signal that you approve of having the Gravity Ethereum <> Cosmos bridge deployed onto the Cosmos Hub.

## Vision

Gravity as an Ethereum-Cosmos bridge is designed for the Cosmos Hub to pull as much value as possible into the orbits of Cosmos via a direct and decentralized bridge. Gravity will be able to bring ERC20 assets from Ethereum into Cosmos, as well as Cosmos assets to Ethereum ERC20 representations. ATOM, and any other asset in the Cosmos ecosystem, will be able to trade on Uniswap and other Ethereum AMMs, and interact with Ethereum DeFi like any ERC20 token. As well, Cosmos Hub users can use the Gravity DEX to trade between ERC20 assets and tokens that are transferred using IBC. This will bring a tremendous amount of liquidity and utility to these multi-chain assets.

## Cosmos, Ethereum, and Gravity

Gravity is a secure and highly efficient bridge between EVM and Cosmos SDK-based blockchains. At a high-level, Gravity enables token transfers from Ethereum to the Cosmos Hub and back again by locking tokens on the Ethereum side and minting equivalent tokens on the Cosmos side.

Gravity is completely non-custodial. Control of the bridge mirrors the active validator set on the Cosmos SDK-based chain, and validator stake on Cosmos can be slashed for misbehavior involving the Gravity bridge.

The Gravity Ethereum contract is highly optimized, utilizing batches to dramatically reduce the cost of Cosmos -> Ethereum transfers. Sending funds from Cosmos back to Ethereum can be up to 50% less costly than a normal ERC20 send.

## How do validators support the Gravity Bridge?

Cosmos Hub validators will run three key software components of the Gravity bridge:

- The Gravity bridge module, integrated into gaiad (the binary that runs the Cosmos Hub)
- The Gravity bridge Orchestrator
- A Geth light client or any Ethereum full node implementing the JSON-rpc standard

### Cosmos to Ethereum:

To send transactions from Cosmos to Ethereum, the Gravity Bridge module of a validator's Gaia instance first packages the transaction data, and makes it available on an endpoint. The Orchestrator then signs this data with the validator’s Ethereum key, and submits it as a message to the Ethereum network. The Ethereum signature of each Cosmos validator needs to be assembled and submitted to the Ethereum chain by relayers.

Validators may be slashed if they fail to submit Ethereum signatures within 10,000 blocks (about twelve to fourteen hours) of their creation.

The current liveness rules require a validator to sign at least 500 of the last 10,000 blocks (about twelve to fourteen hours)

Validators may also be slashed if they sign a message with their Ethereum key that was not created by the Gravity bridge module.

Gravity bridge has no other slashing conditions.

### Ethereum to Cosmos:

The Orchestrator also monitors the Ethereum chain, submitting events that occur on Ethereum to Cosmos as messages. When more than 2/3 of the active voting power has sent a message observing the same Ethereum event, the Gravity module will take action.

This oracle action will not be incentivized, nor will it be enforced with slashing. If validators making up more than 33% of the staked tokens do not participate in the oracle, new deposits and withdrawals will not be processed until those validators resume their oracle obligations.

The oracle was designed to adhere to the Cosmos Hub validator security model without slashing conditions to ensure that a consensus failure on Ethereum does not affect operation of the Cosmos chain.

### Slashing Conditions Spec

https://github.com/cosmos/gravity-bridge/blob/main/spec/slashing-spec.md

## How does it work?

Gravity consists of 4 parts:

- An Ethereum contract called Gravity.sol
- The Gravity Cosmos module of Gaia
- The orchestrator program which is run by Cosmos validators alongside the Gravity module
- A market of relayers, who compete to submit transactions to Ethereum

### Gravity.sol

The Gravity Ethereum contract is a highly compact and efficient representation of weighted powers voting on Ethereum. It contains an Etheruem key from each Cosmos validator, as well as their voting power. This signer set is continuously updated as validation power changes on Cosmos, ensuring that it matches the current Cosmos validator set.

Sending tokens, or updating the validator set, contained in Gravity.sol requires more than 66% of the total voting power to approve the action. In this way Gravity.sol mirrors Tendermint consensus on the Cosmos chain as closely as possible on Ethereum.

### Gravity Cosmos Module

The Gravity module governs and coordinates the bridge. Generating messages for the validators to sign with their Ethereum keys and providing these signatures to relayers who assemble and submit them to the Ethereum chain.

### Orchestrator

The Gravity bridge orchestrator performs all the external tasks the Gravity bridge requires for validators, which includes submission of signatures and submission of Ethereum events.

While the Gravity module concerns itself with the correctness and consensus state of the bridge, the Orchestrator is responsible for locating and creating the correct inputs.

### Market of Relayers

Relayers are an unpermissioned role that observes the Cosmos chain for messages ready to be submitted to Ethereum.

The relayer then packages the validators signatures into an Ethereum transaction and submits that transaction to the Ethereum blockchain. All rewards in the Gravity bridge design are paid to msg.sender on Ethereum. This means that relayers do not require any balance on the Cosmos side and can immediately liquidate their earnings into ETH while continuing to relay newer messages.

## Security assumptions

The Gravity bridge is designed with the assumption that the total amount of funds in Gravity.sol is less than the value of the validator set’s total staked tokens.

If this assumption does not hold true, it would be more profitable for validators to steal the funds in the bridge and simply lose their stake to slashing.

There is no automated enforcement of this assumption. It is up to the $ATOM holders to take action if the amount deposited in the bridge exceeds the total value of all stake on the hub.

It should be noted that this condition is not unique to the Gravity bridge. The same dynamic exists for any IBC connection, and even exists in scenarios other than cross-chain communication. For example, in a hypothetical blockchain keeping domain name records, this same vulnerability would exist if the potential profit from exploiting the domain name system was greater than the value of the validator set’s total staked tokens.

## Ongoing work

The Gravity Bridge has been continuously tested throughout Q1/Q2 2021 by multiple ongoing test nets with a diverse group of validators.

### Testing

The Althea team is committed to playing a long-term role in upgrading, documenting, and supporting Gravity over the coming years.
The Gravity bridge is currently live and running in a testnet, which validators can join by following the instructions [here](https://github.com/althea-net/althea-chain/blob/main/docs/althea/testnet-2.md)

### Audit:

The Gravity bridge module is currently undergoing an audit with Informal Systems estimated to be completed by the end of July, 2021.

Phase one of the audit has been completed, which resulted in the addition of evidence based slashing and several other minor design fixes.

The phase two design audit will be completed by the end of June. To be followed by phrase three, an implementation audit to be completed by the end of July.

### Conclusion:

With this proposal, the Althea team, together with Cosmos ecosystem partners, will expedite the development of the Gravity Bridge with an incentivized testnet and launch in Q3 2021. Althea will be closely shepherding the Gravity Bridge throughout all phases related to the testing, audit, and implementation process on the Cosmos Hub.

## Proposers

_The Althea Gravity bridge team._

Deborah Simpiler, Justin Kilpatrick and, Jehan Tremback

We’d like to share praise and thank you for contributions from the following teams!

Interchain Foundation

All in Bits/Tendermint

Sommelier, Informal, Injective, Confio

_Gravity Readiness Committee:_

Justin Kilpatrick and Jehan Tremback, Althea

Zarko Milosevic, Informal Systems

Zaki Manian, Sommelier/Iqlusion

## Governance Votes

The following items summarize the voting options and what it means for this proposal.

- **YES**: You agree that Gravity Bridge should be deployed to the Cosmos Hub.
- **NO**: You disapprove of deploying Gravity bridge on the Cosmos Hub in its current form (please indicate in the [Cosmos Forum](https://forum.cosmos.network/) why this is the case).
- **NO WITH VETO**: You are strongly opposed to the deployment of Gravity bridge on the Cosmos Hub and will exit the network if this occurred.
- **ABSTAIN**: You are impartial to the outcome of the proposal.

## Appendix

### FAQ

### Is running the Gravity Module difficult for Cosmos Validators?

Soliciting feedback from dozens of Cosmos hub validators and over 100 test net participants, we found that running the Gravity module is not a difficult task or undue burden on Cosmos validators.

### Is the Gravity bridge secure?

The Gravity bridge is undergoing an audit by Informal Systems. It will then be up to ATOM holders to interpret the results of the code audit and weigh implementation risks in another governance proposal before deployment. Fundamentally, the design of the Gravity Bridge means that its security is directly represented by the security of the validator set on Cosmos Hub. On the Cosmos side, the Gravity Bridge security design adheres to the Cosmos network's ideal of safety over liveness assured by 2/3+ validator voting power, but without sacrificing safety or liveness due to issues that may rarely arise with Ethereum.

### Are slashing conditions a problem for validators?

Gravity Bridge slashing conditions closely mirror the slashing conditions which validators are already subject to.

- Uptime: Validators on Cosmos currently must keep their validator software running at all times, or risk slashing. Gravity adds an additional binary which must be run, which is low in difficulty, resource usage, and operational requirements.

- Equivocation: Validators on Cosmos are subject to slashing if they sign two blocks at the same height. It is possible for this to happen through accidental misconfiguration. Gravity adds an additional item which must not be signed, which are the fraudulent bridge transactions that never existed on Cosmos. It is not possible for this to happen by accident, so this slashing condition is much less of a risk than the Hub’s existing slashing conditions.

### What about peg zones?

The concept of a "peg zone" has been around in Cosmos for a while. This is a separate chain which runs the Gravity Bridge and connects to Ethereum or another blockchain. We believe that running Gravity Bridge on the Cosmos Hub and connecting directly is the superior solution. First, let's look at what different forms peg zones could take.

1. The most likely type of peg zone is what will result if this proposal does not pass. There are at least 5 Cosmos SDK chains who will be using the Gravity Bridge module to connect to Ethereum. There will be no official way to bridge Ethereum assets into the Cosmos ecosystem. Instead there will be many dueling representations of Ethereum assets. It goes without saying that this will confuse Cosmos users. If one of these peg zones gains the upper hand and becomes dominant, the Cosmos Hub will miss out on all the transaction fees that it generates.

2. It would also be possible to establish an official peg zone, and airdrop its staking token 1:1 to current Atom holders. This would at least allow the Cosmos Hub stakeholders to keep the economic benefit of activity on the peg zone. However, this is capital inefficient. A dollar staked on this peg zone would not be staked on the hub and vice versa. Splitting stake between these two important chains would make both weaker. Cosmos users will also need to choose whether to put value into Atom or put value into the peg zone token, resulting in less value in Atom.

3. Shared security has been talked about for a long time in Cosmos, but it is not yet in production. Establishing a peg zone with shared security with the Cosmos Hub would allow the same validator set to validate both chains, and put the same Atoms at stake to secure each of them. This would avoid the issues of scenarios 1 and 2 above. However, we cannot afford to wait. Many other PoS blockchains already have Ethereum bridges and Cosmos needs to continue innovating to stay relevant.

This is easier to both design and debug, and is ideal for high-value chains like the Hub.

### Why not use IBC to create a bridge to Ethereum?

Ethereum as it currently exists (ETH1) fundamentally can not implement IBC. It may be possible in the future to create an IBC bridge with ETH2.

Implementing IBC for ETH2 require significant development effort and coordination with the Ethereum team as well as the completion and deployment of ETH2, which is on an uncertain timeline.

Gravity bridge is a design that can be deployed in the forseeable future and will continue to function so long as ETH1 compatibility is maintained.
