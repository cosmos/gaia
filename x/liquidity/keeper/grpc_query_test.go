package keeper_test

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

func (suite *KeeperTestSuite) TestGRPCLiquidityPool() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient
	pool, found := app.LiquidityKeeper.GetPool(ctx, suite.pools[0].Id)
	suite.True(found)

	var req *types.QueryLiquidityPoolRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = &types.QueryLiquidityPoolRequest{}
			},
			false,
		},
		{
			"valid request",
			func() {
				req = &types.QueryLiquidityPoolRequest{PoolId: suite.pools[0].Id}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			res, err := queryClient.LiquidityPool(context.Background(), req)
			if tc.expPass {
				suite.NoError(err)
				suite.Equal(pool.Id, res.Pool.Id)
			} else {
				suite.Error(err)
				suite.Nil(res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCLiquidityPoolByPoolCoinDenom() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient
	pool, found := app.LiquidityKeeper.GetPool(ctx, suite.pools[0].Id)
	suite.True(found)

	var req *types.QueryLiquidityPoolByPoolCoinDenomRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = &types.QueryLiquidityPoolByPoolCoinDenomRequest{}
			},
			false,
		},
		{
			"valid request",
			func() {
				req = &types.QueryLiquidityPoolByPoolCoinDenomRequest{PoolCoinDenom: suite.pools[0].PoolCoinDenom}
			},
			true,
		},
		{
			"invalid request",
			func() {
				req = &types.QueryLiquidityPoolByPoolCoinDenomRequest{PoolCoinDenom: suite.pools[0].PoolCoinDenom[:10]}
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			res, err := queryClient.LiquidityPoolByPoolCoinDenom(context.Background(), req)
			if tc.expPass {
				suite.NoError(err)
				suite.Equal(pool.Id, res.Pool.Id)
			} else {
				suite.Error(err)
				suite.Nil(res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCLiquidityPoolByReserveAcc() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient
	pool, found := app.LiquidityKeeper.GetPool(ctx, suite.pools[0].Id)
	suite.True(found)

	var req *types.QueryLiquidityPoolByReserveAccRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = &types.QueryLiquidityPoolByReserveAccRequest{}
			},
			false,
		},
		{
			"valid request",
			func() {
				req = &types.QueryLiquidityPoolByReserveAccRequest{ReserveAcc: suite.pools[0].ReserveAccountAddress}
			},
			true,
		},
		{
			"invalid request",
			func() {
				req = &types.QueryLiquidityPoolByReserveAccRequest{ReserveAcc: suite.pools[0].ReserveAccountAddress[:10]}
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			res, err := queryClient.LiquidityPoolByReserveAcc(context.Background(), req)
			if tc.expPass {
				suite.NoError(err)
				suite.Equal(pool.Id, res.Pool.Id)
			} else {
				suite.Error(err)
				suite.Nil(res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCQueryLiquidityPools() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient
	pools := app.LiquidityKeeper.GetAllPools(ctx)

	var req *types.QueryLiquidityPoolsRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		numPools int
		hasNext  bool
	}{
		{
			"empty request",
			func() {
				req = &types.QueryLiquidityPoolsRequest{
					Pagination: &query.PageRequest{},
				}
			},
			true,
			2,
			false,
		},
		{
			"valid request",
			func() {
				req = &types.QueryLiquidityPoolsRequest{
					Pagination: &query.PageRequest{Limit: 1, CountTotal: true},
				}
			},
			true,
			1,
			true,
		},
		{
			"valid request",
			func() {
				req = &types.QueryLiquidityPoolsRequest{
					Pagination: &query.PageRequest{Limit: 10, CountTotal: true},
				}
			},
			true,
			2,
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			resp, err := queryClient.LiquidityPools(context.Background(), req)
			if tc.expPass {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal(tc.numPools, len(resp.Pools))
				suite.Equal(uint64(len(pools)), resp.Pagination.Total)

				if tc.hasNext {
					suite.NotNil(resp.Pagination.NextKey)
				} else {
					suite.Nil(resp.Pagination.NextKey)
				}
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCLiquidityPoolBatch() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient
	batch, found := app.LiquidityKeeper.GetPoolBatch(ctx, suite.pools[0].Id)
	suite.True(found)

	var req *types.QueryLiquidityPoolBatchRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = &types.QueryLiquidityPoolBatchRequest{}
			},
			false,
		},
		{
			"valid request",
			func() {
				req = &types.QueryLiquidityPoolBatchRequest{PoolId: suite.pools[0].Id}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			res, err := queryClient.LiquidityPoolBatch(context.Background(), req)
			if tc.expPass {
				suite.NoError(err)
				suite.True(batch.Equal(&res.Batch))
			} else {
				suite.Error(err)
				suite.Nil(res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCQueryBatchDepositMsgs() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient
	msgs := app.LiquidityKeeper.GetAllPoolBatchDepositMsgs(ctx, suite.batches[0])

	var req *types.QueryPoolBatchDepositMsgsRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		numMsgs  int
		hasNext  bool
	}{
		{
			"empty request",
			func() {
				req = &types.QueryPoolBatchDepositMsgsRequest{}
			},
			false,
			0,
			false,
		},
		{
			"returns all the pool batch deposit Msgs",
			func() {
				req = &types.QueryPoolBatchDepositMsgsRequest{
					PoolId: suite.batches[0].PoolId,
				}
			},
			true,
			len(msgs),
			false,
		},
		{
			"valid request",
			func() {
				req = &types.QueryPoolBatchDepositMsgsRequest{
					PoolId:     suite.batches[0].PoolId,
					Pagination: &query.PageRequest{Limit: 1, CountTotal: true},
				}
			},
			true,
			1,
			true,
		},
		{
			"valid request",
			func() {
				req = &types.QueryPoolBatchDepositMsgsRequest{
					PoolId:     suite.batches[0].PoolId,
					Pagination: &query.PageRequest{Limit: 10, CountTotal: true},
				}
			},
			true,
			len(msgs),
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			resp, err := queryClient.PoolBatchDepositMsgs(context.Background(), req)
			if tc.expPass {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal(tc.numMsgs, len(resp.Deposits))
				suite.Equal(uint64(len(msgs)), resp.Pagination.Total)

				if tc.hasNext {
					suite.NotNil(resp.Pagination.NextKey)
				} else {
					suite.Nil(resp.Pagination.NextKey)
				}
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCQueryBatchWithdrawMsgs() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient
	msgs := app.LiquidityKeeper.GetAllPoolBatchWithdrawMsgStates(ctx, suite.batches[0])

	var req *types.QueryPoolBatchWithdrawMsgsRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		numMsgs  int
		hasNext  bool
	}{
		{
			"empty request",
			func() {
				req = &types.QueryPoolBatchWithdrawMsgsRequest{}
			},
			false,
			0,
			false,
		},
		{
			"returns all the pool batch withdraw Msgs",
			func() {
				req = &types.QueryPoolBatchWithdrawMsgsRequest{
					PoolId: suite.batches[0].PoolId,
				}
			},
			true,
			len(msgs),
			false,
		},
		{
			"valid request",
			func() {
				req = &types.QueryPoolBatchWithdrawMsgsRequest{
					PoolId:     suite.batches[0].PoolId,
					Pagination: &query.PageRequest{Limit: 1, CountTotal: true},
				}
			},
			true,
			1,
			true,
		},
		{
			"valid request",
			func() {
				req = &types.QueryPoolBatchWithdrawMsgsRequest{
					PoolId:     suite.batches[0].PoolId,
					Pagination: &query.PageRequest{Limit: 10, CountTotal: true},
				}
			},
			true,
			len(msgs),
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			resp, err := queryClient.PoolBatchWithdrawMsgs(context.Background(), req)
			if tc.expPass {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal(tc.numMsgs, len(resp.Withdraws))
				suite.Equal(uint64(len(msgs)), resp.Pagination.Total)

				if tc.hasNext {
					suite.NotNil(resp.Pagination.NextKey)
				} else {
					suite.Nil(resp.Pagination.NextKey)
				}
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCQueryBatchSwapMsgs() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient
	msgs := app.LiquidityKeeper.GetAllPoolBatchSwapMsgStatesAsPointer(ctx, suite.batches[0])

	var req *types.QueryPoolBatchSwapMsgsRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		numMsgs  int
		hasNext  bool
	}{
		{
			"empty request",
			func() {
				req = &types.QueryPoolBatchSwapMsgsRequest{}
			},
			false,
			0,
			false,
		},
		{
			"returns all the pool batch swap Msgs",
			func() {
				req = &types.QueryPoolBatchSwapMsgsRequest{
					PoolId: suite.batches[0].PoolId,
				}
			},
			true,
			len(msgs),
			false,
		},
		{
			"valid request",
			func() {
				req = &types.QueryPoolBatchSwapMsgsRequest{
					PoolId:     suite.batches[0].PoolId,
					Pagination: &query.PageRequest{Limit: 1, CountTotal: true},
				}
			},
			true,
			1,
			true,
		},
		{
			"valid request",
			func() {
				req = &types.QueryPoolBatchSwapMsgsRequest{
					PoolId:     suite.batches[0].PoolId,
					Pagination: &query.PageRequest{Limit: 10, CountTotal: true},
				}
			},
			true,
			len(msgs),
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			resp, err := queryClient.PoolBatchSwapMsgs(context.Background(), req)
			if tc.expPass {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal(tc.numMsgs, len(resp.Swaps))
				suite.Equal(uint64(len(msgs)), resp.Pagination.Total)

				if tc.hasNext {
					suite.NotNil(resp.Pagination.NextKey)
				} else {
					suite.Nil(resp.Pagination.NextKey)
				}
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
