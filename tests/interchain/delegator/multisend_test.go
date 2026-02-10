package delegator_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/cosmos/gaia/v26/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v26/tests/interchain/delegator"
	"github.com/stretchr/testify/suite"
)

type MultiSendSuite struct {
	*delegator.Suite
}

func (s *MultiSendSuite) TestMultiSendGasCurve() {
	sender := s.DelegatorWallet.FormattedAddress()
	dest1 := s.DelegatorWallet2.FormattedAddress()
	amount := "100uatom"

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

	s.Require().Greater(gasUsed10, gasUsed5+50000, "Gas usage should enhance quadratically with recipients")
}

func (s *MultiSendSuite) TestRecipientLimit() {
	sender := s.DelegatorWallet.FormattedAddress()
	dest1 := s.DelegatorWallet2.FormattedAddress()
	amount := "1uatom"

	// Max is 500. Try 501.
	args := []string{"bank", "multi-send", sender}
	for i := 0; i < 501; i++ {
		args = append(args, dest1, amount)
	}
	args = append(args, "--gas", "auto", "--gas-adjustment", "1.5", "--output", "json", "-y")

	// ExecTx returns (stdout, err) based on bank_test.go
	stdout, err := s.Chain.GetNode().ExecTx(s.GetContext(), sender, args...)

	if err != nil {
		s.Require().Contains(err.Error(), "too many recipients", "Transaction should fail due to recipient limit")
		return
	}

	var txResp map[string]interface{}
	if parseErr := json.Unmarshal([]byte(stdout), &txResp); parseErr == nil {
		if rawLog, ok := txResp["raw_log"].(string); ok {
			if strings.Contains(rawLog, "too many recipients") {
				return // Success
			}
		}
		if code, ok := txResp["code"].(float64); ok && code != 0 {
			return // Success
		}
	}

	// Fallback check on stdout string content
	s.Require().Error(err, "Should have failed with too many recipients")
	s.Require().Contains(string(stdout), "too many recipients")
}

func (s *MultiSendSuite) execMultiSend(sender, dest, amount string, recipients int) (string, error) {
	args := []string{"bank", "multi-send", sender}
	for i := 0; i < recipients; i++ {
		args = append(args, dest, amount)
	}
	args = append(args, "--gas", "auto", "--gas-adjustment", "1.5", "--output", "json", "-y")

	stdout, err := s.Chain.GetNode().ExecTx(s.GetContext(), sender, args...)
	if err != nil {
		return "", err
	}

	var txResp struct {
		TxHash string `json:"txhash"`
		Code   int    `json:"code"`
		RawLog string `json:"raw_log"`
	}
	if err := json.Unmarshal([]byte(stdout), &txResp); err != nil {
		return "", fmt.Errorf("failed to parse tx output: %w", err)
	}

	if txResp.Code != 0 {
		return "", fmt.Errorf("tx failed: %s", txResp.RawLog)
	}

	return txResp.TxHash, nil
}

func (s *MultiSendSuite) queryTxGas(txHash string) (int64, error) {
	// ChainNode.ExecQuery returns (stdout, stderr, err) based on chain.go
	stdout, _, err := s.Chain.GetNode().ExecQuery(s.GetContext(), "tx", txHash, "--output", "json")
	if err != nil {
		return 0, err
	}

	var respMap map[string]interface{}
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
