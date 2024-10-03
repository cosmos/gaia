package v21

import (
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/gaia/v21/app/helpers"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHasExpectedChainIdSanityCheck(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})

	pk := gaiaApp.ProviderKeeper

	// no such consumer chain
	consumerId := "0"
	require.False(t, HasExpectedChainIdSanityCheck(ctx, pk, consumerId, "chain-1"))

	// consumer chain does not have `chain-1` id
	pk.SetConsumerChainId(ctx, consumerId, "chain-2")
	require.False(t, HasExpectedChainIdSanityCheck(ctx, pk, consumerId, "chain-1"))

	pk.SetConsumerChainId(ctx, consumerId, "chain-1")
	require.True(t, HasExpectedChainIdSanityCheck(ctx, pk, consumerId, "chain-1"))
}
