<!-- order: 7 -->

 # Events

The liquidity module emits the following events.

## Handlers

### MsgCreatePool

Type        | Attribute Key   | Attribute Value
----------- | --------------- | ------------------------
create_pool | pool_id         | {poolId}
create_pool | pool_type_id    | {poolTypeId}
create_pool | pool_name       | {AttributeValuePoolName}
create_pool | reserve_account | {reserveAccountAddress}
create_pool | deposit_coins   | {depositCoins}
create_pool | pool_coin_denom | {poolCoinDenom}
message     | module          | liquidity
message     | action          | create_pool
message     | sender          | {senderAddress}

### MsgDepositWithinBatch

Type                 | Attribute Key | Attribute Value
-------------------- | ------------- | --------------------
deposit_within_batch | pool_id       | {poolId}
deposit_within_batch | batch_index   | {batchIndex}
deposit_within_batch | msg_index     | {depositMsgIndex}
deposit_within_batch | deposit_coins | {depositCoins}
message              | module        | liquidity
message              | action        | deposit_within_batch
message              | sender        | {senderAddress}

### MsgWithdrawWithinBatch

Type                  | Attribute Key    | Attribute Value
--------------------- | ---------------- | ---------------------
withdraw_within_batch | pool_id          | {poolId}
withdraw_within_batch | batch_index      | {batchIndex}
withdraw_within_batch | msg_index        | {withdrawMsgIndex}
withdraw_within_batch | pool_coin_denom  | {poolCoinDenom}
withdraw_within_batch | pool_coin_amount | {poolCoinAmount}
message               | module           | liquidity
message               | action           | withdraw_within_batch
message               | sender           | {senderAddress}

### MsgSwapWithinBatch

Type              | Attribute Key     | Attribute Value
----------------- | ----------------- | -----------------
swap_within_batch | pool_id           | {poolId}
swap_within_batch | batch_index       | {batchIndex}
swap_within_batch | msg_index         | {swapMsgIndex}
swap_within_batch | swap_type_id      | {swapTypeId}
swap_within_batch | offer_coin_denom  | {offerCoinDenom}
swap_within_batch | offer_coin_amount | {offerCoinAmount}
swap_within_batch | demand_coin_denom | {demandCoinDenom}
swap_within_batch | order_price       | {orderPrice}
message           | module            | liquidity
message           | action            | swap_within_batch
message           | sender            | {senderAddress}

## EndBlocker

### Batch Result for MsgDepositWithinBatch

Type            | Attribute Key    | Attribute Value
--------------- | ---------------- | ------------------
deposit_to_pool | pool_id          | {poolId}
deposit_to_pool | batch_index      | {batchIndex}
deposit_to_pool | msg_index        | {depositMsgIndex}
deposit_to_pool | depositor        | {depositorAddress}
deposit_to_pool | accepted_coins   | {acceptedCoins}
deposit_to_pool | refunded_coins   | {refundedCoins}
deposit_to_pool | pool_coin_denom  | {poolCoinDenom}
deposit_to_pool | pool_coin_amount | {poolCoinAmount}
deposit_to_pool | success          | {success}

### Batch Result for MsgWithdrawWithinBatch

| Type               | Attribute Key      | Attribute Value     |
| ------------------ | ------------------ | ------------------- |
| withdraw_from_pool | pool_id            | {poolId}            |
| withdraw_from_pool | batch_index        | {batchIndex}        |
| withdraw_from_pool | msg_index          | {withdrawMsgIndex}  |
| withdraw_from_pool | withdrawer         | {withdrawerAddress} |
| withdraw_from_pool | pool_coin_denom    | {poolCoinDenom}     |
| withdraw_from_pool | pool_coin_amount   | {poolCoinAmount}    |
| withdraw_from_pool | withdraw_coins     | {withdrawCoins}     |
| withdraw_from_pool | withdraw_fee_coins | {withdrawFeeCoins}  |
| withdraw_from_pool | success            | {success}           |

### Batch Result for MsgSwapWithinBatch

Type            | Attribute Key                  | Attribute Value
--------------- | ------------------------------ | ----------------------------
swap_transacted | pool_id                        | {poolId}
swap_transacted | batch_index                    | {batchIndex}
swap_transacted | msg_index                      | {swapMsgIndex}
swap_transacted | swap_requester                 | {swapRequesterAddress}
swap_transacted | swap_type_id                   | {swapTypeId}
swap_transacted | offer_coin_denom               | {offerCoinDenom}
swap_transacted | offer_coin_amount              | {offerCoinAmount}
swap_transacted | exchanged_coin_denom           | {exchangedCoinDenom}
swap_transacted | order_price                    | {orderPrice}
swap_transacted | swap_price                     | {swapPrice}
swap_transacted | transacted_coin_amount         | {transactedCoinAmount}
swap_transacted | remaining_offer_coin_amount    | {remainingOfferCoinAmount}
swap_transacted | exchanged_offer_coin_amount    | {exchangedOfferCoinAmount}
swap_transacted | exchanged_demand_coin_amount   | {exchangedDemandCoinAmount}
swap_transacted | offer_coin_fee_amount          | {offerCoinFeeAmount}
swap_transacted | exchanged_coin_fee_amount      | {exchangedCoinFeeAmount}
swap_transacted | reserved_offer_coin_fee_amount | {reservedOfferCoinFeeAmount}
swap_transacted | order_expiry_height            | {orderExpiryHeight}
swap_transacted | success                        | {success}

<!-- remove for v1 ### Cancel Result for MsgSwapWithinBatch on Batch The spec, msg for cancellation of the swap order will be added from v2 | Type | Attribute Key | Attribute Value | | ----------- | ------------------------------ | ---------------------------- | | swap_cancel | pool_id | {poolId} | | swap_cancel | batch_index | {batchIndex} | | swap_cancel | msg_index | {swapMsgIndex} | | swap_cancel | swap_requester | {swapRequesterAddress} | | swap_cancel | swap_type_id | {swapTypeId} | | swap_cancel | offer_coin_denom | {offerCoinDenom} | | swap_cancel | offer_coin_amount | {offerCoinAmount} | | swap_cancel | offer_coin_fee_amount | {offerCoinFeeAmount} | | swap_cancel | reserved_offer_coin_fee_amount | {reservedOfferCoinFeeAmount} | | swap_cancel | order_price | {orderPrice} | | swap_cancel | swap_price | {swapPrice} | | swap_cancel | cancelled_coin_amount | {cancelledOfferCoinAmount} | | swap_cancel | remaining_offer_coin_amount | {remainingOfferCoinAmount} | | swap_cancel | order_expiry_height | {orderExpiryHeight} | | swap_cancel | success | {success} | -->
