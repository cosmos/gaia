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

	icamauthtypes "github.com/cosmos/gaia/v8/x/icamauth/types"
)

var (
	// TODO remove those global vars
	adminAddr   string
	aliceAddr   string
	bobAddr     string
	charlieAddr string
)

var (
	proposalId = 1
	sendAmount = sdk.NewInt64Coin(uatomDenom, 5000000)

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

const (
	groupId = iota + 1
)

const (
	originalMembersFilename  = "members1.json"
	addMemberFilename        = "members2.json"
	removeMemberFilename     = "members3.json"
	thresholdPolicyFilename  = "policy1.json"
	percentagePolicyFilename = "policy2.json"
	ICAGroupMetadata         = "ICA Group"
	ICAGroupPolicyMetadata   = "ICA Group Policy"
	thresholdPolicyMetadata  = "Policy 1"
	percentagePolicyMetadata = "Policy 2"
	proposalMsgSendPath      = "proposal1.json"
)

/*
GroupsSendMsgTest tests group lifecycle, policy creation, and proposal submission.
Test Benchmarks:
1. Create group with 3 members, including the administrator
2. Update group members to add a new member
3. Query and validate group size has increased to 4
4. Update group members to remove recently added member
5. Query and validate group size has returned to 3
6. Create threshold decision policy (threshold = 1, tally of voter weights required to pass a proposal)
7. Query and validate policy successfully created
8. Fund threshold decision policy and validate balanced has increased by expected amount
9. Submit and vote YES with sufficient threshold on a group proposal on behalf of the threshold decision policy to send tokens to Bob
10. Validate that the proposal status is PROPOSAL_STATUS_ACCEPTED
11. Execute passed proposal (execution must happen within MaxExecutionPeriod, fees are paid by executor)
12. Validate proposal has passed by verifying Bob's balance has increased by expected amount
13. Create percentage decision policy (percentage = 0.5, percentage of voting power required to pass a proposal)
14. Query and validate policy successfully created
15. Submit and vote NO with insufficient percentage on a group proposal on behalf of the percentage decision policy to send tokens to bob
16. Validate that the proposal status is PROPOSAL_STATUS_REJECTED
*/
func (s *IntegrationTestSuite) GroupsSendMsgTest() {
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	s.setupGroupsSuite()

	s.T().Logf("Creating Group")
	s.execCreateGroup(s.chainA, 0, adminAddr, "Cosmos Hub Group", filepath.Join(gaiaConfigPath, originalMembersFilename), standardFees.String())
	membersRes, err := queryGroupMembers(chainAAPIEndpoint, 1)
	s.Require().NoError(err)
	s.Assert().Equal(len(membersRes.Members), 3)

	s.T().Logf("Adding New Group Member")
	s.execUpdateGroupMembers(s.chainA, 0, adminAddr, strconv.Itoa(groupId), filepath.Join(gaiaConfigPath, addMemberFilename), standardFees.String())
	membersRes, err = queryGroupMembers(chainAAPIEndpoint, 1)
	s.Require().NoError(err)
	s.Assert().Equal(len(membersRes.Members), 4)

	s.T().Logf("Removing New Group Member")
	s.execUpdateGroupMembers(s.chainA, 0, adminAddr, strconv.Itoa(groupId), filepath.Join(gaiaConfigPath, removeMemberFilename), standardFees.String())
	membersRes, err = queryGroupMembers(chainAAPIEndpoint, 1)
	s.Require().NoError(err)
	s.Assert().Equal(len(membersRes.Members), 3)

	s.T().Logf("Creating Group Threshold Decision Policy")
	s.executeCreateGroupPolicy(s.chainA, 0, adminAddr, strconv.Itoa(groupId), thresholdPolicyMetadata, filepath.Join(gaiaConfigPath, thresholdPolicyFilename), standardFees.String())
	policies, err := queryGroupPolicies(chainAAPIEndpoint, groupId)
	s.Require().NoError(err)
	policy, err := getPolicy(policies.GroupPolicies, thresholdPolicyMetadata, groupId)
	s.Require().NoError(err)

	s.T().Logf("Funding Group Threshold Decision Policy")
	s.execBankSend(s.chainA, 0, adminAddr, policy.Address, depositAmount.String(), standardFees.String(), false)
	s.verifyBalanceChange(chainAAPIEndpoint, depositAmount, policy.Address)

	s.writeGroupProposal(s.chainA, policy.Address, adminAddr, sendAmount, proposalMsgSendPath)
	s.T().Logf("Submitting Group Proposal 1: Send 5 uatom from group to Bob")
	s.executeSubmitGroupProposal(s.chainA, 0, adminAddr, filepath.Join(gaiaConfigPath, proposalMsgSendPath))

	s.T().Logf("Voting Group Proposal 1: Send 5 uatom from group to Bob")
	s.executeVoteGroupProposal(s.chainA, 0, proposalId, adminAddr, group.VOTE_OPTION_YES.String(), "Admin votes yes")
	s.executeVoteGroupProposal(s.chainA, 1, proposalId, aliceAddr, group.VOTE_OPTION_YES.String(), "Alice votes yes")

	s.Require().Eventually(
		func() bool {
			proposalRes, err := queryGroupProposal(chainAAPIEndpoint, groupId)
			s.Require().NoError(err)

			return proposalRes.Proposal.Status == group.PROPOSAL_STATUS_ACCEPTED
		},
		30*time.Second,
		5*time.Second,
	)
	s.T().Logf("Group Proposal 1 Passed: Send 5 uatom from group to Bob")

	s.T().Logf("Executing Group Proposal 1: Send 5 uatom from group to Bob")
	s.executeExecGroupProposal(s.chainA, 1, proposalId, aliceAddr)
	s.verifyBalanceChange(chainAAPIEndpoint, sendAmount, bobAddr)

	proposalId++
	s.T().Logf("Creating Group Percentage Decision Policy")
	s.executeCreateGroupPolicy(s.chainA, 0, adminAddr, strconv.Itoa(groupId), percentagePolicyMetadata, filepath.Join(gaiaConfigPath, percentagePolicyFilename), standardFees.String())
	policies, err = queryGroupPolicies(chainAAPIEndpoint, 1)
	s.Require().NoError(err)
	policy, err = getPolicy(policies.GroupPolicies, percentagePolicyMetadata, groupId)
	s.Require().NoError(err)

	s.writeGroupProposal(s.chainA, policy.Address, adminAddr, sendAmount, proposalMsgSendPath)
	s.T().Logf("Submitting Group Proposal 2: Send 5 uatom from group to Bob")
	s.executeSubmitGroupProposal(s.chainA, 0, adminAddr, filepath.Join(gaiaConfigPath, proposalMsgSendPath))

	s.T().Logf("Voting Group Proposal 2: Send 5 uatom from group to Bob")
	s.executeVoteGroupProposal(s.chainA, 0, proposalId, adminAddr, group.VOTE_OPTION_YES.String(), "Admin votes yes")
	s.executeVoteGroupProposal(s.chainA, 1, proposalId, aliceAddr, group.VOTE_OPTION_ABSTAIN.String(), "Alice votes abstain")

	s.Require().Eventually(
		func() bool {
			proposalRes, err := queryGroupProposalByGroupPolicy(chainAAPIEndpoint, policy.Address)
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
	return policies[0], errors.New("no matching policy found")
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

	// Update weight to 0 to remove member from group
	newMembers[2].Weight = "0"
	s.writeGroupMembers(c, newMembers, removeMemberFilename)
}

func (s *IntegrationTestSuite) creatICAGroupProposal(c *chain) string {
	var (
		portID        = "1317/tcp"
		resourceChain = s.valResources[c.id][0]
		chainAPI      = fmt.Sprintf("http://%s", resourceChain.GetHostPort(portID))
	)

	policies, err := queryGroupPolicies(chainAPI, groupId)
	s.Require().NoError(err)
	policy, err := getPolicy(policies.GroupPolicies, ICAGroupPolicyMetadata, groupId)
	s.Require().NoError(err)

	registerICAMsg := &icamauthtypes.MsgRegisterAccount{
		Owner:        policy.Address,
		ConnectionId: icaConnectionID,
		Version:      icaVersion,
	}
	proposal := &group.Proposal{
		GroupPolicyAddress: policy.Address,
		Proposers:          []string{adminAddr},
		Metadata:           ICAGroupMetadata,
	}

	err = proposal.SetMsgs([]sdk.Msg{registerICAMsg})
	s.Require().NoError(err)

	body, err := cdc.MarshalJSON(proposal)
	s.Require().NoError(err)

	s.writeFile(c, ICAGroupProposal, body)

	s.T().Logf("Submitting Group ICA Proposal")
	s.executeSubmitGroupProposal(c, 0, adminAddr, filepath.Join(gaiaConfigPath, ICAGroupProposal))

	s.T().Logf("Voting Group ICA Proposal")
	s.executeVoteGroupProposal(c, 0, 1, adminAddr, group.VOTE_OPTION_YES.String(), "Admin votes yes")
	s.executeVoteGroupProposal(c, 1, 1, aliceAddr, group.VOTE_OPTION_YES.String(), "Alice votes yes")

	s.Require().Eventually(
		func() bool {
			proposalRes, err := queryGroupProposal(chainAPI, groupId)
			s.Require().NoError(err)

			return proposalRes.Proposal.Status == group.PROPOSAL_STATUS_ACCEPTED
		},
		30*time.Second,
		5*time.Second,
	)
	s.T().Logf("Group ICA Proposal Passed")

	s.T().Logf("Executing Group ICA Proposal")
	s.executeExecGroupProposal(c, 1, 1, aliceAddr)

	return policy.Address
}

func (s *IntegrationTestSuite) writeGroupProposal(c *chain, policyAddress, signingAddress string, sendAmount sdk.Coin, filename string) {
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
	err := proposal.SetMsgs(msgs)
	s.Require().NoError(err)

	body, err := cdc.MarshalJSON(proposal)
	s.Require().NoError(err)

	s.writeFile(c, filename, body)
}

func (s *IntegrationTestSuite) TestICAGroupProposal() {
	var (
		portID        = "1317/tcp"
		chain         = s.chainA
		resourceChain = s.valResources[chain.id][0]
		chainAAPI     = fmt.Sprintf("http://%s", resourceChain.GetHostPort(portID))
		chainBAPI     = fmt.Sprintf("http://%s", resourceChain.GetHostPort(portID))
	)
	s.setupGroupsSuite()

	s.T().Logf("Creating ICA Group")
	s.execCreateGroupWithPolicy(
		chain,
		0,
		adminAddr,
		ICAGroupMetadata,
		ICAGroupPolicyMetadata,
		configFile(originalMembersFilename),
		configFile(thresholdPolicyFilename),
		standardFees.String(),
	)

	owner := s.creatICAGroupProposal(chain)

	var ica string
	s.Require().Eventually(
		func() bool {
			ica, err := queryICAAddress(chainAAPI, owner, icaConnectionID)
			s.Require().NoError(err)

			return err == nil && ica != ""
		},
		time.Minute,
		5*time.Second,
	)

	// step 2: fund ica, send tokens from chain b val to ica on chain b
	senderAddr, err := s.chainB.validators[0].keyInfo.GetAddress()
	s.Require().NoError(err)
	sender := senderAddr.String()

	s.execBankSend(s.chainB, 0, sender, ica, tokenAmount.String(), standardFees.String(), false)

	s.Require().Eventually(
		func() bool {
			afterSenderICABalance, err := getSpecificBalance(chainBAPI, ica, uatomDenom)
			s.Require().NoError(err)
			return afterSenderICABalance.IsEqual(tokenAmount)
		},
		time.Minute,
		5*time.Second,
	)
}

func (s *IntegrationTestSuite) writeGroupPolicies(
	c *chain,
	thresholdFilename,
	percentageFilename string,
	thresholdPolicy *group.ThresholdDecisionPolicy,
	percentagePolicy *group.PercentageDecisionPolicy,
) {
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
	s.writeGroupPolicies(
		s.chainA,
		thresholdPolicyFilename,
		percentagePolicyFilename,
		thresholdPolicy,
		percentagePolicy,
	)
}
