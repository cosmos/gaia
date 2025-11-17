package delegator_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tidwall/sjson"

	"github.com/cosmos/gaia/v26/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v26/tests/interchain/delegator"
)

type AuthSuite struct {
	*delegator.Suite
}

func (s *AuthSuite) SetupSuite() {
	s.Suite.SetupSuite()
	// Delegate >1 ATOM with delegator account
	node := s.Chain.GetNode()
	node.StakingDelegate(s.GetContext(), s.DelegatorWallet.KeyName(), s.Chain.ValidatorWallets[0].ValoperAddress, string(govStakeAmount)+s.Chain.Config().Denom)
	node.StakingDelegate(s.GetContext(), s.DelegatorWallet2.KeyName(), s.Chain.ValidatorWallets[0].ValoperAddress, string(govStakeAmount)+s.Chain.Config().Denom)
}

func (s *AuthSuite) TestParamChange() {
	authParams, err := s.Chain.QueryJSON(s.GetContext(), "params", "auth", "params")
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Params: %s", authParams)
	currentMemoLimit := authParams.Get("max_memo_characters").Int()
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Current max memo characters: %d", currentMemoLimit)
	newLimit := int64(512)

	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Submitting transaction with more than %d characters (must fail).", currentMemoLimit)
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.FormattedAddress(),
		"bank", "send",
		s.DelegatorWallet.FormattedAddress(), s.DelegatorWallet2.FormattedAddress(), txAmountUatom(), "--note=7c97b5bdfefff77ecd8d8ffa2be7e09a9e8b0c599accdae5a74d132c7c19c9c08127742d2ab6547d202be9dc96ef13bb131f482670967bf0de765792462171de524c8f1509b2d1d7ce7f2c473b30b71e4bfe41a85e7f78d02846dfc2f7ae31da29585e8b39547215d143e772ba5be11bbe896e98f3f196dfa2b1a37dc17c3cd4fdd3",
	)
	s.Require().Error(err)

	authority, err := s.Chain.GetGovernanceAddress(s.GetContext())
	s.Require().NoError(err)

	updatedParams, err := sjson.Set(authParams.String(), "max_memo_characters", fmt.Sprintf("%d", newLimit))
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Updated params: %s", updatedParams)

	paramChangeMessage := fmt.Sprintf(`{
		"@type": "/cosmos.auth.v1beta1.MsgUpdateParams",
		"authority": "%s",
		"params": %s
	}`, authority, updatedParams)

	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Message: %s", paramChangeMessage)

	// Submit proposal
	prop, err := s.Chain.BuildProposal(nil, "Auth Param Change Proposal", "Test Proposal", "ipfs://CID", chainsuite.GovDepositAmount, s.DelegatorWallet.KeyName(), false)
	s.Require().NoError(err)
	prop.Messages = []json.RawMessage{json.RawMessage(paramChangeMessage)}
	result, err := s.Chain.SubmitProposal(s.GetContext(), s.DelegatorWallet.KeyName(), prop)
	s.Require().NoError(err)
	proposalId := result.ProposalID

	json, _, err := s.Chain.GetNode().ExecQuery(s.GetContext(), "gov", "proposal", proposalId)
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("%s", string(json))

	// Pass proposal
	s.Require().NoError(s.Chain.PassProposal(s.GetContext(), proposalId))

	// Test
	authParams, err = s.Chain.QueryJSON(s.GetContext(), "params", "auth", "params")
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Params: %s", authParams)
	currentMemoLimit = authParams.Get("max_memo_characters").Int()
	s.Require().Equal(newLimit, currentMemoLimit)

	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Submitting transaction with more than %d characters (must pass).", currentMemoLimit)
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.FormattedAddress(),
		"bank", "send",
		s.DelegatorWallet.FormattedAddress(), s.DelegatorWallet2.FormattedAddress(), txAmountUatom(), "--note=7c97b5bdfefff77ecd8d8ffa2be7e09a9e8b0c599accdae5a74d132c7c19c9c08127742d2ab6547d202be9dc96ef13bb131f482670967bf0de765792462171de524c8f1509b2d1d7ce7f2c473b30b71e4bfe41a85e7f78d02846dfc2f7ae31da29585e8b39547215d143e772ba5be11bbe896e98f3f196dfa2b1a37dc17c3cd4fdd3",
	)
	s.Require().NoError(err)
}

func TestAuthModule(t *testing.T) {
	s := &AuthSuite{Suite: &delegator.Suite{Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
		UpgradeOnSetup: true,
	})}}
	suite.Run(t, s)
}
