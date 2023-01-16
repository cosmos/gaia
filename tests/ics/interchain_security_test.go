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
	ccvSuite := e2e.NewCCVTestSuite[*app.GaiaApp, *appConsumer.App](
		// Pass in ibctesting.AppIniters for gaia (provider) and consumer.
		icstestingutils.GaiaAppIniter, icstestingutils.ConsumerAppIniter, []string{})

	suite.Run(t, ccvSuite)
}
