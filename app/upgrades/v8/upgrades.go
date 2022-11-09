package v8

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/gaia/v8/app/keepers"
	icacontrollertypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/host/types"
)

func fixBank(ctx sdk.Context, keepers *keepers.AppKeepers) {
	key := keepers.GetKey("bank")
	store := ctx.KVStore(key)

	oldDenomMetaDataStore := prefix.NewStore(store, banktypes.DenomMetadataPrefix)

	oldDenomMetaDataIter := oldDenomMetaDataStore.Iterator(nil, nil)
	defer oldDenomMetaDataIter.Close()

	for ; oldDenomMetaDataIter.Valid(); oldDenomMetaDataIter.Next() {
		oldKey := oldDenomMetaDataIter.Key()
		l := len(oldKey) - 1

		newKey := make([]byte, l)
		copy(newKey, oldKey[:l])
		oldDenomMetaDataStore.Set(newKey, oldDenomMetaDataIter.Value())
		oldDenomMetaDataStore.Delete(oldKey)
	}
}

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {

		ctx.Logger().Info("start to run module migrations...")
		mvm, err := mm.RunMigrations(ctx, configurator, vm)

		if err != nil {
			return mvm, err
		}

		// fixes issue https://github.com/cosmos/cosmos-sdk/issues/13797
		// should be removed if upgrading to a patched version of the SDK
		fixBank(ctx, keepers)

		// Add atom name and symbol into the bank keeper
		atomMetaData, found := keepers.BankKeeper.GetDenomMetaData(ctx, "uatom")
		if !found {
			return nil, errors.New("atom denom not found")
		}
		atomMetaData.Name = "Cosmos Hub Atom"
		atomMetaData.Symbol = "ATOM"
		keepers.BankKeeper.SetDenomMetaData(ctx, atomMetaData)

		// Enable controller chain
		controllerParams := icacontrollertypes.Params{
			ControllerEnabled: true,
		}

		// Change hostParams allow_messages = [*] instead of whitelisting individual messages
		hostParams := icahosttypes.Params{
			HostEnabled:   true,
			AllowMessages: []string{"*"},
		}

		// Update params for host & controller keepers
		keepers.ICAHostKeeper.SetParams(ctx, hostParams)
		keepers.ICAControllerKeeper.SetParams(ctx, controllerParams)

		return mvm, err
	}
}
