package ante_test

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	"github.com/cosmos/gaia/v7/ante"
)

func (s *IntegrationTestSuite) TestMempoolFeeDecorator() {
	s.SetupTest()
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()

	mfd := ante.NewMempoolFeeDecorator([]string{
		sdk.MsgTypeURL(&ibcchanneltypes.MsgRecvPacket{}),
		sdk.MsgTypeURL(&ibcchanneltypes.MsgAcknowledgement{}),
		sdk.MsgTypeURL(&ibcclienttypes.MsgUpdateClient{}),
	})
	antehandler := sdk.ChainAnteDecorators(mfd)
	priv1, _, addr1 := testdata.KeyTestPubAddr()

	msg := testdata.NewTestMsg(addr1)
	feeAmount := testdata.NewTestFeeAmount()
	gasLimit := testdata.NewTestGasLimit()
	s.Require().NoError(s.txBuilder.SetMsgs(msg))
	s.txBuilder.SetFeeAmount(feeAmount)
	s.txBuilder.SetGasLimit(gasLimit)

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	s.Require().NoError(err)

	// Set high gas price so standard test fee fails
	feeAmt := sdk.NewDecCoinFromDec("uatom", sdk.NewDec(200).Quo(sdk.NewDec(100000)))
	minGasPrice := []sdk.DecCoin{feeAmt}
	s.ctx = s.ctx.WithMinGasPrices(minGasPrice).WithIsCheckTx(true)

	// antehandler errors with insufficient fees
	_, err = antehandler(s.ctx, tx, false)
	s.Require().Error(err, "expected error due to low fee")

	// ensure no fees for certain IBC msgs
	s.Require().NoError(s.txBuilder.SetMsgs(
		ibcchanneltypes.NewMsgRecvPacket(ibcchanneltypes.Packet{}, nil, ibcclienttypes.Height{}, ""),
	))

	oracleTx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	_, err = antehandler(s.ctx, oracleTx, false)
	s.Require().NoError(err, "expected min fee bypass for IBC messages")

	s.ctx = s.ctx.WithIsCheckTx(false)

	// antehandler should not error since we do not check min gas prices in DeliverTx
	_, err = antehandler(s.ctx, tx, false)
	s.Require().NoError(err, "unexpected error during DeliverTx")
}
