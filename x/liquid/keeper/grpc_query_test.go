package keeper_test

import (
	gocontext "context"
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/gaia/v25/x/liquid/types"
)

func (s *KeeperTestSuite) TestGRPCQueryLiquidValidator() {
	ctx, keeper, queryClient := s.ctx, s.lsmKeeper, s.queryClient
	require := s.Require()

	lVal := types.NewLiquidValidator(sdk.ValAddress(PKs[0].Address().Bytes()).String())
	lVal.LiquidShares = math.LegacyNewDec(10000)
	require.NoError(keeper.SetLiquidValidator(ctx, lVal))
	var req *types.QueryLiquidValidatorRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = &types.QueryLiquidValidatorRequest{}
			},
			false,
		},
		{
			"nil request",
			func() {
				req = nil
			},
			false,
		},
		{
			"with valid and not existing address",
			func() {
				req = &types.QueryLiquidValidatorRequest{
					ValidatorAddr: "cosmosvaloper15jkng8hytwt22lllv6mw4k89qkqehtahd84ptu",
				}
			},
			false,
		},
		{
			"valid request",
			func() {
				req = &types.QueryLiquidValidatorRequest{ValidatorAddr: lVal.OperatorAddress}
			},
			true,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			res, err := queryClient.LiquidValidator(gocontext.Background(), req)
			if tc.expPass {
				require.NoError(err)
				require.Equal(lVal.OperatorAddress, res.LiquidValidator.OperatorAddress)
				require.Equal(lVal.LiquidShares, res.LiquidValidator.LiquidShares)
			} else {
				require.Error(err)
				require.Nil(res)
			}
		})
	}
}
