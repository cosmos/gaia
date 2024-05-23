package gaia

import (
	"github.com/cosmos/cosmos-sdk/std"

	"github.com/cosmos/gaia/v18/app/params"
)

func RegisterEncodingConfig() params.EncodingConfig {
	encConfig := params.MakeEncodingConfig()

	std.RegisterLegacyAminoCodec(encConfig.Amino)
	std.RegisterInterfaces(encConfig.InterfaceRegistry)
	ModuleBasics.RegisterLegacyAminoCodec(encConfig.Amino)
	ModuleBasics.RegisterInterfaces(encConfig.InterfaceRegistry)

	return encConfig
}
