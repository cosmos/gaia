package query

import (
	"fmt"
	"io"
	"net/http"

	"github.com/cosmos/gaia/v27/tests/e2e/common"
	liquidtypes "github.com/cosmos/gaia/v27/x/liquid/types"
)

func LiquidValidator(endpoint string, valAddr string) (liquidtypes.QueryLiquidValidatorResponse, error) {
	resp, err := http.Get(fmt.Sprintf("%s/gaia/liquid/v1beta1/liquid_validator/%s", endpoint, valAddr))
	if err != nil {
		return liquidtypes.QueryLiquidValidatorResponse{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return liquidtypes.QueryLiquidValidatorResponse{}, err
	}

	var lvr liquidtypes.QueryLiquidValidatorResponse
	if err := common.Cdc.UnmarshalJSON(bz, &lvr); err != nil {
		return liquidtypes.QueryLiquidValidatorResponse{}, err
	}

	return lvr, nil
}

func LiquidParams(endpoint string) (liquidtypes.QueryParamsResponse, error) {
	resp, err := http.Get(fmt.Sprintf("%s/gaia/liquid/v1beta1/params", endpoint))
	if err != nil {
		return liquidtypes.QueryParamsResponse{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return liquidtypes.QueryParamsResponse{}, err
	}

	var params liquidtypes.QueryParamsResponse
	if err := common.Cdc.UnmarshalJSON(bz, &params); err != nil {
		return liquidtypes.QueryParamsResponse{}, err
	}

	return params, nil
}

func TokenizeShareRecordByID(endpoint string, recordID int) (liquidtypes.TokenizeShareRecord, error) {
	var res liquidtypes.QueryTokenizeShareRecordByIdResponse

	body, err := common.HTTPGet(fmt.Sprintf("%s/gaia/liquid/v1beta1/tokenize_share_record_by_id/%d", endpoint, recordID))
	if err != nil {
		return liquidtypes.TokenizeShareRecord{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	if err := common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return liquidtypes.TokenizeShareRecord{}, err
	}
	return res.Record, nil
}

func TotalLiquidStaked(endpoint string) (string, error) {
	body, err := common.HTTPGet(fmt.Sprintf("%s/gaia/liquid/v1beta1/total_liquid_staked", endpoint))
	if err != nil {
		return "", fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	var resp liquidtypes.QueryTotalLiquidStakedResponse
	if err := common.Cdc.UnmarshalJSON(body, &resp); err != nil {
		return "", err
	}

	return resp.Tokens, nil
}

func LastTokenizeShareRecordID(endpoint string) (uint64, error) {
	var res liquidtypes.QueryLastTokenizeShareRecordIdResponse
	body, err := common.HTTPGet(fmt.Sprintf("%s/gaia/liquid/v1beta1/last_tokenize_share_record_id", endpoint))
	if err != nil {
		return 0, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	if err := common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return 0, err
	}
	return res.Id, nil
}
