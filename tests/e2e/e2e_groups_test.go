package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	group "github.com/cosmos/cosmos-sdk/x/group"
)

type PolicyRequest struct {
	Type      string
	Threshold string
	Windows   group.DecisionPolicyWindows
}

var (
	aliceAddr                string
	bobAddr                  string
	charlieAddr              string
	members                  []GroupMember
	originalMembersFilename  = "members1.json"
	addMemberFilename        = "members2.json"
	removeMemberFilename     = "members3.json"
	thresholdPolicyFilename  = "policy1.json"
	percentagePolicyFilename = "policy2.json"

	proposalMsgSendPath = "proposal1.json"
	sendAmount          = sdk.NewInt64Coin(photonDenom, 5000000)
	windows             = DecisionPolicyWindow{
		MinExecutionPeriod: (0 * time.Second).String(),
		VotingPeriod:       (30 * time.Second).String(),
	}

	thresholdPolicy = ThresholdPolicy{
		Type:      "/cosmos.group.v1.ThresholdDecisionPolicy",
		Threshold: "1",
		Windows:   windows,
	}

	percentagePolicy = PercentagePolicy{
		Type:       "/cosmos.group.v1.PercentageDecisionPolicy",
		Percentage: "0.5",
		Windows:    windows,
	}
)

func (s *IntegrationTestSuite) TestGroupsSendMsg() {
	thresholdPolicyMetadata := "Policy 1"
	percentagePolicyMetadata := "Policy 2"
	groupId := 1
	proposalId := 1

	var res group.QueryGroupMembersResponse
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	adminAddr, err := s.chainA.validators[0].keyInfo.GetAddress()
	s.Require().NoError(err)

	val2, err := s.chainA.validators[1].keyInfo.GetAddress()
	s.Require().NoError(err)
	aliceAddr = val2.String()

	bobAddr = s.executeGaiaKeysAddCommand(ctx, s.chainA, 0, "bob")
	charlieAddr = s.executeGaiaKeysAddCommand(ctx, s.chainA, 0, "charlie")

	s.prepareGroupFiles(ctx, s.chainA, adminAddr.String(), aliceAddr, bobAddr, charlieAddr)

	s.T().Logf("Creating Group")
	s.execCreateGroup(s.chainA, 0, chainAAPIEndpoint, adminAddr.String(), "Cosmos Hub Group", fmt.Sprintf("/root/.gaia/config/%s", originalMembersFilename), fees.String())
	res, err = s.queryGroupMembers(chainAAPIEndpoint, 1)
	s.Require().NoError(err)
	s.Assert().Equal(len(res.Members), len(members))

	s.T().Logf("Adding New Group Member")
	s.execUpdateGroupMembers(s.chainA, 0, chainAAPIEndpoint, adminAddr.String(), strconv.Itoa(groupId), fmt.Sprintf("/root/.gaia/config/%s", addMemberFilename), fees.String())
	res, err = s.queryGroupMembers(chainAAPIEndpoint, 1)
	s.Require().NoError(err)
	s.Assert().Equal(len(res.Members), len(members)+1)

	s.T().Logf("Removing New Group Member")
	s.execUpdateGroupMembers(s.chainA, 0, chainAAPIEndpoint, adminAddr.String(), strconv.Itoa(groupId), fmt.Sprintf("/root/.gaia/config/%s", removeMemberFilename), fees.String())
	res, err = s.queryGroupMembers(chainAAPIEndpoint, 1)
	s.Require().NoError(err)
	s.Assert().Equal(len(res.Members), len(members))

	s.T().Logf("Creating Group Threshold Decision Policy")
	s.writeGroupPolicies(s.chainA, thresholdPolicyFilename, percentagePolicyFilename, thresholdPolicy, percentagePolicy)
	s.executeCreateGroupPolicy(s.chainA, 0, chainAAPIEndpoint, adminAddr.String(), strconv.Itoa(groupId), thresholdPolicyMetadata, fmt.Sprintf("/root/.gaia/config/%s", thresholdPolicyFilename), fees.String())
	policies, err := s.queryGroupPolicies(chainAAPIEndpoint, groupId)
	s.Require().NoError(err)
	policy, err := getPolicy(policies.GroupPolicies, thresholdPolicyMetadata, groupId)
	s.Require().NoError(err)

	s.T().Logf("Funding Group Threshold Decision Policy")
	s.sendMsgSend(s.chainA, 0, adminAddr.String(), policy.Address, depositAmount.String(), fees.String())
	s.verifyBalanceChange(chainAAPIEndpoint, depositAmount, policy.Address)

	s.writeGroupProposal(s.chainA, policy.Address, adminAddr.String(), sendAmount, proposalMsgSendPath)
	s.T().Logf("Submitting Group Proposal 1: Send 5 photon from group to Bob")
	s.executeSubmitGroupProposal(s.chainA, 0, chainAAPIEndpoint, adminAddr.String(), fmt.Sprintf("/root/.gaia/config/%s", proposalMsgSendPath))

	s.T().Logf("Voting Group Proposal 1: Send 5 photon from group to Bob")
	s.executeVoteGroupProposal(s.chainA, 0, chainAAPIEndpoint, strconv.Itoa(proposalId), adminAddr.String(), group.VOTE_OPTION_YES.String(), "Admin votes yes")
	s.executeVoteGroupProposal(s.chainA, 1, chainAAPIEndpoint, strconv.Itoa(proposalId), aliceAddr, group.VOTE_OPTION_YES.String(), "Admin votes yes")

	s.Require().Eventually(
		func() bool {
			proposalRes, err := s.queryGroupProposal(chainAAPIEndpoint, 1)
			s.Require().NoError(err)

			return proposalRes.Proposal.Status == group.PROPOSAL_STATUS_ACCEPTED
		},
		30*time.Second,
		5*time.Second,
	)
	s.T().Logf("Group Proposal 1 Passed: Send 5 photon from group to Bob")

	s.T().Logf("Executing Group Proposal 1: Send 5 photon from group to Bob")
	s.executeExecGroupProposal(s.chainA, 1, chainAAPIEndpoint, strconv.Itoa(proposalId), aliceAddr)
	s.verifyBalanceChange(chainAAPIEndpoint, sendAmount, bobAddr)

	proposalId++
	s.T().Logf("Creating Group Percentage Decision Policy")
	s.executeCreateGroupPolicy(s.chainA, 0, chainAAPIEndpoint, adminAddr.String(), strconv.Itoa(groupId), percentagePolicyMetadata, fmt.Sprintf("/root/.gaia/config/%s", percentagePolicyFilename), fees.String())
	policies, err = s.queryGroupPolicies(chainAAPIEndpoint, 1)
	s.Require().NoError(err)
	policy, err = getPolicy(policies.GroupPolicies, percentagePolicyMetadata, 1)
	s.Require().NoError(err)

	s.writeGroupProposal(s.chainA, policy.Address, adminAddr.String(), sendAmount, proposalMsgSendPath)
	s.T().Logf("Submitting Group Proposal 2: Send 5 photon from group to Bob")
	s.executeSubmitGroupProposal(s.chainA, 0, chainAAPIEndpoint, adminAddr.String(), fmt.Sprintf("/root/.gaia/config/%s", proposalMsgSendPath))

	s.T().Logf("Voting Group Proposal 2: Send 5 photon from group to Bob")
	s.executeVoteGroupProposal(s.chainA, 0, chainAAPIEndpoint, strconv.Itoa(proposalId), adminAddr.String(), group.VOTE_OPTION_YES.String(), "Admin votes yes")
	s.executeVoteGroupProposal(s.chainA, 1, chainAAPIEndpoint, strconv.Itoa(proposalId), aliceAddr, group.VOTE_OPTION_ABSTAIN.String(), "Admin votes yes")

	s.Require().Eventually(
		func() bool {
			proposalRes, err := s.queryGroupProposalByGroupPolicy(chainAAPIEndpoint, policy.Address)
			s.Require().NoError(err)

			return proposalRes.Proposals[0].Status == group.PROPOSAL_STATUS_REJECTED
		},
		30*time.Second,
		5*time.Second,
	)
	s.T().Logf("Group Proposal Rejected: Send 5 photon from group to Bob")
}

func (s *IntegrationTestSuite) verifyBalanceChange(endpoint string, expectedAmount types.Coin, recipientAddress string) {
	s.Require().Eventually(
		func() bool {
			afterPhotonBalance, err := getSpecificBalance(endpoint, recipientAddress, "photon")
			s.Require().NoError(err)

			return afterPhotonBalance.IsEqual(expectedAmount)
		},
		20*time.Second,
		5*time.Second,
	)
}

func getPolicy(policies []*group.GroupPolicyInfo, metadata string, groupId int) (*group.GroupPolicyInfo, error) {
	for _, p := range policies {
		if p.Metadata == metadata && p.GroupId == uint64(groupId) {
			return p, nil
		}
	}
	return policies[0], errors.New("No matching policy found")
}

func (s *IntegrationTestSuite) prepareGroupFiles(ctx context.Context, c *chain, adminAddr string, member1Address string, member2Address string, member3Address string) {

	members = []GroupMember{
		{
			Address:  adminAddr,
			Weight:   "1",
			Metadata: "Admin",
		},
		{
			Address:  member1Address,
			Weight:   "1",
			Metadata: "Alice",
		},
		{
			Address:  member2Address,
			Weight:   "1",
			Metadata: "Bob",
		},
	}
	s.writeGroupMembers(c, members, originalMembersFilename)

	newMembers := append(members, GroupMember{
		Address:  member3Address,
		Weight:   "1",
		Metadata: "Charlie",
	})
	s.writeGroupMembers(c, newMembers, addMemberFilename)

	removeMembers := append(members, GroupMember{
		Address:  charlieAddr,
		Weight:   "0",
		Metadata: "Charlie",
	})

	s.writeGroupMembers(c, removeMembers, removeMemberFilename)
}

func (s *IntegrationTestSuite) writeGroupProposal(c *chain, policyAddress string, signingAddress string, sendAmount sdk.Coin, filename string) {
	message := MsgSend{
		Type:   "/cosmos.bank.v1beta1.MsgSend",
		From:   policyAddress,
		To:     bobAddr,
		Amount: []sdk.Coin{sendAmount},
	}
	prop := struct {
		GroupPolicyAddress string    `json:"group_policy_address"`
		Proposers          []string  `json:"proposers"`
		Metadata           string    `json:"metadata"`
		Messages           []MsgSend `json:"messages"`
	}{
		GroupPolicyAddress: policyAddress,
		Proposers:          []string{signingAddress},
		Metadata:           "Send 5photon to Bob",
		Messages:           []MsgSend{message},
	}

	body, err := json.MarshalIndent(prop, "", " ")
	s.Require().NoError(err)

	s.writeFile(c, filename, body)
}

func (s *IntegrationTestSuite) writeGroupPolicies(c *chain, thresholdFilename string, percentageFilename string, thresholdPolicyJson ThresholdPolicy, percentagePolicyJson PercentagePolicy) {

	thresholdBody, err := json.MarshalIndent(thresholdPolicyJson, "", " ")
	s.Require().NoError(err)

	percentageBody, err := json.MarshalIndent(percentagePolicyJson, "", " ")
	s.Require().NoError(err)

	s.writeFile(c, thresholdFilename, thresholdBody)
	s.writeFile(c, percentageFilename, percentageBody)

}
