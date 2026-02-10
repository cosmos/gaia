package gaia_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	gaiahelpers "github.com/cosmos/gaia/v26/app/helpers"
)

func TestMultiSendGas(t *testing.T) {
	app := gaiahelpers.Setup(t)
	ctx := app.NewUncachedContext(false, tmproto.Header{Time: time.Now()})

	acc1 := sdk.AccAddress([]byte("addr1_______________"))

	bk := app.BankKeeper
	err := bk.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("uatom", 1000000000)))
	require.NoError(t, err)
	err = bk.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, acc1, sdk.NewCoins(sdk.NewInt64Coin("uatom", 1000000000)))
	require.NoError(t, err)

	runMeasure := func(n int) {
		input := banktypes.Input{Address: acc1.String(), Coins: sdk.NewCoins(sdk.NewInt64Coin("uatom", int64(n*10)))}
		outputs := make([]banktypes.Output, n)
		for i := 0; i < n; i++ {
			addr := sdk.AccAddress([]byte(fmt.Sprintf("dest%d________________", i)))
			outputs[i] = banktypes.Output{
				Address: addr.String(),
				Coins:   sdk.NewCoins(sdk.NewInt64Coin("uatom", 10)),
			}
		}

		// Use struct literal to bypass constructor issues
		msg := &banktypes.MsgMultiSend{
			Inputs:  []banktypes.Input{input},
			Outputs: outputs,
		}

		handler := app.MsgServiceRouter().Handler(msg)
		require.NotNil(t, handler, "Handler for MsgMultiSend not found")

		ctxTest := ctx.WithGasMeter(storetypes.NewGasMeter(100000000))
		_, err := handler(ctxTest, msg)
		require.NoError(t, err)

		fmt.Printf(">>> Gas for %d Recipients MultiSend: %d\n", n, ctxTest.GasMeter().GasConsumed())
	}

	runMeasure(1)
	runMeasure(10)
	runMeasure(50)
	runMeasure(100)
}
