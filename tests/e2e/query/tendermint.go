package query

import (
	"fmt"

	tendermintv1beta1 "cosmossdk.io/api/cosmos/base/tendermint/v1beta1"

	"github.com/cosmos/gaia/v26/tests/e2e/common"
)

func GetLatestBlockHeight(endpoint string) (int, error) {
	body, err := common.HTTPGet(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/blocks/latest", endpoint))
	if err != nil {
		return 0, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	var response tendermintv1beta1.GetLatestBlockResponse
	if err = common.Cdc.UnmarshalJSON(body, &response); err != nil {
		return 0, err
	}
	return int(response.GetBlock().GetLastCommit().Height), nil
}
