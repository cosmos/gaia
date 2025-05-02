package v23_2_0_test

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	_ "embed"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"

	"github.com/cosmos/gaia/v24/app/helpers"
	"github.com/cosmos/gaia/v24/app/upgrades/v23_2_0"
)

const (
	clientStateJSON    = `{"@type":"/ibc.lightclients.wasm.v1.ClientState","data":"eyJjaGFpbl9pZCI6MzE1MTkwOCwiZXBvY2hzX3Blcl9zeW5jX2NvbW1pdHRlZV9wZXJpb2QiOjgsImZvcmtfcGFyYW1ldGVycyI6eyJhbHRhaXIiOnsiZXBvY2giOjAsInZlcnNpb24iOiIweDIwMDAwMDM4In0sImJlbGxhdHJpeCI6eyJlcG9jaCI6MCwidmVyc2lvbiI6IjB4MzAwMDAwMzgifSwiY2FwZWxsYSI6eyJlcG9jaCI6MCwidmVyc2lvbiI6IjB4NDAwMDAwMzgifSwiZGVuZWIiOnsiZXBvY2giOjAsInZlcnNpb24iOiIweDUwMDAwMDM4In0sImVsZWN0cmEiOnsiZXBvY2giOjEsInZlcnNpb24iOiIweDYwMDAwMDM4In0sImdlbmVzaXNfZm9ya192ZXJzaW9uIjoiMHgxMDAwMDAzOCIsImdlbmVzaXNfc2xvdCI6MH0sImdlbmVzaXNfc2xvdCI6MCwiZ2VuZXNpc190aW1lIjoxNzQ0MjI4NzE4LCJnZW5lc2lzX3ZhbGlkYXRvcnNfcm9vdCI6IjB4ZDYxZWE0ODRmZWJhY2ZhZTUyOThkNTJhMmI1ODFmM2UzMDVhNTFmMzExMmE5MjQxYjk2OGRjY2YwMTlmN2IxMSIsImliY19jb21taXRtZW50X3Nsb3QiOiIweDEyNjA5NDQ0ODkyNzI5ODhkOWRmMjg1MTQ5YjVhYTFiMGY0OGYyMTM2ZDZmNDE2MTU5Zjg0MGEzZTA3NDc2MDAiLCJpYmNfY29udHJhY3RfYWRkcmVzcyI6IjB4ODk4MTRmYWU3NjBlYjViNzM5MmNiZGNiYTc1YzM1NWM0NjdiMjFkOSIsImlzX2Zyb3plbiI6ZmFsc2UsImxhdGVzdF9leGVjdXRpb25fYmxvY2tfbnVtYmVyIjozMiwibGF0ZXN0X3Nsb3QiOjMyLCJtaW5fc3luY19jb21taXR0ZWVfcGFydGljaXBhbnRzIjoxLCJzZWNvbmRzX3Blcl9zbG90Ijo2LCJzbG90c19wZXJfZXBvY2giOjgsInN5bmNfY29tbWl0dGVlX3NpemUiOjMyfQ==","checksum":"O+U+9qV98BMWnk+WmX42xg+vqqMv9jH7Brxvh7fZCW8=","latest_height":{"revision_number":"0","revision_height":"32"}}`
	consensusStateJSON = `{"@type":"/ibc.lightclients.wasm.v1.ConsensusState","data":"eyJjdXJyZW50X3N5bmNfY29tbWl0dGVlIjp7ImFnZ3JlZ2F0ZV9wdWJrZXkiOiIweDg2MTMzNzg1Y2M1ZDIzOGFkN2I4YmI1Y2M1M2ViZmJlMzQ2NGNiYTliNjBlNGRhMTBkZTIwN2U4OGQzOTUwNTQ5ZGFkYTFjMzRmZTA1ZmVhMDI2MzFkOWM2ZGEyMTI2OSIsInB1YmtleXNfaGFzaCI6IjB4OTVhZTQyZmZlNzI5NGVlYTYxMDE1NzA2YzIzN2Q4NzAxNWVhOWE2NTZmYTUzZTk0NDJlZTBiM2Q1N2Q4NzczNiJ9LCJuZXh0X3N5bmNfY29tbWl0dGVlIjp7ImFnZ3JlZ2F0ZV9wdWJrZXkiOiIweDg2MTMzNzg1Y2M1ZDIzOGFkN2I4YmI1Y2M1M2ViZmJlMzQ2NGNiYTliNjBlNGRhMTBkZTIwN2U4OGQzOTUwNTQ5ZGFkYTFjMzRmZTA1ZmVhMDI2MzFkOWM2ZGEyMTI2OSIsInB1YmtleXNfaGFzaCI6IjB4OTVhZTQyZmZlNzI5NGVlYTYxMDE1NzA2YzIzN2Q4NzAxNWVhOWE2NTZmYTUzZTk0NDJlZTBiM2Q1N2Q4NzczNiJ9LCJzbG90IjozMiwic3RhdGVfcm9vdCI6IjB4YjgwNGMzMzkzNDJlY2FjYTAzMDQ1ZmM3YjgyYTBjNTE0ZTQ1ZWY5M2YyNjlkOThkZmUxMTZkZjhlODFiMTI3NCIsInN0b3JhZ2Vfcm9vdCI6IjB4MDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMCIsInRpbWVzdGFtcCI6MTc0NDIzMDUxMH0="}`
)

var (
	//go:embed cw_ics08_wasm_eth-1.1.0.wasm.gz
	codeV1 []byte
	//go:embed cw_ics08_wasm_eth-1.2.0.wasm.gz
	codeV2 []byte
)

func TestMigrateIBCWasm(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{
		Time: time.Unix(1740829624, 0),
	})

	// Setup IBC Client
	checksumV1, err := gaiaApp.WasmClientKeeper.StoreCode(ctx, types.NewMsgStoreCode(v23_2_0.SignerAccount, codeV1))
	require.NoError(t, err)
	require.NotNil(t, checksumV1)
	hexV1Checksum := hex.EncodeToString(checksumV1.Checksum)

	queryChecksumsResp, err := gaiaApp.WasmClientKeeper.Checksums(ctx, &types.QueryChecksumsRequest{})
	require.NoError(t, err)
	require.Len(t, queryChecksumsResp.Checksums, 1)
	require.Equal(t, hexV1Checksum, queryChecksumsResp.Checksums[0])
	_, err = gaiaApp.WasmClientKeeper.StoreCode(ctx, types.NewMsgStoreCode(v23_2_0.SignerAccount, codeV2))
	require.NoError(t, err)

	var clientState ibcexported.ClientState
	err = gaiaApp.AppCodec().UnmarshalInterfaceJSON([]byte(clientStateJSON), &clientState)
	require.NoError(t, err)

	var consensusState ibcexported.ConsensusState
	err = gaiaApp.AppCodec().UnmarshalInterfaceJSON([]byte(consensusStateJSON), &consensusState)
	require.NoError(t, err)

	createMsg, err := clienttypes.NewMsgCreateClient(clientState, consensusState, v23_2_0.SignerAccount)
	require.NoError(t, err)
	createClientResp, err := gaiaApp.IBCKeeper.CreateClient(ctx, createMsg)
	require.NoError(t, err)
	require.Equal(t, "08-wasm-0", createClientResp.ClientId)

	err = v23_2_0.MigrateIBCWasm(ctx, gaiaApp.WasmClientKeeper, v23_2_0.HexChecksum, v23_2_0.MigrateMsgBase64,
		"08-wasm-0", v23_2_0.SignerAccount)
	require.NoError(t, err)
}
