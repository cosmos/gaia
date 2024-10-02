package chainsuite

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"

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
	ValidatorCount         = 6
	UpgradeDelta           = 30
	ValidatorFunds         = 11_000_000_000
	ChainSpawnWait         = 155 * time.Second
	SlashingWindowConsumer = 20
	BlocksPerDistribution  = 10
)

func (c SuiteConfig) Merge(other SuiteConfig) SuiteConfig {
	if c.ChainSpec == nil {
		c.ChainSpec = other.ChainSpec
	} else if other.ChainSpec != nil {
		c.ChainSpec.ChainConfig = c.ChainSpec.MergeChainSpecConfig(other.ChainSpec.ChainConfig)
		if other.ChainSpec.Name != "" {
			c.ChainSpec.Name = other.ChainSpec.Name
		}
		if other.ChainSpec.ChainName != "" {
			c.ChainSpec.ChainName = other.ChainSpec.ChainName
		}
		if other.ChainSpec.Version != "" {
			c.ChainSpec.Version = other.ChainSpec.Version
		}
		if other.ChainSpec.NoHostMount != nil {
			c.ChainSpec.NoHostMount = other.ChainSpec.NoHostMount
		}
		if other.ChainSpec.NumValidators != nil {
			c.ChainSpec.NumValidators = other.ChainSpec.NumValidators
		}
		if other.ChainSpec.NumFullNodes != nil {
			c.ChainSpec.NumFullNodes = other.ChainSpec.NumFullNodes
		}
	}
	c.UpgradeOnSetup = other.UpgradeOnSetup
	c.CreateRelayer = other.CreateRelayer
	c.Scope = other.Scope
	return c
}

func DefaultGenesisAmounts(denom string) func(i int) (types.Coin, types.Coin) {
	return func(i int) (types.Coin, types.Coin) {
		return types.Coin{
				Denom:  denom,
				Amount: sdkmath.NewInt(ValidatorFunds),
			}, types.Coin{
				Denom: denom,
				Amount: sdkmath.NewInt([ValidatorCount]int64{
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

func DefaultSuiteConfig(env Environment) SuiteConfig {
	fullNodes := 0
	validators := ValidatorCount
	var repository string
	if env.DockerRegistry == "" {
		repository = env.GaiaImageName
	} else {
		repository = fmt.Sprintf("%s/%s", env.DockerRegistry, env.GaiaImageName)
	}
	return SuiteConfig{
		ChainSpec: &interchaintest.ChainSpec{
			Name:          "gaia",
			NumFullNodes:  &fullNodes,
			NumValidators: &validators,
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
					UidGid:     "1025:1025", // this is the user in heighliner docker images
				}},
				ModifyGenesis:        cosmos.ModifyGenesis(DefaultGenesis()),
				ModifyGenesisAmounts: DefaultGenesisAmounts(Uatom),
			},
		},
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
		cosmos.NewGenesisKV("app_state.provider.params.slash_meter_replenish_period", "2s"),
		cosmos.NewGenesisKV("app_state.provider.params.slash_meter_replenish_fraction", "1.00"),
		cosmos.NewGenesisKV("app_state.provider.params.blocks_per_epoch", "1"),
		cosmos.NewGenesisKV("app_state.feemarket.params.min_base_gas_price", strings.TrimSuffix(GasPrices, Uatom)),
		cosmos.NewGenesisKV("app_state.feemarket.state.base_gas_price", strings.TrimSuffix(GasPrices, Uatom)),
		cosmos.NewGenesisKV("app_state.feemarket.params.fee_denom", Uatom),
		cosmos.NewGenesisKV("app_state.wasm.params.code_upload_access.permission", "Nobody"),
		cosmos.NewGenesisKV("app_state.wasm.params.instantiate_default_permission", "AnyOfAddresses"),
	}
}
