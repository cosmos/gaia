package lcdtest

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	"github.com/cosmos/cosmos-sdk/codec"
	crkeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/genaccounts"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/tendermint/go-amino"
	tmcfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	nm "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	pvm "github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmrpc "github.com/tendermint/tendermint/rpc/lib/server"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	gapp "github.com/cosmos/gaia/app"
)

// TODO: Make InitializeTestLCD safe to call in multiple tests at the same time
// InitializeLCD starts Tendermint and the LCD in process, listening on
// their respective sockets where nValidators is the total number of validators
// and initAddrs are the accounts to initialize with some stake tokens. It
// returns a cleanup function, a set of validator public keys, and a port.
func InitializeLCD(nValidators int, initAddrs []sdk.AccAddress, minting bool, portExt ...string) (
	cleanup func(), valConsPubKeys []crypto.PubKey, valOperAddrs []sdk.ValAddress, port string, err error) {

	config, err := GetConfig()
	if err != nil {
		return
	}
	config.Consensus.TimeoutCommit = 100
	config.Consensus.SkipTimeoutCommit = false
	config.TxIndex.IndexAllTags = true

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	logger = log.NewFilter(logger, log.AllowError())

	db := dbm.NewMemDB()
	app := gapp.NewGaiaApp(logger, db, nil, true, 0, baseapp.SetPruning(store.PruneNothing))
	cdc = gapp.MakeCodec()

	genDoc, valConsPubKeys, valOperAddrs, privVal, err := defaultGenesis(config, nValidators, initAddrs, minting)
	if err != nil {
		return
	}

	var listenAddr string

	if len(portExt) == 0 {
		listenAddr, port, err = server.FreeTCPAddr()
		if err != nil {
			return
		}
	} else {
		listenAddr = fmt.Sprintf("tcp://0.0.0.0:%s", portExt[0])
		port = portExt[0]
	}

	// XXX: Need to set this so LCD knows the tendermint node address!
	viper.Set(client.FlagNode, config.RPC.ListenAddress)
	viper.Set(client.FlagChainID, genDoc.ChainID)
	// TODO Set to false once the upstream Tendermint proof verification issue is fixed.
	viper.Set(client.FlagTrustNode, true)

	node, err := startTM(config, logger, genDoc, privVal, app)
	if err != nil {
		return
	}

	tests.WaitForNextHeightTM(tests.ExtractPortFromAddress(config.RPC.ListenAddress))
	lcdInstance, err := startLCD(logger, listenAddr, cdc)
	if err != nil {
		return
	}

	tests.WaitForLCDStart(port)
	tests.WaitForHeight(1, port)

	cleanup = func() {
		logger.Debug("cleaning up LCD initialization")
		err = node.Stop()
		if err != nil {
			logger.Error(err.Error())
		}

		node.Wait()
		err = lcdInstance.Close()
		if err != nil {
			logger.Error(err.Error())
		}
	}

	return cleanup, valConsPubKeys, valOperAddrs, port, err
}

func defaultGenesis(config *tmcfg.Config, nValidators int, initAddrs []sdk.AccAddress, minting bool) (
	genDoc *tmtypes.GenesisDoc, valConsPubKeys []crypto.PubKey, valOperAddrs []sdk.ValAddress, privVal *pvm.FilePV, err error) {
	privVal = pvm.LoadOrGenFilePV(config.PrivValidatorKeyFile(),
		config.PrivValidatorStateFile())
	privVal.Reset()

	if nValidators < 1 {
		err = errors.New("InitializeLCD must use at least one validator")
		return
	}

	genesisFile := config.GenesisFile()
	genDoc, err = tmtypes.GenesisDocFromFile(genesisFile)
	if err != nil {
		return
	}
	genDoc.Validators = nil
	err = genDoc.SaveAs(genesisFile)
	if err != nil {
		return
	}

	// append any additional (non-proposing) validators
	var genTxs []auth.StdTx
	var accs []genaccounts.GenesisAccount

	totalSupply := sdk.ZeroInt()

	for i := 0; i < nValidators; i++ {
		operPrivKey := secp256k1.GenPrivKey()
		operAddr := operPrivKey.PubKey().Address()
		pubKey := privVal.GetPubKey()

		power := int64(100)
		if i > 0 {
			pubKey = ed25519.GenPrivKey().PubKey()
			power = 1
		}
		startTokens := sdk.TokensFromConsensusPower(power)

		msg := staking.NewMsgCreateValidator(
			sdk.ValAddress(operAddr),
			pubKey,
			sdk.NewCoin(sdk.DefaultBondDenom, startTokens),
			staking.NewDescription(fmt.Sprintf("validator-%d", i+1), "", "", ""),
			staking.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
			sdk.OneInt(),
		)
		stdSignMsg := auth.StdSignMsg{
			ChainID: genDoc.ChainID,
			Msgs:    []sdk.Msg{msg},
		}
		var sig []byte
		sig, err = operPrivKey.Sign(stdSignMsg.Bytes())
		if err != nil {
			return
		}
		transaction := auth.NewStdTx([]sdk.Msg{msg}, auth.StdFee{}, []auth.StdSignature{{Signature: sig, PubKey: operPrivKey.PubKey()}}, "")
		genTxs = append(genTxs, transaction)
		valConsPubKeys = append(valConsPubKeys, pubKey)
		valOperAddrs = append(valOperAddrs, sdk.ValAddress(operAddr))

		accAuth := auth.NewBaseAccountWithAddress(sdk.AccAddress(operAddr))
		accTokens := sdk.TokensFromConsensusPower(150)
		totalSupply = totalSupply.Add(accTokens)

		accAuth.Coins = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, accTokens))
		accs = append(accs, genaccounts.NewGenesisAccount(&accAuth))
	}

	genesisState := simapp.NewDefaultGenesisState()
	genDoc.AppState, err = cdc.MarshalJSON(genesisState)
	if err != nil {
		return
	}

	genesisState, err = genutil.SetGenTxsInAppGenesisState(cdc, genesisState, genTxs)
	if err != nil {
		return
	}

	// add some tokens to init accounts
	stakingDataBz := genesisState[staking.ModuleName]
	var stakingData staking.GenesisState
	cdc.MustUnmarshalJSON(stakingDataBz, &stakingData)

	// add some tokens to init accounts
	for _, addr := range initAddrs {
		accAuth := auth.NewBaseAccountWithAddress(addr)
		accTokens := sdk.TokensFromConsensusPower(100)
		accAuth.Coins = sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, accTokens)}
		totalSupply = totalSupply.Add(accTokens)
		acc := genaccounts.NewGenesisAccount(&accAuth)
		accs = append(accs, acc)
	}

	// distr data
	distrDataBz := genesisState[distr.ModuleName]
	var distrData distr.GenesisState
	cdc.MustUnmarshalJSON(distrDataBz, &distrData)

	commPoolAmt := sdk.NewInt(10)
	distrData.FeePool.CommunityPool = sdk.DecCoins{sdk.NewDecCoin(sdk.DefaultBondDenom, commPoolAmt)}
	distrDataBz = cdc.MustMarshalJSON(distrData)
	genesisState[distr.ModuleName] = distrDataBz

	// staking and genesis accounts
	genesisState[staking.ModuleName] = cdc.MustMarshalJSON(stakingData)
	genesisState[genaccounts.ModuleName] = cdc.MustMarshalJSON(accs)

	// supply data
	supplyDataBz := genesisState[supply.ModuleName]
	var supplyData supply.GenesisState
	cdc.MustUnmarshalJSON(supplyDataBz, &supplyData)

	supplyData.Supply = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, totalSupply.Add(commPoolAmt)))
	supplyDataBz = cdc.MustMarshalJSON(supplyData)
	genesisState[supply.ModuleName] = supplyDataBz

	// mint genesis (none set within genesisState)
	mintData := mint.DefaultGenesisState()
	inflationMin := sdk.ZeroDec()
	if minting {
		inflationMin = sdk.MustNewDecFromStr("10000.0")
		mintData.Params.InflationMax = sdk.MustNewDecFromStr("15000.0")
	} else {
		mintData.Params.InflationMax = inflationMin
	}
	mintData.Minter.Inflation = inflationMin
	mintData.Params.InflationMin = inflationMin
	mintDataBz := cdc.MustMarshalJSON(mintData)
	genesisState[mint.ModuleName] = mintDataBz

	// initialize crisis data
	crisisDataBz := genesisState[crisis.ModuleName]
	var crisisData crisis.GenesisState
	cdc.MustUnmarshalJSON(crisisDataBz, &crisisData)
	crisisData.ConstantFee = sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)
	crisisDataBz = cdc.MustMarshalJSON(crisisData)
	genesisState[crisis.ModuleName] = crisisDataBz

	//// double check inflation is set according to the minting boolean flag
	if minting {
		if !(mintData.Params.InflationMax.Equal(sdk.MustNewDecFromStr("15000.0")) &&
			mintData.Minter.Inflation.Equal(sdk.MustNewDecFromStr("10000.0")) &&
			mintData.Params.InflationMin.Equal(sdk.MustNewDecFromStr("10000.0"))) {
			err = errors.New("Mint parameters does not correspond to their defaults")
			return
		}
	} else {
		if !(mintData.Params.InflationMax.Equal(sdk.ZeroDec()) &&
			mintData.Minter.Inflation.Equal(sdk.ZeroDec()) &&
			mintData.Params.InflationMin.Equal(sdk.ZeroDec())) {
			err = errors.New("Mint parameters not equal to decimal 0")
			return
		}
	}

	appState, err := codec.MarshalJSONIndent(cdc, genesisState)
	if err != nil {
		return
	}
	genDoc.AppState = appState
	return
}

// startTM creates and starts an in-process Tendermint node with memDB and
// in-process ABCI application. It returns the new node or any error that
// occurred.
//
// TODO: Clean up the WAL dir or enable it to be not persistent!
func startTM(
	tmcfg *tmcfg.Config, logger log.Logger, genDoc *tmtypes.GenesisDoc,
	privVal tmtypes.PrivValidator, app *gapp.GaiaApp,
) (*nm.Node, error) {

	genDocProvider := func() (*tmtypes.GenesisDoc, error) { return genDoc, nil }
	dbProvider := func(*nm.DBContext) (dbm.DB, error) { return dbm.NewMemDB(), nil }
	nodeKey, err := p2p.LoadOrGenNodeKey(tmcfg.NodeKeyFile())
	if err != nil {
		return nil, err
	}
	node, err := nm.NewNode(
		tmcfg,
		privVal,
		nodeKey,
		proxy.NewLocalClientCreator(app),
		genDocProvider,
		dbProvider,
		nm.DefaultMetricsProvider(tmcfg.Instrumentation),
		logger.With("module", "node"),
	)
	if err != nil {
		return nil, err
	}

	err = node.Start()
	if err != nil {
		return nil, err
	}

	tests.WaitForRPC(tmcfg.RPC.ListenAddress)
	logger.Info("Tendermint running!")

	return node, err
}

// startLCD starts the LCD.
func startLCD(logger log.Logger, listenAddr string, cdc *codec.Codec) (net.Listener, error) {
	rs := lcd.NewRestServer(cdc)
	registerRoutes(rs)
	listener, err := tmrpc.Listen(listenAddr, tmrpc.DefaultConfig())
	if err != nil {
		return nil, err
	}
	go tmrpc.StartHTTPServer(listener, rs.Mux, logger, tmrpc.DefaultConfig()) //nolint:errcheck
	return listener, nil
}

// NOTE: If making updates here also update cmd/gaia/cmd/gaiacli/main.go
func registerRoutes(rs *lcd.RestServer) {
	client.RegisterRoutes(rs.CliCtx, rs.Mux)
	authrest.RegisterTxRoutes(rs.CliCtx, rs.Mux)
	gapp.ModuleBasics.RegisterRESTRoutes(rs.CliCtx, rs.Mux)
}

var cdc = amino.NewCodec()

func init() {
	ctypes.RegisterAmino(cdc)
}

// CreateAddr adds an address to the key store and returns an address and seed.
// It also requires that the key could be created.
func CreateAddr(name, password string, kb crkeys.Keybase) (sdk.AccAddress, string, error) {
	var (
		err  error
		info crkeys.Info
		seed string
	)
	info, seed, err = kb.CreateMnemonic(name, crkeys.English, password, crkeys.Secp256k1)
	return sdk.AccAddress(info.GetPubKey().Address()), seed, err
}

// CreateAddr adds multiple address to the key store and returns the addresses and associated seeds in lexographical order by address.
// It also requires that the keys could be created.
func CreateAddrs(kb crkeys.Keybase, numAddrs int) (addrs []sdk.AccAddress, seeds, names, passwords []string, errs []error) {
	var (
		err  error
		info crkeys.Info
		seed string
	)

	addrSeeds := AddrSeedSlice{}

	for i := 0; i < numAddrs; i++ {
		name := fmt.Sprintf("test%d", i)
		password := "1234567890"
		info, seed, err = kb.CreateMnemonic(name, crkeys.English, password, crkeys.Secp256k1)
		if err != nil {
			errs = append(errs, err)
		}
		addrSeeds = append(addrSeeds, AddrSeed{Address: sdk.AccAddress(info.GetPubKey().Address()), Seed: seed, Name: name, Password: password})
	}
	if len(errs) > 0 {
		return
	}

	sort.Sort(addrSeeds)

	for i := range addrSeeds {
		addrs = append(addrs, addrSeeds[i].Address)
		seeds = append(seeds, addrSeeds[i].Seed)
		names = append(names, addrSeeds[i].Name)
		passwords = append(passwords, addrSeeds[i].Password)
	}

	return
}

// AddrSeed combines an Address with the mnemonic of the private key to that address
type AddrSeed struct {
	Address  sdk.AccAddress
	Seed     string
	Name     string
	Password string
}

// AddrSeedSlice implements `Interface` in sort package.
type AddrSeedSlice []AddrSeed

func (b AddrSeedSlice) Len() int {
	return len(b)
}

// Less sorts lexicographically by Address
func (b AddrSeedSlice) Less(i, j int) bool {
	// bytes package already implements Comparable for []byte.
	switch bytes.Compare(b[i].Address.Bytes(), b[j].Address.Bytes()) {
	case -1:
		return true
	case 0, 1:
		return false
	default:
		panic("not fail-able with `bytes.Comparable` bounded [-1, 1].")
	}
}

func (b AddrSeedSlice) Swap(i, j int) {
	b[j], b[i] = b[i], b[j]
}

// InitClientHome initialises client home dir.
func InitClientHome(dir string) string {
	var err error
	if dir == "" {
		dir, err = ioutil.TempDir("", "lcd_test")
		if err != nil {
			panic(err)
		}
	}
	// TODO: this should be set in NewRestServer
	// and pass down the CLIContext to achieve
	// parallelism.
	viper.Set(cli.HomeFlag, dir)
	return dir
}

// makePathname creates a unique pathname for each test.
func makePathname() (string, error) {
	p, err := os.Getwd()
	if err != nil {
		return "", err
	}

	sep := string(filepath.Separator)
	return strings.Replace(p, sep, "_", -1), nil
}

// GetConfig returns a Tendermint config for the test cases.
func GetConfig() (*tmcfg.Config, error) {
	pathname, err := makePathname()
	if err != nil {
		return nil, err
	}
	config := tmcfg.ResetTestRoot(pathname)

	tmAddr, _, err := server.FreeTCPAddr()
	if err != nil {
		return nil, err
	}

	rcpAddr, _, err := server.FreeTCPAddr()
	if err != nil {
		return nil, err
	}

	grpcAddr, _, err := server.FreeTCPAddr()
	if err != nil {
		return nil, err
	}

	config.P2P.ListenAddress = tmAddr
	config.RPC.ListenAddress = rcpAddr
	config.RPC.GRPCListenAddress = grpcAddr

	return config, nil
}
