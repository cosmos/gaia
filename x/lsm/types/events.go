package types

// lsm module event types
const (
	EventTypeTokenizeShares              = "tokenize_shares"
	EventTypeRedeemShares                = "redeem_shares"
	EventTypeTransferTokenizeShareRecord = "transfer_tokenize_share_record"
	EventTypeWithdrawTokenizeShareReward = "withdraw_tokenize_share_reward"

	AttributeKeyValidator       = "validator"
	AttributeKeyDelegator       = "delegator"
	AttributeKeyShareOwner      = "share_owner"
	AttributeKeyShareRecordID   = "share_record_id"
	AttributeKeyAmount          = "amount"
	AttributeKeyTokenizedShares = "tokenized_shares"
	AttributeKeyWithdrawAddress = "withdraw_address"
)
