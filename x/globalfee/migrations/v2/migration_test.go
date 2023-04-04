package v2

import (
	gaia "github.com/cosmos/gaia/v9/app"
	"github.com/stretchr/testify/require"

	"testing"
)

func TestMigrateStore(t *testing.T) {
	cdc := gaia.MakeTestEncodingConfig().Codec

	// Run migrations.
	err := MigrateStore(ctx, globalfeeSubspace)
	require.NoError(t, err)

	// Check params
	g
	require.NoError(t, cdc.Unmarshal(bz, &params))
	require.NotNil(t, params)
	require.Equal(t, v1.DefaultParams().ExpeditedMinDeposit, params.ExpeditedMinDeposit)
	require.Equal(t, v1.DefaultParams().ExpeditedThreshold, params.ExpeditedThreshold)
	require.Equal(t, v1.DefaultParams().ExpeditedVotingPeriod, params.ExpeditedVotingPeriod)
}
