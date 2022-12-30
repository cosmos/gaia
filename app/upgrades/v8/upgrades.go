package v8

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibctmtypes "github.com/cosmos/ibc-go/v3/modules/light-clients/07-tendermint/types"

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

func QuicksilverFix(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	// Refund stuck coins from ica address
	sourceAddress, err := sdk.AccAddressFromBech32("cosmos13dqvh4qtg4gzczuktgnw8gc2ewnwmhdwnctekxctyr4azz4dcyysecgq7e")

	if err != nil {
		return errors.New("invalid source address")
	}
	destinationAddress, err := sdk.AccAddressFromBech32("cosmos1jc24kwznud9m3mwqmcz3xw33ndjuufnghstaag")
	if err != nil {
		return errors.New("invalid destination address")
	}

	// Get balance from stuck address and subtract 1 uatom sent by bad actor
	sourceBalance := keepers.BankKeeper.GetBalance(ctx, sourceAddress, "uatom")
	refundBalance := sourceBalance.SubAmount(sdk.NewInt(1))
	keepers.BankKeeper.SendCoins(ctx, sourceAddress, destinationAddress, sdk.NewCoins(refundBalance))

	// Get connection to quicksilver chain
	connectionId, err := getConnectionIdForChainId(ctx, keepers, "quicksilver-1")
	if err != nil {
		return err
	}

	// Close channels
	closeChannel(keepers, ctx, connectionId, "icacontroller-cosmoshub-4.deposit")
	closeChannel(keepers, ctx, connectionId, "icacontroller-cosmoshub-4.withdrawal")
	closeChannel(keepers, ctx, connectionId, "icacontroller-cosmoshub-4.performance")
	closeChannel(keepers, ctx, connectionId, "icacontroller-cosmoshub-4.delegate")

	return nil
}

func closeChannel(keepers *keepers.AppKeepers, ctx sdk.Context, connectionId string, port string) {
	activeChannelId, found := keepers.ICAHostKeeper.GetActiveChannelID(ctx, connectionId, port)
	if found {
		channel, found := keepers.IBCKeeper.ChannelKeeper.GetChannel(ctx, icatypes.PortID, activeChannelId)
		if found {
			channel.State = ibcchanneltypes.CLOSED
			keepers.IBCKeeper.ChannelKeeper.SetChannel(ctx, icatypes.PortID, activeChannelId, channel)
		}
	}
}

func getConnectionIdForChainId(ctx sdk.Context, keepers *keepers.AppKeepers, chainId string) (string, error) {
	connections := keepers.IBCKeeper.ConnectionKeeper.GetAllConnections(ctx)
	for _, conn := range connections {
		clientState, found := keepers.IBCKeeper.ClientKeeper.GetClientState(ctx, conn.ClientId)
		if !found {
			continue
		}
		client, ok := clientState.(*ibctmtypes.ClientState)
		if ok && client.ChainId == chainId {
			return conn.Id, nil
		}
	}

	return "", errors.New("failed to get connection for the chain id")
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

		err = QuicksilverFix(ctx, keepers)
		if err != nil {
			return vm, err
		}

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

		ctx.Logger().Info("upgrade complete")

		return vm, err
	}
}
