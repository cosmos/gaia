package v23_1_1 //nolint:revive

import (
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func AuthzGrantWasmMigrate(ctx sdk.Context, authzKeeper authzkeeper.Keeper, govKeeper govkeeper.Keeper) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	grant, err := authz.NewGrant(
		sdkCtx.BlockTime(),
		authz.NewGenericAuthorization(IBCWasmMigrateTypeURL),
		&GrantExpiration,
	)
	if err != nil {
		return err
	}
	sdkCtx.Logger().Info("Granting IBC Migrate Code", "granter", govKeeper.GetAuthority(), "grantee",
		GranteeAddress)
	resp, err := authzKeeper.Grant(ctx, &authz.MsgGrant{
		Granter: govKeeper.GetAuthority(),
		Grantee: GranteeAddress,
		Grant:   grant,
	})
	if err != nil {
		return err
	}
	if resp != nil {
		sdkCtx.Logger().Info("Authz Keeper Grant", "response", resp.String())
	}
	return nil
}
