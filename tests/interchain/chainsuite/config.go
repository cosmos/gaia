package chainsuite

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/cosmos/interchaintest/v10/testutil"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/types"
)

type ChainScope int

const (
	ChainScopeSuite ChainScope = iota
	ChainScopeTest  ChainScope = iota
)

type SuiteConfig struct {
	ChainSpec      *interchaintest.ChainSpec
	UpgradeOnSetup bool
	CreateRelayer  bool
	Scope          ChainScope
}

const (
	CommitTimeout          = 4 * time.Second
	Uatom                  = "uatom"
	Ucon                   = "ucon"
	NeutronDenom           = "untn"
	StrideDenom            = "ustr"
	GovMinDepositAmount    = 1000
	GovDepositAmount       = "5000000" + Uatom
	GovDepositPeriod       = 60 * time.Second
	GovVotingPeriod        = 80 * time.Second
	DowntimeJailDuration   = 10 * time.Second
	ProviderSlashingWindow = 10
	GasPrices              = "0.005" + Uatom
	// ValidatorCount         = 1
	UpgradeDelta           = 30
	ValidatorFunds         = 11_000_000_000
	ChainSpawnWait         = 155 * time.Second
	SlashingWindowConsumer = 20
	BlocksPerDistribution  = 10
	StrideVersion          = "v24.0.0"
	NeutronVersion         = "v3.0.2"
	TransferPortID         = "transfer"
	// This is needed because not every ics image is in the default heighliner registry
	HyphaICSRepo = "ghcr.io/hyphacoop/ics"
	ICSUidGuid   = "1025:1025"
	// Osmosis chain constants
	OsmosisDenom   = "uosmo"
	OsmosisVersion = "30.0.5"
	OsmosisRepo    = "osmolabs/osmosis"
)

// These have to be vars so we can take their address
var (
	OneValidator   int = 1
	SixValidators  int = 6
	TenValidators  int = 10
)

func MergeChainSpecs(spec, other *interchaintest.ChainSpec) *interchaintest.ChainSpec {
	if spec == nil {
		return other
	}
	if other == nil {
		return spec
	}
	spec.ChainConfig = spec.MergeChainSpecConfig(other.ChainConfig)
	// MergeChainSpecConfig doesn't handle ModifyGenesisAmounts; apply it manually.
	if other.ChainConfig.ModifyGenesisAmounts != nil {
		spec.ChainConfig.ModifyGenesisAmounts = other.ChainConfig.ModifyGenesisAmounts
	}
	if other.Name != "" {
		spec.Name = other.Name
	}
	if other.ChainName != "" {
		spec.ChainName = other.ChainName
	}
	if other.Version != "" {
		spec.Version = other.Version
	}
	if other.NoHostMount != nil {
		spec.NoHostMount = other.NoHostMount
	}
	if other.NumValidators != nil {
		spec.NumValidators = other.NumValidators
	}
	if other.NumFullNodes != nil {
		spec.NumFullNodes = other.NumFullNodes
	}
	return spec
}

func (c SuiteConfig) Merge(other SuiteConfig) SuiteConfig {
	c.ChainSpec = MergeChainSpecs(c.ChainSpec, other.ChainSpec)
	c.UpgradeOnSetup = other.UpgradeOnSetup
	c.CreateRelayer = other.CreateRelayer
	c.Scope = other.Scope
	return c
}

func DefaultGenesisAmounts(denom string) func(i int) (types.Coin, types.Coin) {
	return func(i int) (types.Coin, types.Coin) {
		if i >= SixValidators {
			panic("your chain has too many validators")
		}
		return types.Coin{
				Denom:  denom,
				Amount: sdkmath.NewInt(ValidatorFunds),
			}, types.Coin{
				Denom: denom,
				Amount: sdkmath.NewInt([]int64{
					30_000_000,
					29_000_000,
					20_000_000,
					10_000_000,
					7_000_000,
					4_000_000,
				}[i]),
			}
	}
}

// TenValidatorGenesisAmounts returns genesis amounts for a 10-validator topology:
// 5 in group A (CometBFT consensus), 3 in group B (bonded, not in CometBFT), 2 in group N (unbonded).
// Requires max_provider_consensus_validators=5 and max_validators=8 in genesis.
func TenValidatorGenesisAmounts(denom string) func(i int) (types.Coin, types.Coin) {
	return func(i int) (types.Coin, types.Coin) {
		if i >= TenValidators {
			panic("your chain has too many validators")
		}
		return types.Coin{
				Denom:  denom,
				Amount: sdkmath.NewInt(ValidatorFunds),
			}, types.Coin{
				Denom: denom,
				Amount: sdkmath.NewInt([]int64{
					30_000_000, // A: highest
					29_000_000, // A
					20_000_000, // A
					15_000_000, // A
					12_000_000, // A: lowest in consensus (boundary)
					9_000_000,  // B: bonded, not in CometBFT
					7_000_000,  // B
					5_000_000,  // B: lowest bonded (boundary)
					3_000_000,  // N: not bonded
					1_000_000,  // N: lowest
				}[i]),
			}
	}
}

func DefaultChainSpec(env Environment) *interchaintest.ChainSpec {
	fullNodes := 0
	var repository string
	if env.DockerRegistry == "" {
		repository = env.GaiaImageName
	} else {
		repository = fmt.Sprintf("%s/%s", env.DockerRegistry, env.GaiaImageName)
	}
	return &interchaintest.ChainSpec{
		Name:          "gaia",
		NumFullNodes:  &fullNodes,
		NumValidators: &OneValidator,
		Version:       env.OldGaiaImageVersion,
		ChainConfig: ibc.ChainConfig{
			Denom:         Uatom,
			GasPrices:     GasPrices,
			GasAdjustment: 2.0,
			ConfigFileOverrides: map[string]any{
				"config/config.toml": DefaultConfigToml(),
			},
			Images: []ibc.DockerImage{{
				Repository: repository,
				UIDGID:     "1025:1025", // this is the user in heighliner docker images
			}},
			ModifyGenesis:        cosmos.ModifyGenesis(DefaultGenesis()),
			ModifyGenesisAmounts: DefaultGenesisAmounts(Uatom),
		},
	}
}

func DefaultSuiteConfig(env Environment) SuiteConfig {
	return SuiteConfig{
		ChainSpec: DefaultChainSpec(env),
	}
}

func DefaultConfigToml() testutil.Toml {
	configToml := make(testutil.Toml)
	consensusToml := make(testutil.Toml)
	consensusToml["timeout_commit"] = CommitTimeout
	configToml["consensus"] = consensusToml
	configToml["block_sync"] = false
	configToml["fast_sync"] = false
	return configToml
}

func DefaultGenesis() []cosmos.GenesisKV {
	return []cosmos.GenesisKV{
		cosmos.NewGenesisKV("app_state.gov.params.voting_period", GovVotingPeriod.String()),
		cosmos.NewGenesisKV("app_state.gov.params.max_deposit_period", GovDepositPeriod.String()),
		cosmos.NewGenesisKV("app_state.gov.params.min_deposit.0.denom", Uatom),
		cosmos.NewGenesisKV("app_state.gov.params.min_deposit.0.amount", strconv.Itoa(GovMinDepositAmount)),
		cosmos.NewGenesisKV("app_state.slashing.params.signed_blocks_window", strconv.Itoa(ProviderSlashingWindow)),
		cosmos.NewGenesisKV("app_state.slashing.params.downtime_jail_duration", DowntimeJailDuration.String()),
		cosmos.NewGenesisKV("app_state.feemarket.params.min_base_gas_price", strings.TrimSuffix(GasPrices, Uatom)),
		cosmos.NewGenesisKV("app_state.feemarket.state.base_gas_price", strings.TrimSuffix(GasPrices, Uatom)),
		cosmos.NewGenesisKV("app_state.feemarket.params.fee_denom", Uatom),
		cosmos.NewGenesisKV("app_state.wasm.params.code_upload_access.permission", "Nobody"),
		cosmos.NewGenesisKV("app_state.wasm.params.instantiate_default_permission", "AnyOfAddresses"),
		cosmos.NewGenesisKV("app_state.provider.params.blocks_per_epoch", 1),
	}
}

// OsmosisGenesis returns genesis modifications for Osmosis chain
func OsmosisGenesis() []cosmos.GenesisKV {
	return []cosmos.GenesisKV{
		cosmos.NewGenesisKV("app_state.gov.params.voting_period", GovVotingPeriod.String()),
		cosmos.NewGenesisKV("app_state.gov.params.max_deposit_period", GovDepositPeriod.String()),
		cosmos.NewGenesisKV("app_state.gov.params.min_deposit.0.denom", OsmosisDenom),
		cosmos.NewGenesisKV("app_state.gov.params.min_deposit.0.amount", strconv.Itoa(GovMinDepositAmount)),
		cosmos.NewGenesisKV("app_state.slashing.params.signed_blocks_window", strconv.Itoa(ProviderSlashingWindow)),
		cosmos.NewGenesisKV("app_state.slashing.params.downtime_jail_duration", DowntimeJailDuration.String()),
	}
}

// OsmosisChainSpec returns the chain spec for Osmosis
func OsmosisChainSpec() *interchaintest.ChainSpec {
	fullNodes := 0
	return &interchaintest.ChainSpec{
		Name:          "osmosis",
		ChainName:     "osmosis",
		NumFullNodes:  &fullNodes,
		NumValidators: &OneValidator,
		Version:       OsmosisVersion,
		ChainConfig: ibc.ChainConfig{
			Type:          "cosmos",
			Bin:           "osmosisd",
			Bech32Prefix:  "osmo",
			Denom:         OsmosisDenom,
			GasPrices:     "0.025" + OsmosisDenom,
			GasAdjustment: 2.0,
			ConfigFileOverrides: map[string]any{
				"config/config.toml": DefaultConfigToml(),
			},
			Images: []ibc.DockerImage{{
				Repository: OsmosisRepo,
				Version:    OsmosisVersion,
				UIDGID:     "1025:1025",
			}},
			ModifyGenesis:        cosmos.ModifyGenesis(OsmosisGenesis()),
			ModifyGenesisAmounts: DefaultGenesisAmounts(OsmosisDenom),
		},
	}
}
