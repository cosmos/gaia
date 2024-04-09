package v16

import (
	"errors"

	ratelimitkeeper "github.com/Stride-Labs/ibc-rate-limiting/ratelimit/keeper"
	ratelimittypes "github.com/Stride-Labs/ibc-rate-limiting/ratelimit/types"

	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	providertypes "github.com/cosmos/interchain-security/v4/x/ccv/provider/types"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/cosmos/gaia/v16/app/keepers"
)

var RateLimits = map[string]ratelimittypes.MsgAddRateLimit{
	"osmosis-1": {
		MaxPercentSend: sdkmath.NewInt(5),
		MaxPercentRecv: sdkmath.NewInt(5),
		ChannelId:      "channel-141",
	},
	"neutron-1": {
		MaxPercentSend: sdkmath.NewInt(1),
		MaxPercentRecv: sdkmath.NewInt(1),
		ChannelId:      "channel-569",
	},
	"stride-1": {
		MaxPercentSend: sdkmath.NewInt(1),
		MaxPercentRecv: sdkmath.NewInt(1),
		ChannelId:      "channel-391",
	},
	"kaiyo-1": { // Kujira
		MaxPercentSend: sdkmath.NewInt(1),
		MaxPercentRecv: sdkmath.NewInt(1),
		ChannelId:      "channel-343",
	},
	"injective-1": {
		MaxPercentSend: sdkmath.NewInt(1),
		MaxPercentRecv: sdkmath.NewInt(1),
		ChannelId:      "channel-220",
	},
	"core-1": { // Persistence
		MaxPercentSend: sdkmath.NewInt(1),
		MaxPercentRecv: sdkmath.NewInt(1),
		ChannelId:      "channel-190",
	},
	"secret-4": {
		MaxPercentSend: sdkmath.NewInt(1),
		MaxPercentRecv: sdkmath.NewInt(1),
		ChannelId:      "channel-235",
	},
}

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting module migrations...")

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		// Enable ICA controller
		keepers.ICAControllerKeeper.SetParams(ctx, icacontrollertypes.DefaultParams())

		// Set default blocks per epoch
		providerParams := keepers.ProviderKeeper.GetParams(ctx)
		providerParams.BlocksPerEpoch = providertypes.DefaultBlocksPerEpoch
		keepers.ProviderKeeper.SetParams(ctx, providerParams)

		// Add initial rate limits
		// This operation is permitted to fail and will not halt the upgrade
		// In case of failure, rate limits must be added manually
		addErr := AddRateLimits(ctx, keepers.RatelimitKeeper)
		if addErr != nil && errors.Is(ratelimittypes.ErrChannelNotFound, addErr) {
			ctx.Logger().Error("Unable to add rate limits - all rate limits must be added manually after upgrade")
		} else if addErr != nil {
			return vm, errorsmod.Wrapf(addErr, "unable to add rate limits")
		}

		// Set CosmWasm params
		wasmParams := wasmtypes.DefaultParams()
		wasmParams.CodeUploadAccess = wasmtypes.AllowNobody
		// TODO(reece): only allow specific addresses to instantiate contracts or anyone with AccessTypeEverybody?
		wasmParams.InstantiateDefaultPermission = wasmtypes.AccessTypeAnyOfAddresses
		if err := keepers.WasmKeeper.SetParams(ctx, wasmParams); err != nil {
			return vm, errorsmod.Wrapf(err, "unable to set CosmWasm params")
		}

		ctx.Logger().Info("Upgrade complete")
		return vm, err
	}
}

// Add rate limits as per https://www.mintscan.io/cosmos/proposals/890
func AddRateLimits(ctx sdk.Context, k ratelimitkeeper.Keeper) error {
	ctx.Logger().Info("Adding rate limits...")

	// Osmosis
	msg := RateLimits["osmosis-1"]
	msg.DurationHours = RateLimitDurationHours
	msg.Denom = RateLimitDenom
	if err := k.AddRateLimit(ctx, &msg); err != nil {
		return errorsmod.Wrapf(err, "unable to add rate limit on %s to osmosis-1", msg.ChannelId)
	}

	// Neutron
	msg = RateLimits["neutron-1"]
	msg.DurationHours = RateLimitDurationHours
	msg.Denom = RateLimitDenom
	if err := k.AddRateLimit(ctx, &msg); err != nil {
		return errorsmod.Wrapf(err, "unable to add rate limit on %s to neutron-1", msg.ChannelId)
	}

	// Stride
	msg = RateLimits["stride-1"]
	msg.DurationHours = RateLimitDurationHours
	msg.Denom = RateLimitDenom
	if err := k.AddRateLimit(ctx, &msg); err != nil {
		return errorsmod.Wrapf(err, "unable to add rate limit on %s to stride-1", msg.ChannelId)
	}

	// Kujira
	msg = RateLimits["kaiyo-1"]
	msg.DurationHours = RateLimitDurationHours
	msg.Denom = RateLimitDenom
	if err := k.AddRateLimit(ctx, &msg); err != nil {
		return errorsmod.Wrapf(err, "unable to add rate limit on %s to kaiyo-1", msg.ChannelId)
	}

	// Injective
	msg = RateLimits["injective-1"]
	msg.DurationHours = RateLimitDurationHours
	msg.Denom = RateLimitDenom
	if err := k.AddRateLimit(ctx, &msg); err != nil {
		return errorsmod.Wrapf(err, "unable to add rate limit on %s to injective-1", msg.ChannelId)
	}

	// Persistence
	msg = RateLimits["core-1"]
	msg.DurationHours = RateLimitDurationHours
	msg.Denom = RateLimitDenom
	if err := k.AddRateLimit(ctx, &msg); err != nil {
		return errorsmod.Wrapf(err, "unable to add rate limit on %s to core-1", msg.ChannelId)
	}

	// Secret
	msg = RateLimits["secret-4"]
	msg.DurationHours = RateLimitDurationHours
	msg.Denom = RateLimitDenom
	if err := k.AddRateLimit(ctx, &msg); err != nil {
		return errorsmod.Wrapf(err, "unable to add rate limit on %s to secret-4", msg.ChannelId)
	}

	ctx.Logger().Info("Finished adding rate limits")
	return nil
}
