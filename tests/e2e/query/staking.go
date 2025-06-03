package query

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/gaia/v24/tests/e2e/common"
)

func StakingParams(endpoint string) (types.QueryParamsResponse, error) {
	body, err := common.HTTPGet(fmt.Sprintf("%s/cosmos/staking/v1beta1/params", endpoint))
	if err != nil {
		return types.QueryParamsResponse{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	var params types.QueryParamsResponse
	if err := common.Cdc.UnmarshalJSON(body, &params); err != nil {
		return types.QueryParamsResponse{}, err
	}

	return params, nil
}

func Delegation(endpoint string, validatorAddr string, delegatorAddr string) (types.QueryDelegationResponse, error) {
	var res types.QueryDelegationResponse

	body, err := common.HTTPGet(fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations/%s", endpoint, validatorAddr, delegatorAddr))
	if err != nil {
		return res, err
	}

	if err = common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

func UnbondingDelegation(endpoint string, validatorAddr string, delegatorAddr string) (types.QueryUnbondingDelegationResponse, error) {
	var res types.QueryUnbondingDelegationResponse
	body, err := common.HTTPGet(fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations/%s/unbonding_delegation", endpoint, validatorAddr, delegatorAddr))
	if err != nil {
		return res, err
	}

	if err = common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

func Validator(endpoint, address string) (types.Validator, error) {
	var res types.QueryValidatorResponse

	body, err := common.HTTPGet(fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s", endpoint, address))
	if err != nil {
		return types.Validator{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	if err := common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return types.Validator{}, err
	}
	return res.Validator, nil
}

func Validators(endpoint string) (types.Validators, error) {
	var res types.QueryValidatorsResponse
	body, err := common.HTTPGet(fmt.Sprintf("%s/cosmos/staking/v1beta1/validators", endpoint))
	if err != nil {
		return types.Validators{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	if err := common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return types.Validators{}, err
	}

	return types.Validators{Validators: res.Validators}, nil
}

func Pool(endpoint string) (types.Pool, error) {
	var res types.QueryPoolResponse
	body, err := common.HTTPGet(fmt.Sprintf("%s/cosmos/staking/v1beta1/pool", endpoint))
	if err != nil {
		return types.Pool{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	if err := common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return types.Pool{}, err
	}

	return res.Pool, nil
}
