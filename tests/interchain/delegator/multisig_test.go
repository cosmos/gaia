package delegator_test

import (
	"fmt"
	"path"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/gaia/v23/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v23/tests/interchain/delegator"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/suite"
)

type MultisigTest struct {
	*delegator.Suite
}

func (s *MultisigTest) TestMultisig() {
	multisigName := "multisig"
	_, _, err := s.Chain.GetNode().ExecBin(
		s.GetContext(), "keys", "add", multisigName,
		"--multisig", fmt.Sprintf("%s,%s,%s", s.DelegatorWallet.KeyName(), s.DelegatorWallet2.KeyName(), s.DelegatorWallet3.KeyName()),
		"--multisig-threshold", "2", "--keyring-backend", "test",
	)
	s.Require().NoError(err)
	bogusWallet, err := s.Chain.BuildWallet(s.GetContext(), "bogus", "")
	s.Require().NoError(err)

	multisigAddr, err := s.Chain.GetNode().KeyBech32(s.GetContext(), multisigName, "")
	s.Require().NoError(err)

	err = s.Chain.SendFunds(s.GetContext(), interchaintest.FaucetAccountKeyName, ibc.WalletAmount{
		Denom:   chainsuite.Uatom,
		Amount:  sdkmath.NewInt(chainsuite.ValidatorFunds),
		Address: multisigAddr,
	})
	s.Require().NoError(err)

	balanceBefore, err := s.Chain.GetBalance(s.GetContext(), s.DelegatorWallet3.FormattedAddress(), chainsuite.Uatom)
	s.Require().NoError(err)

	txjson, err := s.Chain.GenerateTx(
		s.GetContext(), 0, "bank", "send", multisigName, s.DelegatorWallet3.FormattedAddress(), txAmountUatom(),
		"--gas", "auto", "--gas-adjustment", fmt.Sprint(s.Chain.Config().GasAdjustment), "--gas-prices", s.Chain.Config().GasPrices,
	)
	s.Require().NoError(err)

	err = s.Chain.GetNode().WriteFile(s.GetContext(), []byte(txjson), "tx.json")
	s.Require().NoError(err)

	signed0, _, err := s.Chain.GetNode().Exec(s.GetContext(),
		s.Chain.GetNode().TxCommand(s.DelegatorWallet.KeyName(),
			"sign",
			path.Join(s.Chain.GetNode().HomeDir(), "tx.json"),
			"--multisig", multisigAddr,
		), nil)
	s.Require().NoError(err)

	signed1, _, err := s.Chain.GetNode().Exec(s.GetContext(),
		s.Chain.GetNode().TxCommand(s.DelegatorWallet2.KeyName(),
			"sign",
			path.Join(s.Chain.GetNode().HomeDir(), "tx.json"),
			"--multisig", multisigAddr,
		), nil)
	s.Require().NoError(err)

	_, _, err = s.Chain.GetNode().Exec(s.GetContext(),
		s.Chain.GetNode().TxCommand(bogusWallet.KeyName(),
			"sign",
			path.Join(s.Chain.GetNode().HomeDir(), "tx.json"),
			"--multisig", multisigAddr,
		), nil)
	s.Require().Error(err)

	err = s.Chain.GetNode().WriteFile(s.GetContext(), signed0, "signed0.json")
	s.Require().NoError(err)
	err = s.Chain.GetNode().WriteFile(s.GetContext(), signed1, "signed1.json")
	s.Require().NoError(err)

	multisign, _, err := s.Chain.GetNode().Exec(s.GetContext(), s.Chain.GetNode().TxCommand(
		multisigName,
		"multisign",
		path.Join(s.Chain.GetNode().HomeDir(), "tx.json"),
		multisigName,
		path.Join(s.Chain.GetNode().HomeDir(), "signed0.json"),
		path.Join(s.Chain.GetNode().HomeDir(), "signed1.json"),
	), nil)
	s.Require().NoError(err)

	err = s.Chain.GetNode().WriteFile(s.GetContext(), multisign, "multisign.json")
	s.Require().NoError(err)

	_, err = s.Chain.GetNode().ExecTx(s.GetContext(), multisigName, "broadcast", path.Join(s.Chain.GetNode().HomeDir(), "multisign.json"))
	s.Require().NoError(err)

	balanceAfter, err := s.Chain.GetBalance(s.GetContext(), s.DelegatorWallet3.FormattedAddress(), chainsuite.Uatom)
	s.Require().NoError(err)
	s.Require().Equal(balanceBefore.Add(sdkmath.NewInt(txAmount)), balanceAfter)
}

func TestMultisig(t *testing.T) {
	s := &MultisigTest{Suite: &delegator.Suite{Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
		UpgradeOnSetup: true,
	})}}
	suite.Run(t, s)
}
