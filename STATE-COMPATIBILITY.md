# State-Compatibility

- [State-Compatibility](#state-compatibility)
  - [Scope](#scope)
  - [Validating State-Compatibility](#validating-state-compatibility)
    - [AppHash](#apphash)
    - [LastResultsHash](#lastresultshash)
  - [Major Sources of State-incompatibility](#major-sources-of-state-incompatibility)
    - [Creating Additional State](#creating-additional-state)
    - [Changing Proto Field Definitions](#changing-proto-field-definitions)
    - [Returning Different Errors Given Same Input](#returning-different-errors-given-same-input)
    - [Variability in Gas Usage](#variability-in-gas-usage)
  - [Secondary Limitations To Keep In Mind](#secondary-limitations-to-keep-in-mind)
    - [Network Requests to External Services](#network-requests-to-external-services)
    - [Randomness](#randomness)
    - [Parallelism and Shared State](#parallelism-and-shared-state)
    - [Hardware Errors](#hardware-errors)


It is critical for the patch and minor releases to be state-machine compatible with prior releases in the same minor version. 
For example, v13.0.2 must be state-machine compatible with v13.0.1. 
_An exception are minor releases that are either emergency releases or replacements of deprecated major releases_. 

This is to ensure **determinism**, i.e., given the same input, the nodes will always produce the same output.

State-incompatibility is allowed for major upgrades because all nodes in the network perform it at the same time. Therefore, after the upgrade, the nodes continue functioning in a deterministic way.
 
## Scope

The state-machine scope includes the following areas:

- All ICS messages including:
  - Every msg's ValidateBasic method
  - Every msg's MsgServer method
    - Net gas usage, in all execution paths
    - Error result returned
    - State changes (namely every store write)
- AnteHandlers in "DeliverTx" mode
- All `BeginBlock`/`EndBlock` logic

The following are **NOT** in the state-machine scope:

- Events
- Queries that are not whitelisted
- CLI interfaces

## Validating State-Compatibility 

CometBFT ensures state compatibility by validating a number of hashes that can be found [here](https://github.com/cometbft/cometbft/blob/v0.38.2/proto/tendermint/types/types.proto#L59-L66).

`AppHash` and `LastResultsHash` are the common sources of problems stemming from our work.
To avoid these problems, let's now examine how these hashes work.

### AppHash

**Note:** The following explanation is simplified for clarity.

An app hash is a hash of hashes of every store's Merkle root that is returned by ABCI's `Commit()` from Cosmos-SDK to CometBFT.
Cosmos-SDK [takes an app hash of the application state](https://github.com/cosmos/cosmos-sdk/blob/v0.47.6/store/rootmulti/store.go#L468), and propagates it to CometBFT which, in turn, compares it to the app hash of the rest of the network.
Then, CometBFT ensures that the app hash of the local node matches the app hash of the network. 

### LastResultsHash

`LastResultsHash` is the root hash of all results from the transactions in the block returned by the ABCI's `DeliverTx`.

The [`LastResultsHash`](https://github.com/cometbft/cometbft/blob/v0.34.29/types/results.go#L47-L54) 
in CometBFT [v0.34.29](https://github.com/cometbft/cometbft/releases/tag/v0.34.29) contains:

1. Tx `GasWanted`

2. Tx `GasUsed`
  > `GasUsed` being Merkelized means that we cannot freely reorder methods that consume gas.
  > We should also be careful of modifying any validation logic since changing the
  > locations where we error or pass might affect transaction gas usage.
  >
  > There are plans to remove this field from being Merkelized in a subsequent CometBFT release, 
  > at which point we will have more flexibility in reordering operations / erroring.

3. Tx response `Data`

  > The `Data` field includes the proto marshalled Tx response. Therefore, we cannot 
  > change these in patch releases.

4. Tx response `Code`

  > This is an error code that is returned by the transaction flow. In the case of
  > success, it is `0`. On a general error, it is `1`. Additionally, each module
  > defines its custom error codes. 
  >
  > As a result, it is important to avoid changing custom error codes or change
  > the semantics of what is valid logic in transaction flows.

Note that all of the above stem from `DeliverTx` execution path, which handles:

- `AnteHandler`'s marked as deliver tx
- `msg.ValidateBasic`
- execution of a message from the message server

The `DeliverTx` return back to the CometBFT is defined [here](https://github.com/cosmos/cosmos-sdk/blob/d11196aad04e57812dbc5ac6248d35375e6603af/baseapp/abci.go#L293-L303).

## Major Sources of State-incompatibility

### Creating Additional State

By erroneously creating database entries that exist in Version A but not in
Version B, we can cause the app hash to differ across nodes running
these versions in the network. Therefore, this must be avoided.

### Changing Proto Field Definitions

For example, if we change a field that gets persisted to the database,
the app hash will differ across nodes running these versions in the network.

Additionally, this affects `LastResultsHash` because it contains a `Data` field that is a marshaled proto message.

### Returning Different Errors Given Same Input

```go
// Version A
func (sk Keeper) validateAmount(ctx context.Context, amount math.Int) error {
    if amount.IsNegative() {
        return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "amount must be positive or zero")
    }
    return nil
}
```

```go
// Version B
func (sk Keeper) validateAmount(ctx context.Context, amount math.Int) error {
    if amount.IsNegative() || amount.IsZero() {
        return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "amount must be positive")
    }
    return nil
}
```

Note that now an amount of 0 can be valid in "Version A", but invalid in "Version B".
Therefore, if some nodes are running "Version A" and others are running "Version B",
the final app hash might not be deterministic.

Additionally, a different error message does not matter because it
is not included in any hash. However, an error code `sdkerrors.ErrInvalidRequest` does.
It translates to the `Code` field in the `LastResultsHash` and participates in
its validation.

### Variability in Gas Usage

For transaction flows (or any other flow that consumes gas), it is important
that the gas usage is deterministic.

Currently, gas usage is being Merklized in the state. As a result, reordering functions
becomes risky.

Suppose my gas limit is 2000 and 1600 is used up before entering
`someInternalMethod`. Consider the following:

```go
func someInternalMethod(ctx sdk.Context) {
  object1 := readOnlyFunction1(ctx) # consumes 1000 gas
  object2 := readOnlyFunction2(ctx) # consumes 500 gas
  doStuff(ctx, object1, object2)
}
```

- It will run out of gas with `gasUsed = 2600` where 2600 getting merkelized
into the tx results.

```go
func someInternalMethod(ctx sdk.Context) {
  object2 := readOnlyFunction2(ctx) # consumes 500 gas
  object1 := readOnlyFunction1(ctx) # consumes 1000 gas
  doStuff(ctx, object1, object2)
}
```

- It will run out of gas with `gasUsed = 2100` where 2100 is getting merkelized
into the tx results.

Therefore, we introduced a state-incompatibility by merkelizing diverging gas
usage.

## Secondary Limitations To Keep In Mind

### Network Requests to External Services

It is critical to avoid performing network requests to external services
since it is common for services to be unavailable or rate-limit.

Imagine a service that returns exchange rates when clients query its HTTP endpoint.
This service might experience downtime or be restricted in some geographical areas.

As a result, nodes may get diverging responses where some
get successful responses while others errors, leading to state breakage.

### Randomness

Randomness cannot be used in the state machine, as the state machine must be deterministic.

**Note:** Iteration order over `map`s is non-deterministic, so to be deterministic 
you must gather the keys, and sort them all prior to iterating over all values.

### Parallelism and Shared State

Threads and Goroutines might preempt differently in different hardware. Therefore,
they should be avoided for the sake of determinism. Additionally, it is hard
to predict when the multi-threaded state can be updated.

### Hardware Errors

This is out of the developer's control but is mentioned for completeness.
