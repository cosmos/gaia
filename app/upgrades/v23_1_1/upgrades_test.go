package v23_1_1_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	"github.com/cosmos/gaia/v23/app/helpers"
	upgrade "github.com/cosmos/gaia/v23/app/upgrades/v23_1_1"
)

func TestGrantIBCWasmAuth(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{
		Time: time.Unix(1740829624, 0),
	})

	err := upgrade.AuthzGrantWasmMigrate(ctx, gaiaApp.AuthzKeeper, *gaiaApp.GovKeeper)
	require.NoError(t, err)

	granteeAddr, err := sdk.AccAddressFromBech32(upgrade.GranteeAddress)
	require.NoError(t, err)
	granterAddr, err := sdk.AccAddressFromBech32(gaiaApp.GovKeeper.GetAuthority())
	require.NoError(t, err)

	auth, _ := gaiaApp.AuthzKeeper.GetAuthorization(
		ctx, granteeAddr,
		granterAddr,
		upgrade.IBCWasmMigrateTypeURL)
	require.NotNil(t, auth)

	resp, err := gaiaApp.AuthzKeeper.GranteeGrants(ctx, &authz.QueryGranteeGrantsRequest{
		Grantee: upgrade.GranteeAddress,
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(resp.Grants))
	require.Equal(t, granterAddr.String(), resp.Grants[0].Granter)
	require.Equal(t, granteeAddr.String(), resp.Grants[0].Grantee)
	require.Equal(t, upgrade.GrantExpiration, *resp.Grants[0].Expiration)
}
