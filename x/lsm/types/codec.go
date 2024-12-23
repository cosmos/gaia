package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the necessary x/lsm interfaces
// and concrete types on the provided LegacyAmino codec. These types are used
// for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "gaia/x/lsm/MsgUpdateParams")
	legacy.RegisterAminoMsg(cdc, &MsgTokenizeShares{}, "gaia/x/lsm/MsgTokenizeShares")
	legacy.RegisterAminoMsg(cdc, &MsgRedeemTokensForShares{}, "gaia/x/lsm/MsgRedeemTokensForShares")
	legacy.RegisterAminoMsg(cdc, &MsgTransferTokenizeShareRecord{}, "gaia/x/lsm/MsgTransferTokenizeShareRecord")
	legacy.RegisterAminoMsg(cdc, &MsgDisableTokenizeShares{}, "gaia/x/lsm/MsgDisableTokenizeShares")
	legacy.RegisterAminoMsg(cdc, &MsgEnableTokenizeShares{}, "gaia/x/lsm/MsgEnableTokenizeShares")
	// TODO eric I haven't included UnbondValidator -- do I need?
	// legacy.RegisterAminoMsg(cdc, &MsgUnbondValidator{}, "cosmos-sdk/MsgUnbondValidator")
	legacy.RegisterAminoMsg(cdc, &MsgWithdrawTokenizeShareRecordReward{}, "gaia/x/lsm/MsgWithdrawTokenizeShareRecordReward")
	legacy.RegisterAminoMsg(cdc, &MsgWithdrawAllTokenizeShareRecordReward{}, "gaia/x/lsm/MsgWithdrawAllTokenizeShareRecordReward")

	cdc.RegisterConcrete(Params{}, "gaia/x/lsm/Params", nil)
}

// RegisterInterfaces registers the x/lsm interfaces with the interface registry
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgUpdateParams{},
		&MsgTokenizeShares{},
		&MsgRedeemTokensForShares{},
		&MsgTransferTokenizeShareRecord{},
		&MsgDisableTokenizeShares{},
		&MsgEnableTokenizeShares{},
		&MsgWithdrawTokenizeShareRecordReward{},
		&MsgWithdrawAllTokenizeShareRecordReward{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
