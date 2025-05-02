package tx

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/gaia/v24/tests/e2e/common"
)

type TestingSuite struct {
	common.TestingSuite
}

func (h *TestingSuite) ExecDecode(
	c *common.Chain,
	txPath string,
	opt ...common.FlagOption,
) string {
	opts := common.ApplyOptions(c.ID, opt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("%s - Executing gaiad decoding with %v", c.ID, txPath)
	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		"decode",
		txPath,
	}
	for flag, value := range opts {
		gaiaCommand = append(gaiaCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	var decoded string
	h.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, 0, func(stdOut []byte, stdErr []byte) bool {
		if stdErr != nil {
			return false
		}
		decoded = strings.TrimSuffix(string(stdOut), "\n")
		return true
	})
	h.Suite.T().Logf("successfully decode %v", txPath)
	return decoded
}

func (h *TestingSuite) ExecEncode(
	c *common.Chain,
	txPath string,
	opt ...common.FlagOption,
) string {
	opts := common.ApplyOptions(c.ID, opt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("%s - Executing gaiad encoding with %v", c.ID, txPath)
	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		"encode",
		txPath,
	}
	for flag, value := range opts {
		gaiaCommand = append(gaiaCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	var encoded string
	h.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, 0, func(stdOut []byte, stdErr []byte) bool {
		if stdErr != nil {
			return false
		}
		encoded = strings.TrimSuffix(string(stdOut), "\n")
		return true
	})
	h.Suite.T().Logf("successfully encode with %v", txPath)
	return encoded
}

func (h *TestingSuite) ExpectErrExecValidation(chain *common.Chain, valIdx int, expectErr bool) func([]byte, []byte) bool {
	return func(stdOut []byte, stdErr []byte) bool {
		var txResp types.TxResponse
		gotErr := common.Cdc.UnmarshalJSON(stdOut, &txResp) != nil
		if gotErr {
			h.Suite.Require().True(expectErr)
		}

		endpoint := fmt.Sprintf("http://%s", h.Resources.ValResources[chain.ID][valIdx].GetHostPort("1317/tcp"))
		// wait for the tx to be committed on chain
		h.Suite.Require().Eventuallyf(
			func() bool {
				gotErr := common.QueryGaiaTx(endpoint, txResp.TxHash) != nil
				return gotErr == expectErr
			},
			time.Minute,
			5*time.Second,
			"stdOut: %s, stdErr: %s",
			string(stdOut), string(stdErr),
		)
		return true
	}
}

func (h *TestingSuite) ExpectTxSubmitError(expectErrString string) func([]byte, []byte) bool {
	return func(stdOut []byte, stdErr []byte) bool {
		var txResp types.TxResponse
		if err := common.Cdc.UnmarshalJSON(stdOut, &txResp); err != nil {
			return false
		}
		if strings.Contains(txResp.RawLog, expectErrString) {
			return true
		}
		return false
	}
}

// SignTxFileOnline signs a transaction file using the gaiacli tx sign command
// the from flag is used to specify the keyring account to sign the transaction
// the from account must be registered in the keyring and exist on chain (have a balance or be a genesis account)
func (h *TestingSuite) SignTxFileOnline(chain *common.Chain, valIdx int, from string, txFilePath string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		"sign",
		filepath.Join(common.GaiaHomePath, txFilePath),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, chain.ID),
		fmt.Sprintf("--%s=%s", flags.FlagHome, common.GaiaHomePath),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	var output []byte
	var erroutput []byte
	captureOutput := func(stdout []byte, stderr []byte) bool {
		output = stdout
		erroutput = stderr
		return true
	}

	h.ExecuteGaiaTxCommand(ctx, chain, gaiaCommand, valIdx, captureOutput)
	if len(erroutput) > 0 {
		return nil, fmt.Errorf("failed to sign tx: %s", string(erroutput))
	}
	return output, nil
}

// BroadcastTxFile broadcasts a signed transaction file using the gaiacli tx broadcast command
// the from flag is used to specify the keyring account to sign the transaction
// the from account must be registered in the keyring and exist on chain (have a balance or be a genesis account)
func (h *TestingSuite) BroadcastTxFile(chain *common.Chain, valIdx int, from string, txFilePath string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	broadcastTxCmd := []string{
		common.GaiadBinary,
		common.TxCommand,
		"broadcast",
		filepath.Join(common.GaiaHomePath, txFilePath),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, chain.ID),
		fmt.Sprintf("--%s=%s", flags.FlagHome, common.GaiaHomePath),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	var output []byte
	var erroutput []byte
	captureOutput := func(stdout []byte, stderr []byte) bool {
		output = stdout
		erroutput = stderr
		return true
	}

	h.ExecuteGaiaTxCommand(ctx, chain, broadcastTxCmd, valIdx, captureOutput)
	if len(erroutput) > 0 {
		return nil, fmt.Errorf("failed to sign tx: %s", string(erroutput))
	}
	return output, nil
}
