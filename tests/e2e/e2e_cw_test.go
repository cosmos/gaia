package e2e

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
)

func (s *IntegrationTestSuite) testCWCounter() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	valIdx := 0
	val := s.chainA.validators[valIdx]
	address, _ := val.keyInfo.GetAddress()
	sender := address.String()
	dirName, err := os.Getwd()
	s.Require().NoError(err)

	// Copy file to container path and store the contract
	src := filepath.Join(dirName, "data/counter.wasm")
	dst := filepath.Join(val.configDir(), "config", "counter.wasm")
	_, err = copyFile(src, dst)
	s.Require().NoError(err)
	storeWasmPath := configFile("counter.wasm")
	s.storeWasm(ctx, s.chainA, valIdx, sender, storeWasmPath)

	// Instantiate the contract
	s.instantiateWasm(ctx, s.chainA, valIdx, sender, "1", "{\"count\":0}", "counter")
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	contractAddr, err := queryWasmContractAddress(chainEndpoint, address.String())
	s.Require().NoError(err)

	// Execute the contract
	s.executeWasm(ctx, s.chainA, valIdx, sender, contractAddr, "{\"increment\":{}}")

	// Validate count increment
	query := map[string]interface{}{
		"get_count": map[string]interface{}{},
	}
	queryJSON, err := json.Marshal(query)
	s.Require().NoError(err)
	queryMsg := base64.StdEncoding.EncodeToString(queryJSON)
	data, err := queryWasmSmartContractState(chainEndpoint, contractAddr, queryMsg)
	s.Require().NoError(err)
	var counterResp map[string]int
	err = json.Unmarshal(data, &counterResp)
	s.Require().NoError(err)
	s.Require().Equal(1, counterResp["count"])
}

func (s *IntegrationTestSuite) storeWasm(ctx context.Context, c *chain, valIdx int, sender, wasmPath string) {
	storeCmd := []string{
		gaiadBinary,
		txCommand,
		"wasm",
		"store",
		wasmPath,
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, standardFees.String()),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		"--gas=2000000",
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}

	s.T().Logf("%s storing wasm on host chain %s", sender, s.chainB.id)
	s.executeGaiaTxCommand(ctx, c, storeCmd, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Log("successfully sent store wasm tx")
}

func (s *IntegrationTestSuite) instantiateWasm(ctx context.Context, c *chain, valIdx int, sender, codeID,
	msg, label string,
) {
	storeCmd := []string{
		gaiadBinary,
		txCommand,
		"wasm",
		"instantiate",
		codeID,
		msg,
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, standardFees.String()),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--label=%s", label),
		"--no-admin",
		"--gas=250000",
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}

	s.T().Logf("%s instantiating wasm on host chain %s", sender, s.chainB.id)
	s.executeGaiaTxCommand(ctx, c, storeCmd, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Log("successfully sent instantiate wasm tx")
}

func (s *IntegrationTestSuite) executeWasm(ctx context.Context, c *chain, valIdx int, sender, addr, msg string) {
	execCmd := []string{
		gaiadBinary,
		txCommand,
		"wasm",
		"execute",
		addr,
		msg,
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, standardFees.String()),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		"--gas=250000",
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}
	s.T().Logf("%s executing wasm on host chain %s", sender, s.chainB.id)
	s.executeGaiaTxCommand(ctx, c, execCmd, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Log("successfully sent execute wasm tx")
}
