package common

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ory/dockertest/v3/docker"

	"github.com/cosmos/cosmos-sdk/types"
)

const (
	// maxTxRetries is the maximum number of retries for sequence mismatch errors
	maxTxRetries = 3
	// initialRetryWait is the initial wait time before retrying (doubles each retry)
	initialRetryWait = 2 * time.Second
)

// isSequenceMismatch checks if the transaction failed due to account sequence mismatch
func isSequenceMismatch(stdOut []byte) bool {
	return bytes.Contains(stdOut, []byte("account sequence mismatch")) ||
		bytes.Contains(stdOut, []byte("incorrect account sequence"))
}

func (h *TestingSuite) ExecuteGaiaTxCommand(ctx context.Context, c *Chain, gaiaCommand []string, valIdx int, validation func([]byte, []byte) bool) {
	if validation == nil {
		validation = h.DefaultExecValidation(h.Resources.ChainA, 0)
	}

	var lastStdOut, lastStdErr []byte
	retryWait := initialRetryWait

	for attempt := 0; attempt <= maxTxRetries; attempt++ {
		var outBuf, errBuf bytes.Buffer

		exec, err := h.Resources.DkrPool.Client.CreateExec(docker.CreateExecOptions{
			Context:      ctx,
			AttachStdout: true,
			AttachStderr: true,
			Container:    h.Resources.ValResources[c.ID][valIdx].Container.ID,
			User:         "nonroot",
			Cmd:          gaiaCommand,
		})
		h.Suite.Require().NoError(err)

		err = h.Resources.DkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
			Context:      ctx,
			Detach:       false,
			OutputStream: &outBuf,
			ErrorStream:  &errBuf,
		})
		h.Suite.Require().NoError(err)

		lastStdOut = outBuf.Bytes()
		lastStdErr = errBuf.Bytes()

		if validation(lastStdOut, lastStdErr) {
			return // Success
		}

		// Check if it's a sequence mismatch - if so, retry with exponential backoff
		if attempt < maxTxRetries && isSequenceMismatch(lastStdOut) {
			h.Suite.T().Logf("Sequence mismatch detected, retrying in %v (attempt %d/%d)",
				retryWait, attempt+1, maxTxRetries)
			time.Sleep(retryWait)
			retryWait *= 2 // Exponential backoff
			continue
		}

		// Not a sequence mismatch or out of retries - fail
		break
	}

	h.Suite.Require().FailNowf("Exec validation failed", "stdout: %s, stderr: %s",
		string(lastStdOut), string(lastStdErr))
}

func (h *TestingSuite) ExecuteHermesCommand(ctx context.Context, hermesCmd []string) ([]byte, error) {
	var outBuf bytes.Buffer
	exec, err := h.Resources.DkrPool.Client.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		AttachStdout: true,
		AttachStderr: true,
		Container:    h.Resources.HermesResource.Container.ID,
		User:         "root",
		Cmd:          hermesCmd,
	})
	h.Suite.Require().NoError(err)

	err = h.Resources.DkrPool.Client.StartExec(exec.ID, docker.StartExecOptions{
		Context:      ctx,
		Detach:       false,
		OutputStream: &outBuf,
	})
	h.Suite.Require().NoError(err)

	// Check that the stdout output contains the expected status
	// and look for errors, e.g "insufficient fees"
	stdOut := []byte{}
	scanner := bufio.NewScanner(&outBuf)
	for scanner.Scan() {
		stdOut = scanner.Bytes()
		var out map[string]interface{}
		err = json.Unmarshal(stdOut, &out)
		h.Suite.Require().NoError(err)
		if err != nil {
			return nil, fmt.Errorf("hermes relayer command returned failed with error: %s", err)
		}
		// errors are caught by observing the logs level in the stderr output
		if lvl := out["level"]; lvl != nil && strings.ToLower(lvl.(string)) == "error" {
			errMsg := out["fields"].(map[string]interface{})["message"]
			return nil, fmt.Errorf("hermes relayer command failed: %s", errMsg)
		}
		if s := out["status"]; s != nil && s != "success" {
			return nil, fmt.Errorf("hermes relayer command returned failed with status: %s", s)
		}
	}

	return stdOut, nil
}

func (h *TestingSuite) DefaultExecValidation(chain *Chain, valIdx int) func([]byte, []byte) bool {
	return func(stdOut []byte, stdErr []byte) bool {
		var txResp types.TxResponse
		if err := Cdc.UnmarshalJSON(stdOut, &txResp); err != nil {
			return false
		}
		if strings.Contains(txResp.String(), "code: 0") || txResp.Code == 0 {
			endpoint := fmt.Sprintf("http://%s", h.Resources.ValResources[chain.ID][valIdx].GetHostPort("1317/tcp"))
			h.Suite.Require().Eventually(
				func() bool {
					return QueryGaiaTx(endpoint, txResp.TxHash) == nil
				},
				time.Minute,
				5*time.Second,
				"stdOut: %s, stdErr: %s",
				string(stdOut), string(stdErr),
			)
			return true
		}
		return false
	}
}

func QueryGaiaTx(endpoint, txHash string) error {
	resp, err := http.Get(fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/%s", endpoint, txHash))
	if err != nil {
		return fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("tx query returned non-200 status: %d", resp.StatusCode)
	}

	var result map[string]interface{}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	txResp := result["tx_response"].(map[string]interface{})
	if v := txResp["code"]; v.(float64) != 0 {
		return fmt.Errorf("tx %s failed with status code %v", txHash, v)
	}

	return nil
}
