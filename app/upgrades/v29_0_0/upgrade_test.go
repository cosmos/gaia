package v29_0_0_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	gaiaapp "github.com/cosmos/gaia/v29/app"
	v290 "github.com/cosmos/gaia/v29/app/upgrades/v29_0_0"
)

// providerStoreKey mirrors the unexported constant of the same name in
// app/upgrades/v29_0_0/constants.go and the raw string mounted in
// app/keepers/keys.go. There's no shared exported constant to reuse, so it's
// duplicated here deliberately: if the store name ever drifts between the
// two, this test should fail rather than silently follow the drift.
const providerStoreKey = "provider"

// Keep the historical key literal independent of the implementation so a typo
// in the cleanup cannot silently leave the v28.3.1 receipt behind.
const legacyBalanceMigrationReceiptKey = "\xffgaia/forks/balance-migration/v1"

func TestLegacyBalanceMigrationReceiptCleanup(t *testing.T) {
	for _, tc := range []struct {
		name       string
		chainID    string
		hasReceipt bool
	}{
		{name: "mainnet receipt", chainID: "cosmoshub-4", hasReceipt: true},
		{name: "testnet without receipt", chainID: "testnet", hasReceipt: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app := newTestApp(t.TempDir(), dbm.NewMemDB())
			ctx := sdk.NewContext(app.CommitMultiStore(), tmproto.Header{
				ChainID: tc.chainID,
				Height:  1,
			}, false, log.NewNopLogger())
			store := ctx.KVStore(app.GetKey(upgradetypes.StoreKey))
			if tc.hasReceipt {
				// The retired receipt is opaque to v29; cleanup must not depend
				// on decoding its contents or importing the old fork code.
				store.Set([]byte(legacyBalanceMigrationReceiptKey), []byte("completed-migration-receipt"))
			}
			require.Equal(t, tc.hasReceipt, store.Has([]byte(legacyBalanceMigrationReceiptKey)))

			neighborKey := []byte(legacyBalanceMigrationReceiptKey + "/keep")
			neighborValue := []byte("unrelated-upgrade-data")
			store.Set(neighborKey, neighborValue)
			wantVM := module.VersionMap{"bank": 4}
			require.NoError(t, app.UpgradeKeeper.SetModuleVersionMap(ctx, wantVM))
			plan := upgradetypes.Plan{Name: v290.UpgradeName, Height: 10}
			require.NoError(t, app.UpgradeKeeper.ScheduleUpgrade(ctx, plan))

			mm := module.NewManager()
			configurator := module.NewConfigurator(app.AppCodec(), app.MsgServiceRouter(), app.GRPCQueryRouter())
			handler := v290.CreateUpgradeHandler(mm, configurator, &app.AppKeepers)
			// Call twice to also exercise cleanup with an already-deleted key.
			for range 2 {
				_, err := handler(ctx, plan, module.VersionMap{})
				require.NoError(t, err)
				require.False(t, store.Has([]byte(legacyBalanceMigrationReceiptKey)))
				require.Equal(t, neighborValue, store.Get(neighborKey))
				gotVM, err := app.UpgradeKeeper.GetModuleVersionMap(ctx)
				require.NoError(t, err)
				require.Equal(t, wantVM, gotVM)
				gotPlan, err := app.UpgradeKeeper.GetUpgradePlan(ctx)
				require.NoError(t, err)
				require.Equal(t, plan, gotPlan)
			}
		})
	}
}

func newTestApp(homeDir string, db dbm.DB) *gaiaapp.GaiaApp {
	appOpts := make(simtestutil.AppOptionsMap)
	appOpts[server.FlagInvCheckPeriod] = 5

	return gaiaapp.NewGaiaApp(
		log.NewNopLogger(),
		db,
		nil,
		true, // loadLatest
		map[int64]bool{},
		homeDir,
		appOpts,
		[]wasmkeeper.Option{},
	)
}

// TestProviderStoreWipedByUpgradeHandler proves the content-deletion itself:
// it calls v29_0_0.CreateUpgradeHandler directly (the same closure app.go
// registers as the "v29.0.0" upgrade handler) against a real GaiaApp's
// store, and asserts every key seeded into the deprecated "provider" store
// beforehand is gone afterward.
//
// The handler is invoked directly rather than by driving x/upgrade's
// PreBlocker through a scheduled plan, since only the delete-loop's
// correctness is under test here; TestProviderStoreSurvivesRestartAfterUpgrade
// below drives the real on-chain plan + PreBlocker path, which is where the
// height/name wiring actually matters.
func TestProviderStoreWipedByUpgradeHandler(t *testing.T) {
	db := dbm.NewMemDB()
	app := newTestApp(t.TempDir(), db)

	testKey, testVal := []byte("leftover-ccv-key"), []byte("leftover-ccv-value")
	providerStore := app.CommitMultiStore().GetKVStore(app.GetKey(providerStoreKey))
	providerStore.Set(testKey, testVal)
	require.Equal(t, testVal, providerStore.Get(testKey))

	// A minimal module.Manager/Configurator is enough here: with an empty
	// VersionMap and no modules, RunMigrations (called after our delete
	// step, inside CreateUpgradeHandler) is a no-op.
	mm := module.NewManager()
	configurator := module.NewConfigurator(app.AppCodec(), app.MsgServiceRouter(), app.GRPCQueryRouter())
	handler := v290.CreateUpgradeHandler(mm, configurator, &app.AppKeepers)

	ctx := sdk.NewContext(app.CommitMultiStore(), tmproto.Header{Height: 1}, false, log.NewNopLogger())
	_, err := handler(ctx, upgradetypes.Plan{Name: v290.UpgradeName, Height: 1}, module.VersionMap{})
	require.NoError(t, err)

	require.Nil(t, providerStore.Get(testKey), "provider store should be empty right after the handler runs")
}

// TestProviderStoreSurvivesRestartAfterUpgrade simulates a production path:
//  1. app1 seeds the "provider" store, schedules the real on-chain upgrade
//     plan, and (mirroring what a halting old binary does) dumps
//     upgrade-info.json to disk.
//  2. app2 boots on the same db (a fresh home dir only to dodge wasmvm's
//     exclusive-lock-per-directory issue when two GaiaApp instances are
//     alive in one test process — see the comment below), reading that
//     upgrade-info.json exactly as a real restarting node would. newTestApp
//     passes loadLatest=true to NewGaiaApp, which internally calls
//     LoadLatestVersion(). It then finalizes the block at the upgrade
//     height, which is what actually invokes the registered
//     CreateUpgradeHandler through x/upgrade's PreBlocker.
//  3. app3 boots again, on the same db, with no pending upgrade at all
//     (an entirely ordinary restart). This would panic using the
//     StoreUpgrades.Deleted mechanism.
func TestProviderStoreSurvivesRestartAfterUpgrade(t *testing.T) {
	db := dbm.NewMemDB()
	testKey, testVal := []byte("leftover-ccv-key"), []byte("leftover-ccv-value")

	// --- Step 1: seed pre-upgrade state and schedule the real upgrade ---
	app1Home := t.TempDir()
	app1 := newTestApp(app1Home, db)

	providerStore := app1.CommitMultiStore().GetKVStore(app1.GetKey(providerStoreKey))
	providerStore.Set(testKey, testVal)
	upgradeStore := app1.CommitMultiStore().GetKVStore(app1.GetKey(upgradetypes.StoreKey))
	upgradeStore.Set([]byte(legacyBalanceMigrationReceiptKey), []byte("completed-migration-receipt"))
	app1.CommitMultiStore().Commit()

	// app1 has committed once (seeding "provider") by this point, so the
	// next commit (scheduling the upgrade) lands at version 2 — the upgrade
	// itself must then be scheduled for version+1, matching what
	// FinalizeBlock will require of app2's next block.
	const upgradeHeight = int64(3)
	scheduleCtx := sdk.NewContext(app1.CommitMultiStore(), tmproto.Header{Height: app1.CommitMultiStore().LastCommitID().Version}, false, log.NewNopLogger())
	require.NoError(t, app1.UpgradeKeeper.ScheduleUpgrade(scheduleCtx, upgradetypes.Plan{
		Name:   v290.UpgradeName,
		Height: upgradeHeight,
	}))
	app1.CommitMultiStore().Commit()

	require.NoError(t, app1.UpgradeKeeper.DumpUpgradeInfoToDisk(upgradeHeight, upgradetypes.Plan{
		Name: v290.UpgradeName,
	}))
	upgradeInfo, err := os.ReadFile(filepath.Join(app1Home, "data", "upgrade-info.json"))
	require.NoError(t, err)

	// --- Step 2: "reboot" into app2 and process the upgrade-height block ---
	//
	// app2 gets its own home dir (upgrade-info.json copied over) rather than
	// reusing app1Home: wasmvm takes an exclusive file lock keyed on the home
	// dir, app1 never releases it (nothing in gaia or wasmvm's Go bindings
	// calls VM.Cleanup()/BaseApp.Close() doesn't touch it), and app1 is still
	// alive in this test process, so sharing a home dir here would be flaky.
	// The chain state itself (including the seeded "provider" data) lives in
	// the shared db, not the home dir, so this doesn't affect what's tested.
	app2Home := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(app2Home, "data"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(app2Home, "data", "upgrade-info.json"), upgradeInfo, 0o600)) //nolint:gosec

	app2 := newTestApp(app2Home, db)
	require.True(t, app2.CommitMultiStore().GetKVStore(app2.GetKey(upgradetypes.StoreKey)).Has([]byte(legacyBalanceMigrationReceiptKey)))
	_, err = app2.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: upgradeHeight,
		Time:   time.Now(),
	})
	require.NoError(t, err)
	_, err = app2.Commit()
	require.NoError(t, err)

	postUpgradeStore := app2.CommitMultiStore().GetKVStore(app2.GetKey(providerStoreKey))
	require.Nil(t, postUpgradeStore.Get(testKey), "provider store should have been wiped by the v29.0.0 upgrade handler")
	require.False(t, app2.CommitMultiStore().GetKVStore(app2.GetKey(upgradetypes.StoreKey)).Has([]byte(legacyBalanceMigrationReceiptKey)),
		"receipt should be deleted by the scheduled upgrade")

	// --- Step 3: an entirely ordinary subsequent restart ---
	app3Home := t.TempDir()
	var app3 *gaiaapp.GaiaApp
	require.NotPanics(t, func() {
		app3 = newTestApp(app3Home, db)
	}, "an ordinary restart after the v29.0.0 upgrade must not panic")

	finalStore := app3.CommitMultiStore().GetKVStore(app3.GetKey(providerStoreKey))
	require.Nil(t, finalStore.Get(testKey), "provider store should remain empty and readable across further restarts")
	require.False(t, app3.CommitMultiStore().GetKVStore(app3.GetKey(upgradetypes.StoreKey)).Has([]byte(legacyBalanceMigrationReceiptKey)),
		"receipt deletion should persist across an ordinary restart")
}

// TestLegacyParamsStoreWipedByUpgradeHandler proves the x/params store
// cleanup added alongside the x/params module removal: it seeds keys under
// several former subspace prefixes — including "ratelimit", which is no
// longer a live consumer now that the ratelimit keeper is constructed with a
// zero-value Subspace (see app/keepers/keepers.go) — directly in the shared
// params kv-store, then asserts the handler wipes the store entirely, while
// a value held in a genuinely different store (staking's own, live params)
// survives untouched.
func TestLegacyParamsStoreWipedByUpgradeHandler(t *testing.T) {
	db := dbm.NewMemDB()
	app := newTestApp(t.TempDir(), db)

	paramsStore := app.CommitMultiStore().GetKVStore(app.GetKey(paramstypes.StoreKey))

	seeded := map[string]string{
		"auth/MaxMemoCharacters": "leftover-auth-param",
		"ratelimit/SomeParam":    "leftover-ratelimit-param",
		"tokenfactory/SomeParam": "leftover-tokenfactory-param",
	}
	for key, val := range seeded {
		paramsStore.Set([]byte(key), []byte(val))
		require.Equal(t, []byte(val), paramsStore.Get([]byte(key)))
	}

	stakingStore := app.CommitMultiStore().GetKVStore(app.GetKey(stakingtypes.StoreKey))
	liveKey, liveVal := []byte("some-live-staking-key"), []byte("should-survive")
	stakingStore.Set(liveKey, liveVal)

	mm := module.NewManager()
	configurator := module.NewConfigurator(app.AppCodec(), app.MsgServiceRouter(), app.GRPCQueryRouter())
	handler := v290.CreateUpgradeHandler(mm, configurator, &app.AppKeepers)

	ctx := sdk.NewContext(app.CommitMultiStore(), tmproto.Header{Height: 1}, false, log.NewNopLogger())
	_, err := handler(ctx, upgradetypes.Plan{Name: v290.UpgradeName, Height: 1}, module.VersionMap{})
	require.NoError(t, err)

	for key := range seeded {
		require.Nil(t, paramsStore.Get([]byte(key)), "legacy params store key %q should be deleted by the handler", key)
	}
	require.Equal(t, liveVal, stakingStore.Get(liveKey), "keys in an unrelated, live store should survive the handler untouched")
}

// TestStakingParamsSurviveUpgradeHandler proves that deleteLegacyParamSubspaces
// wiping the stale "staking/..." keys out of the shared params kv-store has
// no effect on the staking module's own, live parameters: those live in a
// completely separate store (stakingtypes.StoreKey), not the params store,
// so this is really a structural guarantee — this test just makes it
// concrete by round-tripping a non-default value through the real handler.
func TestStakingParamsSurviveUpgradeHandler(t *testing.T) {
	db := dbm.NewMemDB()
	app := newTestApp(t.TempDir(), db)

	ctx := sdk.NewContext(app.CommitMultiStore(), tmproto.Header{Height: 1}, false, log.NewNopLogger())

	wantParams := stakingtypes.DefaultParams()
	wantParams.MaxValidators = 42 // a non-default value, so this couldn't pass by coincidence
	require.NoError(t, app.StakingKeeper.SetParams(ctx, wantParams))

	gotParams, err := app.StakingKeeper.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, wantParams, gotParams, "sanity check: params should read back before the handler even runs")

	mm := module.NewManager()
	configurator := module.NewConfigurator(app.AppCodec(), app.MsgServiceRouter(), app.GRPCQueryRouter())
	handler := v290.CreateUpgradeHandler(mm, configurator, &app.AppKeepers)
	_, err = handler(ctx, upgradetypes.Plan{Name: v290.UpgradeName, Height: 1}, module.VersionMap{})
	require.NoError(t, err)

	gotParams, err = app.StakingKeeper.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, wantParams, gotParams, "staking params should be unchanged after the v29.0.0 upgrade handler runs")
}
