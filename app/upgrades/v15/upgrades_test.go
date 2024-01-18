package v15_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	"github.com/cosmos/gaia/v15/app/helpers"
	v15 "github.com/cosmos/gaia/v15/app/upgrades/v15"
)

func TestUpgradeSigningInfos(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})
	slashingKeeper := gaiaApp.SlashingKeeper

	signingInfosNum := 8
	emptyAddrSigningInfo := make(map[string]struct{})

	// create some dummy signing infos, half of which with an empty address field
	for i := 0; i < signingInfosNum; i++ {
		pubKey, err := mock.NewPV().GetPubKey()
		require.NoError(t, err)

		consAddr := sdk.ConsAddress(pubKey.Address())
		info := slashingtypes.NewValidatorSigningInfo(
			consAddr,
			0,
			0,
			time.Unix(0, 0),
			false,
			0,
		)

		if i < signingInfosNum/2 {
			info.Address = ""
			emptyAddrSigningInfo[consAddr.String()] = struct{}{}
		}

		slashingKeeper.SetValidatorSigningInfo(ctx, consAddr, info)
		require.NoError(t, err)
	}

	require.Equal(t, signingInfosNum/2, len(emptyAddrSigningInfo))

	// check that signing info are correctly set before migration
	slashingKeeper.IterateValidatorSigningInfos(ctx, func(address sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool) {
		if _, ok := emptyAddrSigningInfo[address.String()]; ok {
			require.Empty(t, info.Address)
		} else {
			require.NotEmpty(t, info.Address)
		}

		return false
	})

	// upgrade signing infos
	v15.UpgradeSigningInfos(ctx, slashingKeeper)

	// check that all signing info are updated as expected after migration
	slashingKeeper.IterateValidatorSigningInfos(ctx, func(address sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool) {
		require.NotEmpty(t, info.Address)

		return false
	})
}
