package antetest

import (
	"testing"

	"github.com/stretchr/testify/suite"

	ibcclienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	gaiafeeante "github.com/cosmos/gaia/v12/x/globalfee/ante"
	globfeetypes "github.com/cosmos/gaia/v12/x/globalfee/types"
)

var testGasLimit uint64 = 200_000

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) TestGetDefaultGlobalFees() {
	// set globalfees and min gas price
	feeDecorator, _ := s.SetupTestGlobalFeeStoreAndMinGasPrice([]sdk.DecCoin{}, &globfeetypes.Params{})
	defaultGlobalFees, err := feeDecorator.DefaultZeroGlobalFee(s.ctx)
	s.Require().NoError(err)
	s.Require().Greater(len(defaultGlobalFees), 0)

	if defaultGlobalFees[0].Denom != testBondDenom {
		s.T().Fatalf("bond denom: %s, default global fee denom: %s", testBondDenom, defaultGlobalFees[0].Denom)
	}
}

// Test global fees and min_gas_price with bypass msg types.
// Please note even globalfee=0, min_gas_price=0, we do not let fee=0random_denom pass.
// Paid fees are already sanitized by removing zero coins(through feeFlag parsing), so use sdk.NewCoins() to create it.
func (s *IntegrationTestSuite) TestGlobalFeeMinimumGasFeeAnteHandler() {
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

	globalfeeParamsEmpty := []sdk.DecCoin{}
	minGasPriceEmpty := []sdk.DecCoin{}
	globalfeeParams0 := []sdk.DecCoin{
		sdk.NewDecCoinFromDec("photon", sdk.NewDec(0)),
		sdk.NewDecCoinFromDec("uatom", sdk.NewDec(0)),
	}
	globalfeeParamsContain0 := []sdk.DecCoin{
		sdk.NewDecCoinFromDec("photon", med),
		sdk.NewDecCoinFromDec("uatom", sdk.NewDec(0)),
	}
	minGasPrice0 := []sdk.DecCoin{
		sdk.NewDecCoinFromDec("stake", sdk.NewDec(0)),
		sdk.NewDecCoinFromDec("uatom", sdk.NewDec(0)),
	}
	globalfeeParamsHigh := []sdk.DecCoin{
		sdk.NewDecCoinFromDec("uatom", high),
	}
	minGasPrice := []sdk.DecCoin{
		sdk.NewDecCoinFromDec("uatom", med),
		sdk.NewDecCoinFromDec("stake", med),
	}
	globalfeeParamsLow := []sdk.DecCoin{
		sdk.NewDecCoinFromDec("uatom", low),
	}
	// global fee must be sorted in denom
	globalfeeParamsNewDenom := []sdk.DecCoin{
		sdk.NewDecCoinFromDec("photon", high),
		sdk.NewDecCoinFromDec("quark", high),
	}

	testCases := map[string]struct {
		minGasPrice []sdk.DecCoin
		globalFee   []sdk.DecCoin
		gasPrice    sdk.Coins
		gasLimit    sdk.Gas
		txMsg       sdk.Msg
		txCheck     bool
		expErr      bool
	}{
		// test fees
		// empty min_gas_price or empty global fee
		"empty min_gas_price, nonempty global fee, fee higher/equal than global_fee": {
			minGasPrice: minGasPriceEmpty,
			globalFee:   globalfeeParamsHigh,
			// sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String())
			gasPrice: sdk.NewCoins(sdk.NewCoin("uatom", highFeeAmt)),
			gasLimit: testGasLimit,
			txMsg:    testdata.NewTestMsg(addr1),
			txCheck:  true,
			expErr:   false,
		},
		"empty min_gas_price, nonempty global fee, fee lower than global_fee": {
			minGasPrice: minGasPriceEmpty,
			globalFee:   globalfeeParamsHigh,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("uatom", lowFeeAmt)),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      true,
		},
		"nonempty min_gas_price with defaultGlobalFee denom, empty global fee, fee higher/equal than min_gas_price": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParamsEmpty, // default 0uatom
			gasPrice:    sdk.NewCoins(sdk.NewCoin("uatom", medFeeAmt)),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      false,
		},
		"nonempty min_gas_price  with defaultGlobalFee denom, empty global fee, fee lower than min_gas_price": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParamsEmpty,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("uatom", lowFeeAmt)),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      true,
		},
		"empty min_gas_price, empty global fee, empty fee": {
			minGasPrice: minGasPriceEmpty,
			globalFee:   globalfeeParamsEmpty,
			gasPrice:    sdk.Coins{},
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      false,
		},
		// zero min_gas_price or zero global fee
		"zero min_gas_price, zero global fee, zero fee in global fee denom": {
			minGasPrice: minGasPrice0,
			globalFee:   globalfeeParams0,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("uatom", sdk.ZeroInt()), sdk.NewCoin("photon", sdk.ZeroInt())),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      false,
		},
		"zero min_gas_price, zero global fee, empty fee": {
			minGasPrice: minGasPrice0,
			globalFee:   globalfeeParams0,
			gasPrice:    sdk.Coins{},
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      false,
		},
		// zero global fee
		"zero min_gas_price, zero global fee, zero fee not in globalfee denom": {
			minGasPrice: minGasPrice0,
			globalFee:   globalfeeParams0,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("stake", sdk.ZeroInt())),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      false,
		},
		"zero min_gas_price, zero global fee, zero fees one in, one not in globalfee denom": {
			minGasPrice: minGasPrice0,
			globalFee:   globalfeeParams0,
			gasPrice: sdk.NewCoins(
				sdk.NewCoin("stake", sdk.ZeroInt()),
				sdk.NewCoin("uatom", sdk.ZeroInt())),
			gasLimit: testGasLimit,
			txMsg:    testdata.NewTestMsg(addr1),
			txCheck:  true,
			expErr:   false,
		},
		// zero min_gas_price and empty  global fee
		"zero min_gas_price, empty global fee, zero fee in min_gas_price_denom": {
			minGasPrice: minGasPrice0,
			globalFee:   globalfeeParamsEmpty,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("stake", sdk.ZeroInt())),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      false,
		},
		"zero min_gas_price, empty global fee, zero fee not in min_gas_price denom, not in defaultZeroGlobalFee denom": {
			minGasPrice: minGasPrice0,
			globalFee:   globalfeeParamsEmpty,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("quark", sdk.ZeroInt())),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      false,
		},
		"zero min_gas_price, empty global fee, zero fee in defaultZeroGlobalFee denom": {
			minGasPrice: minGasPrice0,
			globalFee:   globalfeeParamsEmpty,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("uatom", sdk.ZeroInt())),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      false,
		},
		"zero min_gas_price, empty global fee, nonzero fee in defaultZeroGlobalFee denom": {
			minGasPrice: minGasPrice0,
			globalFee:   globalfeeParamsEmpty,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("uatom", lowFeeAmt)),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      false,
		},
		"zero min_gas_price, empty global fee, nonzero fee not in defaultZeroGlobalFee denom": {
			minGasPrice: minGasPrice0,
			globalFee:   globalfeeParamsEmpty,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("quark", highFeeAmt)),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      true,
		},
		// empty min_gas_price, zero global fee
		"empty min_gas_price, zero global fee, zero fee in global fee denom": {
			minGasPrice: minGasPriceEmpty,
			globalFee:   globalfeeParams0,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("uatom", sdk.ZeroInt())),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      false,
		},
		"empty min_gas_price, zero global fee, zero fee not in global fee denom": {
			minGasPrice: minGasPriceEmpty,
			globalFee:   globalfeeParams0,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("stake", sdk.ZeroInt())),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      false,
		},
		"empty min_gas_price, zero global fee, nonzero fee in global fee denom": {
			minGasPrice: minGasPriceEmpty,
			globalFee:   globalfeeParams0,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("uatom", lowFeeAmt)),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      false,
		},
		"empty min_gas_price, zero global fee, nonzero fee not in global fee denom": {
			minGasPrice: minGasPriceEmpty,
			globalFee:   globalfeeParams0,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("stake", highFeeAmt)),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      true,
		},
		// zero min_gas_price, nonzero global fee
		"zero min_gas_price, nonzero global fee, fee is higher than global fee": {
			minGasPrice: minGasPrice0,
			globalFee:   globalfeeParamsLow,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("uatom", lowFeeAmt)),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      false,
		},
		// nonzero min_gas_price, nonzero global fee
		"fee higher/equal than globalfee and min_gas_price": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParamsHigh,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("uatom", highFeeAmt)),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      false,
		},
		"fee lower than globalfee and min_gas_price": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParamsHigh,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("uatom", lowFeeAmt)),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      true,
		},
		"fee with one denom higher/equal, one denom lower than globalfee and min_gas_price": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParamsNewDenom,
			gasPrice: sdk.NewCoins(
				sdk.NewCoin("photon", lowFeeAmt),
				sdk.NewCoin("quark", highFeeAmt)),
			gasLimit: testGasLimit,
			txMsg:    testdata.NewTestMsg(addr1),
			txCheck:  true,
			expErr:   false,
		},
		"globalfee > min_gas_price, fee higher/equal than min_gas_price, lower than globalfee": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParamsHigh,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("uatom", medFeeAmt)),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      true,
		},
		"globalfee < min_gas_price, fee higher/equal than globalfee and lower than min_gas_price": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParamsLow,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("uatom", lowFeeAmt)),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      true,
		},
		//  nonzero min_gas_price, zero global fee
		"nonzero min_gas_price, zero global fee, fee is in global fee denom and lower than min_gas_price": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParams0,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("uatom", lowFeeAmt)),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      true,
		},
		"nonzero min_gas_price, zero global fee, fee is in global fee denom and higher/equal than min_gas_price": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParams0,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("uatom", medFeeAmt)),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      false,
		},
		"nonzero min_gas_price, zero global fee, fee is in min_gas_price denom which is not in global fee default, but higher/equal than min_gas_price": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParams0,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("stake", highFeeAmt)),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      true,
		},
		// fee denom tests
		"min_gas_price denom is not subset of global fee denom , fee paying in global fee denom": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParamsNewDenom,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("photon", highFeeAmt)),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      false,
		},
		"min_gas_price denom is not subset of global fee denom, fee paying in min_gas_price denom": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParamsNewDenom,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("stake", highFeeAmt)),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      true,
		},
		"fees contain denom not in globalfee": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParamsLow,
			gasPrice: sdk.NewCoins(
				sdk.NewCoin("uatom", highFeeAmt),
				sdk.NewCoin("quark", highFeeAmt)),
			gasLimit: testGasLimit,
			txMsg:    testdata.NewTestMsg(addr1),
			txCheck:  true,
			expErr:   true,
		},
		"fees contain denom not in globalfee with zero amount": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParamsLow,
			gasPrice: sdk.NewCoins(sdk.NewCoin("uatom", highFeeAmt),
				sdk.NewCoin("quark", sdk.ZeroInt())),
			gasLimit: testGasLimit,
			txMsg:    testdata.NewTestMsg(addr1),
			txCheck:  true,
			expErr:   false,
		},
		// cases from https://github.com/cosmos/gaia/pull/1570#issuecomment-1190524402
		// note: this is kind of a silly scenario but technically correct
		// if there is a zero coin in the globalfee, the user could pay 0fees
		// if the user includes any fee at all in the non-zero denom, it must be higher than that non-zero fee
		// unlikely we will ever see zero and non-zero together but technically possible
		"globalfee contains zero coin and non-zero coin, fee is lower than the nonzero coin": {
			minGasPrice: minGasPrice0,
			globalFee:   globalfeeParamsContain0,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("photon", lowFeeAmt)),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      true,
		},
		"globalfee contains zero coin, fee contains zero coins of the same denom and a lower fee of the other denom in global fee": {
			minGasPrice: minGasPrice0,
			globalFee:   globalfeeParamsContain0,
			gasPrice: sdk.NewCoins(
				sdk.NewCoin("photon", lowFeeAmt),
				sdk.NewCoin("uatom", sdk.ZeroInt())),
			gasLimit: testGasLimit,
			txMsg:    testdata.NewTestMsg(addr1),
			txCheck:  true,
			expErr:   true,
		},
		"globalfee contains zero coin, fee is empty": {
			minGasPrice: minGasPrice0,
			globalFee:   globalfeeParamsContain0,
			gasPrice:    sdk.Coins{},
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      false,
		},
		"globalfee contains zero coin, fee contains lower fee of zero coins's denom, globalfee also contains nonzero coin,fee contains higher fee of nonzero coins's denom, ": {
			minGasPrice: minGasPrice0,
			globalFee:   globalfeeParamsContain0,
			gasPrice: sdk.NewCoins(
				sdk.NewCoin("photon", lowFeeAmt),
				sdk.NewCoin("uatom", highFeeAmt)),
			gasLimit: testGasLimit,
			txMsg:    testdata.NewTestMsg(addr1),
			txCheck:  true,
			expErr:   false,
		},
		"globalfee contains zero coin, fee is all zero coins but in global fee's denom": {
			minGasPrice: minGasPrice0,
			globalFee:   globalfeeParamsContain0,
			gasPrice: sdk.NewCoins(
				sdk.NewCoin("photon", sdk.ZeroInt()),
				sdk.NewCoin("uatom", sdk.ZeroInt()),
			),
			gasLimit: testGasLimit,
			txMsg:    testdata.NewTestMsg(addr1),
			txCheck:  true,
			expErr:   false,
		},
		"globalfee contains zero coin, fee is higher than the nonzero coin": {
			minGasPrice: minGasPrice0,
			globalFee:   globalfeeParamsContain0,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("photon", highFeeAmt)),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      false,
		},
		"bypass msg type: ibc.core.channel.v1.MsgRecvPacket": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParamsLow,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("uatom", sdk.ZeroInt())),
			gasLimit:    testGasLimit,
			txMsg: ibcchanneltypes.NewMsgRecvPacket(
				ibcchanneltypes.Packet{}, nil, ibcclienttypes.Height{}, ""),
			txCheck: true,
			expErr:  false,
		},
		"bypass msg type: ibc.core.channel.v1.MsgTimeout": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParamsLow,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("uatom", sdk.ZeroInt())),
			gasLimit:    testGasLimit,
			txMsg: ibcchanneltypes.NewMsgTimeout(
				// todo check here
				ibcchanneltypes.Packet{}, 1, nil, ibcclienttypes.Height{}, ""),
			txCheck: true,
			expErr:  false,
		},
		"bypass msg type: ibc.core.channel.v1.MsgTimeoutOnClose": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParamsLow,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("uatom", sdk.ZeroInt())),
			gasLimit:    testGasLimit,
			txMsg: ibcchanneltypes.NewMsgTimeout(
				ibcchanneltypes.Packet{}, 2, nil, ibcclienttypes.Height{}, ""),
			txCheck: true,
			expErr:  false,
		},
		"bypass msg gas usage exceeds maxTotalBypassMinFeeMsgGasUsage": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParamsLow,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("uatom", sdk.ZeroInt())),
			gasLimit:    2 * globfeetypes.DefaultmaxTotalBypassMinFeeMsgGasUsage,
			txMsg: ibcchanneltypes.NewMsgTimeout(
				ibcchanneltypes.Packet{}, 2, nil, ibcclienttypes.Height{}, ""),
			txCheck: true,
			expErr:  true,
		},
		"bypass msg gas usage equals to maxTotalBypassMinFeeMsgGasUsage": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParamsLow,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("uatom", sdk.ZeroInt())),
			gasLimit:    globfeetypes.DefaultmaxTotalBypassMinFeeMsgGasUsage,
			txMsg: ibcchanneltypes.NewMsgTimeout(
				ibcchanneltypes.Packet{}, 3, nil, ibcclienttypes.Height{}, ""),
			txCheck: true,
			expErr:  false,
		},
		"msg type ibc, zero fee not in globalfee denom": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParamsLow,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("photon", sdk.ZeroInt())),
			gasLimit:    testGasLimit,
			txMsg: ibcchanneltypes.NewMsgRecvPacket(
				ibcchanneltypes.Packet{}, nil, ibcclienttypes.Height{}, ""),
			txCheck: true,
			expErr:  false,
		},
		"msg type ibc, nonzero fee in globalfee denom": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParamsLow,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("uatom", highFeeAmt)),
			gasLimit:    testGasLimit,
			txMsg: ibcchanneltypes.NewMsgRecvPacket(
				ibcchanneltypes.Packet{}, nil, ibcclienttypes.Height{}, ""),
			txCheck: true,
			expErr:  false,
		},
		"msg type ibc, nonzero fee not in globalfee denom": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParamsLow,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("photon", highFeeAmt)),
			gasLimit:    testGasLimit,
			txMsg: ibcchanneltypes.NewMsgRecvPacket(
				ibcchanneltypes.Packet{}, nil, ibcclienttypes.Height{}, ""),
			txCheck: true,
			expErr:  true,
		},
		"msg type ibc, empty fee": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParamsLow,
			gasPrice:    sdk.Coins{},
			gasLimit:    testGasLimit,
			txMsg: ibcchanneltypes.NewMsgRecvPacket(
				ibcchanneltypes.Packet{}, nil, ibcclienttypes.Height{}, ""),
			txCheck: true,
			expErr:  false,
		},
		"msg type non-ibc, nonzero fee in globalfee denom": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParamsLow,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("uatom", highFeeAmt)),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      false,
		},
		"msg type non-ibc, empty fee": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParamsLow,
			gasPrice:    sdk.Coins{},
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      true,
		},
		"msg type non-ibc, nonzero fee not in globalfee denom": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParamsLow,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("photon", highFeeAmt)),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     true,
			expErr:      true,
		},
		"disable checkTx: min_gas_price is medium, global fee is low, tx fee is low": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParamsLow,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("uatom", lowFeeAmt)),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     false,
			expErr:      false,
		},
		"disable checkTx: min_gas_price is medium, global fee is low, tx is zero": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParamsLow,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("uatom", sdk.ZeroInt())),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     false,
			expErr:      true,
		},
		"disable checkTx: min_gas_price is low, global fee is low, tx fee's denom is not in global fees denoms set": {
			minGasPrice: minGasPrice,
			globalFee:   globalfeeParamsLow,
			gasPrice:    sdk.NewCoins(sdk.NewCoin("quark", sdk.ZeroInt())),
			gasLimit:    testGasLimit,
			txMsg:       testdata.NewTestMsg(addr1),
			txCheck:     false,
			expErr:      true,
		},
	}

	globalfeeParams := &globfeetypes.Params{
		BypassMinFeeMsgTypes:            globfeetypes.DefaultBypassMinFeeMsgTypes,
		MaxTotalBypassMinFeeMsgGasUsage: globfeetypes.DefaultmaxTotalBypassMinFeeMsgGasUsage,
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			// set globalfees and min gas price
			globalfeeParams.MinimumGasPrices = tc.globalFee
			_, antehandler := s.SetupTestGlobalFeeStoreAndMinGasPrice(tc.minGasPrice, globalfeeParams)

			// set fee decorator to ante handler

			s.Require().NoError(s.txBuilder.SetMsgs(tc.txMsg))
			s.txBuilder.SetFeeAmount(tc.gasPrice)
			s.txBuilder.SetGasLimit(tc.gasLimit)
			tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
			s.Require().NoError(err)

			s.ctx = s.ctx.WithIsCheckTx(tc.txCheck)
			_, err = antehandler(s.ctx, tx, false)
			if !tc.expErr {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

// Test how the operator fees are determined using various min gas prices.
//
// Note that in a real Gaia deployment all zero coins can be removed from minGasPrice.
// This sanitizing happens when the minGasPrice is set into the context.
// (see baseapp.SetMinGasPrices in gaia/cmd/root.go line 221)
func (s *IntegrationTestSuite) TestGetMinGasPrice() {
	expCoins := sdk.Coins{
		sdk.NewCoin("photon", sdk.NewInt(2000)),
		sdk.NewCoin("uatom", sdk.NewInt(3000)),
	}

	testCases := []struct {
		name          string
		minGasPrice   []sdk.DecCoin
		feeTxGasLimit uint64
		expCoins      sdk.Coins
	}{
		{
			"empty min gas price should return empty coins",
			[]sdk.DecCoin{},
			uint64(1000),
			sdk.Coins{},
		},
		{
			"zero coins min gas price should return empty coins",
			[]sdk.DecCoin{
				sdk.NewDecCoinFromDec("stake", sdk.NewDec(0)),
				sdk.NewDecCoinFromDec("uatom", sdk.NewDec(0)),
			},
			uint64(1000),
			sdk.Coins{},
		},
		{
			"zero coins, non-zero coins mix should return zero coin and non-zero coins",
			[]sdk.DecCoin{
				sdk.NewDecCoinFromDec("stake", sdk.NewDec(0)),
				sdk.NewDecCoinFromDec("uatom", sdk.NewDec(1)),
			},
			uint64(1000),
			sdk.Coins{
				sdk.NewCoin("stake", sdk.NewInt(0)),
				sdk.NewCoin("uatom", sdk.NewInt(1000)),
			},
		},

		{
			"unsorted min gas price should return sorted coins",
			[]sdk.DecCoin{
				sdk.NewDecCoinFromDec("uatom", sdk.NewDec(3)),
				sdk.NewDecCoinFromDec("photon", sdk.NewDec(2)),
			},
			uint64(1000),
			expCoins,
		},
		{
			"sorted min gas price should return same conins",
			[]sdk.DecCoin{
				sdk.NewDecCoinFromDec("photon", sdk.NewDec(2)),
				sdk.NewDecCoinFromDec("uatom", sdk.NewDec(3)),
			},
			uint64(1000),
			expCoins,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTestGlobalFeeStoreAndMinGasPrice(tc.minGasPrice, &globfeetypes.Params{})

			fees := gaiafeeante.GetMinGasPrice(s.ctx, int64(tc.feeTxGasLimit))
			s.Require().True(tc.expCoins.Sort().IsEqual(fees))
		})
	}
}

func (s *IntegrationTestSuite) TestContainsOnlyBypassMinFeeMsgs() {
	// set globalfees params and min gas price
	globalfeeParams := &globfeetypes.Params{
		BypassMinFeeMsgTypes:            globfeetypes.DefaultBypassMinFeeMsgTypes,
		MaxTotalBypassMinFeeMsgGasUsage: globfeetypes.DefaultmaxTotalBypassMinFeeMsgGasUsage,
	}
	feeDecorator, _ := s.SetupTestGlobalFeeStoreAndMinGasPrice([]sdk.DecCoin{}, globalfeeParams)
	testCases := []struct {
		name    string
		msgs    []sdk.Msg
		expPass bool
	}{
		{
			"expect empty msgs to pass",
			[]sdk.Msg{},
			true,
		},
		{
			"expect default bypass msg to pass",
			[]sdk.Msg{
				ibcchanneltypes.NewMsgRecvPacket(ibcchanneltypes.Packet{}, nil, ibcclienttypes.Height{}, ""),
				ibcchanneltypes.NewMsgAcknowledgement(ibcchanneltypes.Packet{}, []byte{1}, []byte{1}, ibcclienttypes.Height{}, ""),
			},
			true,
		},
		{
			"expect default bypass msgs to pass",
			[]sdk.Msg{
				ibcchanneltypes.NewMsgRecvPacket(ibcchanneltypes.Packet{}, nil, ibcclienttypes.Height{}, ""),
				ibcchanneltypes.NewMsgAcknowledgement(ibcchanneltypes.Packet{}, []byte{1}, []byte{1}, ibcclienttypes.Height{}, ""),
			},
			true,
		},
		{
			"msgs contain non-bypass msg - should not pass",
			[]sdk.Msg{
				ibcchanneltypes.NewMsgRecvPacket(ibcchanneltypes.Packet{}, nil, ibcclienttypes.Height{}, ""),
				stakingtypes.NewMsgDelegate(sdk.AccAddress{}, sdk.ValAddress{}, sdk.Coin{}),
			},
			false,
		},
		{
			"msgs contain only non-bypass msgs - should not pass",
			[]sdk.Msg{
				stakingtypes.NewMsgDelegate(sdk.AccAddress{}, sdk.ValAddress{}, sdk.Coin{}),
			},
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			res := feeDecorator.ContainsOnlyBypassMinFeeMsgs(s.ctx, tc.msgs)
			s.Require().True(tc.expPass == res)
		})
	}
}

func (s *IntegrationTestSuite) TestGetTxFeeRequired() {
	// create global fee params
	globalfeeParamsEmpty := &globfeetypes.Params{MinimumGasPrices: []sdk.DecCoin{}}

	// setup tests with default global fee i.e. "0uatom" and empty local min gas prices
	feeDecorator, _ := s.SetupTestGlobalFeeStoreAndMinGasPrice([]sdk.DecCoin{}, globalfeeParamsEmpty)

	// set a subspace that doesn't have the stakingtypes.KeyBondDenom key registred
	feeDecorator.StakingSubspace = s.app.GetSubspace(globfeetypes.ModuleName)

	// check that an error is returned when staking bond denom is empty
	_, err := feeDecorator.GetTxFeeRequired(s.ctx, nil)
	s.Require().Equal(err.Error(), "empty staking bond denomination")

	// set non-zero local min gas prices
	localMinGasPrices := sdk.NewCoins(sdk.NewCoin("uatom", sdk.NewInt(1)))

	// setup tests with non-empty local min gas prices
	feeDecorator, _ = s.SetupTestGlobalFeeStoreAndMinGasPrice(
		sdk.NewDecCoinsFromCoins(localMinGasPrices...),
		globalfeeParamsEmpty,
	)

	// mock tx data
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}

	s.Require().NoError(s.txBuilder.SetMsgs(testdata.NewTestMsg(addr1)))
	s.txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin("uatom", sdk.ZeroInt())))

	s.txBuilder.SetGasLimit(uint64(1))
	tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
	s.Require().NoError(err)

	// check that the required fees returned in CheckTx mode are equal to
	// local min gas prices since they're greater than the default global fee values.
	s.Require().True(s.ctx.IsCheckTx())
	res, err := feeDecorator.GetTxFeeRequired(s.ctx, tx)
	s.Require().True(res.IsEqual(localMinGasPrices))
	s.Require().NoError(err)

	// check that the global fee is returned in DeliverTx mode.
	globalFee, err := feeDecorator.GetGlobalFee(s.ctx, tx)
	s.Require().NoError(err)

	ctx := s.ctx.WithIsCheckTx(false)
	res, err = feeDecorator.GetTxFeeRequired(ctx, tx)
	s.Require().NoError(err)
	s.Require().True(res.IsEqual(globalFee))
}
