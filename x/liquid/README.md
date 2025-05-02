# `x/liquid`

## Abstract

This module enables the Cosmos Hub to support the issuance of native liquid staking denoms. 


## Contents

* [State](#state)
    * [TotalLiquidStakedTokens](#totalliquidstakedtokens)
    * [PendingTokenizeShareAuthorizations](#pendingtokenizeshareauthorizations)
* [Messages](#messages)
    * [MsgUpdateParams](#msgupdateparams)
    * [MsgTokenizeShares](#msgtokenizeshares)
    * [MsgRedeemTokensForShares](#msgredeemtokensforshares)
    * [MsgTransferTokenizeShareRecord](#msgtransfertokenizesharerecord)
    * [MsgEnableTokenizeShares](#msgenabletokenizeshares)
    * [MsgDisableTokenizeShares](#msgdisabletokenizeshares)
    * [MsgWithdrawTokenizeShareRecordReward](#msgwithdrawtokenizesharerecordreward)
    * [MsgWithdrawAllTokenizeShareRecordReward](#msgwithdrawalltokenizesharerecordreward)
* [Begin-Block](#begin-block)
    * [Expire Tokenize Share Locks](#removeexpiredtokenizesharelocks)
* [Events](#events)
    * [EndBlocker](#endblocker)
    * [Msg's](#msgs)
* [Parameters](#parameters)
* [Client](#client)
    * [CLI](#cli)
    * [gRPC](#grpc)
    * [REST](#rest)

## State

### Params

The liquid module stores its params in state with the prefix of `0x51`,
it can be updated with governance or the address with authority.

* Params: `0x51 | ProtocolBuffer(Params)`

```protobuf reference
https://github.com/cosmos/gaia/blob/main/proto/gaia/liquid/v1beta1/liquid.proto#L13-L35
```

### TotalLiquidStakedTokens

TotalLiquidStakedTokens stores the total liquid staked tokens monitoring the progress towards the `GlobalLiquidStakingCap`.

* TotalLiquidStakedTokens: `0x5 -> math.Int`. 


### PendingTokenizeShareAuthorizations

PendingTokenizeShareAuthorizations stores a queue of addresses that have their tokenize share re-enablement/unlocking in progress. When an address is enqueued, it will sit for 1 unbonding period before the tokenize share lock is removed.

```go reference
https://github.com/cosmos/gaia/blob/main/proto/gaia/liquid/v1beta1/liquid.proto#L48-L50
```

## Messages

In this section we describe the processing of the staking messages and the corresponding updates to the state. All created/modified state objects specified by each message are defined within the [state](#state) section.

### MsgCreateValidator

A validator is created using the `MsgCreateValidator` message.
The validator must be created with an initial delegation from the operator.

```protobuf reference
https://github.com/cosmos/gaia/blob/main/proto/gaia/liquid/v1beta1/tx.proto#L20-L21
```

```protobuf reference
https://github.com/cosmos/gaia/blob/main/proto/gaia/liquid/v1beta1/tx.proto#L50-L73
```

This message is expected to fail if:

* another validator with this operator address is already registered
* another validator with this pubkey is already registered
* the initial self-delegation tokens are of a denom not specified as the bonding denom
* the commission parameters are faulty, namely:
    * `MaxRate` is either > 1 or < 0
    * the initial `Rate` is either negative or > `MaxRate`
    * the initial `MaxChangeRate` is either negative or > `MaxRate`
* the description fields are too large

This message creates and stores the `Validator` object at appropriate indexes.
Additionally a self-delegation is made with the initial tokens delegation
tokens `Delegation`. The validator always starts as unbonded but may be bonded
in the first end-block.

### MsgEditValidator

The `Description`, `CommissionRate` of a validator can be updated using the
`MsgEditValidator` message.

```protobuf reference
https://github.com/cosmos/gaia/blob/main/proto/gaia/liquid/v1beta1/tx.proto#L23-L24
```

```protobuf reference
https://github.com/cosmos/gaia/blob/main/proto/gaia/liquid/v1beta1/tx.proto#L78-L97
```

This message is expected to fail if:

* the initial `CommissionRate` is either negative or > `MaxRate`
* the `CommissionRate` has already been updated within the previous 24 hours
* the `CommissionRate` is > `MaxChangeRate`
* the description fields are too large

This message stores the updated `Validator` object.

### MsgDelegate

Within this message the delegator provides coins, and in return receives
some amount of their validator's (newly created) delegator-shares that are
assigned to `Delegation.Shares`.

```protobuf reference
https://github.com/cosmos/gaia/blob/main/proto/gaia/liquid/v1beta1/tx.proto#L26-L28
```

```protobuf reference
https://github.com/cosmos/gaia/blob/main/proto/gaia/liquid/v1beta1/tx.proto#L102-L114
```

This message is expected to fail if:

* the validator does not exist
* the `Amount` `Coin` has a denomination different than one defined by `params.BondDenom`
* the exchange rate is invalid, meaning the validator has no tokens (due to slashing) but there are outstanding shares
* the amount delegated is less than the minimum allowed delegation
* the delegator is a liquid staking provider and the delegation exceeds
either the `GlobalLiquidStakingCap`, the `ValidatorLiquidStakingCap` or the validator bond cap.

If an existing `Delegation` object for provided addresses does not already
exist then it is created as part of this message otherwise the existing
`Delegation` is updated to include the newly received shares.

If the delegation if is a validator bond, the `ValidatorBondShares` of the validator is increased.

If the delegator is a liquid staking provider, the `TotalLiquidStakedTokens`
and the validator `LiquidShares` are incremented.

The delegator receives newly minted shares at the current exchange rate.
The exchange rate is the number of existing shares in the validator divided by
the number of currently delegated tokens.

The validator is updated in the `ValidatorByPower` index, and the delegation is
tracked in validator object in the `Validators` index.

It is possible to delegate to a jailed validator, the only difference being it
will not be added to the power index until it is unjailed.

![Delegation sequence](https://raw.githubusercontent.com/cosmos/cosmos-sdk/release/v0.46.x/docs/uml/svg/delegation_sequence.svg)

### MsgUndelegate

The `MsgUndelegate` message allows delegators to undelegate their tokens from
validator.

```protobuf reference
https://github.com/cosmos/gaia/blob/main/proto/gaia/liquid/v1beta1/tx.proto#L34-L36
```

```protobuf reference
https://github.com/cosmos/gaia/blob/main/proto/gaia/liquid/v1beta1/tx.proto#L140-L152
```

This message returns a response containing the completion time of the undelegation:

```protobuf reference
https://github.com/cosmos/gaia/blob/main/proto/gaia/liquid/v1beta1/tx.proto#L154-L158
```

This message is expected to fail if:

* the delegation doesn't exist
* the validator doesn't exist
* the delegation has less shares than the ones worth of `Amount`
* existing `UnbondingDelegation` has maximum entries as defined by `params.MaxEntries`
* the `Amount` has a denomination different than one defined by `params.BondDenom`
* the unbonded delegation is a `ValidatorBond` and the reduction in validator bond would cause the existing liquid delegation to exceed the cap.

When this message is processed the following actions occur:

* if the delegation is a validator bond, the `ValidatorBondShares` of the validator is decreased.
 
* if the delegator is a liquid staking provider, the `TotalLiquidStakedTokens`
and the validator's `LiquidShares` are decreased.

* validator's `DelegatorShares` and the delegation's `Shares` are both reduced by the message `SharesAmount`
* calculate the token worth of the shares remove that amount tokens held within the validator
* with those removed tokens, if the validator is:
    * `Bonded` - add them to an entry in `UnbondingDelegation` (create `UnbondingDelegation` if it doesn't exist) with a completion time a full unbonding period from the current time. Update pool shares to reduce BondedTokens and increase NotBondedTokens by token worth of the shares.
    * `Unbonding` - add them to an entry in `UnbondingDelegation` (create `UnbondingDelegation` if it doesn't exist) with the same completion time as the validator (`UnbondingMinTime`).
    * `Unbonded` - then send the coins the message `DelegatorAddr`
* if there are no more `Shares` in the delegation, then the delegation object is removed from the store
    * under this situation if the delegation is the validator's self-delegation then also jail the validator.

![Unbond sequence](https://raw.githubusercontent.com/cosmos/cosmos-sdk/release/v0.46.x/docs/uml/svg/unbond_sequence.svg)

### MsgCancelUnbondingDelegation

The `MsgCancelUnbondingDelegation` message allows delegators to cancel the `unbondingDelegation` entry and delegate back to a previous validator.

```protobuf reference
https://github.com/cosmos/gaia/blob/main/proto/gaia/liquid/v1beta1/tx.proto#L38-L42
```

```protobuf reference
https://github.com/cosmos/gaia/blob/main/proto/gaia/liquid/v1beta1/tx.proto#L160-L175
```

This message is expected to fail if:

* the `unbondingDelegation` entry is already processed.
* the `cancel unbonding delegation` amount is greater than the `unbondingDelegation` entry balance.
* the `cancel unbonding delegation` height doesn't exist in the `unbondingDelegationQueue` of the delegator.

When this message is processed the following actions occur:

* if the `unbondingDelegation` Entry balance is zero
    * in this condition `unbondingDelegation` entry will be removed from `unbondingDelegationQueue`.
    * otherwise `unbondingDelegationQueue` will be updated with new `unbondingDelegation` entry balance and initial balance
* the validator's `DelegatorShares` and the delegation's `Shares` are both increased by the message `Amount`.

### MsgBeginRedelegate

The redelegation command allows delegators to instantly switch validators. Once
the unbonding period has passed, the redelegation is automatically completed in
the EndBlocker.

```protobuf reference
https://github.com/cosmos/gaia/blob/main/proto/gaia/liquid/v1beta1/tx.proto#L30-L32
```

```protobuf reference
https://github.com/cosmos/gaia/blob/main/proto/gaia/liquid/v1beta1/tx.proto#L119-L132
```

This message returns a response containing the completion time of the redelegation:

```protobuf reference
https://github.com/cosmos/gaia/blob/main/proto/gaia/liquid/v1beta1/tx.proto#L133-L138
```

This message is expected to fail if:

* the delegation doesn't exist
* the source or destination validators don't exist
* the delegation has less shares than the ones worth of `Amount`
* the source validator has a receiving redelegation which is not matured (aka. the redelegation may be transitive)
* existing `Redelegation` has maximum entries as defined by `params.MaxEntries`
* the `Amount` `Coin` has a denomination different than one defined by `params.BondDenom`
* the delegation is a `ValidatorBond` and the reduction in validator bond would cause the existing liquid delegation to exceed the cap.
* the delegator is a liquid staking provider and the delegation exceeds
either the `GlobalLiquidStakingCap`, the `ValidatorLiquidStakingCap` or the validator bond cap.

When this message is processed the following actions occur:

* if the delegation if is a validator bond, the `ValidatorBondShares` of the source validator is decreased.
* if the delegator is a liquid staking provider,
 the source validator's `LiquidShares` increased and the destination validator's `LiquidShares` is decreased.
* the source validator's `DelegatorShares` and the delegations `Shares` are both reduced by the message `SharesAmount`
* calculate the token worth of the shares remove that amount tokens held within the source validator.
* if the source validator is:
    * `Bonded` - add an entry to the `Redelegation` (create `Redelegation` if it doesn't exist) with a completion time a full unbonding period from the current time. Update pool shares to reduce BondedTokens and increase NotBondedTokens by token worth of the shares (this may be effectively reversed in the next step however).
    * `Unbonding` - add an entry to the `Redelegation` (create `Redelegation` if it doesn't exist) with the same completion time as the validator (`UnbondingMinTime`).
    * `Unbonded` - no action required in this step
* Delegate the token worth to the destination validator, possibly moving tokens back to the bonded state.
* if there are no more `Shares` in the source delegation, then the source delegation object is removed from the store
    * under this situation if the delegation is the validator's self-delegation then also jail the validator.

![Begin redelegation sequence](https://raw.githubusercontent.com/cosmos/cosmos-sdk/release/v0.46.x/docs/uml/svg/begin_redelegation_sequence.svg)

## MsgTokenizeShares

The `MsgTokenizeShares` message allows users to tokenize their delegated tokens. Share tokens have denom using the validator address and record id of the underlying delegation with the format `{validatorAddress}/{recordId}`.

```protobuf reference
https://github.com/cosmos/gaia/blob/v0.45.16-ics-lsm/proto/gaia/liquid/v1beta1/tx.proto#L49-L50
```

```protobuf reference
https://github.com/cosmos/gaia/blob/v0.45.16-ics-lsm/proto/gaia/liquid/v1beta1/tx.proto#L190-L199
```

This message returns a response containing the number of tokens generated:

```protobuf reference
https://github.com/cosmos/gaia/blob/v0.45.16-ics-lsm/proto/gaia/liquid/v1beta1/tx.proto#L201-L204
```

This message is expected to fail if:

* the delegation is a `ValidatorBond`
* the delegator sender's address has disabled tokenization, meaning that the account 
lock status is either `LOCKED` or `LOCK_EXPIRING`.
* the account is a vesting account and the free delegation (non-vesting delegation) is exceeding the tokenized share amount.
* the sender is NOT a liquid staking provider and the tokenized shares exceeds 
either the `GlobalLiquidStakingCap`, the `ValidatorLiquidStakingCap` or the validator bond cap.


When this message is processed the following actions occur:

* If delegator is a NOT liquid staking provider (otherwise the shares are already included)
    * Increment the `GlobalLiquidStakingCap`
    * Increment the validator's `ValidatorLiquidStakingCap`
* Unbond the delegation shares and transfer the coins back to delegator
* Create an equivalent amount of tokenized shares that the initial delegation shares
* Mint the liquid coins and send them to delegator
* Create a tokenized share record
* Get validator to whom the sender delegated his shares 
* Send coins to module address and delegate them to the validator

## MsgRedeemTokensForShares

The `MsgRedeemTokensForShares` message allows users to redeem their native delegations from share tokens.


```protobuf reference
https://github.com/cosmos/gaia/blob/v0.45.16-ics-lsm/proto/gaia/liquid/v1beta1/tx.proto#L52-L54
```

```protobuf reference
https://github.com/cosmos/gaia/blob/v0.45.16-ics-lsm/proto/gaia/liquid/v1beta1/tx.proto#L206-L213
```

This message returns a response containing the amount of staked tokens redeemed:

```protobuf reference
https://github.com/cosmos/gaia/blob/v0.45.16-ics-lsm/proto/gaia/liquid/v1beta1/tx.proto#L215-L218
```

This message is expected to fail if:

* if the sender's balance doesn't have enough liquid tokens 


When this message is processed the following actions occur:

* Get the tokenized shares record
* Get the validator that issued the tokenized shares from the record
* Unbond the delegation associated with the tokenized shares
* The delegator is NOT a liquid staking provider:
    * Decrease the `ValidatorLiquidStakingCap`
    * Decrease the validator's `LiquidShares`
* Burn the liquid coins equivalent of the tokenized shares
* Delete the tokenized shares record
* Send equivalent amount of tokens to the delegator
* Delegate sender's tokens to the validator

## MsgTransferTokenizeShareRecord

The `MsgTransferTokenizeShareRecord` message enables users to transfer the ownership of rewards generated from the tokenized amount of delegation.

```protobuf reference
https://github.com/cosmos/gaia/blob/v0.45.16-ics-lsm/proto/gaia/liquid/v1beta1/tx.proto#L56-L58
```

```protobuf reference
https://github.com/cosmos/gaia/blob/v0.45.16-ics-lsm/proto/gaia/liquid/v1beta1/tx.proto#L220-L228
```

This message is expected to fail if:

* the tokenized shares record doesn't exist
* the sender address doesn't match the owner address in the record 

When this message is processed the following actions occur:

* the tokenized shares record is updated with the new owner address

## MsgEnableTokenizeShares

The `MsgEnableTokenizeShares` message begins the countdown after which tokenizing shares by the sender delegator address is re-allowed, which will complete after the unbonding period.


```protobuf reference
https://github.com/cosmos/gaia/blob/v0.45.16-ics-lsm/proto/gaia/liquid/v1beta1/tx.proto#L63-L65
```

```protobuf reference
https://github.com/cosmos/gaia/blob/v0.45.16-ics-lsm/proto/gaia/liquid/v1beta1/tx.proto#L244-L250
```

This message returns a response containing the time at which the lock is completely removed:

```protobuf reference
https://github.com/cosmos/gaia/blob/v0.45.16-ics-lsm/proto/gaia/liquid/v1beta1/tx.proto#L252-L255
```

This message is expected to fail if:

*  if the sender's account lock status is either equal to `UNLOCKED` or `LOCK_EXPIRING`,
meaning that the tokenized shares aren't currently disabled.


When this message is processed the following actions occur:

* queue the unlock authorization.

## MsgDisableTokenizeShares

The `MsgDisableTokenizeShares` message prevents the sender delegator address from tokenizing any of its delegations.

```protobuf reference
https://github.com/cosmos/gaia/blob/v0.45.16-ics-lsm/proto/gaia/liquid/v1beta1/tx.proto#L60-L61
```

```protobuf reference
https://github.com/cosmos/gaia/blob/v0.45.16-ics-lsm/proto/gaia/liquid/v1beta1/tx.proto#L233-L239
```

This message is expected to fail if:

*  the sender's account already has the `LOCKED` lock status


When this message is processed the following actions occur:

* if the sender's account lock status is equal to `LOCK_EXPIRING`,
it cancels the pending unlock authorizations by removing them from the queue.
* Create a new tokenization lock for the sender's account. Note that
if there is a lock expiration in progress, it is overridden.

## MsgValidatorBond

The `MsgValidatorBond` message designates a delegation as a validator bond.
It enables validators to receive more liquid staking delegations

```protobuf reference
https://github.com/cosmos/gaia/blob/v0.45.16-ics-lsm/proto/gaia/liquid/v1beta1/tx.proto#L67-L68
```

```protobuf reference
https://github.com/cosmos/gaia/blob/v0.45.16-ics-lsm/proto/gaia/liquid/v1beta1/tx.proto#L257-L265
```

This message is expected to fail if:

* the delegator is a liquid staking provider

When this message is processed the following actions occur:

* If the delegation is not already a `ValidatorBond`:
    * Enable the delegation's `ValidatorBond` flag
    * Update validator's `ValidatorBondShares`

## MsgUnbondValidator

The `MsgTransferTokenizeShareRecord` message allows validator to change their
status from transfers from `Bonded` to `Unbonding` without experiencing slashing.

```protobuf reference
https://github.com/cosmos/gaia/blob/v0.45.16-ics-lsm/proto/gaia/liquid/v1beta1/tx.proto#L36-L38
```

```protobuf reference
https://github.com/cosmos/gaia/blob/v0.45.16-ics-lsm/proto/gaia/liquid/v1beta1/tx.proto#L165-L169
```

This message is expected to fail if:

* the validator isn't registered or is already jailed

When this message is processed the following actions occur:

* the validator is jailed
* the validator status changes from `Bonded` to `Unbonding`

### MsgUpdateParams

The `MsgUpdateParams` update the staking module parameters.
The params are updated through a governance proposal where the signer is the gov module account address.

```protobuf reference
https://github.com/cosmos/gaia/blob/main/proto/gaia/liquid/v1beta1/tx.proto#L182-L195
```

The message handling can fail if:

* signer is not the authority defined in the staking keeper (usually the gov module account).

## Begin-Block

### RemoveExpiredTokenizeShareLocks
Each abci begin block call, the liquid module will prune expired tokenize share locks.


### Queues

Within staking, certain state-transitions are not instantaneous but take place
over a duration of time (typically the unbonding period). When these
transitions are mature certain operations must take place in order to complete
the state operation. This is achieved through the use of queues which are
checked/processed at the end of each block.

## Events

The liquid module emits the following events:

## Msg's

### MsgTokenizeShares

| Type                          | Attribute Key         | Attribute Value              |
| ----------------------------- |-----------------------|------------------------------|
| tokenize_shares               | delegator_address     | {delegatorAddress}           |
| tokenize_shares               | validator_address     | {validatorAddress}           |
| tokenize_shares               | tokenized_share_owner | {tokenizedShareOwnerAddress} |
| tokenize_shares               | amount                | {tokenizeAmount}             |
| message                       | module                | liquid                       |
| message                       | action                | tokenize_shares              |
| message                       | sender                | {senderAddress}              |

### MsgRedeemTokensForShares

| Type                          | Attribute Key     | Attribute Value    |
| ----------------------------- |-------------------|--------------------|
| redeem_tokens_for_shares      | delegator_address | {delegatorAddress} |
| redeem_tokens_for_shares      | amount            | {redeemAmount}     |
| message                       | module            | liquid             |
| message                       | action            | redeem_tokens      |
| message                       | sender            | {senderAddress}    |

### MsgTransferTokenizeShareRecord

| Type                               | Attribute Key            | Attribute Value                |
| ---------------------------------- |--------------------------|--------------------------------|
| transfer_tokenize_share_record     | tokenize_share_record_id | {shareRecordID}                |
| transfer_tokenize_share_record     | sender                   | {senderAddress}                |
| transfer_tokenize_share_record     | new_owner                | {newShareOwnerAddress}         |
| message                            | module                   | liquid                         |
| message                            | action                   | transfer-tokenize-share-record |
| message                            | sender                   | {senderAddress}                |

### MsgEnableTokenizeShares

| Type                          | Attribute Key     | Attribute Value        |
| ----------------------------- |-------------------|------------------------|
| enable_tokenize_shares        | delegator_address | {delegatorAddress}     |
| message                       | module            | liquid                 |
| message                       | action            | enable_tokenize_shares |
| message                       | sender            | {senderAddress}        |

### MsgDisableTokenizeShares

| Type                          | Attribute Key     | Attribute Value         |
| ----------------------------- |-------------------|-------------------------|
| disable_tokenize_shares       | delegator_address | {delegatorAddress}      |
| message                       | module            | liquid                  |
| message                       | action            | disable_tokenize_shares |
| message                       | sender            | {senderAddress}         |

### MsgWithdrawTokenizeShareRecordReward

| Type                                  | Attribute Key | Attribute Value                       |
|---------------------------------------|---------------|---------------------------------------|
| withdraw_tokenize_share_record_reward | owner_address | {ownerAddress}                        |
| withdraw_tokenize_share_record_reward | record_id     | {recordID}                            |
| message                               | module        | liquid                                |
| message                               | action        | withdraw_tokenize_share_record_reward |
| message                               | sender        | {senderAddress}                       |

### MsgWithdrawAllTokenizeShareRecordReward

| Type                                      | Attribute Key | Attribute Value                           |
|-------------------------------------------|---------------|-------------------------------------------|
| withdraw_all_tokenize_share_record_reward | owner_address | {ownerAddress}                            |
| message                                   | module        | liquid                                    |
| message                                   | action        | withdraw_all_tokenize_share_record_reward |
| message                                   | sender        | {senderAddress}                           |


## Parameters

The liquid module contains the following parameters:

| Key                         | Type             | Example                  |
|-------------------------    |------------------|--------------------------|
| GlobalLiquidStakingCap      | string           | "1.000000000000000000"   | 
| ValidatorLiquidStakingCap   | string           | "0.250000000000000000"   | 


## Client

### CLI

A user can query and interact with the `liquid` module using the CLI.

#### Query

The `query` commands allows users to query `liquid` state.

```bash
gaiad query liquid --help
```

##### all-tokenize-share-records

The `all-tokenize-share-records` command allows users to query all tokenize share records.

Usage:

```bash
gaiad query liquid all-tokenize-share-records [flags]
```

Example:

```bash
gaiad query liquid all-tokenize-share-records
```

Example Output:

```bash
pagination:
  total: "1"
records:
- id: "1"
  module_account: tokenizeshare_1
  owner: cosmos1dw6s9qsz4uh42j3cgapyfm3tu83qafchy2srez
  validator: cosmosvaloper1dw6s9qsz4uh42j3cgapyfm3tu83qafchp7yk43
```

##### last-tokenize-share-record-id

The `last-tokenize-share-record-id` command allows users to query the last tokenize share record ID issued.

Usage:

```bash
gaiad query liquid last-tokenize-share-record-id [flags]
```

Example:

```bash
gaiad query liquid last-tokenize-share-record-id
```

Example Output:

```bash
id: "2"
```

##### params

The `params` command allows users to query the current module params.

Usage:

```bash
gaiad query liquid params [flags]
```

Example:

```bash
gaiad query liquid params
```

Example Output:

```bash
params:
  global_liquid_staking_cap: "0.250000000000000000"
  validator_liquid_staking_cap: "1.000000000000000000"
```

##### tokenize-share-lock-info

The `tokenize-share-lock-info` command allows users to query the current tokenization lock status for a given account.

Usage:

```bash
gaiad query liquid tokenize-share-lock-info [account-addr] [flags]
```

Example:

```bash
gaiad query liquid tokenize-share-lock-info cosmos1dw6s9qsz4uh42j3cgapyfm3tu83qafchy2srez
```

Example Output:

```bash
status: TOKENIZE_SHARE_LOCK_STATUS_UNLOCKED
```


##### tokenize-share-record-by-denom

The `tokenize-share-record-by-denom` command allows users to query the tokenize share record information for the 
provided denom.

Usage:

```bash
gaiad query liquid tokenize-share-record-by-denom [denom] [flags]
```

Example:

```bash
gaiad query liquid tokenize-share-record-by-denom cosmosvaloper1vuvl27z833dksv89vz2205mrwhadez3k3egzrh/1
```

Example Output:

```bash
record:
  id: "1"
  module_account: tokenizeshare_1
  owner: cosmos15ty20clrlwmph2v8k7qzr4lklpz883zdd89ckp
  validator: cosmosvaloper1vuvl27z833dksv89vz2205mrwhadez3k3egzrh
```

##### tokenize-share-record-by-id

The `tokenize-share-record-by-id` command allows users to query the tokenize share record information for the
provided record ID.

Usage:

```bash
gaiad query liquid tokenize-share-record-by-id [ID] [flags]
```

Example:

```bash
gaiad query liquid tokenize-share-record-by-id 1
```

Example Output:

```bash
record:
  id: "1"
  module_account: tokenizeshare_1
  owner: cosmos15ty20clrlwmph2v8k7qzr4lklpz883zdd89ckp
  validator: cosmosvaloper1vuvl27z833dksv89vz2205mrwhadez3k3egzrh
```

##### tokenize-share-record-rewards

The `tokenize-share-record-rewards` command allows users to query the rewards for the provided record owner.

Usage:

```bash
gaiad query liquid tokenize-share-record-rewards [owner-addr] [flags]
```

Example:

```bash
gaiad query liquid tokenize-share-record-rewards cosmos15ty20clrlwmph2v8k7qzr4lklpz883zdd89ckp
```

Example Output:

```bash
rewards:
  - record_id: "1"
  reward:
  - 1496874162.803718702000000000stake
  - 2.155511221800000000uatom
total:
- 1496874162.803718702000000000stake
- 2.155511221800000000uatom
```

##### tokenize-share-records-owned

The `tokenize-share-records-owned` command allows users to query the records account address.

Usage:

```bash
gaiad query liquid tokenize-share-records-owned [owner-addr] [flags]
```

Example:

```bash
gaiad query liquid tokenize-share-records-owned cosmos15ty20clrlwmph2v8k7qzr4lklpz883zdd89ckp
```

Example Output:

```bash
records:
- id: "1"
  module_account: tokenizeshare_1
  owner: cosmos15ty20clrlwmph2v8k7qzr4lklpz883zdd89ckp
  validator: cosmosvaloper1vuvl27z833dksv89vz2205mrwhadez3k3egzrh
```

##### total-liquid-staked

The `total-liquid-staked` command allows users to query the total amount of tokens liquid staked.

Usage:

```bash
gaiad query liquid total-liquid-staked [flags]
```

Example:

```bash
gaiad query liquid total-liquid-staked
```

Example Output:

```bash
tokens: "200000000"
```

##### total-tokenize-share-assets

The `total-tokenize-share-assets` command allows users to query the total amount of tokenized assets.

Usage:

```bash
gaiad query liquid total-tokenize-share-assets [flags]
```

Example:

```bash
gaiad query liquid total-tokenize-share-assets
```

Example Output:

```bash
value:
  amount: "200000000"
  denom: uatom
```

#### Transactions

The `tx` commands allows users to interact with the `liquid` module.

```bash
gaiad tx liquid --help
```

##### disable-tokenize-shares

The command `disable-tokenize-shares` allows users to disable tokenization for their account.

Usage:

```bash
gaiad tx liquid disable-tokenize-shares [flags]
```

Example:

```bash
gaiad tx liquid disable-tokenize-shares --from=mykey
```

##### enable-tokenize-shares

The command `enable-tokenize-shares` allows users to enable tokenization for their account.

Usage:

```bash
gaiad tx liquid enable-tokenize-shares [flags]
```

Example:

```bash
gaiad tx liquid enable-tokenize-shares --from=mykey
```

##### redeem-tokens

The command `redeem-tokens` allows users to convert a specified amount of tokenized shares for the underlying 
delegation.

Usage:

```bash
gaiad tx liquid redeem-tokens [amount] [flags]
```

Example:

```bash
gaiad tx liquid redeem-tokens 10000cosmosvaloper1vuvl27z833dksv89vz2205mrwhadez3k3egzrh/1
```

##### tokenize-share

The command `tokenize-share` allows users to convert a delegation into tokenized shares.

Usage:

```bash
gaiad tx liquid tokenize-share [validator-addr] [amount] [rewardOwner] [flags]
```

Example:

```bash
gaiad tx liquid tokenize-share cosmosvaloper1vuvl27z833dksv89vz2205mrwhadez3k3egzrh 1000uatom cosmos15ty20clrlwmph2v8k7qzr4lklpz883zdd89ckp
```

##### transfer-tokenize-share-record

The command `transfer-tokenize-share-record` allows users to transfer a tokenize share record to another owner.

Usage:

```bash
gaiad tx liquid transfer-tokenize-share-record [record-id] [new-owner] [flags]
```

Example:

```bash
gaiad tx liquid transfer-tokenize-share-record 1 cosmosvaloper1vuvl27z833dksv89vz2205mrwhadez3k3egzrh
```

##### withdraw-all-tokenize-share-rewards

The command `withdraw-all-tokenize-share-rewards` allows users to withdraw all rewards for their tokenized shares.

Usage:

```bash
gaiad tx liquid withdraw-all-tokenize-share-rewards [flags]
```

Example:

```bash
gaiad tx liquid withdraw-all-tokenize-share-rewards --from=myKey
```

##### withdraw-tokenize-share-rewards

The command `withdraw-tokenize-share-rewards` allows users to withdraw rewards for a tokenize share record.

Usage:

```bash
gaiad tx liquid withdraw-tokenize-share-rewards [record-id] [flags]
```

Example:

```bash
gaiad tx liquid withdraw-all-tokenize-share-rewards 1 --from=myKey
```

### gRPC

A user can query the `liquid` module using gRPC endpoints.

#### AllTokenizeShareRecords

The `AllTokenizeShareRecords` endpoint queries all tokenize share records.

```bash
gaia.liquid.v1beta1.Query/AllTokenizeShareRecords
```

Example:

```bash
grpcurl -plaintext localhost:9090 gaia.liquid.v1beta1.Query/AllTokenizeShareRecords
```

Example Output:

```bash
{
  "records": [
    {
      "id": "1",
      "owner": "cosmos12xw6ylce2enratz3m942xd9jnjc4qrkk0yqnmr",
      "moduleAccount": "tokenizeshare_1",
      "validator": "cosmosvaloper1jd9slc386vepwpamrrgzkpflhfy94mhqcf0sre"
    }
  ],
  "pagination": {
    "total": "1"
  }
}
```

#### LastTokenizeShareRecordId

The `LastTokenizeShareRecordId` endpoint queries the last tokenize share record ID issued.

```bash
gaia.liquid.v1beta1.Query/LastTokenizeShareRecordId
```

Example:

```bash
grpcurl -plaintext localhost:9090 gaia.liquid.v1beta1.Query/LastTokenizeShareRecordId
```

Example Output:

```bash
{
  "id": "1"
}
```

#### TokenizeShareLockInfo

The `TokenizeShareLockInfo` endpoint queries the current tokenization lock status for a given account.

```bash
gaia.liquid.v1beta1.Query/TokenizeShareLockInfo
```

Example:

```bash
grpcurl -plaintext -d '{"address": "cosmos12xw6ylce2enratz3m942xd9jnjc4qrkk0yqnmr"}' \
localhost:9090 gaia.liquid.v1beta1.Query/TokenizeShareLockInfo
```

Example Output:

```bash
{
  "status": "TOKENIZE_SHARE_LOCK_STATUS_UNLOCKED"
}
```

#### TokenizeShareRecordByDenom

The `TokenizeShareRecordByDenom` endpoint queries the tokenize share record information for the provided denom.

```bash
gaia.liquid.v1beta1.Query/TokenizeShareRecordByDenom
```

Example:

```bash
grpcurl -plaintext -d '{"denom": "cosmosvaloper1jd9slc386vepwpamrrgzkpflhfy94mhqcf0sre/1"}' \
localhost:9090 gaia.liquid.v1beta1.Query/TokenizeShareRecordByDenom
```

Example Output:

```bash
{
  "record": {
    "id": "1",
    "owner": "cosmos12xw6ylce2enratz3m942xd9jnjc4qrkk0yqnmr",
    "moduleAccount": "tokenizeshare_1",
    "validator": "cosmosvaloper1jd9slc386vepwpamrrgzkpflhfy94mhqcf0sre"
  }
}
```

#### TokenizeShareRecordById

The `TokenizeShareRecordById` endpoint queries the tokenize share record information for the provided record ID.

```bash
gaia.liquid.v1beta1.Query/TokenizeShareRecordById
```

Example:

```bash
grpcurl -plaintext -d '{"id": "1"}' \
localhost:9090 gaia.liquid.v1beta1.Query/TokenizeShareRecordById
```

Example Output:

```bash
{
  "record": {
    "id": "1",
    "owner": "cosmos12xw6ylce2enratz3m942xd9jnjc4qrkk0yqnmr",
    "moduleAccount": "tokenizeshare_1",
    "validator": "cosmosvaloper1jd9slc386vepwpamrrgzkpflhfy94mhqcf0sre"
  }
}
```

#### TokenizeShareRecordReward

The `TokenizeShareRecordReward` endpoint queries the rewards for the provided record owner.

```bash
gaia.liquid.v1beta1.Query/TokenizeShareRecordReward
```

Example:

```bash
grpcurl -plaintext -d '{"owner_address": "cosmos12xw6ylce2enratz3m942xd9jnjc4qrkk0yqnmr"}' \
localhost:9090 gaia.liquid.v1beta1.Query/TokenizeShareRecordReward
```

Example Output:

```bash
{
  "rewards": [
    {
      "recordId": "1",
      "reward": [
        {
          "denom": "stake",
          "amount": "8588380036928696253000000000"
        },
        {
          "denom": "uatom",
          "amount": "2155511221800000000"
        }
      ]
    }
  ],
  "total": [
    {
      "denom": "stake",
      "amount": "8588380036928696253000000000"
    },
    {
      "denom": "uatom",
      "amount": "2155511221800000000"
    }
  ]
}
```

#### TokenizeShareRecordsOwned

The `TokenizeShareRecordsOwned` command allows users to query the records account address.

```bash
gaia.liquid.v1beta1.Query/TokenizeShareRecordsOwned
```

Example:

```bash
grpcurl -plaintext -d '{"owner": "cosmos12xw6ylce2enratz3m942xd9jnjc4qrkk0yqnmr"}' \
localhost:9090 gaia.liquid.v1beta1.Query/TokenizeShareRecordsOwned
```

Example Output:

```bash
{
  "records": [
    {
      "id": "1",
      "owner": "cosmos12xw6ylce2enratz3m942xd9jnjc4qrkk0yqnmr",
      "moduleAccount": "tokenizeshare_1",
      "validator": "cosmosvaloper1jd9slc386vepwpamrrgzkpflhfy94mhqcf0sre"
    }
  ]
}
```

#### TotalLiquidStaked

The `TotalLiquidStaked` endpoint queries the total amount of tokens liquid staked.

```bash
gaia.liquid.v1beta1.Query/TotalLiquidStaked
```

Example:

```bash
grpcurl -plaintext localhost:9090 gaia.liquid.v1beta1.Query/TotalLiquidStaked
```

Example Output:

```bash
{
  "tokens": "200000000"
}
```

#### TotalTokenizeSharedAssets

The `TotalTokenizeSharedAssets` endpoint queries the total amount of tokenized assets.

```bash
gaia.liquid.v1beta1.Query/TotalTokenizeSharedAssets
```

Example:

```bash
grpcurl -plaintext localhost:9090 gaia.liquid.v1beta1.Query/TotalTokenizeSharedAssets
```

Example Output:

```bash
{
  "value": {
    "denom": "uatom",
    "amount": "200000000"
  }
}
```

#### Params

The `Params` endpoint queries the module Params.

```bash
gaia.liquid.v1beta1.Query/Params
```

Example:

```bash
grpcurl -plaintext localhost:9090 gaia.liquid.v1beta1.Query/Params
```

Example Output:

```bash
{
  "params": {
    "globalLiquidStakingCap": "250000000000000000",
    "validatorLiquidStakingCap": "500000000000000000"
  }
}
```

### REST

A user can query the `liquid` module using REST endpoints.

#### AllTokenizeShareRecords

The `AllTokenizeShareRecords` REST endpoint queries all tokenize share records.

```bash
/gaia/liquid/v1beta1/tokenize_share_records
```

Example:

```bash
curl -X GET "http://localhost:1317/gaia/liquid/v1beta1/tokenize_share_records" -H  "accept: application/json"
```

Example Output:

```bash
{
  "delegation_responses": [
    {
      "delegation": {
        "delegator_address": "cosmos1vcs68xf2tnqes5tg0khr0vyevm40ff6zdxatp5",
        "validator_address": "cosmosvaloper1quqxfrxkycr0uzt4yk0d57tcq3zk7srm7sm6r8",
        "shares": "256250000.000000000000000000"
      },
      "balance": {
        "denom": "stake",
        "amount": "256250000"
      }
    },
    {
      "delegation": {
        "delegator_address": "cosmos1vcs68xf2tnqes5tg0khr0vyevm40ff6zdxatp5",
        "validator_address": "cosmosvaloper194v8uwee2fvs2s8fa5k7j03ktwc87h5ym39jfv",
        "shares": "255150000.000000000000000000"
      },
      "balance": {
        "denom": "stake",
        "amount": "255150000"
      }
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "2"
  }
}
```

#### LastTokenizeShareRecordId

The `LastTokenizeShareRecordId` REST endpoint queries the last tokenize share record ID issued.

```bash
/gaia/liquid/v1beta1/last_tokenize_share_record_id
```

Example:

```bash
curl -X GET "http://localhost:1317/gaia/liquid/v1beta1/last_tokenize_share_record_id" -H  "accept: application/json"
```

Example Output:

```bash
{
  "id": "1"
}
```

#### TokenizeShareLockInfo

The `TokenizeShareLockInfo` REST endpoint queries the current tokenization lock status for a given account.

```bash
/gaia/liquid/v1beta1/tokenize_share_lock_info/{address}
```

Example:

```bash
curl -X GET "http://localhost:1317/gaia/liquid/v1beta1/tokenize_share_lock_info/cosmos12xw6ylce2enratz3m942xd9jnjc4qrkk0yqnmr" -H  "accept: application/json"
```

Example Output:

```bash
{
  "status": "TOKENIZE_SHARE_LOCK_STATUS_UNLOCKED",
  "expiration_time": ""
}
```

#### TokenizeShareRecordById

The `TokenizeShareRecordById` REST endpoint queries the tokenize share record information for the provided record ID.

```bash
/gaia/liquid/v1beta1/tokenize_share_record_by_id/{id}
```

Example:

```bash
curl -X GET "http://localhost:1317/gaia/liquid/v1beta1/tokenize_share_record_by_id/1" -H  "accept: application/json"
```

Example Output:

```bash
{
  "record": {
    "id": "1",
    "owner": "cosmos12xw6ylce2enratz3m942xd9jnjc4qrkk0yqnmr",
    "module_account": "tokenizeshare_1",
    "validator": "cosmosvaloper1jd9slc386vepwpamrrgzkpflhfy94mhqcf0sre"
  }
}
```

#### TokenizeShareRecordReward

The `TokenizeShareRecordReward` REST endpoint queries the rewards for the provided record owner.

```bash
/gaia/liquid/v1beta1/{owner_address}/tokenize_share_record_rewards"
```

Example:

```bash
curl -X GET "http://localhost:1317/gaia/liquid/v1beta1/cosmos12xw6ylce2enratz3m942xd9jnjc4qrkk0yqnmr/tokenize_share_record_rewards" -H  "accept: application/json"
```

Example Output:

```bash
{
  "rewards": [
    {
      "record_id": "1",
      "reward": [
        {
          "denom": "stake",
          "amount": "392793740917.315504955400000000"
        },
        {
          "denom": "uatom",
          "amount": "2.155511221800000000"
        }
      ]
    }
  ],
  "total": [
    {
      "denom": "stake",
      "amount": "392793740917.315504955400000000"
    },
    {
      "denom": "uatom",
      "amount": "2.155511221800000000"
    }
  ]
}
```


#### TokenizeShareRecordsOwned

The `TokenizeShareRecordsOwned` REST endpoint queries the records account address.

```bash
/gaia/liquid/v1beta1/tokenize_share_record_owned/{owner}"
```

Example:

```bash
curl -X GET "http://localhost:1317/gaia/liquid/v1beta1/tokenize_share_record_owned/cosmos12xw6ylce2enratz3m942xd9jnjc4qrkk0yqnmr" -H  "accept: application/json"
```

Example Output:

```bash
{
  "records": [
    {
      "id": "1",
      "owner": "cosmos12xw6ylce2enratz3m942xd9jnjc4qrkk0yqnmr",
      "module_account": "tokenizeshare_1",
      "validator": "cosmosvaloper1jd9slc386vepwpamrrgzkpflhfy94mhqcf0sre"
    }
  ]
}
```


#### TotalLiquidStaked

The `TotalLiquidStaked` REST endpoint queries the total amount of tokens liquid staked.

```bash
/gaia/liquid/v1beta1/total_liquid_staked"
```

Example:

```bash
curl -X GET "http://localhost:1317/gaia/liquid/v1beta1/total_liquid_staked" -H  "accept: application/json"
```

Example Output:

```bash
{
  "tokens": "200000000"
}
```

#### TotalTokenizeSharedAssets

The `TotalTokenizeSharedAssets` REST endpoint queries the total amount of tokenized assets.

```bash
/gaia/liquid/v1beta1/total_tokenize_shared_assets"
```

Example:

```bash
curl -X GET "http://localhost:1317/gaia/liquid/v1beta1/total_tokenize_shared_assets" -H  "accept: application/json"
```

Example Output:

```bash
{
  "value": {
    "denom": "uatom",
    "amount": "200000000"
  }
}
```

#### Params

The `Params` REST endpoint queries the module Params.

```bash
/gaia/liquid/v1beta1/params"
```

Example:

```bash
curl -X GET "http://localhost:1317/gaia/liquid/v1beta1/params" -H  "accept: application/json"
```

Example Output:

```bash
{
  "params": {
    "global_liquid_staking_cap": "0.250000000000000000",
    "validator_liquid_staking_cap": "0.500000000000000000"
  }
}
```