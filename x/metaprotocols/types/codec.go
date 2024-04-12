package types

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
	tx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// RegisterInterfaces adds the x/metaprotocols module's interfaces to the provided InterfaceRegistry
// The ExtendedData interface is registered so that the TxExtensionOptionsI can be properly encoded and decoded
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*tx.TxExtensionOptionI)(nil),
		&ExtensionData{},
		// needs to be registered to allow parsing of historic data stored in TxExtensionOptions
		// the app does not interact with this message in any way but it performs an unmarshal which must not fail
		&authz.MsgRevoke{},
	)
}
