<!-- order: 1 -->

 # Concepts

## Liquidity Module

The liquidity module is a module that can be used on any Cosmos SDK-based application. The liquidity module implements a decentralized exchange (DEX) that serves liquidity providing and coin swap functions. Anyone can create a liquidity pool with a pair of coins, provide liquidity by depositing reserve coins into the liquidity pool, and trade coins using the liquidity pool. All of the logic is designed to always protect the pool investors.

## Liquidity Pool

A liquidity pool is a coin reserve that contains two different types of coins in a trading pair. The trading pair has to be unique. A liquidity provider can be anyone (permissionless) who provides liquidity by depositing reserve coins into the pool. The liquidity provider earns the accumulated swap fees with respect to their pool share. The pool share is represented as possession of pool coins. All matchable swap requests are expected to be executed and unmatched swap requests are removed.
## Equivalent Swap Price Model (ESPM)

The liquidity module is a Cosmos SDK implementation of an AMM system with a novel economic model called the Equivalent Swap Price Model (ESPM).

The key distinguishing feature of the ESPM model from the Constant Product Market Maker (CPMM) model is the implementation of a hybrid system that combines an orderbook model exchange with a simple liquidity pool model that governs the order book with a set of order rules and performs execution in batches. In the ESPM, the pool price is always equal to the last swap price which reduces opportunities for arbitrage.

The ESPM model is intended to provide protection against price volatility, transaction ordering vulnerabilities, and losses due to arbitrage. AMMs such as Uniswap do not provide this level of protection.

## Batch Execution

The liquidity module uses a batch execution methodology. Deposits, withdrawals, and swap orders are accumulated in a liquidity pool for a pre-defined period that is one or more blocks in length. Orders are then added to the pool and executed at the end of the batch. The size of each batch is configured by using the `UnitBatchSize` governance parameter.

## Price Discovery

Swap prices in liquidity pools are determined by the current pool coin reserves and the requested swap amount. Arbitrageurs buy or sell coins in liquidity pools to gain instant profit that results in real-time price discovery of liquidity pools.

## Escrow Process

The liquidity module uses a module account that acts as an escrow account. The module account holds and releases the coin amount during batch execution.

## Refund 

The liquidity module has refund functions when deposit, withdraw, or swap batch states are not successfully executed.
Read [the batch transaction logic](https://github.com/tendermint/liquidity/blob/e8ab2f4d75079157d008eba9f310b199573eed28/x/liquidity/keeper/batch.go#L83-L127) in the code for more context.
## Fees

You set liquidity module fees for pool creation, withdrawal, and swap in genesis state. These fees can be updated by the governance proposal.
### PoolCreationFee

The liquidity module pool creation fee set by the `PoolCreationFee` parameter is paid on pool creation. The purpose of this fee is to prevent users from creating useless pools and making limitless transactions. The funds from this fee go to the community fund.
### WithdrawalFeeRate

The liquidity module has `WithdrawFeeRate` parameter that is paid upon withdrawal. The purpose of this fee is to prevent from making limitless withdrawals.

### SwapFeeRate

Swap fees are paid upon swap orders. They are accumulated in the pools and are shared among the liquidity providers. The liquidity module implements half-half-fee mechanism that minimizes the impact of fee payment process. Read [the issue about fees in half offer coins, half exchanged coins](https://github.com/tendermint/liquidity/issues/41) to have more context.
## Pool Identification

The pools in the liquidity module are identified with:
### PoolName

- Concatenate two different reserve coin denoms and pool type id and forward slash `/` separator. 
  - Example: `uatom/stake/1`
### PoolReserveAccount

- `sdk.AccAddress(crypto.AddressHash([]byte(PoolName)))`
  - Example: `cosmos16ddqestwukv0jzcyfn3fdfq9h2wrs83cr4rfm3` (`D35A0CC16EE598F90B044CE296A405BA9C381E38`)
### PoolCoinDenom

- `fmt.Sprintf("%s%X", PoolCoinDenomPrefix, sha256.Sum256([]byte(PoolName)))`
- Use `PoolCoinDenomPrefix` for `pool`
  - Example: `poolD35A0CC16EE598F90B044CE296A405BA9C381E38837599D96F2F70C2F02A23A4`



