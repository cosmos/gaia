package types

// DONTCOVER
// client is excluded from test coverage in the poc phase milestone 1 and will be included in milestone 2 with completeness

// QueryLiquidityPool liquidity query endpoint supported by the liquidity querier
const (
	QueryLiquidityPool  = "liquidityPool"
	QueryLiquidityPools = "liquidityPools"
)

// QueryLiquidityPoolParams is the query parameters for 'custom/liquidity'
type QueryLiquidityPoolParams struct {
	PoolId uint64 `json:"pool_id" yaml:"pool_id"` //nolint:revive
}

// return params of Liquidity Pool Query
func NewQueryLiquidityPoolParams(poolID uint64) QueryLiquidityPoolParams {
	return QueryLiquidityPoolParams{
		PoolId: poolID,
	}
}

// QueryValidatorsParams defines the params for the following queries:
// - 'custom/liquidity/liquidityPools'
type QueryLiquidityPoolsParams struct {
	Page, Limit int
}

// return params of Liquidity Pools Query
func NewQueryLiquidityPoolsParams(page, limit int) QueryLiquidityPoolsParams {
	return QueryLiquidityPoolsParams{page, limit}
}
