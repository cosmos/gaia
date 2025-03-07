package v23_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"

	"github.com/cosmos/gaia/v23/app/helpers"
	v23 "github.com/cosmos/gaia/v23/app/upgrades/v23"
)

const (
	clientStateJSON    = `{"@type":"/ibc.lightclients.wasm.v1.ClientState","data":"eyJjaGFpbl9pZCI6MzE1MTkwOCwiZXBvY2hzX3Blcl9zeW5jX2NvbW1pdHRlZV9wZXJpb2QiOjgsImZvcmtfcGFyYW1ldGVycyI6eyJhbHRhaXIiOnsiZXBvY2giOjAsInZlcnNpb24iOiIyMDAwMDAzOCJ9LCJiZWxsYXRyaXgiOnsiZXBvY2giOjAsInZlcnNpb24iOiIzMDAwMDAzOCJ9LCJjYXBlbGxhIjp7ImVwb2NoIjowLCJ2ZXJzaW9uIjoiNDAwMDAwMzgifSwiZGVuZWIiOnsiZXBvY2giOjAsInZlcnNpb24iOiI1MDAwMDAzOCJ9LCJnZW5lc2lzX2ZvcmtfdmVyc2lvbiI6IjEwMDAwMDM4IiwiZ2VuZXNpc19zbG90IjowfSwiZ2VuZXNpc190aW1lIjoxNzQwODI3NDA3LCJnZW5lc2lzX3ZhbGlkYXRvcnNfcm9vdCI6ImQ2MWVhNDg0ZmViYWNmYWU1Mjk4ZDUyYTJiNTgxZjNlMzA1YTUxZjMxMTJhOTI0MWI5NjhkY2NmMDE5ZjdiMTEiLCJpYmNfY29tbWl0bWVudF9zbG90IjoiMHgxMjYwOTQ0NDg5MjcyOTg4ZDlkZjI4NTE0OWI1YWExYjBmNDhmMjEzNmQ2ZjQxNjE1OWY4NDBhM2UwNzQ3NjAwIiwiaWJjX2NvbnRyYWN0X2FkZHJlc3MiOiIweDI5M2MxOGUwOWU1NTA0ZWJjZTE3ZGFhN2YzZDY2MmM4YjliZjZkNzUiLCJpc19mcm96ZW4iOmZhbHNlLCJsYXRlc3Rfc2xvdCI6MzIsIm1pbl9zeW5jX2NvbW1pdHRlZV9wYXJ0aWNpcGFudHMiOjMyLCJzZWNvbmRzX3Blcl9zbG90Ijo2LCJzbG90c19wZXJfZXBvY2giOjh9","checksum":"+CVJ9byK2u8Y5c5PW2gmmUc0N0LJONrDIvrxWDMZFyw=","latest_height":{"revision_number":"0","revision_height":"32"}}`
	consensusStateJSON = `{"@type":"/ibc.lightclients.wasm.v1.ConsensusState","data":"eyJjdXJyZW50X3N5bmNfY29tbWl0dGVlIjoiMHg4MTQ1ZjkyZDEzMTcxNTNlNmJiODFhOTAzZGI2ZjBiMTBjNTBhOTM2ODNmNGMyYWFiMzVlOWE4YTRiYjI4MzMyMzQyNWZiNTZhNDdkOGIzMGE5ZWZkNTA5YzhhZjE0ZTEiLCJuZXh0X3N5bmNfY29tbWl0dGVlIjoiMHg4MTQ1ZjkyZDEzMTcxNTNlNmJiODFhOTAzZGI2ZjBiMTBjNTBhOTM2ODNmNGMyYWFiMzVlOWE4YTRiYjI4MzMyMzQyNWZiNTZhNDdkOGIzMGE5ZWZkNTA5YzhhZjE0ZTEiLCJzbG90IjozMiwic3RhdGVfcm9vdCI6IjB4MjI4NTQzYWVlYzk0NjA5YjQwOGUyNzI0NjIzZjgyMGExNjFhYmY5OWRkODMyNzQ4MWQ1NGNmYmUyNzAyOTE1ZCIsInN0b3JhZ2Vfcm9vdCI6IjB4MDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMCIsInRpbWVzdGFtcCI6MTc0MDgyNzU5OX0="}`
)

func TestAddEthLightWasmLightClient(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{
		Time: time.Unix(1740829624, 0),
	})

	err := v23.AddEthLightWasmLightClient(ctx, gaiaApp.WasmClientKeeper)
	require.NoError(t, err)

	// check that the checksum is as expected
	queryChecksumsResp, err := gaiaApp.WasmClientKeeper.Checksums(ctx, &ibcwasmtypes.QueryChecksumsRequest{})
	require.NoError(t, err)
	require.Len(t, queryChecksumsResp.Checksums, 1)
	require.Equal(t, v23.ExpectedEthLightClientChecksum, queryChecksumsResp.Checksums[0])

	var clientState ibcexported.ClientState
	err = gaiaApp.AppCodec().UnmarshalInterfaceJSON([]byte(clientStateJSON), &clientState)
	require.NoError(t, err)

	var consensusState ibcexported.ConsensusState
	err = gaiaApp.AppCodec().UnmarshalInterfaceJSON([]byte(consensusStateJSON), &consensusState)
	require.NoError(t, err)

	createMsg, err := clienttypes.NewMsgCreateClient(clientState, consensusState, "")
	require.NoError(t, err)

	// create a light client
	createClientResp, err := gaiaApp.IBCKeeper.CreateClient(ctx, createMsg)
	require.NoError(t, err)
	require.Equal(t, "08-wasm-0", createClientResp.ClientId)

	// Make a call into the actual light client to verify we can call the light client contract
	timestamp, err := gaiaApp.IBCKeeper.ClientKeeper.GetClientTimestampAtHeight(ctx, "08-wasm-0", clienttypes.NewHeight(0, 32))
	require.NoError(t, err)
	require.Equal(t, uint64(1740827599000000000), timestamp)
}

func TestGrantIBCWasmAuth(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{
		Time: time.Unix(1740829624, 0),
	})

	err := v23.AuthzGrantWasmLightClient(ctx, gaiaApp.AuthzKeeper, *gaiaApp.GovKeeper)
	require.NoError(t, err)

	auth, _ := gaiaApp.AuthzKeeper.GetAuthorization(
		ctx, sdk.AccAddress(v23.ClientUploaderAddress),
		sdk.AccAddress(gaiaApp.GovKeeper.GetAuthority()),
		v23.IBCWasmStoreCodeTypeURL)
	require.NotNil(t, auth)
}
