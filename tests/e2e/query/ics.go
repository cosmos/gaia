package query

import (
	"fmt"
	"github.com/cosmos/gaia/v23/tests/e2e/common"
	"github.com/cosmos/interchain-security/v7/x/ccv/provider/types"
)

func QueryBlocksPerEpoch(endpoint string) (int64, error) {
	body, err := common.HttpGet(fmt.Sprintf("%s/interchain_security/ccv/provider/params", endpoint))
	if err != nil {
		return 0, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	var response types.QueryParamsResponse
	if err = common.Cdc.UnmarshalJSON(body, &response); err != nil {
		return 0, err
	}

	return response.Params.BlocksPerEpoch, nil
}
