package common

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"

	tmrand "github.com/cometbft/cometbft/libs/rand"

	dbm "github.com/cosmos/cosmos-db"
	ratelimittypes "github.com/cosmos/ibc-apps/modules/rate-limiting/v10/types"
	wasmclienttypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10/types"
	providertypes "github.com/cosmos/interchain-security/v7/x/ccv/provider/types"

	"cosmossdk.io/log"
	evidencetypes "cosmossdk.io/x/evidence/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distribtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	paramsproptypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	gaia "github.com/cosmos/gaia/v23/app"
	gaiaparams "github.com/cosmos/gaia/v23/app/params"
	metaprotocoltypes "github.com/cosmos/gaia/v23/x/metaprotocols/types"
)

const (
	keyringPassphrase = "testpassphrase"
	KeyringAppName    = "testnet"
)

var (
	EncodingConfig gaiaparams.EncodingConfig
	Cdc            codec.Codec
	TxConfig       client.TxConfig
)

func init() {
	EncodingConfig = gaiaparams.MakeEncodingConfig()
	banktypes.RegisterInterfaces(EncodingConfig.InterfaceRegistry)
	authtypes.RegisterInterfaces(EncodingConfig.InterfaceRegistry)
	authvesting.RegisterInterfaces(EncodingConfig.InterfaceRegistry)
	stakingtypes.RegisterInterfaces(EncodingConfig.InterfaceRegistry)
	evidencetypes.RegisterInterfaces(EncodingConfig.InterfaceRegistry)
	cryptocodec.RegisterInterfaces(EncodingConfig.InterfaceRegistry)
	govv1types.RegisterInterfaces(EncodingConfig.InterfaceRegistry)
	govv1beta1types.RegisterInterfaces(EncodingConfig.InterfaceRegistry)
	paramsproptypes.RegisterInterfaces(EncodingConfig.InterfaceRegistry)
	paramsproptypes.RegisterLegacyAminoCodec(EncodingConfig.Amino)

	upgradetypes.RegisterInterfaces(EncodingConfig.InterfaceRegistry)
	distribtypes.RegisterInterfaces(EncodingConfig.InterfaceRegistry)
	providertypes.RegisterInterfaces(EncodingConfig.InterfaceRegistry)
	metaprotocoltypes.RegisterInterfaces(EncodingConfig.InterfaceRegistry)
	ratelimittypes.RegisterInterfaces(EncodingConfig.InterfaceRegistry)
	wasmclienttypes.RegisterInterfaces(EncodingConfig.InterfaceRegistry)

	Cdc = EncodingConfig.Marshaler
	TxConfig = EncodingConfig.TxConfig
}

type Chain struct {
	DataDir    string
	ID         string
	Validators []*validator
	accounts   []*account //nolint:unused
	// initial accounts in genesis
	GenesisAccounts        []*account
	GenesisVestingAccounts map[string]sdk.AccAddress
}

func NewChain() (*Chain, error) {
	tmpDir, err := os.MkdirTemp("", "gaia-e2e-testnet-")
	if err != nil {
		return nil, err
	}

	return &Chain{
		ID:      "chain-" + tmrand.Str(6),
		DataDir: tmpDir,
	}, nil
}

func (c *Chain) configDir() string {
	return fmt.Sprintf("%s/%s", c.DataDir, c.ID)
}

func (c *Chain) CreateAndInitValidators(count int) error {
	// create a separate app dir for the tempApp so that wasmvm won't complain about file locks
	tempAppDir := filepath.Join(gaia.DefaultNodeHome, strconv.Itoa(rand.Intn(10000)))
	tempApplication := gaia.NewGaiaApp(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		tempAppDir,
		gaia.EmptyAppOptions{},
		gaia.EmptyWasmOptions,
	)
	defer func() {
		if err := tempApplication.Close(); err != nil {
			panic(err)
		}
	}()

	genesisState := tempApplication.ModuleBasics.DefaultGenesis(EncodingConfig.Marshaler)

	for i := 0; i < count; i++ {
		node := c.createValidator(i)

		// generate genesis files
		if err := node.init(genesisState); err != nil {
			return err
		}

		c.Validators = append(c.Validators, node)

		// create keys
		if err := node.createKey("val"); err != nil {
			return err
		}
		if err := node.createNodeKey(); err != nil {
			return err
		}
		if err := node.createConsensusKey(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Chain) createAndInitValidatorsWithMnemonics(count int, mnemonics []string) error { //nolint:unused // this is called during e2e tests
	tempApplication := gaia.NewGaiaApp(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		gaia.DefaultNodeHome,
		gaia.EmptyAppOptions{},
		gaia.EmptyWasmOptions,
	)
	defer func() {
		if err := tempApplication.Close(); err != nil {
			panic(err)
		}
	}()

	genesisState := tempApplication.ModuleBasics.DefaultGenesis(EncodingConfig.Marshaler)

	for i := 0; i < count; i++ {
		// create node
		node := c.createValidator(i)

		// generate genesis files
		if err := node.init(genesisState); err != nil {
			return err
		}

		c.Validators = append(c.Validators, node)

		// create keys
		if err := node.createKeyFromMnemonic("val", mnemonics[i]); err != nil {
			return err
		}
		if err := node.createNodeKey(); err != nil {
			return err
		}
		if err := node.createConsensusKey(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Chain) createValidator(index int) *validator {
	return &validator{
		chain:   c,
		Index:   index,
		Moniker: fmt.Sprintf("%s-gaia-%d", c.ID, index),
	}
}
