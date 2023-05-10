<!-- order: 8 -->

 # Parameters

The liquidity module contains the following parameters:

Key                    | Type             | Example
---------------------- | ---------------- | -------------------------------------------------------------------------------------------------------------------
PoolTypes              | []PoolType            | [{"id":1,"name":"StandardLiquidityPool","min_reserve_coin_num":2,"max_reserve_coin_num":2,"description":"Standard liquidity pool with pool price function X/Y, ESPM constraint, and two kinds of reserve coins"}]
MinInitDepositAmount   | string (sdk.Int)      | "1000000"
InitPoolCoinMintAmount | string (sdk.Int)      | "1000000"
MaxReserveCoinAmount   | string (sdk.Int)      | "0"
PoolCreationFee        | sdk.Coins             | [{"denom":"stake","amount":"40000000"}]
SwapFeeRate            | string (sdk.Dec)      | "0.003000000000000000"
WithdrawFeeRate        | string (sdk.Dec)      | "0.000000000000000000"
MaxOrderAmountRatio    | string (sdk.Dec)      | "0.100000000000000000"
UnitBatchHeight        | uint32                | 1
CircuitBreakerEnabled  | bool                  | false

## PoolTypes

List of available PoolType

```go
type PoolType struct {
    Id                    uint32
    Name                  string
    MinReserveCoinNum     uint32
    MaxReserveCoinNum     uint32
    Description           string
}
```

## MinInitDepositAmount

Minimum number of coins to be deposited to the liquidity pool upon pool creation.

## InitPoolCoinMintAmount

Initial mint amount of pool coin on pool creation.

## MaxReserveCoinAmount

Limit the size of each liquidity pool. The deposit transaction fails if the total reserve coin amount after the deposit is larger than the reserve coin amount. 

The default value of zero means no limit. 

**Note:** Especially in the early phases of liquidity module adoption, set `MaxReserveCoinAmount` to a non-zero value to minimize risk on error or exploitation.

## PoolCreationFee

Fee paid for to create a LiquidityPool creation. This fee prevents spamming and is collected in in the community pool of the distribution module. 

## SwapFeeRate

Swap fee rate for every executed swap. When a swap is requested, the swap fee is reserved: 

- Half reserved as `OfferCoinFee`
- Half reserved as `ExchangedCoinFee`

The swap fee is collected when a batch is executed. 

## WithdrawFeeRate

Reserve coin withdrawal with less proportion by `WithdrawFeeRate`. This fee prevents attack vectors from repeated deposit/withdraw transactions. 

## MaxOrderAmountRatio

Maximum ratio of reserve coins that can be ordered at a swap order.

## UnitBatchHeight

The smallest unit batch size for every liquidity pool.

## CircuitBreakerEnabled

The intention of circuit breaker is to have a contingency plan for a running network which maintains network liveness. This parameter enables or disables `MsgCreatePool`, `MsgDepositWithinBatch` and `MsgSwapWithinBatch` message types in liquidity module.
# Constant Variables

Key                 | Type   | Constant Value
------------------- | ------ | --------------
CancelOrderLifeSpan | int64  | 0
MinReserveCoinNum   | uint32 | 2
MaxReserveCoinNum   | uint32 | 2

## CancelOrderLifeSpan

The life span of swap orders in block heights.

## MinReserveCoinNum, MaxReserveCoinNum

The mininum and maximum number of reserveCoins for `PoolType`.
