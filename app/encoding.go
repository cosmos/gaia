package gaia

import (
	"github.com/cosmos/gaia/v7/app/params"

	"github.com/cosmos/cosmos-sdk/std"
)

// MakeEncodingConfig creates an EncodingConfig for testing
func MakeEncodingConfig() params.EncodingConfig {
	encodingConfig := params.MakeEncodingConfig()
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	// TODO: Figure out whether these need to be added back
	// ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	// ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}
