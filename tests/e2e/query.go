package e2e

import (
	"fmt"
	"io"
	"net/http"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func queryValidator(endpoint, address string) (stakingtypes.Validator, error) {
	var res stakingtypes.QueryValidatorResponse

	url := fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s", endpoint, address)
	resp, err := http.Get(url)
	if err != nil {
		return stakingtypes.Validator{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return stakingtypes.Validator{}, err
	}
	if err := cdc.UnmarshalJSON(bz, &res); err != nil {
		return stakingtypes.Validator{}, err
	}
	return res.Validator, nil
}

func queryValidators(endpoint string) (stakingtypes.Validators, error) {
	var res stakingtypes.QueryValidatorsResponse
	url := fmt.Sprintf("%s/cosmos/staking/v1beta1/validators", endpoint)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := cdc.UnmarshalJSON(bz, &res); err != nil {
		return nil, err
	}
	return res.Validators, nil
}
