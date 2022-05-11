//go:build test_amino
// +build test_amino

package params

// This function should be used only internally (in the SDK).
// App user should'nt create new codecs - use the app.AppCodec instead.
// [DEPRECATED]
// func MakeTestEncodingConfig() EncodingConfig {
// 	cdc := codec.NewLegacyAmino()
// 	interfaceRegistry := types.NewInterfaceRegistry()
// 	marshaler := codec.NewAminoCodec(cdc)
// 	return EncodingConfig{
// 		InterfaceRegistry: interfaceRegistry,
// 		Marshaler:         marshaler,
// 		TxConfig:          legacytx.StdTxConfig{Cdc: cdc},
// 		Amino:             cdc,
// 	}
// }
