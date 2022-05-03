<!--
order: 3
-->

# Validator FAQ

::: warning Disclaimer
This is work in progress. Mechanisms and values are susceptible to change.
:::

## General Concepts

### What is a Cosmos validator?

The [Cosmos Hub](../getting-started/what-is-gaia.md) is based on [Tendermint](https://tendermint.com/docs/introduction/what-is-tendermint.html) that relies on a set of validators to secure the network. The role of validators is to run a full node and participate in consensus by broadcasting votes that contain cryptographic signatures signed by the validator's private key. Validators commit new blocks in the blockchain and receive revenue in exchange for their work. Validators must also participate in governance by voting on proposals. Validators are weighted according to their total stake.

### What is staking?

The Cosmos Hub is a public Proof-Of-Stake (PoS) blockchain, meaning that the weight of validators is determined by the amount of staking tokens (ATOM) bonded as collateral. These ATOM tokens can be self-delegated directly by the validator or delegated to the validator by other ATOM holders.

Any user in the system can declare their intention to become a validator by sending a `create-validator` transaction to become validator candidates.

The weight (i.e. voting power) of a validator determines whether they are an active validator. The active validator set is limited to [an amount](https://www.mintscan.io/cosmos/validators) that changes over time.

### What is a full node?

A full node is a server running a chain's _binary_ (its software) that fully validates transactions and blocks of a blockchain and keeps a full record of all historic activity. A full node is distinct from a pruned node that processes only block headers and a small subset of transactions. Running a full node requires more resources than a pruned node. Validators can decide to run either a full node or a pruned node, but they need to make sure they retain enough blocks to be able to validate new blocks.

Of course, it is possible and encouraged for users to run full nodes even if they do not plan to be validators.

You can find more details about the requirements in the [Joining Mainnet Tutorial](../hub-tutorials/join-mainnet.md).

### What is a delegator?

Delegators are ATOM holders who cannot, or do not want to, run a validator themselves. ATOM holders can delegate ATOM to a validator and obtain a part of their revenue in exchange. For details on how revenue is distributed, see [What is the incentive to stake?](#what-is-the-incentive-to-stake?) and [What are validators commission?](#what-are-validators-commission?) in this document.

Because delegators share revenue with their validators, they also share risks. If a validator misbehaves, each of their delegators are partially slashed in proportion to their delegated stake. This penalty is one of the reasons why delegators must perform due diligence on validators before delegating. Spreading their stake over multiple validators is another layer of protection.

Delegators play a critical role in the system, as they are responsible for choosing validators. Being a delegator is not a passive role. Delegators must actively monitor the actions of their validators and participate in governance. For details on being a delegator, read the [Delegator FAQ](https://hub.cosmos.network/main/delegators/delegator-faq.html).

## Becoming a Validator

### How to become a validator?

Any participant in the network can signal that they want to become a validator by sending a `create-validator` transaction, where they must fill out the following parameters:

- **Validator's `PubKey`:** The private key associated with this Tendermint `PubKey` is used to sign _prevotes_ and _precommits_.
- **Validator's Address:** Application level address that is used to publicly identify your validator. The private key associated with this address is used to delegate, unbond, claim rewards, and participate in governance.
- **Validator's name (moniker)**
- **Validator's website (Optional)**
- **Validator's description (Optional)**
- **Initial commission rate**: The commission rate on block rewards and fees charged to delegators.
- **Maximum commission:** The maximum commission rate that this validator can charge. This parameter is fixed and cannot be changed after the `create-validator` transaction is processed.
- **Commission max change rate:** The maximum daily increase of the validator commission. This parameter is fixed cannot be changed after the `create-validator` transaction is processed.
- **Minimum self-delegation:** Minimum amount of ATOM the validator requires to have bonded at all time. If the validator's self-delegated stake falls below this limit, their entire staking pool is unbonded.

After a validator is created, ATOM holders can delegate ATOM to them, effectively adding stake to the validator's pool. The total stake of an address is the combination of ATOM bonded by delegators and ATOM self-bonded by the validator.

From all validator candidates that signaled themselves, the 150 validators with the most total stake are the designated **validators**. If a validator's total stake falls below the top 150, then that validator loses its validator privileges. The validator cannot participate in consensus or generate rewards until the stake is high enough to be in the top 150. Over time, the maximum number of validators may be increased via on-chain governance proposal.

## Testnet

### How can I join the testnet?

The testnet is a great environment to test your validator setup before launch.

Testnet participation is a great way to signal to the community that you are ready and able to operate a validator. For details, see [Join the Public Testnet](../hub-tutorials/join-testnet.md) documentation.

## Additional Concepts

### What are the different types of keys?

There are two types of keys:

- **Tendermint key**: A unique key that is used to sign consensus votes.
  - It is associated with a public key `cosmosvalconspub` (To get this value, run `gaiad tendermint show-validator`)
  - It is generated when the node is created with `gaiad init`.
- **Application key**: This key is created from the `gaiad` binary and is used to sign transactions. Application keys are associated with a public key that is prefixed by `cosmospub` and an address that is prefixed by `cosmos`. 

The Tendermint key and the application key are derived from account keys that are generated by the `gaiad keys add` command.

**Note:** A validator's operator key is directly tied to an application key and uses the `cosmosvaloper` and `cosmosvaloperpub` prefixes that are reserved solely for this purpose. 

### What are the different states a validator can be in?

After a validator is created with a `create-validator` transaction, the validator is in one of three states:

- `in validator set`: Validator is in the active set and participates in consensus. The validator is earning rewards and can be slashed for misbehavior.
- `jailed`: Validator misbehaved and is in jail, i.e. outside of the validator set. 

  - If the jailing is due to being offline for too long (i.e. having missed more than `95%` out of the last `10,000` blocks), the validator can send an `unjail` transaction in order to re-enter the validator set. 
  - If the jailing is due to double signing, the validator cannot unjail.

- `unbonded`: Validator is not in the active set, and therefore not signing blocs. The validator cannot be slashed and does not earn any reward. It is still possible to delegate ATOM to an unbonded validator. Undelegating from an `unbonded` validator is immediate, meaning that the tokens are not subject to the unbonding period.

### What is self-delegation? How can I increase my self-delegation?

Self-delegation is a delegation of ATOM from a validator to themselves. The delegated amount can be increased by sending a `delegate` transaction from your validator's `application` application key.

### Is there a minimum amount of ATOM that must be delegated to be an active (bonded) validator?

The minimum is 1 ATOM. But the network is currently secured by much higher values. You can check the minimum required ATOM to become part of the active validator set on the [Mintscan validator page](https://www.mintscan.io/cosmos/validators).

### How do delegators choose their validators?

Delegators are free to choose validators according to their own subjective criteria. Selection criteria includes:

- **Amount of self-delegated ATOM:** Number of ATOM a validator self-delegated to themselves. A validator with a higher amount of self-delegated ATOM indicates that the validator is sharing the risk and experienced consequences for their actions.
- **Amount of delegated ATOM:** Total number of ATOM delegated to a validator. A high voting power shows that the community trusts this validator. Larger validators also decrease the decentralization of the network, so delegators are suggested to consider delegating to smaller validators.
- **Commission rate:** Commission applied on revenue by validators before the revenue is distributed to their delegators.
- **Track record:** Delegators review the track record of the validators they plan to delegate to. This track record includes past votes on proposals and historical average uptime.
- **Community contributions:** Another (more subjective) criteria is the work that validators have contributed to the community, such as educational content, participation in the community channels, contributions to open source software, etc.

Apart from these criteria, validators send a `create-validator` transaction to signal a website address to complete their resume. Validators must build reputation one way or another to attract delegators. For example, a good practice for validators is to have a third party audit their setup. Note though, that the Tendermint team does not approve or conduct any audits themselves. For more information on due diligence, see the [A Delegator’s Guide to Staking](https://medium.com/@interchain_io/3d0faf10ce6f) blog post.

## Responsibilities

### Do validators need to be publicly identified?

No, they do not. Each delegator can value validators based on their own criteria. Validators are able to register a website address when they nominate themselves so that they can advertise their operation as they see fit. Some delegators prefer a website that clearly displays the team operating the validator and their resume, while other validators might prefer to be anonymous validators with positive track records.

### What are the responsibilities of a validator?

Validators have two main responsibilities:

- **Be able to constantly run a correct version of the software:** Validators must ensure that their servers are always online and their private keys are not compromised.

- **Actively participate in governance:** Validators are required to vote on every proposal.

Additionally, validators are expected to be active members of the community. Validators must always be up-to-date with the current state of the ecosystem so that they can easily adapt to any change.

### What does 'participate in governance' entail?

Validators and delegators on the Cosmos Hub can vote on proposals to change operational parameters (such as the block gas limit), coordinate upgrades, or make a decision on any given matter.

Validators play a special role in the governance system. As pillars of the system, validators are required to vote on every proposal. It is especially important since delegators who do not vote inherit the vote of their validator.

### What does staking imply?

Staking ATOM can be thought of as a safety deposit on validation activities. When a validator or a delegator wants to retrieve part or all of their deposit, they send an `unbonding` transaction. Then, ATOM undergoes a **3-week unbonding period** during which they are liable to being slashed for potential misbehaviors committed by the validator before the unbonding process started.

Validators, and by association delegators, receive block rewards, fees, and have the right to participate in governance. If a validator misbehaves, a certain portion of their total stake is slashed. This means that every delegator that bonded ATOM to this validator gets penalized in proportion to their bonded stake. Delegators are therefore incentivized to delegate to validators that they anticipate will function safely.

### Can a validator run away with their delegators' ATOM?

By delegating to a validator, a user delegates voting power. The more voting power a validator have, the more weight they have in the consensus and governance processes. This does not mean that the validator has custody of their delegators' ATOM. **A validator cannot run away with its delegator's funds**.

Even though delegated funds cannot be stolen by their validators, delegators' tokens can still be slashed by a small percentage if their validator suffers a [slashing event](#what-are-the-slashing-conditions?), which is why we encourage due diligence when [selecting a validator](#how-do-delegators-choose-their-validators?).

### How often is a validator chosen to propose the next block? Does frequency increase with the quantity of bonded ATOM?

The validator that is selected to propose the next block is called the proposer. Each proposer is selected deterministically. The frequency of being chosen is proportional to the voting power (i.e. amount of bonded ATOM) of the validator. For example, if the total bonded stake across all validators is 100 ATOM and a validator's total stake is 10 ATOM, then this validator is the proposer ~10% of the blocks.

### Are validators of the Cosmos Hub required to validate other zones in the Cosmos ecosystem?

This depends, currently no validators are required to validate other blockchains. But when the first version of [Interchain Security](https://blog.cosmos.network/interchain-security-is-coming-to-the-cosmos-hub-f144c45fb035) is launched on the Cosmos Hub, delegators can vote to have certain blockchains secured via Interchain Security. In those cases, validators are required to validate on these chains as well.

## Incentives

### What is the incentive to stake?

Each member of a validator's staking pool earns different types of revenue:

- **Block rewards:** Native tokens of applications (e.g. ATOM on the Cosmos Hub) run by validators are inflated to produce block provisions. These provisions exist to incentivize ATOM holders to bond their stake. Non-bonded ATOM are diluted over time.
- **Transaction fees:** The Cosmos Hub maintains an allow list of tokens that are accepted as fee payment. The initial fee token is the `atom`.

This total revenue is divided among validators' staking pools according to each validator's weight. Then, within each validator's staking pool the revenue is divided among delegators in proportion to each delegator's stake. A commission on delegators' revenue is applied by the validator before it is distributed.

### What is a validator commission?

Revenue received by a validator's pool is split between the validator and their delegators. The validator can apply a commission on the part of the revenue that goes to their delegators. This commission is set as a percentage. Each validator is free to set their initial commission, maximum daily commission change rate, and maximum commission. The Cosmos Hub enforces the parameter that each validator sets. The maximum commission rate is fixed and cannot be changed. However, the commission rate itself can be changed after the validator is created as long as it does not exceed the maximum commission.

### What is the incentive to run a validator?

Validators earn proportionally more revenue than their delegators because of the commission they take on the staking rewards from their delegators.

Validators also play a major role in governance. If a delegator does not vote, they inherit the vote from their validator. This voting inheritance gives validators a major responsibility in the ecosystem.

### How are block rewards distributed?

Block rewards are distributed proportionally to all validators relative to their voting power. This means that even though each validator gains ATOM with each reward, all validators maintain equal weight over time.

For example, 10 validators have equal voting power and a commission rate of 1%. For this example, the reward for a block is 1000 ATOM and each validator has 20% of self-bonded ATOM. These tokens do not go directly to the proposer. Instead, the tokens are evenly spread among validators. So now each validator's pool has 100 ATOM. These 100 ATOM are distributed according to each participant's stake:

- Commission: `100*80%*1% = 0.8 ATOM`
- Validator gets: `100\*20% + Commission = 20.8 ATOM`
- All delegators get: `100\*80% - Commission = 79.2 ATOM`

Then, each delegator can claim their part of the 79.2 ATOM in proportion to their stake in the validator's staking pool.

### How are fees distributed?

Fees are similarly distributed with the exception that the block proposer can get a bonus on the fees of the block they propose if the proposer includes more than the strict minimum of required precommits.

When a validator is selected to propose the next block, the validator must include at least 2/3 precommits of the previous block. However, an incentive to include more than 2/3 precommits is a bonus. The bonus is linear: it ranges from 1% if the proposer includes 2/3rd precommits (minimum for the block to be valid) to 5% if the proposer includes 100% precommits. Of course the proposer must not wait too long or other validators may timeout and move on to the next proposer. As such, validators have to find a balance between wait-time to get the most signatures and risk of losing out on proposing the next block. This mechanism aims to incentivize non-empty block proposals, better networking between validators, and mitigates censorship.

For a concrete example to illustrate the aforementioned concept, there are 10 validators with equal stake. Each validator applies a 1% commission rate and has 20% of self-delegated ATOM. Now comes a successful block that collects a total of 1025.51020408 ATOM in fees.

First, a 2% tax is applied. The corresponding ATOM go to the reserve pool. The reserve pool's funds can be allocated through governance to fund bounties and upgrades.

- `2% * 1025.51020408 = 20.51020408` ATOM go to the reserve pool.

1005 ATOM now remain. For this example, the proposer included 100% of the signatures in its block so the proposer obtains the full bonus of 5%.

To solve this simple equation to find the reward R for each validator:

`9*R + R + R*5% = 1005 ⇔ R = 1005/10.05 = 100`

- For the proposer validator:
  - The pool obtains `R + R * 5%`: 105 ATOM
  - Commission: `105 * 80% * 1%` = 0.84 ATOM
  - Validator's reward: `105 * 20% + Commission` = 21.84 ATOM
  - Delegators' rewards: `105 * 80% - Commission` = 83.16 ATOM (each delegator is able to claim its portion of these rewards in proportion to their stake)
- For each non-proposer validator:
  - The pool obtains R: 100 ATOM
  - Commission: `100 * 80% * 1%` = 0.8 ATOM
  - Validator's reward: `100 * 20% + Commission` = 20.8 ATOM
  - Delegators' rewards: `100 * 80% - Commission` = 79.2 ATOM (each delegator is able to claim their portion of these rewards in proportion to their stake)

### What are the slashing conditions?

If a validator misbehaves, their delegated stake is partially slashed. Two faults can result in slashing of funds for a validator and their delegators:

- **Double signing:** If someone reports on chain A that a validator signed two blocks at the same height on chain A and chain B, and if chain A and chain B share a common ancestor, then this validator gets slashed by 5% on chain A.
- **Downtime:** If a validator misses more than `95%` of the last `10,000` blocks (roughly ~19 hours), they are slashed by 0.01%.

### Are validators required to self-delegate ATOM?

Yes, they do need to self-delegate at least `1 atom`. Even though there is no obligation for validators to self-delegate more than `1 atom`, delegators want their validator to have more self-delegated ATOM in their staking pool. In other words, validators share the risk.

In order for delegators to have some guarantee about how much shared risk their validator has, the validator can signal a minimum amount of self-delegated ATOM. If a validator's self-delegation goes below the limit that it predefined, this validator and all of its delegators are unbonded.

Note however that it's possible that some validators decide to self-delegate via a different address for security reasons.

### How to prevent concentration of stake in the hands of a few top validators?

The community is expected to behave in a smart and self-preserving way. When a mining pool in Bitcoin gets too much mining power the community usually stops contributing to that pool. The Cosmos Hub relies on the same effect. Additionally, when delegaters switch to another validator, they are not subject to the unbonding period, which removes any barrier to quickly redelegating tokens in service of improving decentralization.

## Technical Requirements

### What are hardware requirements?

A modest level of hardware specifications is initially required and rises as network use increases. Participating in the testnet is the best way to learn more. You can find the current hardware recommendations in the [Joining Mainnet documentation](../hub-tutorials/join-mainnet.md).

Validators are recommended to set up [sentry nodes](https://docs.tendermint.com/master/nodes/validators.html#setting-up-a-validator) to protect your validator node from DDoS attacks.

### What are software requirements?

In addition to running a Cosmos Hub node, validators are expected to implement monitoring, alerting, and management solutions. There are [several tools](https://medium.com/solar-labs-team/cosmos-how-to-monitoring-your-validator-892a46298722) that you can use.

### What are bandwidth requirements?

The Cosmos network has the capacity for very high throughput relative to chains like Ethereum or Bitcoin.

We recommend that the data center nodes connect only to trusted full nodes in the cloud or other validators that know each other socially. This connection strategy relieves the data center node from the burden of mitigating denial-of-service attacks.

Ultimately, as the network becomes more heavily used, multigigabyte per day bandwidth is very realistic.

### How to handle key management?

Validators are expected to run an HSM that supports ed25519 keys. Here are potential options:

- YubiHSM 2
- Ledger Nano S
- Ledger BOLOS SGX enclave
- Thales nShield support

The Interchain Foundation does not recommend one solution above the other. The community is encouraged to bolster the effort to improve HSMs and the security of key management.

### What can validators expect in terms of operations?

Running an effective operation is key to avoiding unexpected unbonding or slashing. Operations must be able to respond to attacks and outages, as well as maintain security and isolation in the data center.

### What are the maintenance requirements?

Validators are expected to perform regular software updates to accommodate chain upgrades and bug fixes. It is suggested to consider using [Cosmovisor](https://docs.cosmos.network/master/run-node/cosmovisor.html) to partially automate this process.

During an chain upgrade, progress is discussed in a private channel in the [Cosmos Developer Discord](https://discord.gg/cosmosnetwork). If your validator is in the active set we encourage you to request access to that channel by contacting a moderator.

### How can validators protect themselves from denial-of-service attacks?

Denial-of-service attacks occur when an attacker sends a flood of internet traffic to an IP address to prevent the server at the IP address from connecting to the internet.

An attacker scans the network, tries to learn the IP address of various validator nodes, and disconnects them from communication by flooding them with traffic.

One recommended way to mitigate these risks is for validators to carefully structure their network topology using a sentry node architecture.

Validator nodes are expected to connect only to full nodes they trust because they operate the full nodes themselves or the trust full nodes are run by other validators they know socially. A validator node is typically run in a data center. Most data centers provide direct links to the networks of major cloud providers. The validator can use those links to connect to sentry nodes in the cloud. This mitigation shifts the burden of denial-of-service from the validator's node directly to its sentry nodes, and can require that new sentry nodes are spun up or activated to mitigate attacks on existing ones.

Sentry nodes can be quickly spun up or change their IP addresses. Because the links to the sentry nodes are in private IP space, an internet-based attack cannot disturb them directly. This strategy ensures that validator block proposals and votes have a much higher change to make it to the rest of the network.

For more sentry node details, see the [Tendermint Documentation](https://docs.tendermint.com/master/nodes/validators.html#setting-up-a-validator) or the [Sentry Node Architecture Overview](https://forum.cosmos.network/t/sentry-node-architecture-overview/454) on the forum.
