package v23_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/gaia/v24/app/helpers"
	v23 "github.com/cosmos/gaia/v24/app/upgrades/v23"
)

func TestGrantIBCWasmAuth(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{
		Time: time.Unix(1740829624, 0),
	})

	err := v23.AuthzGrantWasmLightClient(ctx, gaiaApp.AuthzKeeper, *gaiaApp.GovKeeper)
	require.NoError(t, err)

	granteeAddr, err := sdk.AccAddressFromBech32(v23.ClientUploaderAddress)
	require.NoError(t, err)
	granterAddr, err := sdk.AccAddressFromBech32(gaiaApp.GovKeeper.GetAuthority())
	require.NoError(t, err)

	auth, _ := gaiaApp.AuthzKeeper.GetAuthorization(
		ctx, granteeAddr,
		granterAddr,
		v23.IBCWasmStoreCodeTypeURL)
	require.NotNil(t, auth)
}
