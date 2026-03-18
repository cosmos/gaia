package v28_0_0

import (
	"github.com/cosmos/gaia/v28/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v28.0.0"

	// providerModuleName is the ICS provider module name, used as the IBC port
	// for ICS channels. Hardcoded to avoid importing the ICS provider package.
	providerModuleName = "provider"

	// providerStoreKey is the store key used by the ICS provider module.
	// Hardcoded to avoid importing providertypes in production app wiring.
	providerStoreKey = "provider"
)

// providerParametersKey is the KV store key under which the ICS provider
// module stores its Params. Mirrors providertypes.ParametersKey() — 0xFF as
// per ICS x/ccv/provider/types/keys.go.
var providerParametersKey = []byte{0xFF}

// icsProviderParams is a minimal proto-compatible struct for decoding only
// the MaxProviderConsensusValidators field from the ICS provider Params stored
// in the provider KV store during the v28 upgrade.
// Field 12 (varint) matches the ICS Params protobuf definition.
type icsProviderParams struct {
	MaxProviderConsensusValidators int64 `protobuf:"varint,12,opt,name=max_provider_consensus_validators,proto3"`
}

func (p *icsProviderParams) Reset()         {}
func (p *icsProviderParams) String() string { return "" }
func (p *icsProviderParams) ProtoMessage()  {}

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
}
