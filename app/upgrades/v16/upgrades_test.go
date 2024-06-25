package v16_test

import (
	"strings"
	"testing"

	ratelimittypes "github.com/Stride-Labs/ibc-rate-limiting/ratelimit/types"
	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/cosmos/gaia/v18/app/helpers"
	v16 "github.com/cosmos/gaia/v18/app/upgrades/v16"
)

var AtomSupply = sdkmath.NewInt(1000)

func TestAddRateLimits(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})
	ratelimitkeeper := gaiaApp.RatelimitKeeper

	// mint atoms
	amount := sdk.NewCoin(v16.RateLimitDenom, AtomSupply)
	amountCoins := sdk.NewCoins(amount)
	err := gaiaApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, amountCoins)
	require.NoError(t, err)

	// mock IBC channels
	gaiaApp.IBCKeeper.ChannelKeeper.SetChannel(ctx, transfertypes.PortID, v16.RateLimits["osmosis-1"].ChannelId, channeltypes.Channel{})
	gaiaApp.IBCKeeper.ChannelKeeper.SetChannel(ctx, transfertypes.PortID, v16.RateLimits["neutron-1"].ChannelId, channeltypes.Channel{})
	gaiaApp.IBCKeeper.ChannelKeeper.SetChannel(ctx, transfertypes.PortID, v16.RateLimits["stride-1"].ChannelId, channeltypes.Channel{})
	gaiaApp.IBCKeeper.ChannelKeeper.SetChannel(ctx, transfertypes.PortID, v16.RateLimits["kaiyo-1"].ChannelId, channeltypes.Channel{})
	gaiaApp.IBCKeeper.ChannelKeeper.SetChannel(ctx, transfertypes.PortID, v16.RateLimits["injective-1"].ChannelId, channeltypes.Channel{})
	gaiaApp.IBCKeeper.ChannelKeeper.SetChannel(ctx, transfertypes.PortID, v16.RateLimits["core-1"].ChannelId, channeltypes.Channel{})
	gaiaApp.IBCKeeper.ChannelKeeper.SetChannel(ctx, transfertypes.PortID, v16.RateLimits["secret-4"].ChannelId, channeltypes.Channel{})

	err = v16.AddRateLimits(ctx, ratelimitkeeper)
	require.NoError(t, err)

	for chain, msg := range v16.RateLimits {
		expectedRateLimit := ratelimittypes.RateLimit{
			Path: &ratelimittypes.Path{
				Denom:     v16.RateLimitDenom,
				ChannelId: msg.ChannelId,
			},
			Flow: &ratelimittypes.Flow{
				Inflow:       sdkmath.NewInt(0),
				Outflow:      sdkmath.NewInt(0),
				ChannelValue: AtomSupply,
			},
			Quota: &ratelimittypes.Quota{
				DurationHours: v16.RateLimitDurationHours,
			},
		}
		if strings.Compare("osmosis-1", chain) == 0 {
			expectedRateLimit.Quota.MaxPercentSend = sdkmath.NewInt(5)
			expectedRateLimit.Quota.MaxPercentRecv = sdkmath.NewInt(5)
		} else {
			expectedRateLimit.Quota.MaxPercentSend = sdkmath.NewInt(1)
			expectedRateLimit.Quota.MaxPercentRecv = sdkmath.NewInt(1)
		}
		rateLimit, found := ratelimitkeeper.GetRateLimit(ctx, v16.RateLimitDenom, msg.ChannelId)
		require.True(t, found)
		require.Equal(t, expectedRateLimit, rateLimit)
	}
}
