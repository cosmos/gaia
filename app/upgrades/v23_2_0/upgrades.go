package v23_2_0 //nolint:revive

import (
	"context"
	"encoding/base64"
	"encoding/hex"

	ibcwasmkeeper "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10/keeper"
	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10/types"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/cosmos/gaia/v23/app/keepers"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		ctx.Logger().Info("Starting module migrations...")

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Starting Migrate IBC Wasm...")
		if err = MigrateIBCWasm(ctx, keepers.WasmClientKeeper, HexChecksum, MigrateMsgBase64, ClientId,
			SignerAccount); err != nil {
			ctx.Logger().Info("Error running migrate for IBC Wasm client", "message", err.Error())
		}

		ctx.Logger().Info("Upgrade v23.2.0 complete")
		return vm, nil
	}
}

func MigrateIBCWasm(ctx sdk.Context, wasmClientKeeper ibcwasmkeeper.Keeper, hexChecksum string,
	migrateMsgB64 string, clientId string, signerAcc string) error {
	checksumBz, err := hex.DecodeString(hexChecksum)
	if err != nil {
		return err
	}

	migrateMsgBz, err := base64.RawStdEncoding.DecodeString(migrateMsgB64)
	if err != nil {
		return err
	}

	_, err = wasmClientKeeper.MigrateContract(ctx, &ibcwasmtypes.MsgMigrateContract{
		Signer:   signerAcc,
		ClientId: clientId,
		Checksum: checksumBz,
		Msg:      migrateMsgBz,
	})
	return err
}
