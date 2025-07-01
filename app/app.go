package gaia

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	feemarketkeeper "github.com/skip-mev/feemarket/x/feemarket/keeper"
	"github.com/spf13/cast"
	"github.com/spf13/viper"

	abci "github.com/cometbft/cometbft/abci/types"
	tmcfg "github.com/cometbft/cometbft/config"
	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/privval"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/gogoproto/proto"
	ibcwasm "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10"
	ibcwasmkeeper "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10/keeper"
	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10/types"
	ibctm "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"
	ibctesting "github.com/cosmos/ibc-go/v10/testing"
	providertypes "github.com/cosmos/interchain-security/v7/x/ccv/provider/types"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/appmodule"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/x/tx/signing"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	sigtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	txmodule "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	gaiaante "github.com/cosmos/gaia/v25/ante"
	"github.com/cosmos/gaia/v25/app/keepers"
	"github.com/cosmos/gaia/v25/app/upgrades"
	v25 "github.com/cosmos/gaia/v25/app/upgrades/v25_0_0"
	gaiatelemetry "github.com/cosmos/gaia/v25/telemetry"
	"github.com/cosmos/gaia/v25/x/telemetry"
)

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	Upgrades = []upgrades.Upgrade{v25.Upgrade}
)

var (
	_ runtime.AppI            = (*GaiaApp)(nil)
	_ servertypes.Application = (*GaiaApp)(nil)
	_ ibctesting.TestingApp   = (*GaiaApp)(nil)
)

// GaiaApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type GaiaApp struct { //nolint: revive
	*baseapp.BaseApp
	keepers.AppKeepers

	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry types.InterfaceRegistry

	// the module manager
	mm           *module.Manager
	ModuleBasics module.BasicManager

	// simulation manager
	sm           *module.SimulationManager
	configurator module.Configurator

	otelClient *gaiatelemetry.OtelClient
}

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, ".gaia")
}

// NewGaiaApp returns a reference to an initialized Gaia.
func NewGaiaApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	appOpts servertypes.AppOptions,
	wasmOpts []wasmkeeper.Option,
	baseAppOptions ...func(*baseapp.BaseApp),
) *GaiaApp {
	legacyAmino := codec.NewLegacyAmino()
	interfaceRegistry, err := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
			},
			ValidatorAddressCodec: address.Bech32Codec{
				Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
			},
		},
	})
	if err != nil {
		panic(err)
	}
	appCodec := codec.NewProtoCodec(interfaceRegistry)
	txConfig := authtx.NewTxConfig(appCodec, authtx.DefaultSignModes)

	std.RegisterLegacyAminoCodec(legacyAmino)
	std.RegisterInterfaces(interfaceRegistry)

	bApp := baseapp.NewBaseApp(
		appName,
		logger,
		db,
		txConfig.TxDecoder(),
		baseAppOptions...)

	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetTxEncoder(txConfig.TxEncoder())

	app := &GaiaApp{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		txConfig:          txConfig,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
	}

	vi, err := getValidatorInfo(homePath, appOpts)
	if err != nil {
		logger.Debug("failed to get validator info: unable to determine if this node is a validator", "err", err)
	} else {
		logger.Debug("successfully determined if this node is a validator", "moniker", vi.Moniker)
	}

	otelConfig := getOtelConfig(appOpts)
	app.otelClient = gaiatelemetry.NewOtelClient(otelConfig, vi)

	moduleAccountAddresses := app.ModuleAccountAddrs()

	// Setup keepers
	app.AppKeepers = keepers.NewAppKeeper(
		appCodec,
		bApp,
		legacyAmino,
		maccPerms,
		moduleAccountAddresses,
		app.BlockedModuleAccountAddrs(moduleAccountAddresses),
		skipUpgradeHeights,
		homePath,
		logger,
		appOpts,
		wasmOpts,
	)

	// Create IBC Tendermint Light Client Stack
	clientKeeper := app.IBCKeeper.ClientKeeper
	tmLightClientModule := ibctm.NewLightClientModule(appCodec, clientKeeper.GetStoreProvider())
	clientKeeper.AddRoute(ibctm.ModuleName, &tmLightClientModule)

	// Create WASM Light Client Stack
	wasmLightClientModule := ibcwasm.NewLightClientModule(app.WasmClientKeeper, clientKeeper.GetStoreProvider())
	clientKeeper.AddRoute(ibcwasmtypes.ModuleName, &wasmLightClientModule)

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	app.mm = module.NewManager(appModules(app, appCodec, txConfig, tmLightClientModule)...)
	app.ModuleBasics = newBasicManagerFromManager(app)

	enabledSignModes := append([]sigtypes.SignMode(nil), authtx.DefaultSignModes...)
	enabledSignModes = append(enabledSignModes, sigtypes.SignMode_SIGN_MODE_TEXTUAL)

	txConfigOpts := authtx.ConfigOptions{
		EnabledSignModes:           enabledSignModes,
		TextualCoinMetadataQueryFn: txmodule.NewBankKeeperCoinMetadataQueryFn(app.BankKeeper),
	}
	txConfig, err = authtx.NewTxConfigWithOptions(
		appCodec,
		txConfigOpts,
	)
	if err != nil {
		panic(err)
	}
	app.txConfig = txConfig

	// NOTE: upgrade module is required to be prioritized
	app.mm.SetOrderPreBlockers(
		upgradetypes.ModuleName,
		authtypes.ModuleName,
		telemetry.ModuleName,
	)
	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	// Tell the app's module manager how to set the order of BeginBlockers, which are run at the beginning of every block.
	app.mm.SetOrderBeginBlockers(orderBeginBlockers()...)

	app.mm.SetOrderEndBlockers(orderEndBlockers()...)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: The genutils module must also occur after auth so that it can access the params from auth.
	app.mm.SetOrderInitGenesis(orderInitBlockers()...)

	// Uncomment if you want to set a custom migration order here.
	// app.mm.SetOrderMigrations(custom order)

	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	err = app.mm.RegisterServices(app.configurator)
	if err != nil {
		panic(err)
	}

	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.mm.Modules))

	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)

	// add test gRPC service for testing gRPC queries in isolation
	testdata.RegisterQueryServer(app.GRPCQueryRouter(), testdata.QueryImpl{})

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: this is not required apps that don't use the simulator for fuzz testing
	// transactions
	app.sm = module.NewSimulationManager(simulationModules(app, appCodec)...)

	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(app.GetKVStoreKey())
	app.MountTransientStores(app.GetTransientStoreKey())
	app.MountMemoryStores(app.GetMemoryStoreKey())

	wasmConfig, err := wasm.ReadNodeConfig(appOpts)
	if err != nil {
		panic("error while reading wasm config: " + err.Error())
	}

	anteHandler, err := gaiaante.NewAnteHandler(
		gaiaante.HandlerOptions{
			AccountKeeper:   &app.AccountKeeper,
			BankKeeper:      app.BankKeeper,
			FeegrantKeeper:  app.FeeGrantKeeper,
			SignModeHandler: txConfig.SignModeHandler(),
			SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,

			Codec:                 appCodec,
			IBCkeeper:             app.IBCKeeper,
			StakingKeeper:         app.StakingKeeper,
			FeeMarketKeeper:       app.FeeMarketKeeper,
			WasmConfig:            &wasmConfig,
			TXCounterStoreService: runtime.NewKVStoreService(app.GetKey(wasmtypes.StoreKey)),
			TxFeeChecker: func(ctx sdk.Context, tx sdk.Tx) (sdk.Coins, int64, error) {
				return minTxFeesChecker(ctx, tx, *app.FeeMarketKeeper)
			},
		},
	)
	if err != nil {
		panic(fmt.Errorf("failed to create AnteHandler: %s", err))
	}

	postHandlerOptions := PostHandlerOptions{
		AccountKeeper:   app.AccountKeeper,
		BankKeeper:      app.BankKeeper,
		FeeMarketKeeper: app.FeeMarketKeeper,
	}
	postHandler, err := NewPostHandler(postHandlerOptions)
	if err != nil {
		panic(err)
	}

	// set ante and post handlers
	app.SetAnteHandler(anteHandler)
	app.SetPostHandler(postHandler)

	app.SetInitChainer(app.InitChainer)
	app.SetPreBlocker(app.PreBlocker)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	if manager := app.SnapshotManager(); manager != nil {
		err = manager.RegisterExtensions(
			wasmkeeper.NewWasmSnapshotter(app.CommitMultiStore(), &app.WasmKeeper),
			ibcwasmkeeper.NewWasmSnapshotter(app.CommitMultiStore(), &app.WasmClientKeeper),
		)
		if err != nil {
			panic("failed to register snapshot extension: " + err.Error())
		}
	}

	app.setupUpgradeHandlers()
	app.setupUpgradeStoreLoaders()

	// At startup, after all modules have been registered, check that all prot
	// annotations are correct.
	protoFiles, err := proto.MergedRegistry()
	if err != nil {
		panic(err)
	}
	err = msgservice.ValidateProtoAnnotations(protoFiles)
	if err != nil {
		// Once we switch to using protoreflect-based antehandlers, we might
		// want to panic here instead of logging a warning.
		fmt.Fprintln(os.Stderr, err.Error())
	}

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			panic(fmt.Sprintf("failed to load latest version: %s", err))
		}

		ctx := app.NewUncachedContext(true, tmproto.Header{})

		if err := app.WasmKeeper.InitializePinnedCodes(ctx); err != nil {
			panic(fmt.Sprintf("WasmKeeper failed initialize pinned codes %s", err))
		}

		if err := app.WasmClientKeeper.InitializePinnedCodes(ctx); err != nil {
			panic(fmt.Sprintf("wasmlckeeper failed initialize pinned codes %s", err))
		}
	}

	if otelConfig.CollectorEndpoint != "" && !otelConfig.Disable {
		logger.Debug("creating gaia app with open telemetry")
		if err := app.otelClient.StartExporter(logger); err != nil {
			panic(err)
		}
	} else {
		logger.Debug("creating gaia app without open telemetry")
	}

	return app
}

// Name returns the name of the App
func (app *GaiaApp) Name() string { return app.BaseApp.Name() }

// PreBlocker application updates every pre block
func (app *GaiaApp) PreBlocker(ctx sdk.Context, _ *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	return app.mm.PreBlock(ctx)
}

// BeginBlocker application updates every begin block
func (app *GaiaApp) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.mm.BeginBlock(ctx)
}

// EndBlocker application updates every end block
func (app *GaiaApp) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.mm.EndBlock(ctx)
}

// InitChainer application update at chain initialization
func (app *GaiaApp) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}

	if err := app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap()); err != nil {
		panic(err)
	}

	response, err := app.mm.InitGenesis(ctx, app.appCodec, genesisState)
	if err != nil {
		panic(err)
	}

	return response, nil
}

// LoadHeight loads a particular height
func (app *GaiaApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *GaiaApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// BlockedModuleAccountAddrs returns all the app's blocked module account
// addresses.
func (app *GaiaApp) BlockedModuleAccountAddrs(modAccAddrs map[string]bool) map[string]bool {
	// remove module accounts that are ALLOWED to received funds
	delete(modAccAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	// Remove the ConsumerRewardsPool from the group of blocked recipient addresses in bank
	delete(modAccAddrs, authtypes.NewModuleAddress(providertypes.ConsumerRewardsPool).String())

	return modAccAddrs
}

func getOtelConfig(appOpts servertypes.AppOptions) gaiatelemetry.OtelConfig {
	// if appOpts.Get yields nil, this value was not set.
	// since the user isn't making any intent to disable here, we will use the DefaultOtelConfig.
	disableRaw := appOpts.Get("opentelemetry.disable")
	if disableRaw == nil {
		return gaiatelemetry.DefaultOtelConfig
	}
	// if disableRaw wasn't nil, the user is making the intent to use their config. so we will use their values.
	otelConfig := gaiatelemetry.OtelConfig{
		Disable:                 cast.ToBool(appOpts.Get("opentelemetry.disable")),
		CollectorEndpoint:       cast.ToString(appOpts.Get("opentelemetry.collector-endpoint")),
		CollectorMetricsURLPath: cast.ToString(appOpts.Get("opentelemetry.collector-metrics-url-path")),
		User:                    cast.ToString(appOpts.Get("opentelemetry.user")),
		Token:                   cast.ToString(appOpts.Get("opentelemetry.token")),
		PushInterval:            cast.ToDuration(appOpts.Get("opentelemetry.push-interval")),
	}
	return otelConfig
}

// LegacyAmino returns GaiaApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *GaiaApp) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns Gaia's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *GaiaApp) AppCodec() codec.Codec {
	return app.appCodec
}

// DefaultGenesis returns a default genesis from the registered AppModuleBasic's.
func (app *GaiaApp) DefaultGenesis() map[string]json.RawMessage {
	return app.ModuleBasics.DefaultGenesis(app.appCodec)
}

// InterfaceRegistry returns Gaia's InterfaceRegistry
func (app *GaiaApp) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// SimulationManager implements the SimulationApp interface
func (app *GaiaApp) SimulationManager() *module.SimulationManager {
	return app.sm
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *GaiaApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register legacy and grpc-gateway routes for all modules.
	app.ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register nodeservice grpc-gateway routes.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
}

// RegisterNodeService allows query minimum-gas-prices in app.toml
func (app *GaiaApp) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *GaiaApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.GRPCQueryRouter(), clientCtx, app.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *GaiaApp) RegisterTendermintService(clientCtx client.Context) {
	cmtservice.RegisterTendermintService(
		clientCtx,
		app.GRPCQueryRouter(),
		app.interfaceRegistry,
		app.Query,
	)
}

// configure store loader that checks if version == upgradeHeight and applies store upgrades
func (app *GaiaApp) setupUpgradeStoreLoaders() {
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("failed to read upgrade info from disk %s", err))
	}

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	for _, upgrade := range Upgrades {
		upgrade := upgrade
		if upgradeInfo.Name == upgrade.UpgradeName {
			storeUpgrades := upgrade.StoreUpgrades
			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
		}
	}
}

func (app *GaiaApp) setupUpgradeHandlers() {
	for _, upgrade := range Upgrades {
		app.UpgradeKeeper.SetUpgradeHandler(
			upgrade.UpgradeName,
			upgrade.CreateUpgradeHandler(
				app.mm,
				app.configurator,
				&app.AppKeepers,
			),
		)
	}
}

// RegisterSwaggerAPI registers swagger route with API Server
func RegisterSwaggerAPI(rtr *mux.Router) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))
}

func (app *GaiaApp) OnTxSucceeded(_ sdk.Context, _, _ string, _ []byte, _ []byte) {
}

func (app *GaiaApp) OnTxFailed(_ sdk.Context, _, _ string, _ []byte, _ []byte) {
}

// AutoCliOpts returns the autocli options for the app.
func (app *GaiaApp) AutoCliOpts() autocli.AppOptions {
	modules := make(map[string]appmodule.AppModule, 0)
	for _, m := range app.mm.Modules {
		if moduleWithName, ok := m.(module.HasName); ok {
			moduleName := moduleWithName.Name()
			if appModule, ok := moduleWithName.(appmodule.AppModule); ok {
				modules[moduleName] = appModule
			}
		}
	}

	return autocli.AppOptions{
		Modules:               modules,
		AddressCodec:          authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		ValidatorAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		ConsensusAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	}
}

func getValidatorInfo(homePath string, appOpts servertypes.AppOptions) (gaiatelemetry.ValidatorInfo, error) {
	cfg := &tmcfg.Config{
		BaseConfig:      tmcfg.BaseConfig{},
		RPC:             &tmcfg.RPCConfig{},
		P2P:             &tmcfg.P2PConfig{},
		Mempool:         &tmcfg.MempoolConfig{},
		StateSync:       &tmcfg.StateSyncConfig{},
		BlockSync:       &tmcfg.BlockSyncConfig{},
		Consensus:       &tmcfg.ConsensusConfig{},
		Storage:         &tmcfg.StorageConfig{},
		TxIndex:         &tmcfg.TxIndexConfig{},
		Instrumentation: &tmcfg.InstrumentationConfig{},
	}
	cfg.SetRoot(homePath)

	configPath := filepath.Join(homePath, "config", "config.toml")
	if _, err := os.Stat(configPath); err == nil {
		viper := viper.New()
		viper.SetConfigType("toml")
		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err == nil {
			if err := viper.Unmarshal(cfg); err != nil {
				return gaiatelemetry.ValidatorInfo{}, fmt.Errorf("failed to unmarshal config file: %w", err)
			}
		}
	} else {
		return gaiatelemetry.ValidatorInfo{}, fmt.Errorf("unable to stat config file at %s", configPath)
	}

	chainID := cast.ToString(appOpts.Get(flags.FlagChainID))
	if chainID == "" {
		genDocFile := filepath.Join(homePath, "config", "genesis.json")
		appGenesis, err := genutiltypes.AppGenesisFromFile(genDocFile)
		if err == nil {
			chainID = appGenesis.ChainID
		}
	}

	vi, err := validatorInfoFromCometConfig(cfg, chainID)
	return vi, err
}

// TestingApp functions

// GetBaseApp implements the TestingApp interface.
func (app *GaiaApp) GetBaseApp() *baseapp.BaseApp {
	return app.BaseApp
}

// GetTxConfig implements the TestingApp interface.
func (app *GaiaApp) GetTxConfig() client.TxConfig {
	return app.txConfig
}

// GetTestGovKeeper implements the TestingApp interface.
func (app *GaiaApp) GetTestGovKeeper() *govkeeper.Keeper {
	return app.GovKeeper
}

// EmptyAppOptions is a stub implementing AppOptions
type EmptyAppOptions struct{}

// EmptyWasmOptions is a stub implementing Wasmkeeper Option
var EmptyWasmOptions []wasmkeeper.Option

// Get implements AppOptions
func (ao EmptyAppOptions) Get(_ string) interface{} {
	return nil
}

// minTxFeesChecker will be executed only if the feemarket module is disabled.
// In this case, the auth module's DeductFeeDecorator is executed, and
// we use the minTxFeesChecker to enforce the minimum transaction fees.
// Min tx fees are calculated as gas_limit * feemarket_min_base_gas_price
func minTxFeesChecker(ctx sdk.Context, tx sdk.Tx, feemarketKp feemarketkeeper.Keeper) (sdk.Coins, int64, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, 0, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	// To keep the gentxs with zero fees, we need to skip the validation in the first block
	if ctx.BlockHeight() == 0 {
		return feeTx.GetFee(), 0, nil
	}

	feeMarketParams, err := feemarketKp.GetParams(ctx)
	if err != nil {
		return nil, 0, err
	}

	feeRequired := sdk.NewCoins(
		sdk.NewCoin(
			feeMarketParams.FeeDenom,
			feeMarketParams.MinBaseGasPrice.MulInt(math.NewIntFromUint64(feeTx.GetGas())).Ceil().RoundInt()))

	feeCoins := feeTx.GetFee()
	if len(feeCoins) != 1 {
		return nil, 0, fmt.Errorf(
			"expected exactly one fee coin; got %s, required: %s", feeCoins.String(), feeRequired.String())
	}

	if !feeCoins.IsAnyGTE(feeRequired) {
		return nil, 0, fmt.Errorf(
			"not enough fees provided; got %s, required: %s", feeCoins.String(), feeRequired.String())
	}

	return feeTx.GetFee(), 0, nil
}

var ErrNotValidator = fmt.Errorf("not validator")

func validatorInfoFromCometConfig(cfg *tmcfg.Config, chainID string) (gaiatelemetry.ValidatorInfo, error) {
	vi := gaiatelemetry.ValidatorInfo{
		ChainID: chainID,
	}
	if cfg.PrivValidatorListenAddr != "" {
		listenAddr := cfg.PrivValidatorListenAddr
		pve, err := privval.NewSignerListener(listenAddr, nil)
		if err != nil {
			return vi, fmt.Errorf("failed to start private validator: %w", err)
		}

		pvsc, err := privval.NewSignerClient(pve, chainID)
		if err != nil {
			return vi, fmt.Errorf("failed to start private validator: %w", err)
		}

		pk, err := pvsc.GetPubKey()
		if err != nil {
			return vi, fmt.Errorf("cannot get pubkey from remote signer: %w", err)
		}
		vi.Moniker = cfg.Moniker
		vi.Address = pk.Address()
		return vi, nil
	} else if cfg.PrivValidatorKey != "" {
		vi.Moniker = cfg.Moniker
		_, err := os.Stat(cfg.PrivValidatorKeyFile())
		if err != nil {
			return vi, ErrNotValidator
		}
		pv := privval.LoadFilePV(cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile())
		vi.Address = pv.GetAddress()
		return vi, nil
	}
	return vi, ErrNotValidator
}
