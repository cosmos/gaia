package keeper_test

import (
	gocontext "context"
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

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

func (s *KeeperTestSuite) TestGRPCQueryLiquidValidators() {
	ctx, keeper, queryClient := s.ctx, s.lsmKeeper, s.queryClient
	require := s.Require()

	// Create multiple liquid validators for testing
	lVal1 := types.NewLiquidValidator(sdk.ValAddress(PKs[0].Address().Bytes()).String())
	lVal1.LiquidShares = math.LegacyNewDec(10000)
	require.NoError(keeper.SetLiquidValidator(ctx, lVal1))

	lVal2 := types.NewLiquidValidator(sdk.ValAddress(PKs[1].Address().Bytes()).String())
	lVal2.LiquidShares = math.LegacyNewDec(20000)
	require.NoError(keeper.SetLiquidValidator(ctx, lVal2))

	lVal3 := types.NewLiquidValidator(sdk.ValAddress(PKs[2].Address().Bytes()).String())
	lVal3.LiquidShares = math.LegacyNewDec(30000)
	require.NoError(keeper.SetLiquidValidator(ctx, lVal3))

	var req *types.QueryLiquidValidatorsRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = &types.QueryLiquidValidatorsRequest{}
			},
			true,
		},
		{
			"valid request with pagination limit",
			func() {
				req = &types.QueryLiquidValidatorsRequest{
					Pagination: &query.PageRequest{
						Limit: 2,
					},
				}
			},
			true,
		},
		{
			"valid request with offset pagination",
			func() {
				req = &types.QueryLiquidValidatorsRequest{
					Pagination: &query.PageRequest{
						Offset: 1,
						Limit:  2,
					},
				}
			},
			true,
		},
		{
			"request with count total",
			func() {
				req = &types.QueryLiquidValidatorsRequest{
					Pagination: &query.PageRequest{
						CountTotal: true,
					},
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			res, err := queryClient.LiquidValidators(gocontext.Background(), req)
			if tc.expPass {
				require.NoError(err)
				require.NotNil(res)
				require.NotNil(res.LiquidValidators)

				// Additional checks for specific test cases
				if tc.msg == "empty request" {
					// Should return all validators
					require.Len(res.LiquidValidators, 3)

					// Verify the validators are correct
					foundValidators := make(map[string]bool)
					for _, lv := range res.LiquidValidators {
						foundValidators[lv.OperatorAddress] = true
					}
					require.True(foundValidators[lVal1.OperatorAddress])
					require.True(foundValidators[lVal2.OperatorAddress])
					require.True(foundValidators[lVal3.OperatorAddress])
				}

				if tc.msg == "request with count total" {
					require.NotNil(res.Pagination)
					require.Equal(uint64(3), res.Pagination.Total)
					require.Len(res.LiquidValidators, 3)
				}

				if tc.msg == "valid request with pagination limit" {
					require.NotNil(res.Pagination)
					require.Len(res.LiquidValidators, 2)
				}

				if tc.msg == "valid request with offset pagination" {
					require.NotNil(res.Pagination)
					require.Len(res.LiquidValidators, 2)
				}
			} else {
				require.Error(err)
				require.Nil(res)
			}
		})
	}
}
