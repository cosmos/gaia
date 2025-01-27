package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	apphelpers "github.com/strangelove-ventures/tokenfactory/app/helpers"
	appparams "github.com/strangelove-ventures/tokenfactory/app/params"
	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtypes "github.com/cometbft/cometbft/types"

	dbm "github.com/cosmos/cosmos-db"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types" //nolint:all
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	pruningtypes "cosmossdk.io/store/pruning/types"
	"cosmossdk.io/store/snapshots"
	snapshottypes "cosmossdk.io/store/snapshots/types"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// SimAppChainID hardcoded chainID for simulation
const (
	SimAppChainID = "testing"
)

// EmptyBaseAppOptions is a stub implementing AppOptions
type EmptyBaseAppOptions struct{}

// Get implements AppOptions
func (ao EmptyBaseAppOptions) Get(_ string) interface{} {
	return nil
}

// DefaultConsensusParams defines the default Tendermint consensus params used
// in app testing.
var DefaultConsensusParams = &tmproto.ConsensusParams{
	Block: &tmproto.BlockParams{
		MaxBytes: 200000,
		MaxGas:   2000000,
	},
	Evidence: &tmproto.EvidenceParams{
		MaxAgeNumBlocks: 302400,
		MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
		MaxBytes:        10000,
	},
	Validator: &tmproto.ValidatorParams{
		PubKeyTypes: []string{
			tmtypes.ABCIPubKeyTypeEd25519,
		},
	},
	Version: &tmproto.VersionParams{
		App: 0,
	},
	Abci: &tmproto.ABCIParams{
		VoteExtensionsEnableHeight: 1,
	},
}

type EmptyAppOptions struct{}

func (EmptyAppOptions) Get(_ string) interface{} { return nil }

func Setup(t *testing.T) (sdk.Context, *TokenFactoryApp) {
	t.Helper()

	privVal := apphelpers.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	// create validator set with single validator
	validator := tmtypes.NewValidator(pubKey, 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})

	sdkAddr, err := sdk.AccAddressFromHexUnsafe(pubKey.Address().String())
	require.NoError(t, err)

	// generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(sdkAddr, senderPrivKey.PubKey(), 0, 0)
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(appparams.BondDenom, sdkmath.NewInt(100000000000000))),
	}

	ctx, app := SetupWithGenesisValSet(t, valSet, []authtypes.GenesisAccount{acc}, balance)

	return ctx, app
}

// SetupWithGenesisValSet initializes a new App with a validator set and genesis accounts
// that also act as delegators. For simplicity, each validator is bonded with a delegation
// of one consensus engine unit in the default token of the App from first genesis
// account. A Nop logger is set in app.
func SetupWithGenesisValSet(t *testing.T, valSet *tmtypes.ValidatorSet, genAccs []authtypes.GenesisAccount, balances ...banktypes.Balance) (sdk.Context, *TokenFactoryApp) {
	t.Helper()

	appparams.SetAddressPrefixes()

	app, genesisState := setup(t, true)

	ctx := app.BaseApp.NewUncachedContext(true, tmproto.Header{Height: 1, ChainID: "testing", Time: time.Now().UTC()})

	genesisState = genesisStateWithValSet(t, app, genesisState, valSet, genAccs, balances...)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	require.NoError(t, err)

	// init chain will set the validator set and initialize the genesis accounts
	_, err = app.InitChain(
		&abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: DefaultConsensusParams,
			AppStateBytes:   stateBytes,
			ChainId:         SimAppChainID,
			Time:            time.Now().UTC(),
			InitialHeight:   1,
		},
	)
	require.NoError(t, err)

	// commit genesis changes
	_, err = app.Commit()
	require.NoError(t, err)

	// checking the error here throws for standard collection types. Likely something with encoding
	app.BeginBlocker(ctx) // nolint:errcheck

	return ctx, app
}

//nolint:all
func setup(t *testing.T, withGenesis bool, opts ...wasmkeeper.Option) (*TokenFactoryApp, GenesisState) {
	db := dbm.NewMemDB()
	nodeHome := t.TempDir()
	snapshotDir := filepath.Join(nodeHome, "data", "snapshots")

	snapshotDB, err := dbm.NewDB("metadata", dbm.GoLevelDBBackend, snapshotDir)
	require.NoError(t, err)
	t.Cleanup(func() { snapshotDB.Close() })
	snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
	require.NoError(t, err)

	// var emptyWasmOpts []wasm.Option
	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = nodeHome // ensure unique folder

	app := NewApp(
		log.NewNopLogger(),
		db,
		nil,
		true,
		EmptyAppOptions{},
		opts,
		bam.SetChainID(SimAppChainID),
		bam.SetSnapshot(snapshotStore, snapshottypes.SnapshotOptions{KeepRecent: 2}),
	)

	// Setup Context
	ctx := app.BaseApp.NewUncachedContext(false, tmproto.Header{
		ChainID: SimAppChainID,
		Height:  1,
		Time:    time.Now().UTC(),
	})

	// Set Default Params
	app.MintKeeper.Minter.Set(ctx, minttypes.DefaultInitialMinter())
	app.MintKeeper.Params.Set(ctx, minttypes.DefaultParams())
	app.CrisisKeeper.ConstantFee.Set(ctx, sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100000)))
	app.DistrKeeper.Params.Set(ctx, distrtypes.DefaultParams())
	app.DistrKeeper.FeePool.Set(ctx, distrtypes.FeePool{
		CommunityPool: sdk.NewDecCoins(),
	})
	app.TransferKeeper.SetParams(ctx, transfertypes.DefaultParams())
	app.StakingKeeper.SetParams(ctx, stakingtypes.DefaultParams())
	app.IBCKeeper.ClientKeeper.SetParams(ctx, clienttypes.DefaultParams())
	app.IBCKeeper.ClientKeeper.SetNextClientSequence(ctx, 0)
	app.IBCKeeper.ConnectionKeeper.SetNextConnectionSequence(ctx, 0)
	app.IBCKeeper.ConnectionKeeper.SetParams(ctx, connectiontypes.DefaultParams())
	app.IBCKeeper.ChannelKeeper.SetNextChannelSequence(ctx, 0)
	app.WasmKeeper.SetParams(ctx, wasm.DefaultParams())
	app.ICAControllerKeeper.SetParams(ctx, icatypes.DefaultParams())
	app.ICAHostKeeper.SetParams(ctx, icahosttypes.DefaultParams())
	app.GovKeeper.Constitution.Set(ctx, "")
	app.GovKeeper.Params.Set(ctx, govv1types.DefaultParams())
	app.ConsensusParamsKeeper.ParamsStore.Set(ctx, *simtestutil.DefaultConsensusParams)

	if withGenesis {
		return app, NewDefaultGenesisState(t)
	}

	return app, GenesisState{}
}

func genesisStateWithValSet(t *testing.T,
	app *TokenFactoryApp, genesisState GenesisState,
	valSet *tmtypes.ValidatorSet, genAccs []authtypes.GenesisAccount,
	balances ...banktypes.Balance,
) GenesisState {
	codec := app.AppCodec()

	// set genesis accounts
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = codec.MustMarshalJSON(authGenesis)

	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))

	bondAmt := sdk.DefaultPowerReduction

	for _, val := range valSet.Validators {
		pk, err := cryptocodec.FromCmtPubKeyInterface(val.PubKey)
		require.NoError(t, err)
		pkAny, err := codectypes.NewAnyWithValue(pk)
		require.NoError(t, err)

		valAddr, err := sdk.ValAddressFromHex(val.Address.String())
		require.NoError(t, err)

		validator := stakingtypes.Validator{
			OperatorAddress:   valAddr.String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   sdkmath.LegacyOneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec()),
			MinSelfDelegation: sdkmath.ZeroInt(),
		}
		validators = append(validators, validator)
		delegations = append(delegations, stakingtypes.NewDelegation(genAccs[0].GetAddress().String(), valAddr.String(), sdkmath.LegacyOneDec()))

	}

	defaultStParams := stakingtypes.DefaultParams()
	stParams := stakingtypes.NewParams(
		defaultStParams.UnbondingTime,
		defaultStParams.MaxValidators,
		defaultStParams.MaxEntries,
		defaultStParams.HistoricalEntries,
		appparams.BondDenom,
		defaultStParams.MinCommissionRate, // 5%
	)

	// set validators and delegations
	stakingGenesis := stakingtypes.NewGenesisState(stParams, validators, delegations)
	genesisState[stakingtypes.ModuleName] = codec.MustMarshalJSON(stakingGenesis)

	// add bonded amount to bonded pool module account
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin(appparams.BondDenom, bondAmt.MulRaw(int64(len(valSet.Validators))))},
	})

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// add genesis acc tokens to total supply
		totalSupply = totalSupply.Add(b.Coins...)
	}

	// update total supply
	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, totalSupply, []banktypes.Metadata{}, []banktypes.SendEnabled{})
	genesisState[banktypes.ModuleName] = codec.MustMarshalJSON(bankGenesis)

	return genesisState
}

func keyPubAddr() (crypto.PrivKey, crypto.PubKey, sdk.AccAddress) {
	key := ed25519.GenPrivKey()
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	return key, pub, addr
}

func RandomAccountAddress() sdk.AccAddress {
	_, _, addr := keyPubAddr()
	return addr
}

func ExecuteRawCustom(t *testing.T, ctx sdk.Context, app *TokenFactoryApp, contract sdk.AccAddress, sender sdk.AccAddress, msg json.RawMessage, funds sdk.Coin) error {
	t.Helper()
	oracleBz, err := json.Marshal(msg)
	require.NoError(t, err)
	// no funds sent if amount is 0
	var coins sdk.Coins
	if !funds.Amount.IsNil() {
		coins = sdk.Coins{funds}
	}

	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(app.WasmKeeper)
	_, err = contractKeeper.Execute(ctx, contract, sender, oracleBz, coins)
	return err
}

var emptyWasmOptions []wasmkeeper.Option

// NewTestNetworkFixture returns a new app AppConstructor for network simulation tests
func NewTestNetworkFixture() network.TestFixture {
	dir, err := os.MkdirTemp("", "simapp")
	if err != nil {
		panic(fmt.Sprintf("failed creating temporary directory: %v", err))
	}
	defer os.RemoveAll(dir)

	app := NewApp(log.NewNopLogger(), dbm.NewMemDB(), nil, true, simtestutil.NewAppOptionsWithFlagHome(dir), emptyWasmOptions)
	appCtr := func(val network.ValidatorI) servertypes.Application {
		return NewApp(
			val.GetCtx().Logger, dbm.NewMemDB(), nil, true,
			simtestutil.NewAppOptionsWithFlagHome(val.GetCtx().Config.RootDir),
			emptyWasmOptions,
			bam.SetPruning(pruningtypes.NewPruningOptionsFromString(val.GetAppConfig().Pruning)),
			bam.SetMinGasPrices(val.GetAppConfig().MinGasPrices),
			bam.SetChainID(val.GetCtx().Viper.GetString(flags.FlagChainID)),
		)
	}

	return network.TestFixture{
		AppConstructor: appCtr,
		GenesisState:   app.DefaultGenesis(),
		EncodingConfig: testutil.TestEncodingConfig{
			InterfaceRegistry: app.InterfaceRegistry(),
			Codec:             app.AppCodec(),
			TxConfig:          app.TxConfig(),
			Amino:             app.LegacyAmino(),
		},
	}
}
