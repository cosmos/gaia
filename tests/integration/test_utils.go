package integration

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/stretchr/testify/require"

	tmdb "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"

	gaiaApp "github.com/cosmos/gaia/v18/app"
)

var app *gaiaApp.GaiaApp

// GaiaAppIniter implements ibctesting.AppIniter for the gaia app
func GaiaAppIniter() (ibctesting.TestingApp, map[string]json.RawMessage) {
	encoding := gaiaApp.RegisterEncodingConfig()
	app = gaiaApp.NewGaiaApp(
		log.NewNopLogger(),
		tmdb.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		gaiaApp.DefaultNodeHome,
		encoding,
		gaiaApp.EmptyAppOptions{},
		gaiaApp.EmptyWasmOptions)

	testApp := ibctesting.TestingApp(app)

	return testApp, gaiaApp.NewDefaultGenesisState(encoding)
}

// SendMsgs() behavior must be changed since the default one uses zero fees
func OverrideSendMsgs(chains map[string]*ibctesting.TestChain, feeAmount sdk.Coin, gasLimit uint64) {
	for _, chain := range chains {
		chain := chain
		chain.SendMsgsOverride = func(msgs ...sdk.Msg) (*sdk.Result, error) {
			return SendMsgsOverride(chain, feeAmount, gasLimit, msgs...)
		}
	}
}

func SendMsgsOverride(chain *ibctesting.TestChain, feeAmount sdk.Coin, gasLimit uint64, msgs ...sdk.Msg) (*sdk.Result, error) {
	// ensure the chain has the latest time
	chain.Coordinator.UpdateTimeForChain(chain)

	_, r, err := SignAndDeliver(
		chain,
		chain.TxConfig,
		chain.App.GetBaseApp(),
		chain.GetContext().BlockHeader(),
		msgs,
		chain.ChainID,
		[]uint64{chain.SenderAccount.GetAccountNumber()},
		[]uint64{chain.SenderAccount.GetSequence()},
		true, true,
		feeAmount, gasLimit,
		chain.SenderPrivKey,
	)
	if err != nil {
		return nil, err
	}

	// NextBlock calls app.Commit()
	chain.NextBlock()

	// increment sequence for successful transaction execution
	err = chain.SenderAccount.SetSequence(chain.SenderAccount.GetSequence() + 1)
	if err != nil {
		return nil, err
	}

	chain.Coordinator.IncrementTime()

	return r, nil
}

func SignAndDeliver(
	chain *ibctesting.TestChain, txCfg client.TxConfig, app *baseapp.BaseApp, header tmproto.Header, msgs []sdk.Msg,
	chainID string, accNums, accSeqs []uint64, expSimPass, expPass bool, feeAmount sdk.Coin, gasLimit uint64, priv ...cryptotypes.PrivKey,
) (sdk.GasInfo, *sdk.Result, error) {
	tx, err := simtestutil.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		txCfg,
		msgs,
		sdk.Coins{feeAmount},
		gasLimit,
		chainID,
		accNums,
		accSeqs,
		priv...,
	)
	require.NoError(chain.T, err)

	// Simulate a sending a transaction
	gInfo, res, err := app.SimDeliver(txCfg.TxEncoder(), tx)

	if expPass {
		require.NoError(chain.T, err)
		require.NotNil(chain.T, res)
	} else {
		require.Error(chain.T, err)
		require.Nil(chain.T, res)
	}

	return gInfo, res, err
}
