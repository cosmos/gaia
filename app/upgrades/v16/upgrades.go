package v16

import (
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	ratelimitkeeper "github.com/Stride-Labs/ibc-rate-limiting/ratelimit/keeper"
	ratelimittypes "github.com/Stride-Labs/ibc-rate-limiting/ratelimit/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/cosmos/gaia/v16/app/keepers"
)

var (
	RateLimits = map[string]ratelimittypes.MsgAddRateLimit{
		"Osmosis": {
			MaxPercentSend: sdkmath.NewInt(5),
			MaxPercentRecv: sdkmath.NewInt(5),
			ChannelId:      "channel-141",
		},
		"Neutron": {
			MaxPercentSend: sdkmath.NewInt(1),
			MaxPercentRecv: sdkmath.NewInt(1),
			ChannelId:      "channel-569",
		},
		"Stride": {
			MaxPercentSend: sdkmath.NewInt(1),
			MaxPercentRecv: sdkmath.NewInt(1),
			ChannelId:      "channel-391",
		},
		"Kujira": {
			MaxPercentSend: sdkmath.NewInt(1),
			MaxPercentRecv: sdkmath.NewInt(1),
			ChannelId:      "channel-343",
		},
		"Injective": {
			MaxPercentSend: sdkmath.NewInt(1),
			MaxPercentRecv: sdkmath.NewInt(1),
			ChannelId:      "channel-220",
		},
		"Persistence": {
			MaxPercentSend: sdkmath.NewInt(1),
			MaxPercentRecv: sdkmath.NewInt(1),
			ChannelId:      "channel-190",
		},
		"Secret": {
			MaxPercentSend: sdkmath.NewInt(1),
			MaxPercentRecv: sdkmath.NewInt(1),
			ChannelId:      "channel-235",
		},
	}
)

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

		// Add initial rate limits
		if err := AddRateLimits(ctx, keepers.RatelimitKeeper); err != nil {
			return vm, err
		}

		ctx.Logger().Info("Upgrade complete")
		return vm, err
	}
}

// Add rate limits as per https://www.mintscan.io/cosmos/proposals/890
func AddRateLimits(ctx sdk.Context, k ratelimitkeeper.Keeper) error {
	ctx.Logger().Info("Adding rate limits...")

	// Osmosis
	msg := RateLimits["Osmosis"]
	msg.DurationHours = RateLimitDurationHours
	msg.Denom = RateLimitDenom
	if err := k.AddRateLimit(ctx, &msg); err != nil {
		return errorsmod.Wrapf(err, "unable to add rate limit on %s to Osmosis", msg.ChannelId)
	}

	// Neutron
	msg = RateLimits["Neutron"]
	msg.DurationHours = RateLimitDurationHours
	msg.Denom = RateLimitDenom
	if err := k.AddRateLimit(ctx, &msg); err != nil {
		return errorsmod.Wrapf(err, "unable to add rate limit on %s to Neutron", msg.ChannelId)
	}

	// Stride
	msg = RateLimits["Stride"]
	msg.DurationHours = RateLimitDurationHours
	msg.Denom = RateLimitDenom
	if err := k.AddRateLimit(ctx, &msg); err != nil {
		return errorsmod.Wrapf(err, "unable to add rate limit on %s to Stride", msg.ChannelId)
	}

	// Kujira
	msg = RateLimits["Kujira"]
	msg.DurationHours = RateLimitDurationHours
	msg.Denom = RateLimitDenom
	if err := k.AddRateLimit(ctx, &msg); err != nil {
		return errorsmod.Wrapf(err, "unable to add rate limit on %s to Kujira", msg.ChannelId)
	}

	// Injective
	msg = RateLimits["Injective"]
	msg.DurationHours = RateLimitDurationHours
	msg.Denom = RateLimitDenom
	if err := k.AddRateLimit(ctx, &msg); err != nil {
		return errorsmod.Wrapf(err, "unable to add rate limit on %s to Injective", msg.ChannelId)
	}

	// Persistence
	msg = RateLimits["Persistence"]
	msg.DurationHours = RateLimitDurationHours
	msg.Denom = RateLimitDenom
	if err := k.AddRateLimit(ctx, &msg); err != nil {
		return errorsmod.Wrapf(err, "unable to add rate limit on %s to Persistence", msg.ChannelId)
	}

	// Secret
	msg = RateLimits["Secret"]
	msg.DurationHours = RateLimitDurationHours
	msg.Denom = RateLimitDenom
	if err := k.AddRateLimit(ctx, &msg); err != nil {
		return errorsmod.Wrapf(err, "unable to add rate limit on %s to Secret", msg.ChannelId)
	}

	ctx.Logger().Info("Finished adding rate limits")
	return nil
}
