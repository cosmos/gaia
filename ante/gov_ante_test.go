package ante_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	tmrand "github.com/cometbft/cometbft/libs/rand"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"

	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/cosmos/gaia/v15/ante"
	gaiaapp "github.com/cosmos/gaia/v15/app"
	gaiahelpers "github.com/cosmos/gaia/v15/app/helpers"
)

var (
	insufficientCoins           = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100))
	insufficientMultiDenomCoins = sdk.NewCoins(
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 100),
		sdk.NewInt64Coin("ibc/example", 100))
	minCoins                   = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000000))
	moreThanMinCoins           = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 2500000))
	moreThanMinMultiDenomCoins = sdk.NewCoins(
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000000),
		sdk.NewInt64Coin("ibc/example", 100))
	testAddr = sdk.AccAddress("test1")
)

type GovAnteHandlerTestSuite struct {
	suite.Suite

	app       *gaiaapp.GaiaApp
	ctx       sdk.Context
	clientCtx client.Context
}

func (s *GovAnteHandlerTestSuite) SetupTest() {
	app := gaiahelpers.Setup(s.T())
	ctx := app.BaseApp.NewContext(false, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
		Height:  1,
	})

	legacyAmino := app.LegacyAmino()
	legacyAmino.Amino.RegisterConcrete(&testdata.TestMsg{}, "testdata.TestMsg", nil)
	testdata.RegisterInterfaces(app.InterfaceRegistry())

	s.app = app
	s.ctx = ctx
	s.clientCtx = client.Context{}.WithTxConfig(app.GetTxConfig())
}

func TestGovSpamPreventionSuite(t *testing.T) {
	suite.Run(t, new(GovAnteHandlerTestSuite))
}

func (s *GovAnteHandlerTestSuite) TestGlobalFeeMinimumGasFeeAnteHandler() {
	// setup test
	s.SetupTest()
	tests := []struct {
		title, description string
		proposalType       string
		proposerAddr       sdk.AccAddress
		initialDeposit     sdk.Coins
		expectPass         bool
	}{
		{"Passing proposal 1", "the purpose of this proposal is to pass", govv1beta1.ProposalTypeText, testAddr, minCoins, true},
		{"Passing proposal 2", "the purpose of this proposal is to pass with more coins than minimum", govv1beta1.ProposalTypeText, testAddr, moreThanMinCoins, true},
		{"Passing proposal 3", "the purpose of this proposal is to pass with multi denom coins", govv1beta1.ProposalTypeText, testAddr, moreThanMinMultiDenomCoins, true},
		{"Failing proposal 1", "the purpose of this proposal is to fail", govv1beta1.ProposalTypeText, testAddr, insufficientCoins, false},
		{"Failing proposal 2", "the purpose of this proposal is to fail with multi denom coins", govv1beta1.ProposalTypeText, testAddr, insufficientMultiDenomCoins, false},
	}

	decorator := ante.NewGovPreventSpamDecorator(s.app.AppCodec(), s.app.GovKeeper)

	for _, tc := range tests {
		content, _ := govv1beta1.ContentFromProposalType(tc.title, tc.description, tc.proposalType)
		s.Require().NotNil(content)

		msg, err := govv1beta1.NewMsgSubmitProposal(
			content,
			tc.initialDeposit,
			tc.proposerAddr,
		)

		s.Require().NoError(err)

		err = decorator.ValidateGovMsgs(s.ctx, []sdk.Msg{msg})
		if tc.expectPass {
			s.Require().NoError(err, "expected %v to pass", tc.title)
		} else {
			s.Require().Error(err, "expected %v to fail", tc.title)
		}
	}
}
