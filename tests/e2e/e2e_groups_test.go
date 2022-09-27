package e2e

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/group"
)

var (
	adminAddr                string
	aliceAddr                string
	bobAddr                  string
	charlieAddr              string
	groupId                  = 1
	proposalId               = 1
	originalMembersFilename  = "members1.json"
	addMemberFilename        = "members2.json"
	removeMemberFilename     = "members3.json"
	thresholdPolicyFilename  = "policy1.json"
	percentagePolicyFilename = "policy2.json"
	thresholdPolicyMetadata  = "Policy 1"
	percentagePolicyMetadata = "Policy 2"
	proposalMsgSendPath      = "proposal1.json"
	sendAmount               = sdk.NewInt64Coin(uatomDenom, 5000000)

	windows = &group.DecisionPolicyWindows{
		MinExecutionPeriod: 0 * time.Second,
		VotingPeriod:       30 * time.Second,
	}

	thresholdPolicy = &group.ThresholdDecisionPolicy{
		Threshold: "1",
		Windows:   windows,
	}

	percentagePolicy = &group.PercentageDecisionPolicy{
		Percentage: "0.5",
		Windows:    windows,
	}
)

func (s *IntegrationTestSuite) GroupsSendMsgTest() {
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	s.setupGroupsSuite()

	s.T().Logf("Creating Group")
	s.execCreateGroup(s.chainA, 0, adminAddr, "Cosmos Hub Group", filepath.Join(gaiaConfigPath, originalMembersFilename), fees.String())
	membersRes, err := s.queryGroupMembers(chainAAPIEndpoint, 1)
	s.Require().NoError(err)
	s.Assert().Equal(len(membersRes.Members), 3)

	s.T().Logf("Adding New Group Member")
	s.execUpdateGroupMembers(s.chainA, 0, adminAddr, strconv.Itoa(groupId), filepath.Join(gaiaConfigPath, addMemberFilename), fees.String())
	membersRes, err = s.queryGroupMembers(chainAAPIEndpoint, 1)
	s.Require().NoError(err)
	s.Assert().Equal(len(membersRes.Members), 4)

	s.T().Logf("Removing New Group Member")
	s.execUpdateGroupMembers(s.chainA, 0, adminAddr, strconv.Itoa(groupId), filepath.Join(gaiaConfigPath, removeMemberFilename), fees.String())
	membersRes, err = s.queryGroupMembers(chainAAPIEndpoint, 1)
	s.Require().NoError(err)
	s.Assert().Equal(len(membersRes.Members), 3)

	s.T().Logf("Creating Group Threshold Decision Policy")
	s.writeGroupPolicies(s.chainA, thresholdPolicyFilename, percentagePolicyFilename, thresholdPolicy, percentagePolicy)
	s.executeCreateGroupPolicy(s.chainA, 0, adminAddr, strconv.Itoa(groupId), thresholdPolicyMetadata, filepath.Join(gaiaConfigPath, thresholdPolicyFilename), fees.String())
	policies, err := s.queryGroupPolicies(chainAAPIEndpoint, groupId)
	s.Require().NoError(err)
	policy, err := getPolicy(policies.GroupPolicies, thresholdPolicyMetadata, groupId)
	s.Require().NoError(err)

	s.T().Logf("Funding Group Threshold Decision Policy")
	s.execBankSend(s.chainA, 0, adminAddr, policy.Address, depositAmount.String(), fees.String(), false)
	s.verifyBalanceChange(chainAAPIEndpoint, depositAmount, policy.Address)

	s.writeGroupProposal(s.chainA, policy.Address, adminAddr, sendAmount, proposalMsgSendPath)
	s.T().Logf("Submitting Group Proposal 1: Send 5 uatom from group to Bob")
	s.executeSubmitGroupProposal(s.chainA, 0, adminAddr, filepath.Join(gaiaConfigPath, proposalMsgSendPath))

	s.T().Logf("Voting Group Proposal 1: Send 5 uatom from group to Bob")
	s.executeVoteGroupProposal(s.chainA, 0, strconv.Itoa(proposalId), adminAddr, group.VOTE_OPTION_YES.String(), "Admin votes yes")
	s.executeVoteGroupProposal(s.chainA, 1, strconv.Itoa(proposalId), aliceAddr, group.VOTE_OPTION_YES.String(), "Alice votes yes")

	s.Require().Eventually(
		func() bool {
			proposalRes, err := s.queryGroupProposal(chainAAPIEndpoint, 1)
			s.Require().NoError(err)

			return proposalRes.Proposal.Status == group.PROPOSAL_STATUS_ACCEPTED
		},
		30*time.Second,
		5*time.Second,
	)
	s.T().Logf("Group Proposal 1 Passed: Send 5 uatom from group to Bob")

	s.T().Logf("Executing Group Proposal 1: Send 5 uatom from group to Bob")
	s.executeExecGroupProposal(s.chainA, 1, strconv.Itoa(proposalId), aliceAddr)
	s.verifyBalanceChange(chainAAPIEndpoint, sendAmount, bobAddr)

	proposalId++
	s.T().Logf("Creating Group Percentage Decision Policy")
	s.executeCreateGroupPolicy(s.chainA, 0, adminAddr, strconv.Itoa(groupId), percentagePolicyMetadata, filepath.Join(gaiaConfigPath, percentagePolicyFilename), fees.String())
	policies, err = s.queryGroupPolicies(chainAAPIEndpoint, 1)
	s.Require().NoError(err)
	policy, err = getPolicy(policies.GroupPolicies, percentagePolicyMetadata, 1)
	s.Require().NoError(err)

	s.writeGroupProposal(s.chainA, policy.Address, adminAddr, sendAmount, proposalMsgSendPath)
	s.T().Logf("Submitting Group Proposal 2: Send 5 uatom from group to Bob")
	s.executeSubmitGroupProposal(s.chainA, 0, adminAddr, filepath.Join(gaiaConfigPath, proposalMsgSendPath))

	s.T().Logf("Voting Group Proposal 2: Send 5 uatom from group to Bob")
	s.executeVoteGroupProposal(s.chainA, 0, strconv.Itoa(proposalId), adminAddr, group.VOTE_OPTION_YES.String(), "Admin votes yes")
	s.executeVoteGroupProposal(s.chainA, 1, strconv.Itoa(proposalId), aliceAddr, group.VOTE_OPTION_ABSTAIN.String(), "Alice votes abstain")

	s.Require().Eventually(
		func() bool {
			proposalRes, err := s.queryGroupProposalByGroupPolicy(chainAAPIEndpoint, policy.Address)
			s.Require().NoError(err)

			return proposalRes.Proposals[0].Status == group.PROPOSAL_STATUS_REJECTED
		},
		30*time.Second,
		5*time.Second,
	)
	s.T().Logf("Group Proposal Rejected: Send 5 uatom from group to Bob")
}

func getPolicy(policies []*group.GroupPolicyInfo, metadata string, groupId int) (*group.GroupPolicyInfo, error) {
	for _, p := range policies {
		if p.Metadata == metadata && p.GroupId == uint64(groupId) {
			return p, nil
		}
	}
	return policies[0], errors.New("No matching policy found")
}

func (s *IntegrationTestSuite) prepareGroupFiles(c *chain, adminAddr string, member1Address string, member2Address string, member3Address string) {
	members := []group.MemberRequest{
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

	newMembers := append(members, group.MemberRequest{
		Address:  member3Address,
		Weight:   "1",
		Metadata: "Charlie",
	})

	s.writeGroupMembers(c, newMembers, addMemberFilename)

	removeMembers := append(members, group.MemberRequest{
		Address:  charlieAddr,
		Weight:   "0",
		Metadata: "Charlie",
	})

	s.writeGroupMembers(c, removeMembers, removeMemberFilename)
}

func (s *IntegrationTestSuite) writeGroupProposal(c *chain, policyAddress string, signingAddress string, sendAmount sdk.Coin, filename string) {
	msg := &banktypes.MsgSend{
		FromAddress: policyAddress,
		ToAddress:   bobAddr,
		Amount:      []sdk.Coin{sendAmount},
	}

	proposal := &group.Proposal{
		GroupPolicyAddress: policyAddress,
		Proposers:          []string{signingAddress},
		Metadata:           "Send 5uatom to Bob",
	}

	msgs := []sdk.Msg{msg}
	proposal.SetMsgs(msgs)

	body, err := cdc.MarshalJSON(proposal)
	s.Require().NoError(err)

	s.writeFile(c, filename, body)
}

func (s *IntegrationTestSuite) writeGroupPolicies(c *chain, thresholdFilename string, percentageFilename string, thresholdPolicy *group.ThresholdDecisionPolicy, percentagePolicy *group.PercentageDecisionPolicy) {
	thresholdBody, err := cdc.MarshalInterfaceJSON(thresholdPolicy)
	s.Require().NoError(err)

	percentageBody, err := cdc.MarshalInterfaceJSON(percentagePolicy)
	s.Require().NoError(err)

	s.writeFile(c, thresholdFilename, thresholdBody)
	s.writeFile(c, percentageFilename, percentageBody)
}

func (s *IntegrationTestSuite) setupGroupsSuite() {
	admin, err := s.chainA.validators[0].keyInfo.GetAddress()
	s.Require().NoError(err)
	adminAddr = admin.String()

	alice, err := s.chainA.validators[1].keyInfo.GetAddress()
	s.Require().NoError(err)
	aliceAddr = alice.String()

	bobAddr = s.executeGKeysAddCommand(s.chainA, 0, "bob", gaiaHomePath)
	charlieAddr = s.executeGKeysAddCommand(s.chainA, 0, "charlie", gaiaHomePath)

	s.prepareGroupFiles(s.chainA, adminAddr, aliceAddr, bobAddr, charlieAddr)
}
