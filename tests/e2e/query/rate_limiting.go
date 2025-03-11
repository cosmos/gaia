package query

import (
	"fmt"
	"github.com/cosmos/gaia/v23/tests/e2e/common"
	"github.com/cosmos/ibc-apps/modules/rate-limiting/v10/types"
)

func QueryAllRateLimits(endpoint string) ([]types.RateLimit, error) {
	var res types.QueryAllRateLimitsResponse

	body, err := common.HttpGet(fmt.Sprintf("%s/Stride-Labs/ibc-rate-limiting/ratelimit/ratelimits", endpoint))
	if err != nil {
		return []types.RateLimit{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	if err := common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return []types.RateLimit{}, err
	}
	return res.RateLimits, nil
}

func QueryRateLimit(endpoint, channelID, denom string) (types.QueryRateLimitResponse, error) {
	var res types.QueryRateLimitResponse

	body, err := common.HttpGet(fmt.Sprintf("%s/Stride-Labs/ibc-rate-limiting/ratelimit/ratelimit/%s/by_denom?denom=%s", endpoint, channelID, denom))
	if err != nil {
		return types.QueryRateLimitResponse{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	if err := common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return types.QueryRateLimitResponse{}, err
	}
	return res, nil
}

func QueryRateLimitsByChainID(endpoint, channelID string) ([]types.RateLimit, error) {
	var res types.QueryRateLimitsByChainIdResponse

	body, err := common.HttpGet(fmt.Sprintf("%s/Stride-Labs/ibc-rate-limiting/ratelimit/ratelimits/%s", endpoint, channelID))
	if err != nil {
		return []types.RateLimit{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	if err := common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return []types.RateLimit{}, err
	}
	return res.RateLimits, nil
}
