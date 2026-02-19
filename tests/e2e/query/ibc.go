package query

import (
	"fmt"

	"github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10/types"

	"github.com/cosmos/gaia/v27/tests/e2e/common"
)

func IbcWasmChecksums(endpoint string) ([]string, error) {
	body, err := common.HTTPGet(fmt.Sprintf("%s/ibc/lightclients/wasm/v1/checksums", endpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	var response types.QueryChecksumsResponse
	if err = common.Cdc.UnmarshalJSON(body, &response); err != nil {
		return nil, err
	}

	return response.Checksums, nil
}
