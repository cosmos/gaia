package v23

import (
	"context"
	"encoding/hex"
	"fmt"

	ibcwasmkeeper "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10/keeper"
	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10/types"
	ibctmtypes "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"

	errorsmod "cosmossdk.io/errors"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/cosmos/gaia/v23/app/keepers"
)

// CreateUpgradeHandler returns an upgrade handler for Gaia v23.
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
			return vm, errorsmod.Wrapf(err, "running module migrations")
		}

		// Set IBC Client AllowedClients
		ctx.Logger().Info("Setting IBC Client AllowedClients")
		params := keepers.IBCKeeper.ClientKeeper.GetParams(ctx)
		params.AllowedClients = []string{ibctmtypes.ModuleName, ibcwasmtypes.ModuleName}
		keepers.IBCKeeper.ClientKeeper.SetParams(ctx, params)

		// Add Eth Light Wasm Light Client
		ctx.Logger().Info("Adding Eth Light Wasm Light Client")
		if err := AddEthLightWasmLightClient(ctx, keepers.WasmClientKeeper); err != nil {
			ctx.Logger().Error("Error adding Eth Light Wasm Light Client", "message", err.Error())
			return nil, err
		}

		ctx.Logger().Info("Upgrade v23 complete")
		return vm, nil
	}
}

func AddEthLightWasmLightClient(ctx context.Context, wasmKeeper ibcwasmkeeper.Keeper) error {
	resp, err := wasmKeeper.StoreCode(ctx, &ibcwasmtypes.MsgStoreCode{
		Signer:       wasmKeeper.GetAuthority(),
		WasmByteCode: ethWasmLightClient,
	})
	if err != nil {
		return errorsmod.Wrap(err, "failed to store eth wasm light client during upgrade")
	}

	actualChecksum := hex.EncodeToString(resp.Checksum)

	if hex.EncodeToString(resp.Checksum) != ExpectedEthLightClientChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", ExpectedEthLightClientChecksum, actualChecksum)
	}

	return nil
}
