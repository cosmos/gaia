package v22_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	testutil "github.com/cosmos/interchain-security/v7/testutil/keeper"
	providertypes "github.com/cosmos/interchain-security/v7/x/ccv/provider/types"

	"cosmossdk.io/math"

	v22 "github.com/cosmos/gaia/v23/app/upgrades/v22"
)

func TestSetDefaultConsumerInfractionParams(t *testing.T) {
	t.Helper()
	inMemParams := testutil.NewInMemKeeperParams(t)
	pk, ctx, ctrl, _ := testutil.GetProviderKeeperAndCtx(t, inMemParams)
	defer ctrl.Finish()

	// Add consumer chains
	initConsumerID := pk.FetchAndIncrementConsumerId(ctx)
	pk.SetConsumerChainId(ctx, initConsumerID, "init-1")
	pk.SetConsumerPhase(ctx, initConsumerID, providertypes.CONSUMER_PHASE_INITIALIZED)
	launchedConsumerID := pk.FetchAndIncrementConsumerId(ctx)
	pk.SetConsumerChainId(ctx, launchedConsumerID, "launched-1")
	pk.SetConsumerPhase(ctx, launchedConsumerID, providertypes.CONSUMER_PHASE_LAUNCHED)
	stoppedConsumerID := pk.FetchAndIncrementConsumerId(ctx)
	pk.SetConsumerChainId(ctx, stoppedConsumerID, "stopped-1")
	pk.SetConsumerPhase(ctx, stoppedConsumerID, providertypes.CONSUMER_PHASE_STOPPED)

	consumerIDs := pk.GetAllConsumerIds(ctx)
	require.Equal(t, 3, len(consumerIDs))

	for _, consumerID := range consumerIDs {
		_, err := pk.GetInfractionParameters(ctx, consumerID)
		require.Error(t, err)
	}

	testParams := testInfractionParams()
	err := v22.SetConsumerInfractionParams(ctx, pk, testParams)
	require.NoError(t, err)

	for _, consumerID := range consumerIDs {
		infractionParams, err := pk.GetInfractionParameters(ctx, consumerID)
		require.NoError(t, err)
		require.Equal(t, testParams, infractionParams)
	}
}

func testInfractionParams() providertypes.InfractionParameters {
	return providertypes.InfractionParameters{
		DoubleSign: &providertypes.SlashJailParameters{
			JailDuration:  time.Duration(1<<63 - 1),        // the largest value a time.Duration can hold 9223372036854775807 (approximately 292 years)
			SlashFraction: math.LegacyNewDecWithPrec(5, 2), // 0.05
		},
		Downtime: &providertypes.SlashJailParameters{
			JailDuration:  600 * time.Second,
			SlashFraction: math.LegacyNewDec(0), // no slashing for downtime on the consumer
		},
	}
}
