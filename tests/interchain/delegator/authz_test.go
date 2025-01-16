package delegator_test

import (
	"context"
	"fmt"
	"path"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/gaia/v23/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v23/tests/interchain/delegator"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const txAmount = 1_000_000_000

func txAmountUatom() string {
	return fmt.Sprintf("%d%s", txAmount, chainsuite.Uatom)
}

type AuthSuite struct {
	*delegator.Suite
}

func (s *AuthSuite) TestSend() {
	balanceBefore, err := s.Chain.GetBalance(s.GetContext(), s.DelegatorWallet3.FormattedAddress(), chainsuite.Uatom)
	s.Require().NoError(err)
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.FormattedAddress(),
		"authz", "grant", s.DelegatorWallet2.FormattedAddress(), "send",
		"--spend-limit", fmt.Sprintf("%d%s", txAmount*2, chainsuite.Uatom),
		"--allow-list", s.DelegatorWallet3.FormattedAddress(),
	)
	s.Require().NoError(err)

	s.Require().Error(s.authzGenExec(s.GetContext(), s.DelegatorWallet2, "bank", "send", s.DelegatorWallet.FormattedAddress(), s.DelegatorWallet2.FormattedAddress(), txAmountUatom()))

	s.Require().NoError(s.authzGenExec(s.GetContext(), s.DelegatorWallet2, "bank", "send", s.DelegatorWallet.FormattedAddress(), s.DelegatorWallet3.FormattedAddress(), txAmountUatom()))
	balanceAfter, err := s.Chain.GetBalance(s.GetContext(), s.DelegatorWallet3.FormattedAddress(), chainsuite.Uatom)
	s.Require().NoError(err)
	s.Require().Equal(balanceBefore.Add(sdkmath.NewInt(int64(txAmount))), balanceAfter)

	s.Require().Error(s.authzGenExec(s.GetContext(), s.DelegatorWallet2, "bank", "send", s.DelegatorWallet.FormattedAddress(), s.DelegatorWallet3.FormattedAddress(), fmt.Sprintf("%d%s", txAmount+200, chainsuite.Uatom)))

	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.FormattedAddress(),
		"authz", "revoke", s.DelegatorWallet2.FormattedAddress(), "/cosmos.bank.v1beta1.MsgSend",
	)
	s.Require().NoError(err)

	s.Require().Error(s.authzGenExec(s.GetContext(), s.DelegatorWallet2, "bank", "send", s.DelegatorWallet.FormattedAddress(), s.DelegatorWallet3.FormattedAddress(), txAmountUatom()))
}

func (s *AuthSuite) TestDelegate() {
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.FormattedAddress(),
		"authz", "grant", s.DelegatorWallet2.FormattedAddress(), "delegate",
		"--allowed-validators", s.Chain.ValidatorWallets[0].ValoperAddress,
	)
	s.Require().NoError(err)

	s.Require().NoError(s.authzGenExec(s.GetContext(), s.DelegatorWallet2, "staking", "delegate", s.Chain.ValidatorWallets[0].ValoperAddress, txAmountUatom(), "--from", s.DelegatorWallet.FormattedAddress()))
	s.Require().Error(s.authzGenExec(s.GetContext(), s.DelegatorWallet2, "staking", "delegate", s.Chain.ValidatorWallets[1].ValoperAddress, txAmountUatom(), "--from", s.DelegatorWallet.FormattedAddress()))

	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.FormattedAddress(),
		"authz", "revoke", s.DelegatorWallet2.FormattedAddress(), "/cosmos.staking.v1beta1.MsgDelegate",
	)
	s.Require().NoError(err)
	s.Require().Error(s.authzGenExec(s.GetContext(), s.DelegatorWallet2, "staking", "delegate", s.Chain.ValidatorWallets[0].ValoperAddress, txAmountUatom(), "--from", s.DelegatorWallet.FormattedAddress()))
}

func (s *AuthSuite) TestUnbond() {
	valHex, err := s.Chain.GetValidatorHex(s.GetContext(), 0)
	s.Require().NoError(err)
	powerBefore, err := s.Chain.GetValidatorPower(s.GetContext(), valHex)
	s.Require().NoError(err)
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.FormattedAddress(),
		"staking", "delegate", s.Chain.ValidatorWallets[0].ValoperAddress, txAmountUatom(),
	)
	s.Require().NoError(err)
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		powerAfter, err := s.Chain.GetValidatorPower(s.GetContext(), valHex)
		s.Require().NoError(err)
		assert.NoError(c, err)
		assert.Greater(c, powerAfter, powerBefore)
	}, 15*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.FormattedAddress(),
		"authz", "grant", s.DelegatorWallet2.FormattedAddress(), "unbond",
		"--allowed-validators", s.Chain.ValidatorWallets[0].ValoperAddress,
	)
	s.Require().NoError(err)

	s.Require().NoError(s.authzGenExec(s.GetContext(), s.DelegatorWallet2, "staking", "unbond", s.Chain.ValidatorWallets[0].ValoperAddress, txAmountUatom(), "--from", s.DelegatorWallet.FormattedAddress()))
	s.Require().Error(s.authzGenExec(s.GetContext(), s.DelegatorWallet2, "staking", "unbond", s.Chain.ValidatorWallets[1].ValoperAddress, txAmountUatom(), "--from", s.DelegatorWallet.FormattedAddress()))

	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		powerAfter, err := s.Chain.GetValidatorPower(s.GetContext(), valHex)
		s.Require().NoError(err)
		assert.NoError(c, err)
		assert.Equal(c, powerAfter, powerBefore)
	}, 15*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.FormattedAddress(),
		"authz", "revoke", s.DelegatorWallet2.FormattedAddress(), "/cosmos.staking.v1beta1.MsgUndelegate",
	)
	s.Require().NoError(err)
	s.Require().Error(s.authzGenExec(s.GetContext(), s.DelegatorWallet2, "staking", "unbond", s.Chain.ValidatorWallets[0].ValoperAddress, txAmountUatom(), "--from", s.DelegatorWallet.FormattedAddress()))
}

func (s AuthSuite) TestRedelegate() {
	val0Hex, err := s.Chain.GetValidatorHex(s.GetContext(), 0)
	s.Require().NoError(err)
	val2Hex, err := s.Chain.GetValidatorHex(s.GetContext(), 1)
	s.Require().NoError(err)
	val0PowerBefore, err := s.Chain.GetValidatorPower(s.GetContext(), val0Hex)
	s.Require().NoError(err)
	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.Chain.ValidatorWallets[0].Moniker,
		"staking", "delegate", s.Chain.ValidatorWallets[0].ValoperAddress, txAmountUatom(),
	)
	s.Require().NoError(err)
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		val0PowerAfter, err := s.Chain.GetValidatorPower(s.GetContext(), val0Hex)
		s.Require().NoError(err)
		s.Require().NoError(err)
		s.Require().Greater(val0PowerAfter, val0PowerBefore)
	}, 15*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.Chain.ValidatorWallets[0].Address,
		"authz", "grant", s.DelegatorWallet2.FormattedAddress(), "redelegate",
		"--allowed-validators", s.Chain.ValidatorWallets[1].ValoperAddress,
	)
	s.Require().NoError(err)

	s.Require().Error(s.authzGenExec(s.GetContext(), s.DelegatorWallet2, "staking", "redelegate", s.Chain.ValidatorWallets[1].ValoperAddress, s.Chain.ValidatorWallets[0].ValoperAddress, txAmountUatom(), "--from", s.Chain.ValidatorWallets[0].Address))

	val2PowerBefore, err := s.Chain.GetValidatorPower(s.GetContext(), val2Hex)
	s.Require().NoError(err)
	s.Require().NoError(s.authzGenExec(s.GetContext(), s.DelegatorWallet2, "staking", "redelegate", s.Chain.ValidatorWallets[0].ValoperAddress, s.Chain.ValidatorWallets[1].ValoperAddress, txAmountUatom(), "--from", s.Chain.ValidatorWallets[0].Address))
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		val2PowerAfter, err := s.Chain.GetValidatorPower(s.GetContext(), val2Hex)
		s.Require().NoError(err)
		s.Require().Greater(val2PowerAfter, val2PowerBefore)
	}, 15*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	_, err = s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.Chain.ValidatorWallets[0].Address,
		"authz", "revoke", s.DelegatorWallet2.FormattedAddress(), "/cosmos.staking.v1beta1.MsgBeginRedelegate",
	)
	s.Require().NoError(err)

	s.Require().Error(s.authzGenExec(s.GetContext(), s.DelegatorWallet2, "staking", "redelegate", s.Chain.ValidatorWallets[0].ValoperAddress, s.Chain.ValidatorWallets[1].ValoperAddress, txAmountUatom(), "--from", s.Chain.ValidatorWallets[0].Address))
}

func (s AuthSuite) TestGeneric() {
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.Chain.ValidatorWallets[0].Moniker,
		"authz", "grant", s.DelegatorWallet.FormattedAddress(), "generic",
		"--msg-type", "/cosmos.gov.v1.MsgVote",
	)
	s.Require().NoError(err)

	prop, err := s.Chain.BuildProposal(nil, "Test Proposal", "Test Proposal", "ipfs://CID", chainsuite.GovDepositAmount, "", false)
	s.Require().NoError(err)
	result, err := s.Chain.SubmitProposal(s.GetContext(), s.DelegatorWallet.FormattedAddress(), prop)
	s.Require().NoError(err)
	s.Require().NoError(s.authzGenExec(s.GetContext(), s.DelegatorWallet, "gov", "vote", result.ProposalID, "yes", "--from", s.Chain.ValidatorWallets[0].Address))
}

func TestAuthz(t *testing.T) {
	two := 2
	s := &AuthSuite{Suite: &delegator.Suite{Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
		UpgradeOnSetup: true,
		ChainSpec: &interchaintest.ChainSpec{
			NumValidators: &two,
		},
	})}}
	suite.Run(t, s)
}

func (s AuthSuite) authzGenExec(ctx context.Context, grantee ibc.Wallet, command ...string) error {
	txjson, err := s.Chain.GenerateTx(ctx, 1, command...)
	s.Require().NoError(err)

	err = s.Chain.GetNode().WriteFile(ctx, []byte(txjson), "tx.json")
	s.Require().NoError(err)

	_, err = s.Chain.GetNode().ExecTx(
		ctx,
		grantee.FormattedAddress(),
		"authz", "exec", path.Join(s.Chain.Validators[1].HomeDir(), "tx.json"),
	)
	return err
}
