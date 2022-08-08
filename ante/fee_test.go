package ante_test

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gaia/v8/ante"
	globfeetypes "github.com/cosmos/gaia/v8/x/globalfee/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
)

// test global fees and min_gas_price with bypass msg types.
// please note even globalfee=0, min_gas_price=0, we do not let fee=0random_denom pass
// paid fees are already sanitized by removing zero coins(through feeFlag parsing), so use sdk.NewCoins() to create it.
func (s *IntegrationTestSuite) TestGlobalFeeMinimumGasFeeAnteHandler() {
	// setup test
	s.SetupTest()
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}

	denominator := int64(100000)
	high := sdk.NewDec(400).Quo(sdk.NewDec(denominator)) // 0.004
	med := sdk.NewDec(200).Quo(sdk.NewDec(denominator))  // 0.002
	low := sdk.NewDec(100).Quo(sdk.NewDec(denominator))  // 0.001

	highFeeAmt := sdk.NewInt(high.MulInt64(int64(2) * denominator).RoundInt64())
	medFeeAmt := sdk.NewInt(med.MulInt64(int64(2) * denominator).RoundInt64())
	lowFeeAmt := sdk.NewInt(low.MulInt64(int64(2) * denominator).RoundInt64())

	globalfeeParamsEmpty := &globfeetypes.Params{MinimumGasPrices: []sdk.DecCoin{}}
	minGasPriceEmpty := []sdk.DecCoin{}
	globalfeeParams0 := &globfeetypes.Params{MinimumGasPrices: []sdk.DecCoin{
		sdk.NewDecCoinFromDec("photon", sdk.NewDec(0)),
		sdk.NewDecCoinFromDec("uatom", sdk.NewDec(0)),
	}}
	globalfeeParamsContain0 := &globfeetypes.Params{MinimumGasPrices: []sdk.DecCoin{
		sdk.NewDecCoinFromDec("photon", med),
		sdk.NewDecCoinFromDec("uatom", sdk.NewDec(0)),
	}}
	minGasPrice0 := []sdk.DecCoin{
		sdk.NewDecCoinFromDec("stake", sdk.NewDec(0)),
		sdk.NewDecCoinFromDec("uatom", sdk.NewDec(0))}
	globalfeeParamsHigh := &globfeetypes.Params{
		MinimumGasPrices: []sdk.DecCoin{
			sdk.NewDecCoinFromDec("uatom", high),
		},
	}
	minGasPrice := []sdk.DecCoin{
		sdk.NewDecCoinFromDec("uatom", med),
		sdk.NewDecCoinFromDec("stake", med),
	}
	globalfeeParamsLow := &globfeetypes.Params{
		MinimumGasPrices: []sdk.DecCoin{
			sdk.NewDecCoinFromDec("uatom", low),
		},
	}
	// global fee must be sorted in denom
	globalfeeParamsNewDenom := &globfeetypes.Params{
		MinimumGasPrices: []sdk.DecCoin{
			sdk.NewDecCoinFromDec("photon", high),
			sdk.NewDecCoinFromDec("quark", high),
		},
	}
	testCases := map[string]struct {
		minGasPrice     []sdk.DecCoin
		globalFeeParams *globfeetypes.Params
		gasPrice        sdk.Coins
		gasLimit        sdk.Gas
		txMsg           sdk.Msg
		txCheck         bool
		expErr          bool
	}{
		// test fees
		// empty min_gas_price or empty global fee
		"empty min_gas_price, nonempty global fee, fee higher/equal than global_fee": {
			minGasPrice:     minGasPriceEmpty,
			globalFeeParams: globalfeeParamsHigh,
			// sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String())
			gasPrice: sdk.NewCoins(sdk.NewCoin("uatom", highFeeAmt)),
			gasLimit: testdata.NewTestGasLimit(),
			txMsg:    testdata.NewTestMsg(addr1),
			txCheck:  true,
			expErr:   false,
		},
		"empty min_gas_price, nonempty global fee, fee lower than global_fee": {
			minGasPrice:     minGasPriceEmpty,
			globalFeeParams: globalfeeParamsHigh,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("uatom", lowFeeAmt)),
			gasLimit:        testdata.NewTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          true,
		},
		"nonempty min_gas_price with defaultGlobalFee denom, empty global fee, fee higher/equal than min_gas_price": {
			minGasPrice:     minGasPrice,
			globalFeeParams: globalfeeParamsEmpty, //default 0uatom
			gasPrice:        sdk.NewCoins(sdk.NewCoin("uatom", medFeeAmt)),
			gasLimit:        testdata.NewTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          false,
		},
		"nonempty min_gas_price  with defaultGlobalFee denom, empty global fee, fee lower than min_gas_price": {
			minGasPrice:     minGasPrice,
			globalFeeParams: globalfeeParamsEmpty,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("uatom", lowFeeAmt)),
			gasLimit:        testdata.NewTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          true,
		},
		"empty min_gas_price, empty global fee, empty fee": {
			minGasPrice:     minGasPriceEmpty,
			globalFeeParams: globalfeeParamsEmpty,
			gasPrice:        sdk.Coins{},
			gasLimit:        testdata.NewTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          false,
		},
		// zero min_gas_price or zero global fee
		"zero min_gas_price, zero global fee, zero fee in global fee denom": {
			minGasPrice:     minGasPrice0,
			globalFeeParams: globalfeeParams0,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("uatom", sdk.ZeroInt()), sdk.NewCoin("photon", sdk.ZeroInt())),
			gasLimit:        testdata.NewTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          false,
		},
		"zero min_gas_price, zero global fee, empty fee": {
			minGasPrice:     minGasPrice0,
			globalFeeParams: globalfeeParams0,
			gasPrice:        sdk.Coins{},
			gasLimit:        testdata.NewTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          false,
		},
		// zero global fee
		"zero min_gas_price, zero global fee, zero fee not in globalfee denom": {
			minGasPrice:     minGasPrice0,
			globalFeeParams: globalfeeParams0,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("stake", sdk.ZeroInt())),
			gasLimit:        testdata.NewTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          false,
		},
		"zero min_gas_price, zero global fee, zero fees one in, one not in globalfee denom": {
			minGasPrice:     minGasPrice0,
			globalFeeParams: globalfeeParams0,
			gasPrice: sdk.NewCoins(
				sdk.NewCoin("stake", sdk.ZeroInt()),
				sdk.NewCoin("uatom", sdk.ZeroInt())),
			gasLimit: testdata.NewTestGasLimit(),
			txMsg:    testdata.NewTestMsg(addr1),
			txCheck:  true,
			expErr:   false,
		},
		// zero min_gas_price and empty  global fee
		"zero min_gas_price, empty global fee, zero fee in min_gas_price_denom": {
			minGasPrice:     minGasPrice0,
			globalFeeParams: globalfeeParamsEmpty,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("stake", sdk.ZeroInt())),
			gasLimit:        testdata.NewTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          false,
		},
		"zero min_gas_price, empty global fee, zero fee not in min_gas_price denom, not in defaultZeroGlobalFee denom": {
			minGasPrice:     minGasPrice0,
			globalFeeParams: globalfeeParamsEmpty,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("quark", sdk.ZeroInt())),
			gasLimit:        testdata.NewTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          false,
		},
		"zero min_gas_price, empty global fee, zero fee in defaultZeroGlobalFee denom": {
			minGasPrice:     minGasPrice0,
			globalFeeParams: globalfeeParamsEmpty,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("uatom", sdk.ZeroInt())),
			gasLimit:        newTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          false,
		},
		"zero min_gas_price, empty global fee, nonzero fee in defaultZeroGlobalFee denom": {
			minGasPrice:     minGasPrice0,
			globalFeeParams: globalfeeParamsEmpty,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("uatom", lowFeeAmt)),
			gasLimit:        newTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          false,
		},
		"zero min_gas_price, empty global fee, nonzero fee not in defaultZeroGlobalFee denom": {
			minGasPrice:     minGasPrice0,
			globalFeeParams: globalfeeParamsEmpty,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("quark", highFeeAmt)),
			gasLimit:        newTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          true,
		},
		// empty min_gas_price, zero global fee
		"empty min_gas_price, zero global fee, zero fee in global fee denom": {
			minGasPrice:     minGasPriceEmpty,
			globalFeeParams: globalfeeParams0,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("uatom", sdk.ZeroInt())),
			gasLimit:        newTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          false,
		},
		"empty min_gas_price, zero global fee, zero fee not in global fee denom": {
			minGasPrice:     minGasPriceEmpty,
			globalFeeParams: globalfeeParams0,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("stake", sdk.ZeroInt())),
			gasLimit:        newTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          false,
		},
		"empty min_gas_price, zero global fee, nonzero fee in global fee denom": {
			minGasPrice:     minGasPriceEmpty,
			globalFeeParams: globalfeeParams0,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("uatom", lowFeeAmt)),
			gasLimit:        newTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          false,
		},
		"empty min_gas_price, zero global fee, nonzero fee not in global fee denom": {
			minGasPrice:     minGasPriceEmpty,
			globalFeeParams: globalfeeParams0,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("stake", highFeeAmt)),
			gasLimit:        newTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          true,
		},
		// zero min_gas_price, nonzero global fee
		"zero min_gas_price, nonzero global fee, fee is higher than global fee": {
			minGasPrice:     minGasPrice0,
			globalFeeParams: globalfeeParamsLow,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("uatom", lowFeeAmt)),
			gasLimit:        newTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          false,
		},
		// nonzero min_gas_price, nonzero global fee
		"fee higher/equal than globalfee and min_gas_price": {
			minGasPrice:     minGasPrice,
			globalFeeParams: globalfeeParamsHigh,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("uatom", highFeeAmt)),
			gasLimit:        newTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          false,
		},
		"fee lower than globalfee and min_gas_price": {
			minGasPrice:     minGasPrice,
			globalFeeParams: globalfeeParamsHigh,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("uatom", lowFeeAmt)),
			gasLimit:        newTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          true,
		},
		"fee with one denom higher/equal, one denom lower than globalfee and min_gas_price": {
			minGasPrice:     minGasPrice,
			globalFeeParams: globalfeeParamsNewDenom,
			gasPrice: sdk.NewCoins(
				sdk.NewCoin("photon", lowFeeAmt),
				sdk.NewCoin("quark", highFeeAmt)),
			gasLimit: newTestGasLimit(),
			txMsg:    testdata.NewTestMsg(addr1),
			txCheck:  true,
			expErr:   false,
		},
		"globalfee > min_gas_price, fee higher/equal than min_gas_price, lower than globalfee": {
			minGasPrice:     minGasPrice,
			globalFeeParams: globalfeeParamsHigh,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("uatom", medFeeAmt)),
			gasLimit:        newTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          true,
		},
		"globalfee < min_gas_price, fee higher/equal than globalfee and lower than min_gas_price": {
			minGasPrice:     minGasPrice,
			globalFeeParams: globalfeeParamsLow,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("uatom", lowFeeAmt)),
			gasLimit:        newTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          true,
		},
		//  nonzero min_gas_price, zero global fee
		"check!!! nonzero min_gas_price, zero global fee, fee is in global fee denom and lower than min_gas_price": {
			minGasPrice:     minGasPrice,
			globalFeeParams: globalfeeParams0,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("uatom", lowFeeAmt)),
			gasLimit:        newTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          true,
		},
		"nonzero min_gas_price, zero global fee, fee is in global fee denom and higher/equal than min_gas_price": {
			minGasPrice:     minGasPrice,
			globalFeeParams: globalfeeParams0,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("uatom", medFeeAmt)),
			gasLimit:        newTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          false,
		},
		"nonzero min_gas_price, zero global fee, fee is in min_gas_price denom and higher/equal than min_gas_price": {
			minGasPrice:     minGasPrice,
			globalFeeParams: globalfeeParams0,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("stake", highFeeAmt)),
			gasLimit:        newTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          true,
		},
		// fee denom tests
		"min_gas_price denom is not subset of global fee denom , fee paying in global fee denom": {
			minGasPrice:     minGasPrice,
			globalFeeParams: globalfeeParamsNewDenom,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("photon", highFeeAmt)),
			gasLimit:        newTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          false,
		},
		"min_gas_price denom is not subset of global fee denom, fee paying in min_gas_price denom": {
			minGasPrice:     minGasPrice,
			globalFeeParams: globalfeeParamsNewDenom,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("stake", highFeeAmt)),
			gasLimit:        newTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          true,
		},
		"fees contain denom not in globalfee": {
			minGasPrice:     minGasPrice,
			globalFeeParams: globalfeeParamsLow,
			gasPrice: sdk.NewCoins(
				sdk.NewCoin("uatom", highFeeAmt),
				sdk.NewCoin("quark", highFeeAmt)),
			gasLimit: newTestGasLimit(),
			txMsg:    testdata.NewTestMsg(addr1),
			txCheck:  true,
			expErr:   true,
		},
		"fees contain denom not in globalfee with zero amount, fail": {
			minGasPrice:     minGasPrice,
			globalFeeParams: globalfeeParamsLow,
			gasPrice: sdk.NewCoins(sdk.NewCoin("uatom", highFeeAmt),
				sdk.NewCoin("quark", sdk.ZeroInt())),
			gasLimit: newTestGasLimit(),
			txMsg:    testdata.NewTestMsg(addr1),
			txCheck:  true,
			expErr:   false,
		},
		// cases from https://github.com/cosmos/gaia/pull/1570#issuecomment-1190524402
		"globalfee contains zero coin, fee is lower than the nonzero coin": {
			minGasPrice:     minGasPrice0,
			globalFeeParams: globalfeeParamsContain0,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("photon", lowFeeAmt)),
			gasLimit:        newTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          true,
		},
		"globalfee contains zero coin, fee contains zero coins of the same denom and a lower fee of the other denom in global fee": {
			minGasPrice:     minGasPrice0,
			globalFeeParams: globalfeeParamsContain0,
			gasPrice: sdk.NewCoins(
				sdk.NewCoin("photon", lowFeeAmt),
				sdk.NewCoin("uatom", sdk.ZeroInt())),
			gasLimit: newTestGasLimit(),
			txMsg:    testdata.NewTestMsg(addr1),
			txCheck:  true,
			expErr:   true,
		},
		"globalfee contains zero coin, fee is empty": {
			minGasPrice:     minGasPrice0,
			globalFeeParams: globalfeeParamsContain0,
			gasPrice:        sdk.Coins{},
			gasLimit:        newTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          false,
		},
		"globalfee contains zero coin, fee contains lower fee of zero coins's denom, lobalfee also contains nonzero coin,fee contains higher fee of nonzero coins's denom, ": {
			minGasPrice:     minGasPrice0,
			globalFeeParams: globalfeeParamsContain0,
			gasPrice: sdk.NewCoins(
				sdk.NewCoin("photon", lowFeeAmt),
				sdk.NewCoin("uatom", highFeeAmt)),
			gasLimit: newTestGasLimit(),
			txMsg:    testdata.NewTestMsg(addr1),
			txCheck:  true,
			expErr:   false,
		},
		"globalfee contains zero coin, fee is all zero coins but in global fee's denom": {
			minGasPrice:     minGasPrice0,
			globalFeeParams: globalfeeParamsContain0,
			gasPrice: sdk.NewCoins(
				sdk.NewCoin("photon", sdk.ZeroInt()),
				sdk.NewCoin("uatom", sdk.ZeroInt()),
			),
			gasLimit: newTestGasLimit(),
			txMsg:    testdata.NewTestMsg(addr1),
			txCheck:  true,
			expErr:   false,
		},
		"globalfee contains zero coin, fee is higher than the nonzero coin": {
			minGasPrice:     minGasPrice0,
			globalFeeParams: globalfeeParamsContain0,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("photon", highFeeAmt)),
			gasLimit:        newTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         true,
			expErr:          false,
		},
		// test bypass msg
		"msg type ibc, zero fee in globalfee denom": {
			minGasPrice:     minGasPrice,
			globalFeeParams: globalfeeParamsLow,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("uatom", sdk.ZeroInt())),
			gasLimit:        newTestGasLimit(),
			txMsg: ibcchanneltypes.NewMsgRecvPacket(
				ibcchanneltypes.Packet{}, nil, ibcclienttypes.Height{}, ""),
			txCheck: true,
			expErr:  false,
		},
		"msg type ibc, zero fee not in globalfee denom": {
			minGasPrice:     minGasPrice,
			globalFeeParams: globalfeeParamsLow,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("photon", sdk.ZeroInt())),
			gasLimit:        newTestGasLimit(),
			txMsg: ibcchanneltypes.NewMsgRecvPacket(
				ibcchanneltypes.Packet{}, nil, ibcclienttypes.Height{}, ""),
			txCheck: true,
			expErr:  false,
		},
		"msg type ibc, nonzero fee in globalfee denom": {
			minGasPrice:     minGasPrice,
			globalFeeParams: globalfeeParamsLow,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("uatom", highFeeAmt)),
			gasLimit:        newTestGasLimit(),
			txMsg: ibcchanneltypes.NewMsgRecvPacket(
				ibcchanneltypes.Packet{}, nil, ibcclienttypes.Height{}, ""),
			txCheck: true,
			expErr:  false,
		},
		"msg type ibc, nonzero fee not in globalfee denom": {
			minGasPrice:     minGasPrice,
			globalFeeParams: globalfeeParamsLow,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("photon", highFeeAmt)),
			gasLimit:        newTestGasLimit(),
			txMsg: ibcchanneltypes.NewMsgRecvPacket(
				ibcchanneltypes.Packet{}, nil, ibcclienttypes.Height{}, ""),
			txCheck: true,
			expErr:  true,
		},
		"msg type ibc, empty fee": {
			minGasPrice:     minGasPrice,
			globalFeeParams: globalfeeParamsLow,
			gasPrice:        sdk.Coins{},
			gasLimit:        newTestGasLimit(),
			txMsg: ibcchanneltypes.NewMsgRecvPacket(
				ibcchanneltypes.Packet{}, nil, ibcclienttypes.Height{}, ""),
			txCheck: true,
			expErr:  false,
		},
		"disable checkTx: no fee check": {
			minGasPrice:     minGasPrice,
			globalFeeParams: globalfeeParamsLow,
			gasPrice:        sdk.NewCoins(sdk.NewCoin("uatom", sdk.ZeroInt())),
			gasLimit:        newTestGasLimit(),
			txMsg:           testdata.NewTestMsg(addr1),
			txCheck:         false,
			expErr:          false,
		},
	}
	// todo check msg with bypass msg and need to pay msg
	for name, testCase := range testCases {
		s.Run(name, func() {
			// set globalfees and min gas price
			subspace := s.setupTestGlobalFeeStoreAndMinGasPrice(testCase.minGasPrice, testCase.globalFeeParams)
			// setup antehandler
			mfd := ante.NewBypassMinFeeDecorator([]string{
				sdk.MsgTypeURL(&ibcchanneltypes.MsgRecvPacket{}),
				sdk.MsgTypeURL(&ibcchanneltypes.MsgAcknowledgement{}),
				sdk.MsgTypeURL(&ibcclienttypes.MsgUpdateClient{}),
			}, subspace)
			antehandler := sdk.ChainAnteDecorators(mfd)

			s.Require().NoError(s.txBuilder.SetMsgs(testCase.txMsg))
			s.txBuilder.SetFeeAmount(testCase.gasPrice)
			s.txBuilder.SetGasLimit(testCase.gasLimit)
			tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
			s.Require().NoError(err)

			s.ctx = s.ctx.WithIsCheckTx(testCase.txCheck)
			_, err = antehandler(s.ctx, tx, false)
			if !testCase.expErr {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

// helpers
func newTestGasLimit() uint64 {
	return 200000
}
