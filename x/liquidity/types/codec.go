package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers concrete types on the codec.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreatePool{}, "liquidity/MsgCreatePool", nil)
	cdc.RegisterConcrete(&MsgDepositWithinBatch{}, "liquidity/MsgDepositWithinBatch", nil)
	cdc.RegisterConcrete(&MsgWithdrawWithinBatch{}, "liquidity/MsgWithdrawWithinBatch", nil)
	cdc.RegisterConcrete(&MsgSwapWithinBatch{}, "liquidity/MsgSwapWithinBatch", nil)
}

// RegisterInterfaces registers the x/liquidity interface types with the
// interface registry
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreatePool{},
		&MsgDepositWithinBatch{},
		&MsgWithdrawWithinBatch{},
		&MsgSwapWithinBatch{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// legacy amino codecs
var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/liquidity module codec. Note, the
	// codec should ONLY be used in certain instances of tests and for JSON
	// encoding as Amino is still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/liquidity
	// and defined at the application level.
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
