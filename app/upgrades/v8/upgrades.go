package v8

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icahosttypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host/types"

	"github.com/cosmos/gaia/v8/app/keepers"
)

func FixBankMetadata(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	malformedDenom := "uatomu"
	correctDenom := "uatom"

	atomMetaData, foundMalformed := keepers.BankKeeper.GetDenomMetaData(ctx, malformedDenom)
	if foundMalformed {
		// save it with the correct denom
		keepers.BankKeeper.SetDenomMetaData(ctx, atomMetaData)

		// delete the old format
		key := keepers.GetKey(banktypes.ModuleName)
		store := ctx.KVStore(key)
		oldDenomMetaDataStore := prefix.NewStore(store, banktypes.DenomMetadataPrefix)
		oldDenomMetaDataStore.Delete([]byte(malformedDenom))

		// confirm whether the old key is still accessible
		_, foundMalformed = keepers.BankKeeper.GetDenomMetaData(ctx, malformedDenom)
		if foundMalformed {
			return errors.New("malformed 'uatomu' denom not fixed")
		}
	}

	// proceed with the original intention of populating the missing Name and Symbol fields
	atomMetaData, foundCorrect := keepers.BankKeeper.GetDenomMetaData(ctx, correctDenom)
	if !foundCorrect {
		return errors.New("atom denom not found")
	}

	atomMetaData.Name = "Cosmos Hub Atom"
	atomMetaData.Symbol = "ATOM"
	keepers.BankKeeper.SetDenomMetaData(ctx, atomMetaData)

	return nil
}

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("start to run module migrations...")

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("running the rest of the upgrade handler...")

		err = FixBankMetadata(ctx, keepers)
		if err != nil {
			return vm, err
		}

		// Change hostParams allow_messages = [*] instead of whitelisting individual messages
		hostParams := icahosttypes.Params{
			HostEnabled:   true,
			AllowMessages: []string{"*"},
		}

		// Update params for host & controller keepers
		keepers.ICAHostKeeper.SetParams(ctx, hostParams)

		ctx.Logger().Info("upgrade complete")

		return vm, err
	}
}
