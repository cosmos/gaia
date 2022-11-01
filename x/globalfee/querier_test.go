package globalfee_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gaia/v8/x/globalfee"
	"github.com/cosmos/gaia/v8/x/globalfee/types"
	"github.com/stretchr/testify/suite"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/gaia/v8/app"
	gaiahelpers "github.com/cosmos/gaia/v8/app/helpers"
)

type testSuite struct {
	suite.Suite

	app         *gaia.GaiaApp
	ctx         sdk.Context
	queryClient globalfee.GrpcQuerier
}

func TestGRPCQueryMinimumGasPricesSuite(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) TestGRPCQueryMinimumGasPrices() {
	s.setup()

	// MinimumGasPrices: empty coins
	emptyCoins := sdk.DecCoins{}
	globalFeeParams := types.Params{
		MinimumGasPrices: emptyCoins,
	}
	s.setupGrpcQuerier(&globalFeeParams)
	response, err := s.queryClient.MinimumGasPrices(s.ctx, &types.QueryMinimumGasPricesRequest{})
	s.Require().NoError(err)
	s.Require().Equal(response.MinimumGasPrices.String(), sdk.DecCoins{}.String())

	// MinimumGasPrices: zero coin
	zeroStake := sdk.DecCoins{sdk.NewDecCoin("stake", sdk.ZeroInt())}
	globalFeeParams = types.Params{
		MinimumGasPrices: zeroStake,
	}
	s.setupGrpcQuerier(&globalFeeParams)
	response, err = s.queryClient.MinimumGasPrices(s.ctx, &types.QueryMinimumGasPricesRequest{})
	s.Require().NoError(err)
	s.Require().Equal(response.MinimumGasPrices.String(), zeroStake.String())

	// MinimumGasPrices: mix coins
	mixCoins := sdk.DecCoins{
		sdk.NewDecCoin("photon", sdk.OneInt()),
		sdk.NewDecCoin("stake", sdk.ZeroInt()),
	}
	globalFeeParams = types.Params{
		MinimumGasPrices: mixCoins,
	}
	s.setupGrpcQuerier(&globalFeeParams)
	response, err = s.queryClient.MinimumGasPrices(s.ctx, &types.QueryMinimumGasPricesRequest{})
	s.Require().NoError(err)
	s.Require().Equal(response.MinimumGasPrices.String(), mixCoins.String())
}

func (s *testSuite) setup() {
	app := gaiahelpers.Setup(s.T(), false, 1)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
		Height:  1,
	})

	encodingConfig := gaia.MakeTestEncodingConfig()
	encodingConfig.Amino.RegisterConcrete(&testdata.TestMsg{}, "testdata.TestMsg", nil)
	testdata.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	s.app = app
	s.ctx = ctx
}

func (s *testSuite) setupGrpcQuerier(globalFeeParams *types.Params) {
	subspace := s.app.GetSubspace(globalfee.ModuleName)
	subspace.SetParamSet(s.ctx, globalFeeParams)

	s.queryClient = globalfee.NewGrpcQuerier(subspace)
}
