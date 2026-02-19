package query

import (
	"fmt"

	"github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/cosmos/gaia/v27/tests/e2e/common"
)

func WasmContractAddress(endpoint, creator string, idx uint64) (string, error) {
	body, err := common.HTTPGet(fmt.Sprintf("%s/cosmwasm/wasm/v1/contracts/creator/%s", endpoint, creator))
	if err != nil {
		return "", fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	var response types.QueryContractsByCreatorResponse
	if err = common.Cdc.UnmarshalJSON(body, &response); err != nil {
		return "", err
	}

	return response.ContractAddresses[idx], nil
}

func WasmSmartContractState(endpoint, address, msg string) ([]byte, error) {
	body, err := common.HTTPGet(fmt.Sprintf("%s/cosmwasm/wasm/v1/contract/%s/smart/%s", endpoint, address, msg))
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	var response types.QuerySmartContractStateResponse
	if err = common.Cdc.UnmarshalJSON(body, &response); err != nil {
		return nil, err
	}

	return response.Data, nil
}
