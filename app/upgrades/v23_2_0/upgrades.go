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
		if err = MigrateIBCWasm(ctx, keepers.WasmClientKeeper); err != nil {
			ctx.Logger().Info("Error running migrate for IBC Wasm client", "message", err.Error())
		}

		ctx.Logger().Info("Upgrade v23.2.0 complete")
		return vm, nil
	}
}

func MigrateIBCWasm(ctx sdk.Context, wasmClientKeeper ibcwasmkeeper.Keeper) error {
	hexChecksum := "b92e9904aab2292916507f0db04b7ab6d024c2fdb57a9d52e6725f69b2e684c1"
	checksumBz, err := hex.DecodeString(hexChecksum)
	if err != nil {
		return err
	}

	migrateMsgBase64 := "eyJtaWdyYXRpb24iOnsidXBkYXRlX2ZvcmtfcGFyYW1ldGVycyI6eyJnZW5lc2lzX2ZvcmtfdmVyc2lvbiI6IjB4MDAwMDAwMDAiLCJnZW5lc2lzX3Nsb3QiOjAsImFsdGFpciI6eyJ2ZXJzaW9uIjoiMHgwMTAwMDAwMCIsImVwb2NoIjo3NDI0MH0sImJlbGxhdHJpeCI6eyJ2ZXJzaW9uIjoiMHgwMjAwMDAwMCIsImVwb2NoIjoxNDQ4OTZ9LCJjYXBlbGxhIjp7InZlcnNpb24iOiIweDAzMDAwMDAwIiwiZXBvY2giOjE5NDA0OH0sImRlbmViIjp7InZlcnNpb24iOiIweDA0MDAwMDAwIiwiZXBvY2giOjI2OTU2OH0sImVsZWN0cmEiOnsidmVyc2lvbiI6IjB4MDUwMDAwMDAiLCJlcG9jaCI6MzY0MDMyfX19fQ"
	migrateMsgBz, err := base64.StdEncoding.DecodeString(migrateMsgBase64)
	if err != nil {
		return err
	}

	_, err = wasmClientKeeper.MigrateContract(ctx, &ibcwasmtypes.MsgMigrateContract{
		Signer:   "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
		ClientId: "08-wasm-1369",
		Checksum: checksumBz,
		Msg:      migrateMsgBz,
	})
	return err
}
