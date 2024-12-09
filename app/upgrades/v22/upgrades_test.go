package v22_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testutil "github.com/cosmos/interchain-security/v6/testutil/keeper"
	providertypes "github.com/cosmos/interchain-security/v6/x/ccv/provider/types"

	v22 "github.com/cosmos/gaia/v22/app/upgrades/v22"
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

	activeConsumerIDs := pk.GetAllActiveConsumerIds(ctx)
	require.Equal(t, 2, len(activeConsumerIDs))

	for _, consumerID := range activeConsumerIDs {
		_, err := pk.GetInfractionParameters(ctx, consumerID)
		require.Error(t, err)
	}

	err := v22.SetConsumerInfractionParams(ctx, pk)
	require.NoError(t, err)

	defaultInfractionParams := v22.DefaultInfractionParams()
	for _, consumerID := range activeConsumerIDs {
		infractionParams, err := pk.GetInfractionParameters(ctx, consumerID)
		require.NoError(t, err)
		require.Equal(t, defaultInfractionParams, infractionParams)
	}
}
