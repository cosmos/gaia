package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/group"
)

func (s *IntegrationTestSuite) queryGovProposal(endpoint string, proposalId uint64) (govv1beta1.QueryProposalResponse, error) {
	var emptyProp govv1beta1.QueryProposalResponse

	path := fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals/%d", endpoint, proposalId)
	resp, err := http.Get(path)
	if err != nil {
		return emptyProp, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return emptyProp, err
	}
	var govProposalResp govv1beta1.QueryProposalResponse

	if err := cdc.UnmarshalJSON(body, &govProposalResp); err != nil {
		return emptyProp, err
	}
	s.T().Logf("This is the gov response: %s", govProposalResp)

	return govProposalResp, nil
}

func (s *IntegrationTestSuite) getLatestBlockHeight(c *chain, valIdx int) int {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	type syncInfo struct {
		SyncInfo struct {
			LatestHeight string `json:"latest_block_height"`
		} `json:"SyncInfo"`
	}

	var currentHeight int
	gaiaCommand := []string{gaiadBinary, "status"}
	s.executeGaiaTxCommand(ctx, c, gaiaCommand, valIdx, func(stdOut []byte, stdErr []byte) bool {
		var (
			err   error
			block syncInfo
		)
		s.Require().NoError(json.Unmarshal(stdErr, &block))
		currentHeight, err = strconv.Atoi(block.SyncInfo.LatestHeight)
		s.Require().NoError(err)
		return currentHeight > 0
	})
	return currentHeight
}

func (s *IntegrationTestSuite) queryGroupMembers(endpoint string, groupId int) (group.QueryGroupMembersResponse, error) {
	var res group.QueryGroupMembersResponse
	path := fmt.Sprintf("%s/cosmos/group/v1/group_members/%d", endpoint, groupId)

	resp, err := http.Get(path)
	if err != nil {
		return res, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	if err := cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}

	return res, nil
}

func (s *IntegrationTestSuite) queryGroupInfo(endpoint string, groupId int) (group.QueryGroupInfoResponse, error) {
	var res group.QueryGroupInfoResponse
	path := fmt.Sprintf("%s/cosmos/group/v1/group_info/%d", endpoint, groupId)

	resp, err := http.Get(path)
	if err != nil {
		return res, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	if err := cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}

	return res, nil
}

func (s *IntegrationTestSuite) queryGroupsbyAdmin(endpoint string, adminAddress string) (group.QueryGroupsByAdminResponse, error) {
	var res group.QueryGroupsByAdminResponse
	path := fmt.Sprintf("%s/cosmos/group/v1/groups_by_admin/%s", endpoint, adminAddress)

	resp, err := http.Get(path)
	if err != nil {
		return res, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	if err := cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}

	return res, nil
}

func (s *IntegrationTestSuite) queryGroupPolicies(endpoint string, groupId int) (group.QueryGroupPoliciesByGroupResponse, error) {
	var res group.QueryGroupPoliciesByGroupResponse
	path := fmt.Sprintf("%s/cosmos/group/v1/group_policies_by_group/%d", endpoint, groupId)

	resp, err := http.Get(path)
	if err != nil {
		return res, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	if err := cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}

	return res, nil
}

func (s *IntegrationTestSuite) queryGroupProposal(endpoint string, groupId int) (group.QueryProposalResponse, error) {
	var res group.QueryProposalResponse
	path := fmt.Sprintf("%s/cosmos/group/v1/proposal/%d", endpoint, groupId)

	resp, err := http.Get(path)
	if err != nil {
		return res, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	if err := cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}

	return res, nil
}

func (s *IntegrationTestSuite) queryGroupProposalByGroupPolicy(endpoint string, policyAddress string) (group.QueryProposalsByGroupPolicyResponse, error) {
	var res group.QueryProposalsByGroupPolicyResponse
	path := fmt.Sprintf("%s/cosmos/group/v1/proposals_by_group_policy/%s", endpoint, policyAddress)

	resp, err := http.Get(path)
	if err != nil {
		return res, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	if err := cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}

	return res, nil
}

func (s *IntegrationTestSuite) verifyBalanceChange(endpoint string, expectedAmount sdk.Coin, recipientAddress string) {
	s.Require().Eventually(
		func() bool {
			afterAtomBalance, err := getSpecificBalance(endpoint, recipientAddress, uatomDenom)
			s.Require().NoError(err)

			return afterAtomBalance.IsEqual(expectedAmount)
		},
		20*time.Second,
		5*time.Second,
	)
}
