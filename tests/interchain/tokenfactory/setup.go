package interchaintest

import (
	tokenfactorytypes "github.com/strangelove-ventures/tokenfactory/x/tokenfactory/types"

	sdkmath "cosmossdk.io/math"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos/wasm"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"

	sdktestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

var (
	votingPeriod     = "15s"
	maxDepositPeriod = "10s"

	accAddr     = "cosmos1hj5fveer5cjtn4wd6wstzugjfdxzl0xpxvjjvr"
	accMnemonic = "decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry"

	CosmosGovModuleAcc = "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"

	vals      = 1
	fullNodes = 0

	DefaultGenesis = []cosmos.GenesisKV{
		cosmos.NewGenesisKV("app_state.gov.params.voting_period", votingPeriod),
		cosmos.NewGenesisKV("app_state.gov.params.max_deposit_period", maxDepositPeriod),
		cosmos.NewGenesisKV("app_state.tokenfactory.params.denom_creation_fee", nil),
		cosmos.NewGenesisKV("app_state.tokenfactory.params.denom_creation_gas_consume", "1"),
		cosmos.NewGenesisKV("consensus.params.abci.vote_extensions_enable_height", "1"),
		// inflation of 0 allows for SudoMints. This is enabled by default
		cosmos.NewGenesisKV("app_state.mint.minter.inflation", sdkmath.LegacyZeroDec()),
		cosmos.NewGenesisKV("app_state.mint.params.inflation_rate_change", sdkmath.LegacyZeroDec()), // else it will increase slowly
		cosmos.NewGenesisKV("app_state.mint.params.inflation_min", sdkmath.LegacyZeroDec()),
		// TODO: inflation_max, blocks_per_year?
	}

	// `make local-image`
	LocalChainConfig = ibc.ChainConfig{
		Type:    "cosmos",
		Name:    "tokenfactory",
		ChainID: "tokenfactory-2",
		Images: []ibc.DockerImage{
			{
				Repository: "tokenfactory",
				Version:    "local",
				UidGid:     "1025:1025",
			},
		},
		Bin:            "tokend",
		Bech32Prefix:   "cosmos",
		Denom:          "token",
		GasPrices:      "0token",
		GasAdjustment:  1.3,
		TrustingPeriod: "508h",
		NoHostMount:    false,
		EncodingConfig: AppEncoding(),
		ModifyGenesis:  cosmos.ModifyGenesis(DefaultGenesis),
	}

	DefaultGenesisAmt = sdkmath.NewInt(10_000_000)
)

func AppEncoding() *sdktestutil.TestEncodingConfig {
	enc := wasm.WasmEncoding()

	tokenfactorytypes.RegisterInterfaces(enc.InterfaceRegistry)

	return enc
}
