package ics

import (
	"testing"

	app "github.com/cosmos/gaia/v8/app"
	appConsumer "github.com/cosmos/interchain-security/app/consumer"
	"github.com/cosmos/interchain-security/tests/e2e"
	icstestingutils "github.com/cosmos/interchain-security/testutil/ibc_testing"
	"github.com/stretchr/testify/suite"
)

func TestCCVTestSuite(t *testing.T) {
	// Pass in concrete app types that implement the interfaces defined in https://github.com/cosmos/interchain-security/testutil/e2e/interfaces.go
	// IMPORTANT: the concrete app types passed in as type parameters here must match the
	// concrete app types returned by the relevant app initers.
	ccvSuite := e2e.NewCCVTestSuite[*app.GaiaApp, *appConsumer.App](
		// Pass in ibctesting.AppIniters for gaia (provider) and consumer.
		icstestingutils.GaiaAppIniter, icstestingutils.ConsumerAppIniter, []string{})

	// Run tests
	suite.Run(t, ccvSuite)
}
