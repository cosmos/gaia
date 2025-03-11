package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	appConsumer "github.com/cosmos/interchain-security/v7/app/consumer"
	"github.com/cosmos/interchain-security/v7/tests/integration"
	icstestingutils "github.com/cosmos/interchain-security/v7/testutil/ibc_testing"
	"github.com/cosmos/interchain-security/v7/x/ccv/types"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	crysistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/gaia/v23/ante"
	gaiaApp "github.com/cosmos/gaia/v23/app"
)

var ccvSuite *integration.CCVTestSuite

func init() {
	// Pass in concrete app types that implement the interfaces defined in https://github.com/cosmos/interchain-security/testutil/integration/interfaces.go
	// IMPORTANT: the concrete app types passed in as type parameters here must match the
	// concrete app types returned by the relevant app initers.
	ccvSuite = integration.NewCCVTestSuite[*gaiaApp.GaiaApp, *appConsumer.App](
		// Pass in ibctesting.AppIniters for gaia (provider) and consumer.
		GaiaAppIniterTempDir, icstestingutils.ConsumerAppIniter, []string{})

	ante.UseFeeMarketDecorator = false
}

func TestCCVTestSuite(t *testing.T) {
	ante.UseFeeMarketDecorator = false
	// Run tests
	suite.Run(t, ccvSuite)
}

// TestVerifyInvariant checks the integration of the crisis module by
// sending a message to verify an invariant
func TestVerifyInvariant(t *testing.T) {
	ccvSuite.SetT(t)
	ccvSuite.SetupTest()

	delAddr := ccvSuite.GetProviderChain().SenderAccount.GetAddress()
	msg := crysistypes.MsgVerifyInvariant{
		Sender:              delAddr.String(),
		InvariantModuleName: "gov",
		InvariantRoute:      "module-account",
	}

	var err error
	var resp *crysistypes.MsgVerifyInvariantResponse
	provCtx := ccvSuite.GetProviderChain().GetContext()
	require.NotPanics(t, func() {
		resp, err = app.CrisisKeeper.VerifyInvariant(provCtx, &msg)
	})
	require.NoError(t, err, "verify invariant returned an error")
	require.NotNil(t, resp, "verify invariant returned an invalid response")
}

func TestICSEpochs(t *testing.T) {
	// a bit hacky but cannot be called
	//  in SetupTest() since it requires `t`
	ccvSuite.SetT(t)
	ccvSuite.SetupTest()

	providerKeeper := app.GetProviderKeeper()
	stakingKeeper := app.StakingKeeper
	provCtx := ccvSuite.GetProviderChain().GetContext()

	delegateFn := func(ctx sdk.Context) {
		delAddr := ccvSuite.GetProviderChain().SenderAccount.GetAddress()
		consAddr := sdk.ConsAddress(ccvSuite.GetProviderChain().Vals.Validators[0].Address)
		validator, err := stakingKeeper.ValidatorByConsAddr(ctx, consAddr)
		require.NoError(t, err)
		_, err = stakingKeeper.Delegate(
			ctx,
			delAddr,
			math.NewInt(1000000),
			stakingtypes.Unbonded,
			validator.(stakingtypes.Validator),
			true,
		)
		require.NoError(t, err)
	}

	getVSCPacketsFn := func() []types.ValidatorSetChangePacketData {
		consumerID := icstestingutils.FirstConsumerID
		return providerKeeper.GetPendingVSCPackets(provCtx, consumerID)
	}

	nextEpoch := func(ctx sdk.Context) sdk.Context {
		for {
			if providerKeeper.BlocksUntilNextEpoch(ctx) == 0 {
				return ctx
			}
			ccvSuite.GetProviderChain().NextBlock()
			ctx = ccvSuite.GetProviderChain().GetContext()
		}
	}

	// Bond some tokens on provider to change validator powers
	delegateFn(provCtx)

	// VSCPacket should only be created at the end of the current epoch
	require.Empty(t, getVSCPacketsFn())
	provCtx = nextEpoch(provCtx)
	// Expect to create a VSC packet
	// without sending it since CCV channel isn't established
	_, err := app.EndBlocker(provCtx)
	require.NoError(t, err)
	require.NotEmpty(t, getVSCPacketsFn())

	// Expect the VSC packet to send after setting up the CCV channel
	ccvSuite.SetupCCVChannel(ccvSuite.GetCCVPath())
	require.Empty(t, getVSCPacketsFn())
	// Expect VSC Packet to be committed
	require.Len(t, ccvSuite.GetProviderChain().App.GetIBCKeeper().ChannelKeeper.GetAllPacketCommitmentsAtChannel(
		provCtx,
		ccvSuite.GetCCVPath().EndpointB.ChannelConfig.PortID,
		ccvSuite.GetCCVPath().EndpointB.ChannelID,
	), 1)

	// Bond some tokens on provider to change validator powers
	delegateFn(provCtx)
	// Second VSCPacket should only be created at the end of the current epoch
	require.Empty(t, getVSCPacketsFn())

	provCtx = nextEpoch(provCtx)
	_, err = app.EndBlocker(provCtx)
	require.NoError(t, err)
	// Expect second VSC Packet to be committed
	require.Len(t, ccvSuite.GetProviderChain().App.GetIBCKeeper().ChannelKeeper.GetAllPacketCommitmentsAtChannel(
		provCtx,
		ccvSuite.GetCCVPath().EndpointB.ChannelConfig.PortID,
		ccvSuite.GetCCVPath().EndpointB.ChannelID,
	), 2)
}
