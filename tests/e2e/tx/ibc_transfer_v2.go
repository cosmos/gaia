package tx

import (
	types2 "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v10/modules/core/04-channel/v2/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/gaia/v26/tests/e2e/common"
)

func (h *TestingSuite) CreateIBCV2RecvPacketTx(timeoutTimestamp uint64, amount, submitterAddress, recipientAddress, memo string) ([]byte, error) {
	transferData := types2.NewFungibleTokenPacketData(
		"uatom",
		amount,
		submitterAddress,
		recipientAddress,
		memo,
	)
	transferBz := common.Cdc.MustMarshal(&transferData)
	transferPayload := types.NewPayload(
		types2.PortID,
		types2.PortID,
		types2.V1,
		types2.EncodingProtobuf,
		transferBz,
	)

	payloads := []types.Payload{transferPayload}

	packet := types.MsgRecvPacket{
		Packet: types.Packet{
			Sequence:          uint64(h.TestCounters.IBCV2PacketSequence), //nolint:gosec
			SourceClient:      common.CounterpartyID,
			DestinationClient: common.V2TransferClient,
			TimeoutTimestamp:  timeoutTimestamp,
			Payloads:          payloads,
		},
		ProofCommitment: []byte("mock_commitment"),
		ProofHeight: clienttypes.Height{
			RevisionNumber: 0,
			RevisionHeight: 0,
		},
		Signer: submitterAddress,
	}

	builder := common.TxConfig.NewTxBuilder()
	err := builder.SetMsgs(&packet)
	if err != nil {
		return nil, err
	}

	builder.SetGasLimit(uint64(500000))
	builder.SetFeeAmount(sdk.NewCoins(sdk.NewInt64Coin("uatom", 500000)))

	builtTx := builder.GetTx()
	bz, err := common.EncodingConfig.TxConfig.TxEncoder()(builtTx)
	if err != nil {
		return nil, err
	}

	decodedTx, err := common.DecodeTx(bz)
	if err != nil {
		return nil, err
	}

	rawTx, err := common.Cdc.MarshalJSON(decodedTx)
	if err != nil {
		return nil, err
	}

	return rawTx, nil
}
