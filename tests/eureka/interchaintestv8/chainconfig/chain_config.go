package chainconfig

import (
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
)

var DefaultChainSpecs = []*interchaintest.ChainSpec{
	// -- IBC-Go --
	IbcGoChainSpec("ibc-go-simd-1", "simd-1"),
}

func IbcGoChainSpec(name, chainId string) *interchaintest.ChainSpec {
	return &interchaintest.ChainSpec{
		ChainConfig: ibc.ChainConfig{
			Type:    "cosmos",
			Name:    name,
			ChainID: chainId,
			Images: []ibc.DockerImage{{
				Repository: "gaia",
				UidGid:     "1025:1025", // this is the user in heighliner docker images
			}},
			Bin:            "simd",
			Bech32Prefix:   "cosmos",
			Denom:          "stake",
			GasPrices:      "0.00stake",
			GasAdjustment:  1.3,
			EncodingConfig: SDKEncodingConfig(),
			ModifyGenesis:  defaultModifyGenesis(),
			TrustingPeriod: "508h",
			NoHostMount:    false,
		},
	}
}
