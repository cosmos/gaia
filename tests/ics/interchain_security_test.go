package ics

import (
	"encoding/json"
	"testing"

	tmdb "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	appConsumer "github.com/cosmos/interchain-security/v4/app/consumer"
	"github.com/cosmos/interchain-security/v4/tests/integration"
	icstestingutils "github.com/cosmos/interchain-security/v4/testutil/ibc_testing"
	"github.com/cosmos/interchain-security/v4/x/ccv/types"

	gaiaApp "github.com/cosmos/gaia/v16/app"
)

var (
	app      *gaiaApp.GaiaApp
	ccvSuite *integration.CCVTestSuite
)

func init() {
	// Pass in concrete app types that implement the interfaces defined in https://github.com/cosmos/interchain-security/testutil/integration/interfaces.go
	// IMPORTANT: the concrete app types passed in as type parameters here must match the
	// concrete app types returned by the relevant app initers.
	ccvSuite = integration.NewCCVTestSuite[*gaiaApp.GaiaApp, *appConsumer.App](
		// Pass in ibctesting.AppIniters for gaia (provider) and consumer.
		GaiaAppIniter, icstestingutils.ConsumerAppIniter, []string{})
}

func TestCCVTestSuite(t *testing.T) {
	// Run tests
	suite.Run(t, ccvSuite)
}

// GaiaAppIniter implements ibctesting.AppIniter for the gaia app
func GaiaAppIniter() (ibctesting.TestingApp, map[string]json.RawMessage) {
	encoding := gaiaApp.RegisterEncodingConfig()
	app = gaiaApp.NewGaiaApp(
		log.NewNopLogger(),
		tmdb.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		gaiaApp.DefaultNodeHome,
		encoding,
		gaiaApp.EmptyAppOptions{})

	testApp := ibctesting.TestingApp(app)

	return testApp, gaiaApp.NewDefaultGenesisState(encoding)
}

func TestICSEpochs(t *testing.T) {
	// a bit hacky but cannot be called
	//  in SetupTest() since it requires `t`
	ccvSuite.SetT(t)
	ccvSuite.SetupTest()

	providerKeeper := app.GetProviderKeeper()
	stakingKeeper := app.StakingKeeper
	ctx := ccvSuite.GetProviderChain().GetContext()

	delegateFn := func() {
		delAddr := ccvSuite.GetProviderChain().SenderAccount.GetAddress()
		consAddr := sdk.ConsAddress(ccvSuite.GetProviderChain().Vals.Validators[0].Address)
		validator := stakingKeeper.ValidatorByConsAddr(ctx, consAddr)
		_, err := stakingKeeper.Delegate(
			ctx,
			delAddr,
			sdk.NewInt(1000000),
			stakingtypes.Unbonded,
			validator.(stakingtypes.Validator),
			true,
		)
		require.NoError(t, err)
	}

	getVSCPacketsFn := func() []types.ValidatorSetChangePacketData {
		return providerKeeper.GetPendingVSCPackets(ctx, ccvSuite.GetCCVPath().EndpointA.Chain.ChainID)
	}

	nextEpochs := func(ctx sdk.Context) sdk.Context {
		blockPerEpochs := providerKeeper.GetBlocksPerEpoch(ctx)
		for {
			if ctx.BlockHeight()%blockPerEpochs == 0 {
				return ctx
			}
			ccvSuite.GetProviderChain().NextBlock()
			ctx = ccvSuite.GetProviderChain().GetContext()
		}
	}

	// Bond some tokens on provider to change validator powers
	delegateFn()
	// VSCPacket should only be created at the end of the current epoch
	require.Empty(t, getVSCPacketsFn())
	ctx = nextEpochs(ctx)
	// Expect to create a VSC packet
	// without sending it since CCV channel isn't established
	app.EndBlocker(ctx, abci.RequestEndBlock{})
	require.NotEmpty(t, getVSCPacketsFn())

	ccvSuite.SetupCCVChannel(ccvSuite.GetCCVPath())
	require.NotEmpty(t, getVSCPacketsFn())

	ctx = nextEpochs(ctx)
	require.NotEmpty(t, getVSCPacketsFn())

	app.EndBlocker(ctx, abci.RequestEndBlock{})
	// Expect VSC Packet to be sent
	require.Empty(t, getVSCPacketsFn())
}
