package v29_2_0_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/gaia/v29/app/helpers"
	v292 "github.com/cosmos/gaia/v29/app/upgrades/v29_2_0"
)

// TestUpgradeKeepsConsensusParams runs the scheduled v29.2.0 upgrade through
// the PreBlocker and checks that the handler leaves the consensus params, in
// particular validator.pub_key_types, and the module version map unchanged.
//
// v29.2.0 adds ML-DSA-65 verification for counterparty chains (CometBFT
// v0.38.26). The Hub's own validators must stay on ed25519 until enabling
// another key type is decided separately. This test covers what the upgrade
// handler does; it does not cover later governance changes to the params.
func TestUpgradeKeepsConsensusParams(t *testing.T) {
	app := helpers.Setup(t)
	_, err := app.Commit()
	require.NoError(t, err)

	upgradeHeight := app.LastBlockHeight() + 1
	ctx := app.NewUncachedContext(false, cmtproto.Header{Height: app.LastBlockHeight()})

	// Cosmos Hub mainnet values at height 34,020,292.
	hubParams := cmttypes.DefaultConsensusParams().ToProto()
	hubParams.Block.MaxBytes = 2_000_000
	hubParams.Block.MaxGas = 75_000_000
	hubParams.Evidence.MaxAgeNumBlocks = 1_000_000
	hubParams.Evidence.MaxAgeDuration = 48 * time.Hour
	hubParams.Evidence.MaxBytes = 50_000
	hubParams.Validator.PubKeyTypes = []string{cmttypes.ABCIPubKeyTypeEd25519}
	require.NoError(t, app.ConsensusParamsKeeper.ParamsStore.Set(ctx, hubParams))

	wantVersionMap, err := app.UpgradeKeeper.GetModuleVersionMap(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, wantVersionMap)

	require.NoError(t, app.UpgradeKeeper.ScheduleUpgrade(ctx, upgradetypes.Plan{
		Name:   v292.UpgradeName,
		Height: upgradeHeight,
	}))

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: upgradeHeight,
		Time:   time.Now(),
	})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)

	ctx = app.NewUncachedContext(false, cmtproto.Header{Height: app.LastBlockHeight()})

	doneHeight, err := app.UpgradeKeeper.GetDoneHeight(ctx, v292.UpgradeName)
	require.NoError(t, err)
	require.Equal(t, upgradeHeight, doneHeight, "the v29.2.0 handler must run at the plan height")

	gotParams, err := app.ConsensusParamsKeeper.ParamsStore.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, hubParams, gotParams)
	require.Equal(t, []string{cmttypes.ABCIPubKeyTypeEd25519}, gotParams.Validator.PubKeyTypes)

	gotVersionMap, err := app.UpgradeKeeper.GetModuleVersionMap(ctx)
	require.NoError(t, err)
	require.Equal(t, wantVersionMap, gotVersionMap, "v29.2.0 must not change module consensus versions")
}
