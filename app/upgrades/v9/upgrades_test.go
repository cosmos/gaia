package v9_test

import (
	"fmt"
	"testing"

	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	gaiahelpers "github.com/cosmos/gaia/v9/app/helpers"
)

func TestLambdaUpgrade(t *testing.T) {
	app := gaiahelpers.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
		Height:  1,
	})

	rhoUpgrade := upgradetypes.Plan{
		Name:   "v9-Lambda",
		Info:   "ICS Upgrade",
		Height: 100,
	}
	app.AppKeepers.UpgradeKeeper.ApplyUpgrade(ctx, rhoUpgrade)

}
