package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the necessary x/liquid interfaces
// and concrete types on the provided LegacyAmino codec. These types are used
// for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "gaia/x/liquid/MsgUpdateParams")
	legacy.RegisterAminoMsg(cdc, &MsgTokenizeShares{}, "gaia/MsgTokenizeShares")
	legacy.RegisterAminoMsg(cdc, &MsgRedeemTokensForShares{}, "gaia/MsgRedeemTokensForShares")
	legacy.RegisterAminoMsg(cdc, &MsgTransferTokenizeShareRecord{}, "gaia/MsgTransferTokenizeShareRecord")
	legacy.RegisterAminoMsg(cdc, &MsgDisableTokenizeShares{}, "gaia/MsgDisableTokenizeShares")
	legacy.RegisterAminoMsg(cdc, &MsgEnableTokenizeShares{}, "gaia/MsgEnableTokenizeShares")
	// TODO eric I haven't included UnbondValidator
	// legacy.RegisterAminoMsg(cdc, &MsgUnbondValidator{}, "cosmos-sdk/MsgUnbondValidator")
	legacy.RegisterAminoMsg(cdc, &MsgWithdrawTokenizeShareRecordReward{}, "gaia/MsgWithdrawTokenizeReward")
	legacy.RegisterAminoMsg(cdc, &MsgWithdrawAllTokenizeShareRecordReward{}, "gaia/MsgWithdrawAllTokenizeReward")

	cdc.RegisterConcrete(Params{}, "gaia/x/liquid/Params", nil)
}

// RegisterInterfaces registers the x/liquid interfaces with the interface registry
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
