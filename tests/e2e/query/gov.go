package query

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/gaia/v23/tests/e2e/common"
)

func QueryGovProposal(endpoint string, proposalID int) (v1beta1.QueryProposalResponse, error) {
	var govProposalResp v1beta1.QueryProposalResponse

	path := fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals/%d", endpoint, proposalID)

	body, err := common.HttpGet(path)
	if err != nil {
		return govProposalResp, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	if err := common.Cdc.UnmarshalJSON(body, &govProposalResp); err != nil {
		return govProposalResp, err
	}

	return govProposalResp, nil
}

func QueryGovProposalV1(endpoint string, proposalID int) (v1.QueryProposalResponse, error) {
	var govProposalResp v1.QueryProposalResponse

	path := fmt.Sprintf("%s/cosmos/gov/v1/proposals/%d", endpoint, proposalID)

	body, err := common.HttpGet(path)
	if err != nil {
		return govProposalResp, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	if err := common.Cdc.UnmarshalJSON(body, &govProposalResp); err != nil {
		return govProposalResp, err
	}

	return govProposalResp, nil
}
