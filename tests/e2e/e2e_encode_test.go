package e2e

import (
	"encoding/base64"
	"path/filepath"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/gaia/v26/tests/e2e/common"
)

const (
	rawTxFile = "tx_raw.json"
)

func (s *IntegrationTestSuite) testEncode() {
	chain := s.Resources.ChainA
	_, encoded, err := buildRawTx()
	s.Require().NoError(err)

	got := s.ExecEncode(chain, filepath.Join(common.GaiaHomePath, rawTxFile))
	s.T().Logf("encoded tx: %s", got)
	s.Require().Equal(encoded, got)
}

func (s *IntegrationTestSuite) testDecode() {
	chain := s.Resources.ChainA
	rawTx, encoded, err := buildRawTx()
	s.Require().NoError(err)

	got := s.ExecDecode(chain, encoded)
	s.T().Logf("raw tx: %s", got)
	s.Require().Equal(string(rawTx), got)
}

// buildRawTx build a dummy tx using the TxBuilder and
// return the JSON and encoded tx's
func buildRawTx() ([]byte, string, error) {
	builder := common.TxConfig.NewTxBuilder()
	builder.SetGasLimit(common.Gas)
	builder.SetFeeAmount(sdk.NewCoins(common.StandardFees))
	builder.SetMemo("foomemo")
	tx, err := common.TxConfig.TxJSONEncoder()(builder.GetTx())
	if err != nil {
		return nil, "", err
	}
	txBytes, err := common.TxConfig.TxEncoder()(builder.GetTx())
	if err != nil {
		return nil, "", err
	}
	return tx, base64.StdEncoding.EncodeToString(txBytes), err
}
