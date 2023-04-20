package liquidity_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	lapp "github.com/cosmos/gaia/v9/app"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	app := lapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	app.InitChain(
		abcitypes.RequestInitChain{
			AppStateBytes: []byte("{}"),
			ChainId:       "test-chain-id",
		},
	)
	params := app.LiquidityKeeper.GetParams(ctx)
	require.NotNil(t, params)
}
