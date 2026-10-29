// Package params preserves decoding of historical x/params governance content
// after the module was removed from Gaia in v29.0.0.
package params

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
)

// RegisterInterfaces keeps historical ParameterChangeProposal values readable.
// SDK v0.53 still provides their complete types and descriptors. When upgrading
// to SDK v0.55, replace this delegation with field-preserving local types while
// retaining the original type URLs and this registration entry point.
// This does not restore the params module or its governance execution handler.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	paramproposal.RegisterInterfaces(registry)
}
