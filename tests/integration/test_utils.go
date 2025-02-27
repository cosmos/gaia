package integration

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"

	dbm "github.com/cosmos/cosmos-db"
	ibctesting "github.com/cosmos/ibc-go/v10/testing"

	"cosmossdk.io/log"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"

	gaiaApp "github.com/cosmos/gaia/v23/app"
)

var app *gaiaApp.GaiaApp

// Some tests require a random directory to be created when running IBC testing suite with gaia.
// This is due to how CosmWasmVM initializes the VM - all IBC testing apps must have different dirs so they don't conflict.
func GaiaAppIniterTempDir() (ibctesting.TestingApp, map[string]json.RawMessage) {
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		panic(err)
	}
	app = gaiaApp.NewGaiaApp(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		tmpDir,
		gaiaApp.EmptyAppOptions{},
		gaiaApp.EmptyWasmOptions)

	testApp := ibctesting.TestingApp(app)

	return testApp, app.ModuleBasics.DefaultGenesis(app.AppCodec())
}

// GaiaAppIniter implements ibctesting.AppIniter for the gaia app
func GaiaAppIniter() (ibctesting.TestingApp, map[string]json.RawMessage) {
	app = gaiaApp.NewGaiaApp(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		gaiaApp.DefaultNodeHome,
		gaiaApp.EmptyAppOptions{},
		gaiaApp.EmptyWasmOptions)

	testApp := ibctesting.TestingApp(app)

	return testApp, app.ModuleBasics.DefaultGenesis(app.AppCodec())
}

// SendMsgs() behavior must be changed since the default one uses zero fees
func OverrideSendMsgs(chains map[string]*ibctesting.TestChain, feeAmount sdk.Coin, gasLimit uint64) {
	for _, chain := range chains {
		chain := chain
		chain.SendMsgsOverride = func(msgs ...sdk.Msg) (*abci.ExecTxResult, error) {
			return SendMsgsOverride(chain, feeAmount, gasLimit, msgs...)
		}
	}
}

func SendMsgsOverride(chain *ibctesting.TestChain, feeAmount sdk.Coin, gasLimit uint64, msgs ...sdk.Msg) (*abci.ExecTxResult, error) {
	// ensure the chain has the latest time
	chain.Coordinator.UpdateTimeForChain(chain)

	// increment acc sequence regardless of success or failure tx execution
	defer func() {
		err := chain.SenderAccount.SetSequence(chain.SenderAccount.GetSequence() + 1)
		if err != nil {
			panic(err)
		}
	}()

	resp, err := SignAndDeliver(
		chain.TB,
		chain.TxConfig,
		chain.App.GetBaseApp(),
		msgs,
		chain.ChainID,
		[]uint64{chain.SenderAccount.GetAccountNumber()},
		[]uint64{chain.SenderAccount.GetSequence()},
		true,
		chain.ProposedHeader.GetTime(),
		chain.NextVals.Hash(),
		feeAmount,
		gasLimit,
		chain.SenderPrivKey,
	)
	if err != nil {
		return nil, err
	}

	require.Len(chain.TB, resp.TxResults, 1)
	txResult := resp.TxResults[0]

	if txResult.Code != 0 {
		return txResult, fmt.Errorf("%s/%d: %q", txResult.Codespace, txResult.Code, txResult.Log)
	}

	chain.Coordinator.IncrementTime()

	return txResult, nil
}

func SignAndDeliver(
	tb testing.TB, txCfg client.TxConfig, app *bam.BaseApp, msgs []sdk.Msg,
	chainID string, accNums, accSeqs []uint64, expPass bool, blockTime time.Time, nextValHash []byte, feeAmount sdk.Coin, gasLimit uint64, priv ...cryptotypes.PrivKey,
) (*abci.ResponseFinalizeBlock, error) {
	tb.Helper()
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
	require.NoError(tb, err)

	txBytes, err := txCfg.TxEncoder()(tx)
	require.NoError(tb, err)

	return app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             app.LastBlockHeight() + 1,
		Time:               blockTime,
		NextValidatorsHash: nextValHash,
		Txs:                [][]byte{txBytes},
	})
}
