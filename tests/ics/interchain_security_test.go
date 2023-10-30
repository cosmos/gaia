package ics

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"

	appConsumer "github.com/cosmos/interchain-security/v2/app/consumer"
	ibctesting "github.com/cosmos/interchain-security/v2/legacy_ibc_testing/testing"
	"github.com/cosmos/interchain-security/v2/tests/integration"
	icstestingutils "github.com/cosmos/interchain-security/v2/testutil/ibc_testing"

	gaiaApp "github.com/cosmos/gaia/v14/app"
)

func TestCCVTestSuite(t *testing.T) {
	// Pass in concrete app types that implement the interfaces defined in https://github.com/cosmos/interchain-security/testutil/integration/interfaces.go
	// IMPORTANT: the concrete app types passed in as type parameters here must match the
	// concrete app types returned by the relevant app initers.
	ccvSuite := integration.NewCCVTestSuite[*gaiaApp.GaiaApp, *appConsumer.App](
		// Pass in ibctesting.AppIniters for gaia (provider) and consumer.
		GaiaAppIniter, icstestingutils.ConsumerAppIniter, []string{})

	// Run tests
	suite.Run(t, ccvSuite)
}

// GaiaAppIniter implements ibctesting.AppIniter for the gaia app
func GaiaAppIniter() (ibctesting.TestingApp, map[string]json.RawMessage) {
	encoding := gaiaApp.MakeTestEncodingConfig()
	app := gaiaApp.NewGaiaApp(log.NewNopLogger(), tmdb.NewMemDB(), nil, true, map[int64]bool{},
		gaiaApp.DefaultNodeHome, 5, encoding, gaiaApp.EmptyAppOptions{})
	testApp := ibctesting.TestingApp(app)
	return testApp, gaiaApp.NewDefaultGenesisState()
}
