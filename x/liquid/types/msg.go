package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = (*MsgWithdrawTokenizeShareRecordReward)(nil)
	_ sdk.Msg = (*MsgWithdrawAllTokenizeShareRecordReward)(nil)
)

func NewMsgWithdrawTokenizeShareRecordReward(ownerAddr string, recordID uint64) *MsgWithdrawTokenizeShareRecordReward {
	return &MsgWithdrawTokenizeShareRecordReward{
		OwnerAddress: ownerAddr,
		RecordId:     recordID,
	}
}

func NewMsgWithdrawAllTokenizeShareRecordReward(ownerAddr string) *MsgWithdrawAllTokenizeShareRecordReward {
	return &MsgWithdrawAllTokenizeShareRecordReward{
		OwnerAddress: ownerAddr,
	}
}
