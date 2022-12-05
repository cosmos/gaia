package e2e

import (
	"encoding/base64"
	"path/filepath"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	rawTxFile = "tx_raw.json"
)

func (s *IntegrationTestSuite) testEncode() {
	chain := s.chainA
	_, encoded, err := buildRawTx()
	s.Require().NoError(err)

	got := s.execEncode(chain, filepath.Join(gaiaHomePath, rawTxFile))
	s.T().Logf("encoded tx: %s", got)
	s.Require().Equal(encoded, got)
}

func (s *IntegrationTestSuite) testDecode() {
	chain := s.chainA
	rawTx, encoded, err := buildRawTx()
	s.Require().NoError(err)

	got := s.execDecode(chain, encoded)
	s.T().Logf("raw tx: %s", got)
	s.Require().Equal(string(rawTx), got)
}

// buildRawTx build a dummy tx using the TxBuilder and
// return the JSON and encoded tx's
func buildRawTx() ([]byte, string, error) {
	builder := txConfig.NewTxBuilder()
	builder.SetGasLimit(gas)
	builder.SetFeeAmount(sdk.NewCoins(standardFees))
	builder.SetMemo("foomemo")
	tx, err := txConfig.TxJSONEncoder()(builder.GetTx())
	if err != nil {
		return nil, "", err
	}
	txBytes, err := txConfig.TxEncoder()(builder.GetTx())
	if err != nil {
		return nil, "", err
	}
	return tx, base64.StdEncoding.EncodeToString(txBytes), err
}
