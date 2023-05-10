<!-- order: 2 -->

 # State

The liquidity module `x/liquidity` keeps track of the Pool and PoolBatch states. The state represents your app at a given moment.
## Pool

Pool stores information about the liquidity pool.

Pool type has the following structure.

```go
type Pool struct {
    Id                     uint64         // index of this liquidity pool
    TypeId                 uint32         // pool type of this liquidity pool
    ReserveCoinDenoms      []string       // list of reserve coin denoms for this liquidity pool
    ReserveAccountAddress  string         // reserve account address for this liquidity pool to store reserve coins
    PoolCoinDenom          string         // denom of pool coin for this liquidity pool
}
```

The parameters of the Pool state are:

- Pool: `0x11 | Id -> ProtocolBuffer(Pool)`

- PoolByReserveAccIndex: `0x12 | ReserveAccLen (1 byte) | ReserveAcc -> ProtocolBuffer(uint64)`

- GlobalLiquidityPoolIdKey: `[]byte("globalLiquidityPoolId")`

- ModuleName, RouterKey, StoreKey, QuerierRoute: `liquidity`

- PoolCoinDenomPrefix: `pool`
## PoolBatch

PoolBatch stores information about the liquidity pool batch states.

PoolBatch type has the following structure.

```go
type PoolBatch struct {
    PoolId           uint64  // id of target liquidity pool
    Index            uint64  // index of this batch
    BeginHeight      uint64  // block height when batch is created
    DepositMsgIndex  uint64  // last index of DepositMsgStates
    WithdrawMsgIndex uint64  // last index of WithdrawMsgStates
    SwapMsgIndex     uint64  // last index of SwapMsgStates
    Executed         bool    // true if executed, false if not executed
}
```

## Batch Messages

Deposit, withdrawal, or swap orders are accumulated in a liquidity pool for a pre-defined period, which can be one or more blocks in length. Orders are then added to the pool and executed at the end of the batch. The following messages are executed in batch-style. 

### DepositMsgState

`DepositMsgState` defines the state of deposit message as it is processed in the next batch or batches.

When a user sends `MsgDepositWithinBatch` transaction to the network, it is accumulated in a batch. `DepositMsgState` contains the state information about the message; if the transaction is executed, successfully matched, and if it is to be deleted in the next block.

```go
type DepositMsgState struct {
    MsgHeight  int64  // block height where this message is appended to the batch
    MsgIndex   uint64 // index of this deposit message in this liquidity pool
    Executed   bool   // true if executed on this batch, false if not executed
    Succeeded  bool   // true if executed successfully on this batch, false if failed
    ToBeDelete bool   // true if ready to be deleted on kvstore, false if not ready to be deleted
    Msg        MsgDepositWithinBatch
}
```
### WithdrawMsgState

`WithdrawMsgState` defines the state of the withdraw message as it is processed in the next batch or batches.

When a user sends a `MsgWithdrawWithinBatch` transaction to the network, it is accumulated in a batch. `WithdrawMsgState` contains the state information about the message:

- If the transaction is executed
- If the transaction is successfully matched
- If the transaction will be deleted in the next block

```go
type WithdrawMsgState struct {
    MsgHeight  int64  // block height where this message is appended to the batch
    MsgIndex   uint64 // index of this withdraw message in this liquidity pool
    Executed   bool   // true if executed on this batch, false if not executed
    Succeeded  bool   // true if executed successfully on this batch, false if failed
    ToBeDelete bool   // true if ready to be deleted on kvstore, false if not ready to be deleted
    Msg        MsgWithdrawWithinBatch
}
```
### SwapMsgState

`SwapMsgState` defines the state of swap message as it is processed in the next batch or batches.

When a user sends a `MsgSwapWithinBatch` transaction to the network, it is accumulated in a batch. `SwapMsgState` contains the state information about the message:

- If the transaction is executed
- If the transaction is successfully matched
- If the transaction will be deleted in the next block

```go
type SwapMsgState struct {
    MsgHeight          int64  // block height where this message is appended to the batch
    MsgIndex           uint64 // index of this swap message in this liquidity pool
    Executed           bool   // true if executed on this batch, false if not executed
    Succeeded          bool   // true if executed successfully on this batch, false if failed
    ToBeDelete         bool   // true if ready to be deleted on kvstore, false if not ready to be deleted
    OrderExpiryHeight  int64  // swap orders are cancelled when current height is equal to or greater than ExpiryHeight
    ExchangedOfferCoin sdk.Coin // offer coin exchanged so far
    RemainingOfferCoin sdk.Coin // offer coin  remaining to be exchanged
    Msg                MsgSwapWithinBatch
}
```

The parameters of the PoolBatch, DepositMsgState, WithdrawMsgState, and SwapMsgState states are:

- PoolBatch: `0x22 | PoolId -> ProtocolBuffer(PoolBatch)`

- PoolBatchDepositMsgStates: `0x31 | PoolId | MsgIndex -> ProtocolBuffer(DepositMsgState)`

- PoolBatchWithdrawMsgStates: `0x32 | PoolId | MsgIndex -> ProtocolBuffer(WithdrawMsgState)`

- PoolBatchSwapMsgStates: `0x33 | PoolId | MsgIndex -> ProtocolBuffer(SwapMsgState)`
