package keeper_test

import (
	"encoding/json"
	"github.com/cosmos/gaia/v8/app/helpers"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"

	// icaapp "github.com/cosmos/interchain-accounts/app"
	gaiaapp "github.com/cosmos/gaia/v8/app"
)

var (
	// TestAccAddress defines a resuable bech32 address for testing purposes
	// TODO: update crypto.AddressHash() when sdk uses address.Module()
	TestAccAddress = icatypes.GenerateAddress(sdk.AccAddress(crypto.AddressHash([]byte(icatypes.ModuleName))), ibctesting.FirstConnectionID, TestPortID)
	// TestOwnerAddress defines a reusable bech32 address for testing purposes
	TestOwnerAddress = "cosmos17dtl0mjt3t77kpuhg2edqzjpszulwhgzuj9ljs"
	// TestPortID defines a resuable port identifier for testing purposes
	TestPortID, _ = icatypes.NewControllerPortID(TestOwnerAddress)
	// TestVersion defines a resuable interchainaccounts version string for testing purposes
	TestVersion = string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: ibctesting.FirstConnectionID,
		HostConnectionId:       ibctesting.FirstConnectionID,
	}))
)

func init() {
	ibctesting.DefaultTestingAppInit = SetupICATestingApp
}

func SetupICATestingApp() (ibctesting.TestingApp, map[string]json.RawMessage) {
	db := dbm.NewMemDB()
	encCdc := gaiaapp.MakeTestEncodingConfig()
	app := gaiaapp.NewGaiaApp(log.NewNopLogger(), db, nil, true, map[int64]bool{}, gaiaapp.DefaultNodeHome, 5, encCdc, helpers.EmptyAppOptions{})
	return app, gaiaapp.NewDefaultGenesisState()
}

// KeeperTestSuite is a testing suite to test keeper functions
type KeeperTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain
}

func (suite *KeeperTestSuite) GetICAApp(chain *ibctesting.TestChain) *gaiaapp.GaiaApp {
	app, ok := chain.App.(*gaiaapp.GaiaApp)
	if !ok {
		panic("not ica app")
	}

	return app
}

// TestKeeperTestSuite runs all the tests within this package.
func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// SetupTest creates a coordinator with 2 test chains.
func (suite *KeeperTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(1))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(2))
}

func NewICAPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = icatypes.PortID
	path.EndpointB.ChannelConfig.PortID = icatypes.PortID
	path.EndpointA.ChannelConfig.Order = channeltypes.ORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.ORDERED
	path.EndpointA.ChannelConfig.Version = TestVersion
	path.EndpointB.ChannelConfig.Version = TestVersion

	return path
}
