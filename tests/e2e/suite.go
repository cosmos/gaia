package e2e

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/suite"
	"path/filepath"
)

var (
	gaiaConfigPath    = filepath.Join(gaiaHomePath, "config")
	stakingAmount     = math.NewInt(100000000000)
	stakingAmountCoin = sdk.NewCoin(uatomDenom, stakingAmount)
	tokenAmount       = sdk.NewCoin(uatomDenom, math.NewInt(3300000000)) // 3,300uatom
	standardFees      = sdk.NewCoin(uatomDenom, math.NewInt(330000))     // 0.33uatom
	depositAmount     = sdk.NewCoin(uatomDenom, math.NewInt(330000000))  // 3,300uatom
	distModuleAddress = authtypes.NewModuleAddress(distrtypes.ModuleName).String()
	govModuleAddress  = authtypes.NewModuleAddress(govtypes.ModuleName).String()
)

type TestCounters struct {
	proposalCounter           int
	contractsCounter          int
	contractsCounterPerSender map[string]uint64
}

type IntegrationTestSuite struct {
	suite.Suite

	tmpDirs        []string
	chainA         *chain
	chainB         *chain
	dkrPool        *dockertest.Pool
	dkrNet         *dockertest.Network
	hermesResource *dockertest.Resource

	valResources map[string][]*dockertest.Resource

	testCounters TestCounters
}

type AddressResponse struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Address  string `json:"address"`
	Mnemonic string `json:"mnemonic"`
}
