package v21_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	govparams "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/cosmos/gaia/v23/app/helpers"
	v21 "github.com/cosmos/gaia/v23/app/upgrades/v21"
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

func TestInitializeGovParams(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})

	govKeeper := gaiaApp.GovKeeper

	// sets the params to "" so we can confirm that the migration
	// function actually changes the parameter to 0.5 from ""
	setupParams, err := govKeeper.Params.Get(ctx)
	require.NoError(t, err)
	setupParams.ProposalCancelRatio = "" // mainnet value
	setupParams.ProposalCancelDest = ""  // mainnet value
	if err := govKeeper.Params.Set(ctx, setupParams); err != nil {
		t.Fatalf("error setting params: %s", err)
	}

	pre, err := govKeeper.Params.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, "", pre.ProposalCancelRatio) // mainnet value
	require.NotEqual(t, govparams.DefaultProposalCancelRatio.String(), pre.ProposalCancelRatio)

	err = v21.InitializeGovParams(ctx, *govKeeper)
	require.NoError(t, err)

	post, err := govKeeper.Params.Get(ctx)
	require.NoError(t, err)

	require.NotEqual(t, pre.ProposalCancelRatio, post.ProposalCancelRatio)
	require.Equal(t, govparams.DefaultProposalCancelRatio.String(), post.ProposalCancelRatio) // confirm change to sdk default

	require.Equal(t, pre.ProposalCancelDest, post.ProposalCancelDest) // does not change (it was already default)
	require.Equal(t, "", post.ProposalCancelDest)
}
