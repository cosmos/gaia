package e2e

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/gaia/v23/types"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/suite"
	"path/filepath"
)

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

var (
	gaiaConfigPath            = filepath.Join(gaiaHomePath, "config")
	stakingAmount             = math.NewInt(100000000000)
	stakingAmountCoin         = sdk.NewCoin(uatomDenom, stakingAmount)
	tokenAmount               = sdk.NewCoin(uatomDenom, math.NewInt(3300000000)) // 3,300uatom
	standardFees              = sdk.NewCoin(uatomDenom, math.NewInt(330000))     // 0.33uatom
	depositAmount             = sdk.NewCoin(uatomDenom, math.NewInt(330000000))  // 3,300uatom
	distModuleAddress         = authtypes.NewModuleAddress(distrtypes.ModuleName).String()
	govModuleAddress          = authtypes.NewModuleAddress(govtypes.ModuleName).String()
	proposalCounter           = 0
	contractsCounter          = 0
	contractsCounterPerSender = map[string]uint64{}
)

type IntegrationTestSuite struct {
	suite.Suite

	tmpDirs        []string
	chainA         *chain
	chainB         *chain
	dkrPool        *dockertest.Pool
	dkrNet         *dockertest.Network
	hermesResource *dockertest.Resource

	valResources map[string][]*dockertest.Resource
}

type AddressResponse struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Address  string `json:"address"`
	Mnemonic string `json:"mnemonic"`
}
