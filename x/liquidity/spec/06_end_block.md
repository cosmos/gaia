<!-- order: 6 -->

 # Before-End-Block

These operations occur before the end-block operations for the liquidity module.

## Append messages to LiquidityPoolBatch

After successful message verification and coin `escrow` process, the incoming `MsgDepositWithinBatch`, `MsgWithdrawWithinBatch`, and `MsgSwapWithinBatch` messages are appended to the current `PoolBatch` of the corresponding `Pool`.

# End-Block

End-block operations for the Liquidity Module.

## Execute LiquidityPoolBatch upon execution heights

If there are `{*action}MsgState` messages that have not yet executed in the `PoolBatch` for each `Pool`, the `PoolBatch` is executed. This batch contains one or more `DepositLiquidityPool`, `WithdrawLiquidityPool`, and `SwapExecution` processes.

### Transact and refund for each message

A liquidity module escrow account holds coins temporarily and releases them when state changes. Refunds from the escrow account are made for cancellations, partial cancellations, expiration, and failed messages.

### Set states for each message according to the results

After transact and refund transactions occur for each message, update the state of each `{*action}MsgState` message according to the results.

Even if the message is completed or expired:

1. Set the `ToBeDeleted` state value as true instead of deleting the message directly from the `end-block`
2. Delete the messages that have `ToBeDeleted` state from the begin-block in the next block so that each message with result state in the block can be stored to kvstore.

This process allows searching for the past messages that have this result state. Searching is supported when the kvstore is not pruning.
