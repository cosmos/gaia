package common

import (
	"path/filepath"

	"cosmossdk.io/math"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/cosmos/gaia/v26/types"
)

// ica tests
const (
	ICASendTransactionFileName = "execute_ica_transaction.json"
	ConnectionID               = "connection-0"
	IcaChannel                 = "channel-1"
)

// rate limit tests
const (
	ProposalAddRateLimitAtomFilename     = "proposal_add_rate_limit_atom.json"
	ProposalAddRateLimitStakeFilename    = "proposal_add_rate_limit_stake.json"
	ProposalUpdateRateLimitAtomFilename  = "proposal_update_rate_limit_atom.json"
	ProposalResetRateLimitAtomFilename   = "proposal_reset_rate_limit_atom.json"
	ProposalRemoveRateLimitAtomFilename  = "proposal_remove_rate_limit_atom.json"
	ProposalRemoveRateLimitStakeFilename = "proposal_remove_rate_limit_stake.json"
)

// light client tests
const (
	ProposalStoreWasmLightClientFilename = "proposal_store_wasm_light_client.json"
)

// general testing
const (
	GaiadBinary    = "gaiad"
	TxCommand      = "tx"
	QueryCommand   = "query"
	GaiaHomePath   = "/home/nonroot/.gaia"
	UAtomDenom     = types.UAtomDenom
	StakeDenom     = "stake"
	InitBalanceStr = "110000000000stake,100000000000000000photon,100000000000000000uatom"
	MinGasPrice    = "0.005"
	// the test basefee in genesis is the same as minGasPrice
	// global fee lower/higher than min_gas_price
	InitialBaseFeeAmt               = "0.005"
	Gas                             = 200000
	GovProposalBlockBuffer          = 35
	RelayerAccountIndexHermes       = 0
	NumberOfEvidences               = 10
	SlashingShares            int64 = 10000

	ProposalCommunitySpendFilename    = "proposal_community_spend.json"
	ProposalLiquidParamUpdateFilename = "proposal_liquid_param_update.json"
	ProposalBlocksPerEpochFilename    = "proposal_blocks_per_epoch.json"
	ProposalFailExpedited             = "proposal_fail_expedited.json"
	ProposalExpeditedSoftwareUpgrade  = "proposal_expedited_software_upgrade.json"
	ProposalSoftwareUpgrade           = "proposal_software_upgrade.json"
	ProposalCancelSoftwareUpgrade     = "proposal_cancel_software_upgrade.json"

	// proposalAddConsumerChainFilename    = "proposal_add_consumer.json"
	// proposalRemoveConsumerChainFilename = "proposal_remove_consumer.json"

	hermesBinary              = "hermes"
	HermesConfigWithGasPrices = "/root/.hermes/config.toml"
	TransferPort              = "transfer"
	TransferChannel           = "channel-0"

	V2TransferClient = "08-wasm-1"
	CounterpartyID   = "client-0"

	GovAuthority = "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"
)

var (
	GaiaConfigPath    = filepath.Join(GaiaHomePath, "config")
	StakingAmountCoin = sdktypes.NewCoin(UAtomDenom, stakingAmount)
	TokenAmount       = sdktypes.NewCoin(UAtomDenom, math.NewInt(3300000000)) // 3,300uatom
	StandardFees      = sdktypes.NewCoin(UAtomDenom, math.NewInt(330000))     // 0.33uatom
	DepositAmount     = sdktypes.NewCoin(UAtomDenom, math.NewInt(330000000))  // 3,300uatom
	DistModuleAddress = authtypes.NewModuleAddress(distributiontypes.ModuleName).String()
	GovModuleAddress  = authtypes.NewModuleAddress(govtypes.ModuleName).String()
	stakingAmount     = math.NewInt(100000000000)
	AdapterAddress    = ""
	EntrypointAddress = ""
)
