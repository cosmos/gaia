# Interchain Security


# Introduction

Interchain Security has been referred to in many different ways: Shared Security, Cross Chain Validation, Cross Chain Collateralization, Shared Staking.  This document will restrict use to the following three terms:



*   **Shared Security**
    *   Shared security is a family of technologies that include optimistic rollups, zk-rollups, sharding and interchain security.
*   **Interchain Security**
    *   Interchain Security is the Cosmos specific category of Shared Security that uses IBC (Inter-Blockchain Communication).
*   **Cross Chain Validation**
    *   Cross Chain Validation is the specific IBC level protocol that enables Interchain Security.

While there are many ways that Interchain Security could take place, this document will focus on one instance of Interchain Security that has particularly valuable attributes for the ATOM token and the Cosmos Hub. The resulting technology may be applied to other scenarios with little to no modification, but I will leave those out for now (or only dedicate a small section) since the current priority is to implement this feature for the Cosmos Hub.

At a very high level, Interchain Security is the ability for staking tokens that have been delegated to validators on a parent chain to inform the composition of a validator set on a child chain. Inter-Blockchain Communication is utilized to relay updates of validator stake delegations from the parent chain to the child chain so that the child chain will have an up-to-date model of which validators can produce blocks on the child chain. The inclusion of parent chain validators can be mandatory or opt-in depending on the requirements of the child chain. The parent chain will honor any proof of validator misbehavior produced by the child chain as evidence that results in slashing the stake of misbehaving validators on the parent chain. In this way the security gained from the value of the stake locked on the parent chain will be shared with the child chain.


# Cosmos Hub User Story

There are two primary reasons that Interchain Security is valuable to the Cosmos Hub. The first reason is because it allows for hub minimalism and the second is to lower the barrier to launching and running secure sovereign decentralized public blockchains. 


## Practical Hub Minimalism

Hub Minimalism is the strategic philosophy that posits that the Cosmos Hub should have as few features as possible in order to decrease the surface area for security vulnerabilities and to reduce the chance of a conflict of interest between user groups. A hub minimalist might be against the governance module from being on the same blockchain as a DEX since users of the governance module must now accommodate users of the DEX even when they have different interests. At best, divergent user groups can peacefully coexist and at worst they may result in hard-forks that diverge in the features of an application.

The current Cosmos Hub is adding more features, which carries some of the risks that hub minimalism is concerned with. Should Interchain Security become available, it would be possible to satisfy  hub minimalists by allowing for each distinct feature of the Cosmos Hub to be an independent chain validated by the same set of ATOM delegated validators. This way the operation of each function could occur independently without affecting the operation of other ATOM secured hub-specific applications.


## Lowering the Barrier to Security

The security of a network is often described as a function of the cost for attacking that network. In Tendermint consensus we target ⅓ and ⅔ of locked stake for various guarantees about liveness and correctness. This means that in order to do any of a variety of attacks against the network, you would need to acquire ⅓+ or ⅔+ of all stake. The crude way to calculate the cost of an attack is to take the quantity of tokens needed to achieve these proportions and multiply it by the current market price for that token. We'll call this the Cost of Corruption.

The Cost of Corruption calculation doesn't account for the availability of any specific token but it does give a very rough estimate for how secure a chain is. It's important that the total value locked (TVL) on a chain remains less than the Cost of Corruption, otherwise the chain should be considered insecure. Since the ability of a chain to serve a valuable purpose is often dependent on the TVL it can handle, it's important to find ways to increase the Cost of Corruption for chains in the Cosmos ecosystem. One method of doing this is to lend the value of the Cosmos ATOM to the Cost of Corruption of any chain. This becomes possible with Interchain Security as the ATOM, which already has a sizeable market cap, can increase the Cost of Corruption of a new chain.


# The Interchain Security Stack

How Interchain Security works at a technical level is still in the process of development but the stack at a high level is well mapped out. It requires new functionality and modifications to current functionality on both parent and child chains. The technology can be developed progressively so that a minimum viable set of functionality can be launched as a V1 before an extended set of functionality is launched as a V2.


## Chain Registry

The first piece of technology to consider is a Chain Registry on the parent chain. This module stores a list of all blockchains that want to use Interchain Security. There are various paths for a chain to join this list or be updated within this list. Much of the questions around authorization and authentication for these actions is the focus of the Cosmos Name Service. The Cosmos Name Service is a general chain registry that allows more information than whether or not the chain uses Interchain Security to be stored and managed. For details about how these problems are being approached please refer to the CNS repository. For the sake of this document, it assumes there exists a credible method for adding and updating entries to this list with regard to a chain's intention to use Interchain Security. This Chain Registry may use the chain-id as a unique identifier.

Within the Chain Registry, each registered child chain stores a list of parent chain validators that have opted in to participating in Interchain Security on behalf of that child chain. It is up to each child chain to advertise the benefits that parent chain validators are receiving by participating. This could be purely the chance to receive fees on a new network, it could be paired with some genesis distribution of tokens, it could be paired with some off chain legal agreement to pay for the staking services, or any other kind of benefit. The Interchain Security technical stack is not required to facilitate this stage of negotiation but may in the future participate by helping with the creation of the genesis file for a new chain.

Regardless of the exact reason, one must assume that each validator has considered the work it takes to run a validator for the child chain in order to produce blocks on that network, and considered the risk of having their parent chain stake (ATOMs) slashed should they do a poor job validating on that network. In order to opt-in, the validator must use their validator key to submit a transaction to the Chain Registry with the intention of being included in a set of validators of a relevant child chain.

The final piece of information relevant to the Chain Registry is the Time Til Launch (TTL) for each child chain. The TTL designates the point of no return for Validators to join or leave the Chain Registry. At that point the list of validators contained within the registry for a specific network can be exported to an initial genesis file for that network. If this network is transitioning to use Interchain Security from a sovereign validator set, this will be seen as a network upgrade where the validator set gets completely redefined.

It will be possible for a Validator to join an already running Interchain Staked network but if they intend to be part of the original genesis validator set they should join before the TTL. A chain within the registry may not produce a TTL until there is a threshold number of validators or amount of stake reached. Future versions of Interchain Security may help automate these thresholds but an initial version will be manually controlled.


## Parent Chain Staking Module

Tendermint uses ABCI to get a set of validators and voting powers from the state machine in order to perform consensus on block production. This information is stored within the staking module of the Cosmos SDK application. In the configuration of Interchain Security, the child chain also has an instance of Tendermint that uses ABCI to ask for a set of validators and their voting powers. However instead of coming directly from the staking module of the child chain, in a sense it needs to come from the staking module of the parent chain. Practically speaking the state of the staking module of the parent chain is relayed periodically via IBC to the child chain, where it is stored in the child chain staking module and accessible to Tendermint via ABCI. 

In order for the parent chain to relay validator sets and voting powers to the child chain, it needs to be able to distinguish between validator sets relevant to different networks. Not all validators within a parent chain staking module will have opted-in to being part of each child chain set. In order to support Interchain Security, the parent chain will need to be extended to differentiate between sets of validators and their voting powers with respect to various chains as designated within the Chain Registry.

To provide the necessary functionality on the parent chain, a wrapper module may need to be implemented that will collate staking module validators with regard to their inclusion in sets stored within the Chain Registry. A module like this will need to import both the Chain Registry Keeper and the Staking Keeper in order to make chain specific staking queries. These queries will be requested periodically by the Cross Chain Validation module and relayed to the child chain via IBC. This module can be referred to as the "xStaking Module".


### Validator set limits

The current Cosmos Hub has a limit of 125 validators. This limit is imposed on validators who are interested in producing blocks as part of a validator set of the Cosmos Hub itself. This limits the number of eligible validators for child chains to the same top 125 Cosmos Hub participants. However, just because a validator doesn't have enough staked ATOM to be eligible to validate on the Cosmos Hub, doesn't mean that they shouldn't qualify to validate on another child chain.

Interchain Security should increase the diversity of the validator ecosystem by lowering the barrier to running a profitable validator business. This will go far in creating a healthy ecosystem of diverse validators that will result in anti-fragile and robustly operated networks. In order to make it possible for the top 125 validators to remain eligible as block producers for the Cosmos Hub while increasing the number of eligible validators for child chains, the Staking Module needs to stop forcing validators to undelegate when they leave the top set of 125 validators. This will result in a longer list of validators with ATOM delegations that are not participating in block production on the parent chain (Cosmos Hub). These extra validators will however be eligible to produce blocks on child chains and use their delegated ATOMs to earn rewards on the child chains as well as risk their parent chain ATOMs to slashable events should they misbehave on child chains.


### Chain-Specific Delegations

In order to further fulfil the goal of creating a diverse set of validators with healthy competition it is important to work towards chain-specific delegations. As described above and for the initial version of Interchain Security, the chain-specific validation calculation is determined by the validator opting in to being included in the child chain validator set. That means that all the individual delegations made to that validator are included as part of the decision. Similar to how a validator is able to decide its own commission rate, one may decide that it is the prerogative of that validator to make this decision on behalf of its delegators. Luckily if a delegator disagrees with the choice a validator has made on their behalf, they can redelegate to a validator that is better aligned with the wishes of the delegator.

Initially it is up to validators to evaluate whether participation in a child chain validator set is worth the extra work and risk it imposes. Subsequently it is up to a delegator to decide whether the choices of their validator are aligned with their own. However it may be the case that a delegator agrees with how a validator operates on the parent chain and wishes to stay delegated there, but disagrees with the choice to participate (or not) on a child chain.

Initially the cost to validators to operate a new node instance on a new network may prohibit smaller validators from participating. It could be imagined that validating is a business of scale and the larger the operation the easier it is to scale further. This may result in only the largest validators from participating in Interchain Security. In order to ensure that delegators don't all redelegate their ATOMs to these super validators, it's important to build out the ability to delegate to validators on a per-chain basis. This would allow a delegator to delegate ATOMs to a small but well run validator on the Hub while also delegating the same ATOMs to a larger and less risk averse validator for a child chain at the same time. Should the larger validator misbehave and be slashed on the parent chain, the stake of that specific delegator would also be slashed. This would decrease the amount of stake the original small validator has delegated, but only with regard to that one delegator.

The optionality around complex delegations would eventually increase the possible diversity of validator operations. However due to the complexity of such delegations, this functionality is assumed to be part of a V2 of Interchain Security.


### Epoch Staking

The current Staking Module of the Cosmos SDK is moving towards Epoch based staking. This means that instead of validator set delegation amounts being calculated on a per block basis, they will be calculated over some group of time (or blocks) called an Epoch. This will decrease the number of times staking is calculated and generally decrease the complexity involved in staking. The additional complexity of chain relevant stake calculations will similarly benefit from a general simplification of stake calculations, and it will require less packets to be sent between the chains.


## Parent Chain Distribution Module

The distribution module of the Cosmos SDK uses a system called F1 to keep track of how much delegators have delegated to different validators and for how long they've been doing it. When a delegator wants to withdraw rewards, the distribution module calculates the number of rewards received since the delegator last withdrew, and calculates the outcome based on the amount of stake that belongs to the delegator compared to the total stake of the validator. All of this takes place only with respect to the validation of the native chain and the native rewards.

In order for Interchain Security to properly reward validators and their delegators with rewards from child chains, the distribution module will need to be extended. Currently the distribution module imports the staking module, and when `WithdrawDelegationRewards` is called the total delegations are pulled from the staking module keeper before being provided as a parameter to the eventual distribution keeper method `withdrawDelegationRewards`. This method should be extended to contain a chain-id as recorded in the Chain Registry in order to differentiate which pool of rewards are being distributed.

If and when the xStaking module contains complex delegations, the stakingKeeper method for calculating delegations would also need to include the Chain Registry reference to chain-id in order to properly calculate the delegation rewards with regard to the specific child chain. In the meantime, for Interchain Security V1 this is unnecessary.


## IBC & Cross Chain Validation

There are a number of IBC application layer modules and packets that need to be developed to fully realize the IBC component of Interchain Security. This work has begun with a spec draft from Informal Systems that is visible at [@informalsystems/cross-chain-validation](https://github.com/informalsystems/cross-chain-validation/). Instead of diving into the details of what they are and exactly how they work this section will be reserved for high level responsibilities of these mechanisms.

There are three types of operations within Cross Chain Validation which must be present for Interchain Security to take place:



*   Validator Set Updates
*   Reward Distribution
*   Evidence


### Validator Set Updates

The primary duty of Cross Chain Validation is to relay the set of validators and their voting power. The inclusion of a validator within a set relevant to a specific child chain is designated within the Chain Registry. The voting power denominated in parent chain staking token is designated within the staking module. The xStaking module allows for the collation of validators and their delegations on a chain specific designation. This collation is what must be relayed to the child chain via Cross Chain Validation.

The rate at which these validator set updates are relayed is a function of safety. At one extreme you could imagine the validator set being collated and relayed with every single epoch that is produced on the parent chain. This would ensure that the child chain has an absolutely accurate representation of validator weights at every potential state update within the parent chain. However, since delegations are subject to unbonding periods it is possible to approach state updates more conservatively. At another extreme, one may reason that if there is no active stake unbonding happening on the parent chain, it may be assumed that a validator set update will not be possible within a maximum duration of that unbonding period and so a validator set update can wait until just before that moment. Based on the possibility of instant redelegations this assumption may need to be further adjusted. 

The process of recording and relaying validator set updates within safe and correct periods is the focus of the spec and research at Informal Systems. We can assume that the design will be aware of the necessary cadence that validator set updates must take place to ensure safe operation. When there are no updates it may be necessary to simply acknowledge this with something like a heartbeat IBC packet.


### Reward Distribution

While Interchain Security may remove the necessity for a staking token to play a role in the token design of new blockchains, it can be assumed that there will be some sort of economic system in place to reward validators for producing blocks. To follow the default capabilities of the Cosmos SDK one would assume that there is a simple inflationary reward token attached to the production of blocks. This token may also be a governance token and the value may be implied by the ability to govern an otherwise useful and valuable network. The token could have many uses or reasons to exist, but being the responsibility of each blockchain to design its own purpose, we will assume some sort of reward token exists and is used to reward validators for the production of blocks.

In the current system rewards are pooled into the distribution module account. The distribution module imports the staking module to keep track of validator weights in order to calculate and distribute rewards on a per-delegator basis. For this to take place on the child chain it would be necessary for not only the validator set and validator voting power to be relayed with cross chain validation, but also all constituent delegators that compose each validator. This would result in an extremely large IBC packet which would make regular communication difficult if not impossible. Rather than relay all delegators, this information should stay on the parent chain within the parent chain staking module. 

Instead of distributing rewards on the child chain, the rewards collected for each block produced should be transferred to the parent chain at some interval. It may be a regular interval or a user initiated IBC packet, but would be similar to an ics-20 token transfer from the distribution module account of the child chain to the distribution module account of the parent chain. The difference between a standard ics-20 transfer being that the parent chain distribution module account needs to know which chain the rewards came from and over which period they were collected. This information will be necessary to perform the parent chain distribution module's responsibility of distributed rewards to delegators and validators on a per chain basis informed by the Chain Registry.


### Evidence

In a single chain system, a validator may misbehave in various ways that result in the stake attached to that validator being slashed. This can occur automatically within the state machine or only after evidence of the misbehaviour has been collected and submitted to the chain. In a scenario where the validator set for a child chain is secured by stake that exists only on the parent chain, the evidence of misbehaviour needs to be submitted to the parent chain, where the tokens at stake are able to be slashed. Similar to a single chain system, this may take place automatically or only with the manual submission of evidence. If it were to occur automatically it would mean an outgoing IBC packet could be automatically generated but would still need to be manually relayed to the parent chain where punishment could follow. The more manual scenario would mean that the misbehaviour on the child chain results in evidence which is submitted directly to the parent chain, or submitted to the child chain which if successfully processed would result in an outgoing IBC packet containing the instruction to slash at the level of the parent chain.

Either way comes with the question of whether it should be the parent chain that is verifying the evidence of slashable events, or simply trusting the child chain's commands to slash validators at the level of the parent. While there may be the ability to both, trusting the child chain makes the job of the parent chain much easier. In that scenario the parent chain doesn't need to have knowledge of the logic contained within the child chain that may be necessary for determining whether a slashable offense took place. Trusting the child chain to enforce its own rules puts a lot of trust in the child chain to not abuse the position and submit a wave of slashable commands that may reduce a validator to nothing. However this is ultimately the responsibility of the validator to ensure that the slashable risk of validating on a child chain is worth the expected rewards. This includes determining whether the state machine of the child chain includes logic that the validator is comfortable with.

Limits on a child chain's ability to slash a parent chain validator may also be imposed at the parent chain level. For instance it is possible to make it a required parameter of a child chain within the Chain Registry to provide a maximum slashing rate. This would ensure that a validator knows even if they violate all the rules of a child chain, they can still limit the damage to their stake by some amount. It is also possible to enforce a parameter at the level of the parent chain that prevents a validator from joining any combination of child chain validator sets that results in a combined slashable rate over some threshold. For instance if there are 3 child chains that each have maximum slashable rates of 33% over X blocks, and the parent chain has a safety limit for validators that require them to stay below a total of 90% slashable events over X blocks, no validator would be allowed to join the Chain Registry for all 3 of those child chains. This parameter on the parent chain should be a governance parameter that can be adjusted by token staked voting to reflect the risk appetite of the actual validator set of the parent chain.

Consideration should be made to the incentives around submitting evidence in order to ensure that punishable offenses do not go unnoticed. This is a similar issue to Relayer Incentivization, where currently it costs some fee to relay IBC packets but there is no reward or payment possible as part of the core IBC logic. As a result IBC packets are currently relayed in an altruistic manner. It's important to ensure that for integrity of the operation of the child and parent chain that slashable offenses are always submitted. It may be that the slashable amounts of tokens are used as rewards for submitting the evidence (assuming it's possible to ensure it is not the culprit submitting their own evidence and regaining their stake). Maybe some flat fee to at least pay for the transactions are required, although this may become redundant with the addition of any Relayer incentivizations into core IBC.


## Child Chain Staking Module

With Interchain Security, the child chain receives updates of the parent chain's set of validator's voting power via cross chain validation IBC packets. These updates are used to populate the child chain's staking module. On the child chain, Tendermint consensus is running which asks the child chain staking module for the current list of validators and their voting power. This could work virtually the same as a traditional configuration without Interchain Security, as from the perspective of Tendermint the flow is the same (ABCI asks the staking module for this information).

Instead of aligning as closely as possible with the current staking module, a wrapper staking module could be made with additional functionality. This would allow child chains to create their own staking design that extends the validator set received from the parent chain. For example the validator set from the parent chain could be stored as it is received in a staking-like module. In addition another module could be used to track delegations to an overlapping set of validators using a secondary staking token. The actual set of validators and their voting power could be a combination of these two sets and passed to Tendermint via ABCI. This combination could be customized per chain or as per need, for instance it may just boost the power of the parent chain validators or eclipse the set as desired. This set of functionality is possible but should not be considered for V1 of Interchain Security.


## Child Chain Distribution Module

Delegators and validators on the parent chain would not risk their ATOMs in order to be included in the validator set on a child chain unless there was some incentive to do so. At a minimum, transaction fees gained from processing transactions could be considered as a possible incentive. There may also be incentives completely outside the state machine, like a service agreement that comes with regular payments through traditional means of money transmission. More likely it is expected that child chains include some type of block reward as seen in traditional validation schemes. This could be in a child-chain specific token used for gas, governance or some other use.

Regardless of how exactly a reward is calculated it is left up to the child chain to design a system that is attractive enough for parent chain validators to risk their staking token in order to be eligible. Once that reward is calculated it needs to be distributed back to the validators that have earned it. These rewards could be deposited on the child chain or the parent chain. If deposited on the child chain it would only be possible to record to which validator they were rewarded and over which time period since the per-delegator metrics are stored only on the parent chain.

In order to allow delegators to have a similar reward distribution as they currently experience, rewards should be regularly transferred to the parent chain where they can utilize the Parent Chain Distribution Module that tracks Cross Chain Validation client connections. These would be sent to the parent chain via a CCV IBC transaction similar to a traditional IBC Transfer packet but send to the Distribution Module Account with relevant information as to which validator and chain they were earned on and over what period of time. This information would allow the Parent Chain Distribution Module to distribute the rewards to the constituent delegators of each validator.


# Roll-Out and Open Questions

Interchain Security consists of many moving pieces, each of which has a variable scope of functionality. Throughout this document the various capabilities are referred to as V1 and V2. Realistically there is an even simpler version of Interchain Security that could be referred to as V0. This would remove the need for a Registry Module by making cross chain validation mandatory for all validators. Participation of a child chain could be determined in the genesis file of the parent chain or via governance. It would raise the barrier of entry to being a Cosmos Hub validator but potentially shorten the time to deliver a working implementation of Interchain Security. This could be contentious with validators as it removes their autonomy in deciding on which networks they participate as valiators, but it also increases the Cost of Corruption for the child chain to be equal to that of the parent chain.

There are further outstanding questions including the exact implementation details for each of the modules included in the Interchain Security stack. These further give rise to expected user flows for each step as well as edge cases like:



*   Child or parent chain halting
*   Child of parent chain upgrading
*   Contentious forks of either parent of child chain
*   Versions of IBC on each side fall out of sync

Other open questions include addressing the degree of risk this configuration adds to the parent chain. Should it be possible for a parent chain validator to validate on a large number of child chains? To what extent should this be a choice of the Validator or a limit imposed by the parent chain state machine? If the Validator is exposed to slashing conditions of too many child chains, could this endanger the security of the parent chain or is it the responsibility of the delegator to take that risk into account?

The economic cost of validating a new chain is another open question that will be important in determining the viability of this offering. Will the cost to validate on a large number of child chains be prohibitive to smaller validators, or will it be the edge that smaller validators can take to compete against larger validators and exchanges?


# Conclusion

Interchain Security is a new shared security primitive that has implications for the security and scalability of single blockchains like the Cosmos Hub as well as the potential to dramatically lower the barrier to running secure public blockchains for new applications. It could be thought of as a competitive configuration to sharding and put the Cosmos Hub on par with Eth 2.0 or Polkadot in terms of their security offerings to applications included in their environment. The design philosophy around the Cosmos ecosystem has always prioritized autonomy and sovereignty over guarantees around security—"Bring your own security" is a term that has been used in the past. The design of Interchain Security discussed here tries to incorporate the Cosmos design philosophy of autonomy and sovereignty with the offering of shared security. Even within the balance between those considerations there is a spectrum of possibilities and it might be the case that multiple versions and configurations are implemented in parallel in order to satisfy as many needs as possible.
