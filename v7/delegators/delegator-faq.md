<!--
order: 2
-->

# Delegator FAQ

## What is a delegator?

People that cannot or do not want to operate [validator nodes](../validators/overview.md) can still participate in the staking process as delegators. Indeed, validators are not chosen based on their self-delegated stake but based on their total stake, which is the sum of their self-delegated stake and of the stake that is delegated to them. This is an important property, as it makes delegators a safeguard against validators that exhibit bad behavior. If a validator misbehaves, their delegators will move their Atoms away from them, thereby reducing their stake. Eventually, if a validator's stake falls under the top 125 addresses with highest stake, they will exit the validator set.

**Delegators share the revenue of their validators, but they also share the risks.** In terms of revenue, validators and delegators differ in that validators can apply a commission on the revenue that goes to their delegator before it is distributed. This commission is known to delegators beforehand and can only change according to predefined constraints (see [section](#choosing-a-validator) below). In terms of risk, delegators' Atoms can be slashed if their validator misbehaves. For more, see [Risks](#risks) section.

To become delegators, Atom holders need to send a ["Delegate transaction"](./delegator-guide-cli.md#sending-transactions) where they specify how many Atoms they want to bond and to which validator. A list of validator candidates will be displayed in Cosmos Hub explorers. Later, if a delegator wants to unbond part or all of their stake, they needs to send an "Unbond transaction". From there, the delegator will have to wait 3 weeks to retrieve their Atoms. Delegators can also send a "Rebond Transaction" to switch from one validator to another, without having to go through the 3 weeks waiting period. 

For a practical guide on how to become a delegator, click [here](./delegator-guide-cli.md).

## Choosing a validator

In order to choose their validators, delegators have access to a range of information directly in [Lunie](https://lunie.io) or other Cosmos block explorers.

- **Validator's moniker**: Name of the validator candidate.
- **Validator's description**: Description provided by the validator operator.
- **Validator's website**: Link to the validator's website.
- **Initial commission rate**: The commission rate on revenue charged to any delegator by the validator (see below for more detail).
- **Commission max change rate:** The maximum daily increase of the validator's commission. This parameter cannot be changed by the validator operator. 
- **Maximum commission:** The maximum commission rate this validator candidate can charge. This parameter cannot be changed by the validator operator. 
- **Minimum self-bond amount**: Minimum amount of Atoms the validator candidate need to have bonded at all time. If the validator's self-bonded stake falls below this limit, their entire staking pool (i.e. all its delegators) will unbond. This parameter exists as a safeguard for delegators. Indeed, when a validator misbehaves, part of their total stake gets slashed. This included the validator's self-delegateds stake as well as their delegators' stake. Thus, a validator with a high amount of self-delegated Atoms has more skin-in-the-game than a validator with a low amount. The minimum self-bond amount parameter guarantees to delegators that a validator will never fall below a certain amount of self-bonded stake, thereby ensuring a minimum level of skin-in-the-game. This parameter can only be increased by the validator operator. 

## Directives of delegators

Being a delegator is not a passive task. Here are the main directives of a delegator:

- **Perform careful due diligence on validators before delegating.** If a validator misbehaves, part of their total stake, which includes the stake of their delegators, can be slashed. Delegators should therefore carefully select validators they think will behave correctly.
- **Actively monitor their validator after having delegated.** Delegators should ensure that the validators they delegate to behave correctly, meaning that they have good uptime, do not double sign or get compromised, and participate in governance. They should also monitor the commission rate that is applied. If a delegator is not satisfied with its validator, they can unbond or switch to another validator (Note: Delegators do not have to wait for the unbonding period to switch validators. Rebonding takes effect immediately).
- **Participate in governance.** Delegators can and are expected to actively participate in governance. A delegator's voting power is proportional to the size of their bonded stake. If a delegator does not vote, they will inherit the vote of their validator(s). If they do vote, they override the vote of their validator(s). Delegators therefore act as a counterbalance to their validators.

## Revenue

Validators and delegators earn revenue in exchange for their services. This revenue is given in three forms:

- **Block provisions (Atoms):** They are paid in newly created Atoms. Block provisions exist to incentivize Atom holders to stake. The yearly inflation rate is calculated to target 2/3 bonded stake. If the total bonded stake in the network is less than 2/3 of the total Atom supply, inflation increases until it reaches 20%. If the total bonded stake is more than 2/3 of the Atom supply, inflation decreases until it reaches 7%. This means that if total bonded stake stays less than 2/3 of the total Atom supply for a prolonged period of time, unbonded Atom holders can expect their Atom value to deflate by 20% (compounded) per year.
- **Transaction fees (various tokens):** Each transfer on the Cosmos Hub comes with transactions fees. These fees can be paid in any currency that is whitelisted by the Hub's governance. Fees are distributed to bonded Atom holders in proportion to their stake. The first whitelisted token at launch is the ATOM.

## Validator Commission

Each validator receives revenue based on their total stake. Before this revenue is distributed to delegators, the validator can apply a commission. In other words, delegators have to pay a commission to their validators on the revenue they earn. Let us look at a concrete example:

We consider a validator whose stake (i.e. self-delegated stake + delegated stake) is 10% of the total stake of all validators. This validator has 20% self-delegated stake and applies a commission of 10%. Now let us consider a block with the following revenue:

- 990 Atoms in block provisions
- 10 Atoms in transaction fees.

This amounts to a total of 1000 Atoms and 100 Photons to be distributed among all staking pools.

Our validator's staking pool represents 10% of the total stake, which means the pool obtains 100 Atoms and 10 Photons. Now let us look at the internal distribution of revenue:

- Commission = `10% * 80% * 100` Atoms = 8 Atoms 
- Validator's revenue = `20% * 100` Atoms + Commission = 28 Atoms
- Delegators' total revenue = `80% * 100` Atoms - Commission = 72 Atoms

Then, each delegator in the staking pool can claim their portion of the delegators' total revenue.

## Risks

Staking Atoms is not free of risk. First, staked Atoms are locked up, and retrieving them requires a 3 week waiting period called unbonding period. Additionally, if a validator misbehaves, a portion of their total stake can be slashed (i.e. destroyed). This includes the stake of their delegators.

There is one main slashing condition:

- **Double signing:** If someone reports on that a validator signed two different blocks with the same chain ID at the same height, this validator will get slashed. 

This is why Atom holders should perform careful due diligence on validators before delegating. It is also important that delegators actively monitor the activity of their validators. If a validator behaves suspiciously or is too often offline, delegators can choose to unbond from them or switch to another validator. **Delegators can also mitigate risk by distributing their stake across multiple validators.**s
