package e2e

import (
	"encoding/base64"
	"path/filepath"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *IntegrationTestSuite) TestEncode() {
	var (
		rawTxFilePath = "tx_raw.json"
		valIdx        = 0
		chain         = s.chainA
		val           = chain.validators[valIdx]
	)
	rawTx, encoded, err := buildRawTx()
	s.Require().NoError(err)

	err = writeFile(filepath.Join(val.configDir(), rawTxFilePath), rawTx)
	s.Require().NoError(err)

	got := s.execEncode(chain, filepath.Join(gaiaHomePath, rawTxFilePath))
	s.T().Logf("encoded tx: %s", got)
	s.Require().Equal(encoded, got)
}

func (s *IntegrationTestSuite) TestDecode() {
	chain := s.chainA
	rawTx, encoded, err := buildRawTx()
	s.Require().NoError(err)

	got := s.execDecode(chain, encoded)
	s.T().Logf("raw tx: %s", got)
	s.Require().Equal(string(rawTx), got)
}

// buildRawTx build a dummy tx using the TxBuilder and
// return the JSON and encoded tx's
func buildRawTx() (tx []byte, encoded string, err error) {
	builder := txConfig.NewTxBuilder()
	builder.SetGasLimit(gas)
	builder.SetFeeAmount(sdk.NewCoins(fees))
	builder.SetMemo("foomemo")
	tx, err = txConfig.TxJSONEncoder()(builder.GetTx())
	if err != nil {
		return tx, encoded, err
	}
	txBytes, err := txConfig.TxEncoder()(builder.GetTx())
	if err != nil {
		return tx, encoded, err
	}
	encoded = base64.StdEncoding.EncodeToString(txBytes)
	return tx, encoded, err
}
