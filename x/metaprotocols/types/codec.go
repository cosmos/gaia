package types

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
	tx "github.com/cosmos/cosmos-sdk/types/tx"
)

// RegisterInterfaces adds the x/metaprotocols module's interfaces to the provided InterfaceRegistry
// The ExtendedData interface is registered so that the TxExtensionOptionsI can be properly encoded and decoded
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*tx.TxExtensionOptionI)(nil),
		&ExtensionData{},
	)
}
