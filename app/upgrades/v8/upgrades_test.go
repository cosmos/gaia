package v8_test

import (
	"fmt"
	"testing"

	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	gaiahelpers "github.com/cosmos/gaia/v8/app/helpers"
	v8 "github.com/cosmos/gaia/v8/app/upgrades/v8"
	"github.com/stretchr/testify/require"
)

func TestFixBankMetadata(t *testing.T) {
	app := gaiahelpers.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
		Height:  1,
	})

	cdc := app.AppCodec()

	malformedDebom := "uatomu"
	denomMetaData := banktypes.Metadata{
		Name:        "Cosmos Hub Atom",
		Symbol:      "ATOM",
		Description: "The native staking token of the Cosmos Hub.",
		DenomUnits: []*banktypes.DenomUnit{
			{"uatom", uint32(0), []string{"microatom"}},
			{"matom", uint32(3), []string{"milliatom"}},
			{"atom", uint32(6), nil},
		},
		Base:    "uatom",
		Display: "atom",
	}

	// add the old format
	key := app.AppKeepers.GetKey(banktypes.ModuleName)
	store := ctx.KVStore(key)
	oldDenomMetaDataStore := prefix.NewStore(store, banktypes.DenomMetadataPrefix)
	m := cdc.MustMarshal(&denomMetaData)
	oldDenomMetaDataStore.Set([]byte(malformedDebom), m)

	correctDenom := "uatom"

	_, foundCorrect := app.AppKeepers.BankKeeper.GetDenomMetaData(ctx, correctDenom)
	require.False(t, foundCorrect)

	err := v8.FixBankMetadata(ctx, &app.AppKeepers)
	require.NoError(t, err)

	_, foundCorrect = app.AppKeepers.BankKeeper.GetDenomMetaData(ctx, correctDenom)
	require.True(t, foundCorrect)

}
