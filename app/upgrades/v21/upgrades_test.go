package v21_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/gaia/v21/app/helpers"
	v21 "github.com/cosmos/gaia/v21/app/upgrades/v21"
)

func TestHasExpectedChainIDSanityCheck(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})

	pk := gaiaApp.ProviderKeeper

	// no such consumer chain
	consumerID := "0"
	require.False(t, v21.HasExpectedChainIDSanityCheck(ctx, pk, consumerID, "chain-1"))

	// consumer chain does not have `chain-1` id
	pk.SetConsumerChainId(ctx, consumerID, "chain-2")
	require.False(t, v21.HasExpectedChainIDSanityCheck(ctx, pk, consumerID, "chain-1"))

	pk.SetConsumerChainId(ctx, consumerID, "chain-1")
	require.True(t, v21.HasExpectedChainIDSanityCheck(ctx, pk, consumerID, "chain-1"))
}

func TestInitializeConstitutionCollection(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})

	govKeeper := gaiaApp.GovKeeper

	pre, err := govKeeper.Constitution.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, "", pre)
	err = v21.InitializeConstitutionCollection(ctx, *govKeeper)
	require.NoError(t, err)
	post, err := govKeeper.Constitution.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, "This chain has no constitution.", post)
}
