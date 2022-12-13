package v8_test

import (
	"fmt"
	"testing"

	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	gaiahelpers "github.com/cosmos/gaia/v8/app/helpers"
	"github.com/stretchr/testify/require"
)

func TestFixBankMetadata(t *testing.T) {
	app := gaiahelpers.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
		Height:  1,
	})
	cdc := app.AppCodec()

	v7Key := "uatomuatom"
	denomMetaData := banktypes.Metadata{
		Name:        "foo",
		Symbol:      "bar",
		Description: "The native staking token of the Cosmos Hub.",
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: "uatom", Exponent: uint32(0), Aliases: []string{"microatom"}},
			{Denom: "matom", Exponent: uint32(3), Aliases: []string{"milliatom"}},
			{Denom: "atom", Exponent: uint32(6), Aliases: nil},
		},
		Base:    "uatom",
		Display: "atom",
	}

	// add the old format
	key := app.AppKeepers.GetKey(banktypes.ModuleName)
	store := ctx.KVStore(key)
	oldDenomMetaDataStore := prefix.NewStore(store, banktypes.DenomMetadataPrefix)
	m := cdc.MustMarshal(&denomMetaData)
	oldDenomMetaDataStore.Set([]byte(v7Key), m)

	rhoUpgrade := upgradetypes.Plan{
		Name:   "v8-Rho",
		Info:   "some text here",
		Height: 100,
	}
	app.AppKeepers.UpgradeKeeper.ApplyUpgrade(ctx, rhoUpgrade)

	correctDenom := "uatom"

	metadata, foundCorrect := app.AppKeepers.BankKeeper.GetDenomMetaData(ctx, correctDenom)
	require.True(t, foundCorrect)

	require.Equal(t, metadata.Name, "Cosmos Hub Atom")
	require.Equal(t, metadata.Symbol, "ATOM")

	malformedDenom := "uatomu"
	_, foundMalformed := app.AppKeepers.BankKeeper.GetDenomMetaData(ctx, malformedDenom)
	require.False(t, foundMalformed)

}
