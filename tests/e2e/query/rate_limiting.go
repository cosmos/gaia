package query

import (
	"fmt"

	"github.com/cosmos/ibc-apps/modules/rate-limiting/v10/types"

	"github.com/cosmos/gaia/v24/tests/e2e/common"
)

func AllRateLimits(endpoint string) ([]types.RateLimit, error) {
	var res types.QueryAllRateLimitsResponse

	body, err := common.HTTPGet(fmt.Sprintf("%s/Stride-Labs/ibc-rate-limiting/ratelimit/ratelimits", endpoint))
	if err != nil {
		return []types.RateLimit{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	if err := common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return []types.RateLimit{}, err
	}
	return res.RateLimits, nil
}

func RateLimit(endpoint, channelID, denom string) (types.QueryRateLimitResponse, error) {
	var res types.QueryRateLimitResponse

	body, err := common.HTTPGet(fmt.Sprintf("%s/Stride-Labs/ibc-rate-limiting/ratelimit/ratelimit/%s/by_denom?denom=%s", endpoint, channelID, denom))
	if err != nil {
		return types.QueryRateLimitResponse{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	if err := common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return types.QueryRateLimitResponse{}, err
	}
	return res, nil
}

func RateLimitsByChainID(endpoint, channelID string) ([]types.RateLimit, error) {
	var res types.QueryRateLimitsByChainIdResponse

	body, err := common.HTTPGet(fmt.Sprintf("%s/Stride-Labs/ibc-rate-limiting/ratelimit/ratelimits/%s", endpoint, channelID))
	if err != nil {
		return []types.RateLimit{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	if err := common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return []types.RateLimit{}, err
	}
	return res.RateLimits, nil
}
