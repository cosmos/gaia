package v8

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icahosttypes "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"

	"github.com/cosmos/gaia/v9/app/keepers"
)

func FixBankMetadata(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	ctx.Logger().Info("Starting fix bank metadata...")

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

	ctx.Logger().Info("Fix bank metadata complete")

	return nil
}

func QuicksilverFix(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	ctx.Logger().Info("Starting fix quicksilver...")

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
	if sourceBalance.IsGTE(sdk.NewCoin("uatom", sdk.NewInt(1))) {
		refundBalance := sourceBalance.SubAmount(sdk.NewInt(1))
		err = keepers.BankKeeper.SendCoins(ctx, sourceAddress, destinationAddress, sdk.NewCoins(refundBalance))
		if err != nil {
			return errors.New("unable to refund coins")
		}
	}

	// Close channels
	closeChannel(keepers, ctx, "channel-462")
	closeChannel(keepers, ctx, "channel-463")
	closeChannel(keepers, ctx, "channel-464")
	closeChannel(keepers, ctx, "channel-465")
	closeChannel(keepers, ctx, "channel-466")

	ctx.Logger().Info("Fix quicksilver complete")

	return nil
}

func closeChannel(keepers *keepers.AppKeepers, ctx sdk.Context, channelID string) {
	channel, found := keepers.IBCKeeper.ChannelKeeper.GetChannel(ctx, icatypes.PortID, channelID)
	if found {
		channel.State = ibcchanneltypes.CLOSED
		keepers.IBCKeeper.ChannelKeeper.SetChannel(ctx, icatypes.PortID, channelID, channel)
	}
}

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Running upgrade fixes...")

		err := FixBankMetadata(ctx, keepers)
		if err != nil {
			ctx.Logger().Info(fmt.Sprintf("error fixing bank metadata: %s", err.Error()))
		}

		err = QuicksilverFix(ctx, keepers)
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

		ctx.Logger().Info("Starting module migrations...")

		vm, err = mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Upgrade complete")
		return vm, err
	}
}
