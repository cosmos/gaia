package delegator_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/cosmos/gaia/v27/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v27/tests/interchain/delegator"
	"github.com/stretchr/testify/suite"
)

type MultiSendSuite struct {
	*delegator.Suite
}

func (s *MultiSendSuite) TestMultiSendGasCurve() {
	sender := s.DelegatorWallet.FormattedAddress()
	dest1 := s.DelegatorWallet2.FormattedAddress()
	amount := "100uatom"

	gas_factor := 300
	// Expected gas increase: A * (N^2 - M^2) where A=300, N=10, M=5
	expectedIncrease := gas_factor * (10*10 - 5*5) // 300 * (100 - 25) = 300 * 75 = 22500

	// Execute two multi-send transactions with different recipient counts and compare gas usage.

	// 1. Send with 5 recipients
	txHash5, err := s.execMultiSend(sender, dest1, amount, 5)
	s.Require().NoError(err)

	gasUsed5, err := s.queryTxGas(txHash5)
	s.Require().NoError(err)

	// 2. Send with 10 recipients
	txHash10, err := s.execMultiSend(sender, dest1, amount, 10)
	s.Require().NoError(err)

	gasUsed10, err := s.queryTxGas(txHash10)
	s.Require().NoError(err)

	fmt.Printf("Gas Used (5 recipients): %d\n", gasUsed5)
	fmt.Printf("Gas Used (10 recipients): %d\n", gasUsed10)

	s.Require().Greater(gasUsed10, gasUsed5+int64(expectedIncrease), "Gas usage should enhance quadratically with recipients")
}

func (s *MultiSendSuite) TestRecipientLimit() {
	sender := s.DelegatorWallet.FormattedAddress()
	dest1 := s.DelegatorWallet2.FormattedAddress()
	amount := "1uatom"

	// Max is 500. Try 501.
	args := []string{"bank", "multi-send", sender}
	for range 501 {
		args = append(args, dest1)
	}
	args = append(args, amount, "--gas", "auto", "--gas-adjustment", "1.5")

	_, err := s.Chain.GetNode().ExecTx(s.GetContext(), sender, args...)
	s.Require().Error(err, "Should have failed with too many recipients")
}

func (s *MultiSendSuite) execMultiSend(sender, dest, amount string, recipients int) (string, error) {
	args := []string{"bank", "multi-send", sender}
	for range recipients {
		args = append(args, dest)
	}
	args = append(args, amount, "--gas", "auto", "--gas-adjustment", "1.5")

	return s.Chain.GetNode().ExecTx(s.GetContext(), sender, args...)
}

func (s *MultiSendSuite) queryTxGas(txHash string) (int64, error) {
	// ChainNode.ExecQuery returns (stdout, stderr, err) based on chain.go
	stdout, _, err := s.Chain.GetNode().ExecQuery(s.GetContext(), "tx", txHash, "--output", "json")
	if err != nil {
		return 0, err
	}

	var respMap map[string]any
	if err := json.Unmarshal(stdout, &respMap); err != nil {
		return 0, err
	}

	if val, ok := respMap["gas_used"]; ok {
		switch v := val.(type) {
		case string:
			return strconv.ParseInt(v, 10, 64)
		case float64:
			return int64(v), nil
		}
	}
	return 0, fmt.Errorf("gas_used not found in tx query response")
}

func TestMultiSend(t *testing.T) {
	s := &MultiSendSuite{Suite: &delegator.Suite{Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
		UpgradeOnSetup: true,
	})}}
	suite.Run(t, s)
}
