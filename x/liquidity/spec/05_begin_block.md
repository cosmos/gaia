<!-- order: 5 -->

 # Begin-Block

Begin block operations for the liquidity module reinitialize batch messages that were not executed in the previous batch and delete batch messages that were executed or ready to be deleted.

## Delete pool batch messages and reset states for pool batch messages

- Delete `{*action}MsgState` messages that have `ToBeDeleted` state
- Reset states for the remaining `{*action}MsgState` messages to execute on `end-block` of the next batch index

## Reinitialize executed pool batch to next liquidity pool batch

Reinitialize the executed `PoolBatch` for the next batch. The reinitialization process includes the following actions:

- Increase state `BatchIndex` of the batch
- Reset state `BeginHeight` as current block height
- Reset state `Executed` as `false`
