package query

import (
	"fmt"

	"cosmossdk.io/x/evidence/exported"
	"cosmossdk.io/x/evidence/types"

	"github.com/cosmos/gaia/v27/tests/e2e/common"
)

func evidence(endpoint, hash string) (types.QueryEvidenceResponse, error) { //nolint:unused // this is called during e2e tests
	var res types.QueryEvidenceResponse
	body, err := common.HTTPGet(fmt.Sprintf("%s/cosmos/evidence/v1beta1/evidence/%s", endpoint, hash))
	if err != nil {
		return res, err
	}

	if err = common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

func AllEvidence(endpoint string) (types.QueryAllEvidenceResponse, error) {
	var res types.QueryAllEvidenceResponse
	body, err := common.HTTPGet(fmt.Sprintf("%s/cosmos/evidence/v1beta1/evidence", endpoint))
	if err != nil {
		return res, err
	}

	if err = common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

func ExecQueryEvidence(endpoint, hash string) (types.Equivocation, error) { // vlad todo: look into fixing this unmarshalling
	body, err := common.HTTPGet(fmt.Sprintf("%s/cosmos/evidence/v1beta1/evidence/%s", endpoint, hash))
	if err != nil {
		return types.Equivocation{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	var response types.QueryEvidenceResponse
	if err = common.Cdc.UnmarshalJSON(body, &response); err != nil {
		return types.Equivocation{}, err
	}

	var evidence exported.Evidence
	err = common.Cdc.UnpackAny(response.Evidence, &evidence)
	if err != nil {
		return types.Equivocation{}, err
	}

	eq, ok := evidence.(*types.Equivocation)
	if !ok {
		return types.Equivocation{}, fmt.Errorf("evidence is not an Equivocation")
	}

	return *eq, nil
}
