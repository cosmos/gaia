package delegator_test

import (
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/cosmos/gaia/v24/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v24/tests/interchain/delegator"
	"github.com/stretchr/testify/suite"
)

type SigningTest struct {
	*delegator.Suite
}

// Color codes
var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Magenta = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[37m"
var White = "\033[97m"

func (s *SigningTest) TestBatchSign() {
	WalletbalanceBefore, err := s.Chain.GetBalance(s.GetContext(), s.DelegatorWallet.FormattedAddress(), chainsuite.Uatom)
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof(Yellow+"Starting balances of DelegatorWallet: %s"+Reset, WalletbalanceBefore)
	Wallet2balanceBefore, err := s.Chain.GetBalance(s.GetContext(), s.DelegatorWallet2.FormattedAddress(), chainsuite.Uatom)
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof(Yellow+"Starting balances of DelegatorWallet2: %s"+Reset, Wallet2balanceBefore)
	Wallet3balanceBefore, err := s.Chain.GetBalance(s.GetContext(), s.DelegatorWallet3.FormattedAddress(), chainsuite.Uatom)
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof(Yellow+"Starting balances of DelegatorWallet3: %s"+Reset, Wallet3balanceBefore)

	// Generate tx for wallet 2
	const txAmountWallet2 = txAmount + 2
	txjsonWallet2, err := s.Chain.GenerateTx(
		s.GetContext(), 0, "bank", "send", s.DelegatorWallet.FormattedAddress(), s.DelegatorWallet2.FormattedAddress(), fmt.Sprintf("%d%s", txAmountWallet2, chainsuite.Uatom), "--gas-prices", s.Chain.Config().GasPrices,
	)
	s.Require().NoError(err)
	err = s.Chain.GetNode().WriteFile(s.GetContext(), []byte(txjsonWallet2), "tx-wallet2.json")
	s.Require().NoError(err)

	// Generate tx for wallet 3
	const txAmountWallet3 = txAmount + 3
	txjsonWallet3, err := s.Chain.GenerateTx(
		s.GetContext(), 0, "bank", "send", s.DelegatorWallet.FormattedAddress(), s.DelegatorWallet3.FormattedAddress(), fmt.Sprintf("%d%s", txAmountWallet3, chainsuite.Uatom), "--gas-prices", s.Chain.Config().GasPrices,
	)
	s.Require().NoError(err)
	err = s.Chain.GetNode().WriteFile(s.GetContext(), []byte(txjsonWallet3), "tx-wallet3.json")
	s.Require().NoError(err)

	// Batch sign test
	signed0, _, err := s.Chain.GetNode().Exec(s.GetContext(),
		s.Chain.GetNode().TxCommand(s.DelegatorWallet.KeyName(),
			"sign-batch",
			path.Join(s.Chain.GetNode().HomeDir(), "tx-wallet2.json"),
			path.Join(s.Chain.GetNode().HomeDir(), "tx-wallet3.json"),
			"--append",
		), nil)
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Signed transection %s", string(signed0))
	err = s.Chain.GetNode().WriteFile(s.GetContext(), []byte(signed0), "signed0.json")
	s.Require().NoError(err)
	// validate-signatures
	signinvstdout, signinvstderr, err := s.Chain.GetNode().Exec(s.GetContext(),
		s.Chain.GetNode().TxCommand(s.DelegatorWallet.KeyName(),
			"validate-signatures",
			path.Join(s.Chain.GetNode().HomeDir(), "tx-wallet2.json"),
		), nil)
	s.Require().Error(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof(Magenta+"Sign invalid std err: %s"+Reset, signinvstderr)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof(Magenta+"Sign invalid std out: %s"+Reset, signinvstdout)
	signstdout, _, err := s.Chain.GetNode().Exec(s.GetContext(),
		s.Chain.GetNode().TxCommand(s.DelegatorWallet.KeyName(),
			"validate-signatures",
			path.Join(s.Chain.GetNode().HomeDir(), "signed0.json"),
		), nil)
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof(Magenta+"Verify signed std out: %s"+Reset, signstdout)

	// Send tx
	batchsign, _, err := s.Chain.GetNode().Exec(s.GetContext(),
		s.Chain.GetNode().TxCommand(s.DelegatorWallet.KeyName(),
			"broadcast",
			path.Join(s.Chain.GetNode().HomeDir(), "signed0.json"),
		), nil)
	s.Require().NoError(err)
	err = s.Chain.GetNode().WriteFile(s.GetContext(), []byte(signed0), "signed0.json")
	s.Require().NoError(err)
	time.Sleep(20 * time.Second)
	err = s.Chain.GetNode().WriteFile(s.GetContext(), batchsign, "batchsign.json")
	s.Require().NoError(err)

	// Verify balances
	Wallet1balanceAfter, err := s.Chain.GetBalance(s.GetContext(), s.DelegatorWallet.FormattedAddress(), chainsuite.Uatom)
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof(Yellow+"Ending balances of DelegatorWallet: %s"+Reset, Wallet1balanceAfter)

	Wallet2balanceAfter, err := s.Chain.GetBalance(s.GetContext(), s.DelegatorWallet2.FormattedAddress(), chainsuite.Uatom)
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof(Yellow+"Ending balances of DelegatorWallet2: %s"+Reset, Wallet2balanceAfter)
	balanceDifferenceWallet2 := int(Wallet2balanceAfter.Uint64() - Wallet2balanceBefore.Uint64())
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof(Yellow+"DelegatorWallet2 difference: %d"+Reset, balanceDifferenceWallet2)
	s.Require().Equal(txAmountWallet2, balanceDifferenceWallet2)

	Wallet3balanceAfter, err := s.Chain.GetBalance(s.GetContext(), s.DelegatorWallet3.FormattedAddress(), chainsuite.Uatom)
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof(Yellow+"Ending balances of DelegatorWallet3: %s"+Reset, Wallet3balanceAfter)
	balanceDifferenceWallet3 := int(Wallet3balanceAfter.Uint64() - Wallet3balanceBefore.Uint64())
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof(Yellow+"DelegatorWallet3 difference: %d"+Reset, balanceDifferenceWallet3)
	s.Require().Equal(txAmountWallet3, balanceDifferenceWallet3)
}

func TestBatchSign(t *testing.T) {
	s := &SigningTest{Suite: &delegator.Suite{Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
		UpgradeOnSetup: true,
	})}}
	suite.Run(t, s)
}
