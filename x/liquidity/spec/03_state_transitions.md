<!-- order: 3 -->

 # State Transitions

These messages (Msg) in the liquidity module trigger state transitions.

## Coin Escrow for Liquidity Module Messages

Transaction confirmation causes state transition on the [Bank](https://docs.cosmos.network/master/modules/bank/) module. Some messages on the liquidity module require coin escrow before confirmation.

The coin escrow processes for each message type are:

### MsgDepositWithinBatch

To deposit coins into an existing `Pool`, the depositor must escrow `DepositCoins` into `LiquidityModuleEscrowAccount`.

### MsgWithdrawWithinBatch

To withdraw coins from a `Pool`, the withdrawer must escrow `PoolCoin` into `LiquidityModuleEscrowAccount`.

### MsgSwapWithinBatch

To request a coin swap, the swap requestor must escrow `OfferCoin` into `LiquidityModuleEscrowAccount`.

## LiquidityPoolBatch Execution

Batch execution causes state transitions on the `Bank` module. The following categories describe state transition executed by each process in the `PoolBatch` execution.

### Coin Swap

After a successful coin swap, coins accumulated in `LiquidityModuleEscrowAccount` for coin swaps are sent to other swap requestors(self-swap) or to the `Pool`(pool-swap). Fees are also sent to the liquidity `Pool`.

### LiquidityPool Deposit

After a successful deposit transaction, escrowed coins are sent to the `ReserveAccount` of the targeted `Pool` and new pool coins are minted and sent to the depositor.

### LiquidityPool Withdrawal

After a successful withdraw transaction, escrowed pool coins are burned and a corresponding amount of reserve coins are sent to the withdrawer from the liquidity `Pool`.

## Pseudo Algorithm for LiquidityPoolBatch Execution

If you are curious, you can see a Python simulation script on the B-Harvest [GitHub repo](https://github.com/b-harvest/Liquidity-Module-For-the-Hub/blob/master/pseudo-batch-execution-logic/batch.py).

## Swap Price Calculation

Swap execution applies a universal swap ratio for all swap requests.

Swap price calculations are used for these cases.

**Find price direction**

Variables:

- `X`: Reserve of X coin
- `Y`: Reserve of Y coin before this batch execution
- `PoolPrice` = `X`/`Y`
- `XOverLastPrice`: amount of orders that swap X for Y with order price higher than the last `PoolPrice`
- `XAtLastPrice`: amount of orders that swap X for Y with order price equal to the last `PoolPrice`
- `YUnderLastPrice`: amount of orders that swap Y for X with order price lower than last `PoolPrice`
- `YAtLastPrice`: amount of orders that swap Y for X with order price equal to the last `PoolPrice`

- **Increase**: swap price is increased from the last `PoolPrice`

  - `XOverLastPrice` > (`YUnderLastPrice`+`YAtLastPrice`)*`PoolPrice`

- **Decrease**: swap price is decreased from the last `PoolPrice`

  - `YUnderLastPrice` > (`XOverLastPrice`+`XAtLastPrice`)/`PoolPrice`

- **Stay**: swap price is not changed from the last `PoolPrice` when the increase and decrease inequalities do not hold

### Stay case

Variables:

- `swapPrice` = last `PoolPrice`
- `EX`: All executable orders that swap X for Y with order price equal to or greater than last `PoolPrice`
- `EY`: All executable orders that swap Y for X with order price equal or lower than last `PoolPrice`

- **ExactMatch**: If `EX` == `EY`*`swapPrice`

  - Amount of X coins matched from swap orders = `EX`
  - Amount of Y coins matched from swap orders = `EY`

- **FractionalMatch**

  - If `EX` > `EY`*`swapPrice`: Residual X order amount remains

    - Amount of X coins matched from swap orders = `EY`*`swapPrice`
    - Amount of Y coins matched from swap orders = `EY`

  - If `EY` > `EX`/`swapPrice`: Residual Y order amount remains

    - Amount of X coins matched from swap orders = `EX`
    - Amount of Y coins matched from swap orders = `EX`/`swapPrice`

### Increase case

Iteration: iterate `orderPrice(i)` of all swap orders from low to high.

Variables:

- `EX(i)`: Sum of all order amount of swap orders that swap X for Y with order price equal or higher than this `orderPrice(i)`
- `EY(i)`: Sum of all order amounts of swap orders that swap Y for X with order price equal or lower than this `orderPrice(i)`

- ExactMatch: SwapPrice is found between two orderPrices

  - `swapPrice(i)` = (`X` + 2_`EX(i)`)/(`Y` + 2_`EY(i-1)`)

    - condition1) `orderPrice(i-1)` < `swapPrice(i)` < `orderPrice(i)`

  - `PoolY(i)` = (`swapPrice(i)`_`Y` - `X`) / (2_`swapPrice(i)`)

    - condition2) `PoolY(i)` >= 0

  - If both above conditions are met, `swapPrice` is the swap price for this iteration

    - Amount of X coins matched = `EX(i)`

  - If one of these conditions doesn't hold, go to FractionalMatch

- FractionalMatch: SwapPrice is found at an orderPrice

  - `swapPrice(i)` = `orderPrice(i)`
  - `PoolY(i)` = (`swapPrice(i)`_`Y` - `X`) / (2_`swapPrice(i)`)
  - Amount of X coins matched:

    - `EX(i)` ← min[ `EX(i)`, (`EY(i)`+`PoolY(i)`)*`swapPrice(i)` ]

- Find optimized swapPrice:

  - Find `swapPrice(k)` that has the largest amount of X coins matched

    - this is our optimized swap price
    - corresponding swap result variables

      - `swapPrice(k)`, `EX(k)`, `EY(k)`, `PoolY(k)`

### Decrease case

Iteration: iterate `orderPrice(i)` of all swap orders from high to low.

Variables:

- `EX(i)`: Sum of all order amount of swap orders that swap X for Y with order price equal or higher than this `orderPrice(i)`
- `EY(i)`: Sum of all order amount of swap orders that swap Y for X with order price equal or lower than this `orderPrice(i)`

- ExactMatch: SwapPrice is found between two orderPrices

- `swapPrice(i)` = (`X` + 2_`EX(i)`)/(`Y` + 2_`EY(i-1)`)

  - condition1) `orderPrice(i)` < `swapPrice(i)` < `orderPrice(i-1)`

- `PoolX(i)` = (`X` - `swapPrice(i)`*`Y`)/2

  - condition2) `PoolX(i)` >= 0

- If both above conditions are met, `swapPrice` is the swap price for this iteration

  - Amount of Y coins matched = `EY(i)`

- If one of these conditions doesn't hold, go to FractionalMatch

- FractionalMatch: SwapPrice is found at an orderPrice

- `swapPrice(i)` = `orderPrice(i)`

- `PoolX(i)` = (`X` - `swapPrice(i)`*`Y`)/2

- Amount of Y coins matched:

  - `EY(i)` ← min[ `EY(i)`, (`EX(i)`+`PoolX(i)`)/`swapPrice(i)` ]

- Find optimized swapPrice

  - Find `swapPrice(k)` that has the largest amount of Y coins matched

    - this is our optimized swap price
    - corresponding swap result variables

      - `swapPrice(k)`, `EX(k)`, `EY(k)`, `PoolX(k)`

### Calculate matching result

- for swap orders from X to Y

  - Iteration: iterate `orderPrice(i)` of swap orders from X to Y (high to low)

    - sort by order price (high to low), sum all order amount with each `orderPrice(i)`
    - if `EX(i)` ≤ `EX(k)`

      - `fractionalRatio` = 1

    - if `EX(i)` > `EX(k)`

      - `fractionalRatio(i)` = (`EX(k)` - `EX(i-1)`) / (`EX(i)` - `EX(i-1)`)
      - break the iteration

    - matching amount for swap orders with this `orderPrice(i)`:

      - `matchingAmt` = `offerAmt` * `fractionalRatio(i)`

- for swap orders from Y to X

  - Iteration: iterate `orderPrice(i)` of swap orders from Y to X (low to high)

    - sort by order price (low to high), sum all order amount with each `orderPrice(i)`
    - if `EY(i)` ≤ `EY(k)`

      - `fractionalRatio` = 1

    - if `EY(i)` > `EY(k)`

      - `fractionalRatio(i)` = (`EY(k)` - `EY(i-1)`) / (`EY(i)` - `EY(i-1)`)
      - break the iteration

    - matching amount for swap orders with this `orderPrice(i)`:

      - `matchingAmt` = `offerAmt` * `fractionalRatio(i)`

### Swap Fee Payment

Rather than taking fee solely from `OfferCoin`, liquidity module is designed to take fees half from `OfferCoin`, and the other half from `ExchangedCoin`. This smooths out an impact of the fee payment process.
- **OfferCoin Fee Reservation ( fee before batch process, in OfferCoin )**
    - when user orders 100 Xcoin, the swap message demands
        - `OfferCoin`(in Xcoin) : 100
        - `ReservedOfferCoinFeeAmount`(in Xcoin) = `OfferCoin`*(`SwapFeeRate`/2)
    - user needs to have at least 100+100*(`SwapFeeRate`/2) amount of Xcoin to successfully commit this swap message
        - the message fails when user's balance is below this amount
- **Actual Fee Payment**
    - if 10 Xcoin is executed
        - **OfferCoin Fee Payment from Reserved OfferCoin Fee**
            - `OfferCoinFeeAmount`(in Xcoin) = (10/100)*`ReservedOfferCoinFeeAmount`
            - `ReservedOfferCoinFeeAmount` is reduced from this fee payment
        - **ExchangedCoin Fee Payment ( fee after batch process, in ExchangedCoin )**
            - `ExchangedCoinFeeAmount`(in Ycoin) = `OfferCoinFeeAmount` / `SwapPrice`
                - this is exactly equal value compared to advance fee payment assuming the current SwapPrice, to minimize the pool price impact from fee payment process

- Swap fees are proportional to the coins received from matched swap orders.
- Swap fees are sent to the liquidity pool.
- The decimal points of the swap fees are rounded up.

## Cancel unexecuted swap orders with expired CancelHeight

After execution of `PoolBatch`, all remaining swap orders with `CancelHeight` equal to or higher than current height are cancelled.

## Refund escrowed coins

Refunds are issued for escrowed coins for cancelled swap order and failed create pool, deposit, and withdraw messages.
