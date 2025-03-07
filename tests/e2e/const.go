package e2e

import "github.com/cosmos/gaia/v23/types"

// ica tests
const (
	ICASendTransactionFileName = "execute_ica_transaction.json"
	connectionID               = "connection-0"
	icaChannel                 = "channel-1"
)

// rate limit tests
const (
	proposalAddRateLimitAtomFilename     = "proposal_add_rate_limit_atom.json"
	proposalAddRateLimitStakeFilename    = "proposal_add_rate_limit_stake.json"
	proposalUpdateRateLimitAtomFilename  = "proposal_update_rate_limit_atom.json"
	proposalResetRateLimitAtomFilename   = "proposal_reset_rate_limit_atom.json"
	proposalRemoveRateLimitAtomFilename  = "proposal_remove_rate_limit_atom.json"
	proposalRemoveRateLimitStakeFilename = "proposal_remove_rate_limit_stake.json"
)

// light client tests
const (
	proposalStoreWasmLightClientFilename = "proposal_store_wasm_light_client.json"
)

// general testing
const (
	gaiadBinary    = "gaiad"
	txCommand      = "tx"
	queryCommand   = "query"
	keysCommand    = "keys"
	gaiaHomePath   = "/home/nonroot/.gaia"
	photonDenom    = "photon"
	uatomDenom     = types.UAtomDenom
	stakeDenom     = "stake"
	initBalanceStr = "110000000000stake,100000000000000000photon,100000000000000000uatom"
	minGasPrice    = "0.005"
	// the test basefee in genesis is the same as minGasPrice
	// global fee lower/higher than min_gas_price
	initialBaseFeeAmt               = "0.005"
	gas                             = 200000
	govProposalBlockBuffer          = 35
	relayerAccountIndexHermes       = 0
	numberOfEvidences               = 10
	slashingShares            int64 = 10000

	proposalMaxTotalBypassFilename   = "proposal_max_total_bypass.json"
	proposalCommunitySpendFilename   = "proposal_community_spend.json"
	proposalLSMParamUpdateFilename   = "proposal_lsm_param_update.json"
	proposalBlocksPerEpochFilename   = "proposal_blocks_per_epoch.json"
	proposalFailExpedited            = "proposal_fail_expedited.json"
	proposalExpeditedSoftwareUpgrade = "proposal_expedited_software_upgrade.json"
	proposalSoftwareUpgrade          = "proposal_software_upgrade.json"
	proposalCancelSoftwareUpgrade    = "proposal_cancel_software_upgrade.json"

	// proposalAddConsumerChainFilename    = "proposal_add_consumer.json"
	// proposalRemoveConsumerChainFilename = "proposal_remove_consumer.json"

	hermesBinary              = "hermes"
	hermesConfigWithGasPrices = "/root/.hermes/config.toml"
	hermesConfigNoGasPrices   = "/root/.hermes/config-zero.toml"
	transferPort              = "transfer"
	transferChannel           = "channel-0"

	v2TransferClient = "08-wasm-1"

	govAuthority = "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"
)
