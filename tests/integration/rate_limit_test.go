package integration

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gaiaApp "github.com/cosmos/gaia/v23/app"
	"github.com/cosmos/ibc-apps/modules/rate-limiting/v10/types"
	ibctesting "github.com/cosmos/ibc-go/v10/testing"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

const ()

type RateLimitTestSuite struct {
	suite.Suite
	coordinator *ibctesting.Coordinator
	chain       *ibctesting.TestChain
	app         *gaiaApp.GaiaApp
}

func TestRateLimitTestSuite(t *testing.T) {
	ratelimitsuite := &RateLimitTestSuite{}
	suite.Run(t, ratelimitsuite)
}

func (suite *RateLimitTestSuite) SetupTest() {
	ibctesting.DefaultTestingAppInit = GaiaAppIniter
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 1)

	chain, ok := suite.coordinator.Chains[ibctesting.GetChainID(1)]
	suite.Require().True(ok, "chain not found")
	suite.chain = chain
	suite.chain.ProposedHeader.ProposerAddress = sdk.ConsAddress(suite.chain.Vals.Validators[0].Address)

	app, ok := chain.App.(*gaiaApp.GaiaApp)
	suite.Require().True(ok, "expected App to be GaiaApp")
	suite.app = app
}

func (suite *RateLimitTestSuite) TestRateLimitResetAfterWindow() {
	ctx := suite.chain.GetContext()

	suite.app.RatelimitKeeper.SetRateLimit(ctx, types.RateLimit{
		Path: &types.Path{
			Denom:             "uatom",
			ChannelOrClientId: "channel-0",
		},
		Quota: &types.Quota{
			MaxPercentSend: math.NewInt(50),
			MaxPercentRecv: math.NewInt(50),
			DurationHours:  6,
		},
		Flow: &types.Flow{
			Inflow:       math.NewInt(1),
			Outflow:      math.NewInt(1),
			ChannelValue: math.ZeroInt(),
		},
	})

	suite.app.RatelimitKeeper.SetHourEpoch(ctx, types.HourEpoch{
		EpochNumber:      0,
		Duration:         1 * time.Hour,
		EpochStartTime:   suite.chain.LatestCommittedHeader.GetTime(),
		EpochStartHeight: int64(suite.chain.LatestCommittedHeader.GetHeight().GetRevisionHeight()),
	}) //set epoch start at current time

	currentRateLimit, found := suite.app.RatelimitKeeper.GetRateLimit(ctx, "uatom", "channel-0")
	suite.Require().True(found)
	suite.Require().Equal(uint64(6), currentRateLimit.Quota.DurationHours)
	suite.Require().Equal(int64(1), currentRateLimit.Flow.Outflow.Int64())
	suite.Require().Equal(int64(1), currentRateLimit.Flow.Inflow.Int64())

	currentEpoch := suite.app.RatelimitKeeper.GetHourEpoch(ctx)
	suite.Require().Equal(uint64(0), currentEpoch.EpochNumber)

	suite.coordinator.IncrementTimeBy(5 * time.Hour)
	suite.coordinator.CommitNBlocks(suite.chain, 6) // one incremented per hour every beginblock, should only have +5 (==hours elapsed) even with 5 blocks committed

	currentEpoch = suite.app.RatelimitKeeper.GetHourEpoch(ctx)
	suite.Require().Equal(uint64(5), currentEpoch.EpochNumber)

	currentRateLimit, found = suite.app.RatelimitKeeper.GetRateLimit(ctx, "uatom", "channel-0")
	suite.Require().True(found)
	suite.Require().Equal(int64(1), currentRateLimit.Flow.Outflow.Int64())
	suite.Require().Equal(int64(1), currentRateLimit.Flow.Inflow.Int64())

	suite.coordinator.IncrementTimeBy(1 * time.Hour)
	suite.coordinator.CommitNBlocks(suite.chain, 1)

	currentEpoch = suite.app.RatelimitKeeper.GetHourEpoch(ctx)
	suite.Require().Equal(uint64(6), currentEpoch.EpochNumber)

	currentRateLimit, found = suite.app.RatelimitKeeper.GetRateLimit(ctx, "uatom", "channel-0")
	suite.Require().True(found)
	suite.Require().Equal(int64(0), currentRateLimit.Flow.Outflow.Int64())
	suite.Require().Equal(int64(0), currentRateLimit.Flow.Inflow.Int64())
}
