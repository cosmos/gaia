<!-- order: 4 -->

 # Messages

Messages (Msg) are objects that trigger state transitions. Msgs are wrapped in transactions (Txs) that clients submit to the network. The Cosmos SDK wraps and unwraps liquidity module messages from transactions.

## MsgCreatePool

A liquidity pool is created and initial coins are deposited with the `MsgCreatePool` message.

```go
type MsgCreatePool struct {
    PoolCreatorAddress  string         // account address of the origin of this message
    PoolTypeId          uint32         // id of the new liquidity pool
    DepositCoins         sdk.Coins      // deposit initial coins for new liquidity pool
}
```

### Validity Checks

Validity checks are performed for MsgCreatePool messages. The transaction that is triggered with `MsgCreatePool` fails if:

- if `params.CircuitBreakerEnabled` is true
- `PoolCreator` address does not exist
- `PoolTypeId` does not exist in parameters
- A duplicate `LiquidityPool` with same `PoolTypeId` and `ReserveCoinDenoms` exists
- One or more coins in `ReserveCoinDenoms` do not exist in `bank` module
- The balance of `PoolCreator` does not have enough amount of coins for `DepositCoins`
- The balance of `PoolCreator` does not have enough coins for `PoolCreationFee`

## MsgDepositWithinBatch

Coins are deposited in a batch to a liquidity pool with the `MsgDepositWithinBatch` message.

```go
type MsgDepositWithinBatch struct {
    DepositorAddress    string         // account address of depositor that originated this message
    PoolId              uint64         // id of the liquidity pool to receive deposit
    DepositCoins         sdk.Coins      // deposit coins
}
```

## Validity Checks

The MsgDepositWithinBatch message performs validity checks. The transaction that is triggered with the `MsgDepositWithinBatch` message fails if:

- if `params.CircuitBreakerEnabled` is true
- `Depositor` address does not exist
- `PoolId` does not exist
- The denoms of `DepositCoins` are not composed of existing `ReserveCoinDenoms` of the specified `LiquidityPool`
- The balance of `Depositor` does not have enough coins for `DepositCoins`

## MsgWithdrawWithinBatch

Withdraw coins in batch from liquidity pool with the `MsgWithdrawWithinBatch` message.

```go
type MsgWithdrawWithinBatch struct {
    WithdrawerAddress string         // account address of the origin of this message
    PoolId            uint64         // id of the liquidity pool to withdraw the coins from
    PoolCoin          sdk.Coin       // pool coin sent for reserve coin withdrawal
}
```

## Validity Checks

The MsgWithdrawWithinBatch message performs validity checks. The transaction that is triggered with the `MsgWithdrawWithinBatch` message fails if:

- `Withdrawer` address does not exist
- `PoolId` does not exist
- The denom of `PoolCoin` are not equal to the `PoolCoinDenom` of the `LiquidityPool`
- The balance of `Depositor` does not have enough coins for `PoolCoin`

## MsgSwapWithinBatch

Swap coins between liquidity pools in batch with the `MsgSwapWithinBatch` message.

Offer coins are swapped with demand coins for the given order price.

```go
type MsgSwapWithinBatch struct {
    SwapRequesterAddress string     // account address of the origin of this message
    PoolId               uint64     // id of the liquidity pool
    SwapTypeId           uint32     // swap type id of this swap message, default 1: InstantSwap, requesting instant swap
    OfferCoin            sdk.Coin   // offer coin of this swap
    DemandCoinDenom      string     // denom of demand coin of this swap
    OfferCoinFee         sdk.Coin   // offer coin fee for pay fees in half offer coin
    OrderPrice           sdk.Dec    // limit order price where the price is the exchange ratio of X/Y where X is the amount of the first coin and Y is the amount of the second coin when their denoms are sorted alphabetically
}
```

## Validity checks

The MsgSwapWithinBatch message performs validity checks. The transaction that is triggered with the `MsgSwapWithinBatch` message fails if:

- if `params.CircuitBreakerEnabled` is true
- `SwapRequester` address does not exist
- `PoolId` does not exist
- `SwapTypeId` does not exist
- Denoms of `OfferCoin` or `DemandCoin` do not exist in `bank` module
- The balance of `SwapRequester` does not have enough coins for `OfferCoin`
- `OrderPrice` <= zero
- `OfferCoinFee` equals `OfferCoin` * `params.SwapFeeRate` * `0.5` with ceiling
- Has sufficient balance `OfferCoinFee` to reserve offer coin fee
