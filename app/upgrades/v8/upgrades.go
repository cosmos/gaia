package v8

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/gaia/v8/app/keepers"
	icacontrollertypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/host/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {

		// // retrieve metadata
		// actualMetadata := make([]banktypes.Metadata, 0)
		// keepers.BankKeeper.IterateAllDenomMetaData(ctx, func(metadata banktypes.Metadata) bool {
		// 	actualMetadata = append(actualMetadata, metadata)
		// 	return false
		// })
		// fmt.Println("actualMetadata", actualMetadata)

		store := ctx.KVStore(sdk.NewKVStoreKey(banktypes.StoreKey))
		denomMetaDataStore := prefix.NewStore(store, banktypes.DenomMetadataPrefix)

		iterator := denomMetaDataStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			fmt.Printf("iterator.Key() is '%s'\n", string(iterator.Key()))
			fmt.Printf("iterator.Value() is '%s'\n", string(iterator.Value()))
		}

		keepers.BankKeeper.IterateAllDenomMetaData(ctx, func(metadata banktypes.Metadata) bool {
			fmt.Printf("base is: '%s'\n", metadata.Base)

			actualMetadata, found := keepers.BankKeeper.GetDenomMetaData(ctx, metadata.Base)
			if !found {
				fmt.Println("wasn't able to retrieve with the same string that was just retrieved!!!")
			} else {
				fmt.Println("SUCCESS: actualMetadata", actualMetadata)
			}
			return false
		})

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

		ctx.Logger().Info("start to run module migrations...")

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
