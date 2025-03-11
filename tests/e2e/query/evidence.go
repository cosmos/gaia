package query

import (
	"cosmossdk.io/x/evidence/types"
	"fmt"
	"github.com/cosmos/gaia/v23/tests/e2e/common"
)

func queryEvidence(endpoint, hash string) (types.QueryEvidenceResponse, error) { //nolint:unused // this is called during e2e tests
	var res types.QueryEvidenceResponse
	body, err := common.HttpGet(fmt.Sprintf("%s/cosmos/evidence/v1beta1/evidence/%s", endpoint, hash))
	if err != nil {
		return res, err
	}

	if err = common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

func QueryAllEvidence(endpoint string) (types.QueryAllEvidenceResponse, error) {
	var res types.QueryAllEvidenceResponse
	body, err := common.HttpGet(fmt.Sprintf("%s/cosmos/evidence/v1beta1/evidence", endpoint))
	if err != nil {
		return res, err
	}

	if err = common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

func ExecQueryEvidence(endpoint, hash string) (types.Equivocation, error) { // vlad todo: look into fixing this unmarshalling
	_, err := common.HttpGet(fmt.Sprintf("%s/cosmos/evidence/v1beta1/evidence/%s", endpoint, hash))
	if err != nil {
		return types.Equivocation{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	//var response evidencetypes.QueryEvidenceResponse
	//if err = common.Cdc.UnmarshalJSON(body, &response); err != nil {
	//	return evidencetypes.Equivocation{}, err
	//}

	var evidence types.Equivocation
	//err = common.Cdc.UnpackAny(response.Evidence, &evidence)
	//if err != nil {
	//	return evidencetypes.Equivocation{}, err
	//}

	return evidence, nil
}
