package e2e

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gaia/v23/tests/e2e/common"
	"github.com/cosmos/gaia/v23/tests/e2e/query"
	types2 "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v10/modules/core/04-channel/v2/types"
	"path/filepath"
	"time"
)

func (s *IntegrationTestSuite) TestV2RecvPacket() {
	chain := s.commonHelper.Resources.ChainA

	submitterAccount := chain.GenesisAccounts[1]
	submitterAddress, err := submitterAccount.KeyInfo.GetAddress()
	s.Require().NoError(err)

	endpoint := fmt.Sprintf("http://%s", s.commonHelper.Resources.ValResources[chain.ID][0].GetHostPort("1317/tcp"))

	timeoutTimestamp := uint64(time.Now().Add(time.Minute * 5).Unix())

	transferData := types2.NewFungibleTokenPacketData(
		"uatom",
		"1",
		RecipientAddress,
		RecipientAddress,
		"memo",
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
			Sequence:          1,
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
		Signer: submitterAddress.String(),
	}

	builder := common.TxConfig.NewTxBuilder()
	err = builder.SetMsgs(&packet)
	s.Require().NoError(err)

	builder.SetGasLimit(uint64(500000))
	builder.SetFeeAmount(sdk.NewCoins(sdk.NewInt64Coin("uatom", 500000)))
	builder.SetMemo("test")

	builtTx := builder.GetTx()
	bz, err := common.EncodingConfig.TxConfig.TxEncoder()(builtTx)
	s.Require().NoError(err)
	s.Require().NotNil(bz)

	decodedTx, err := common.DecodeTx(bz)
	s.Require().NoError(err)
	s.Require().NotNil(decodedTx)

	rawTx, err := common.Cdc.MarshalJSON(decodedTx)
	s.Require().NoError(err)
	s.Require().NotNil(rawTx)

	unsignedFname := "unsigned_recv_tx.json"
	unsignedJSONFile := filepath.Join(chain.Validators[0].ConfigDir(), unsignedFname)
	err = common.WriteFile(unsignedJSONFile, rawTx)
	s.Require().NoError(err)

	signedTx, err := s.tx.SignTxFileOnline(chain, 0, submitterAddress.String(), unsignedFname)
	s.Require().NoError(err)
	s.Require().NotNil(signedTx)

	signedFname := "signed_recv_tx.json"
	signedJSONFile := filepath.Join(chain.Validators[0].ConfigDir(), signedFname)
	err = common.WriteFile(signedJSONFile, signedTx)
	s.Require().NoError(err)

	// if there's no errors the non_critical_extension_options field was properly encoded and decoded
	out, err := s.tx.BroadcastTxFile(chain, 0, submitterAddress.String(), signedFname)
	s.Require().NoError(err)
	s.Require().NotNil(out)

	balances, err := query.AllBalances(endpoint, RecipientAddress)
	if err != nil {
		return
	}

	s.Require().Equal(balances[0].String(), "1ibc/1FBF3660E6387150C8BBDAA82EF8CE3C0AADE1F1BD921AE7529D892A53321A74")
}
