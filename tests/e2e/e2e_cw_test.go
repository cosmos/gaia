package e2e

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"strconv"
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
	contractAddr := s.instantiateWasm(ctx, s.chainA, valIdx, sender, strconv.Itoa(contractsCounter), "{\"count\":0}", "counter")
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

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

func (s *IntegrationTestSuite) storeWasm(ctx context.Context, c *chain, valIdx int, sender, wasmPath string) string {
	storeCmd := []string{
		gaiadBinary,
		txCommand,
		"wasm",
		"store",
		wasmPath,
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, standardFees.String()),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		"--gas=5000000",
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}

	s.T().Logf("%s storing wasm on host chain %s", sender, s.chainB.id)
	s.executeGaiaTxCommand(ctx, c, storeCmd, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Log("successfully sent store wasm tx")
	contractsCounter++
	return strconv.Itoa(contractsCounter)
}

func (s *IntegrationTestSuite) instantiateWasm(ctx context.Context, c *chain, valIdx int, sender, codeID,
	msg, label string,
) string {
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
		"--gas=500000",
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}

	s.T().Logf("%s instantiating wasm on host chain %s", sender, s.chainB.id)
	s.executeGaiaTxCommand(ctx, c, storeCmd, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Log("successfully sent instantiate wasm tx")
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[c.id][0].GetHostPort("1317/tcp"))
	address, err := queryWasmContractAddress(chainEndpoint, sender, contractsCounterPerSender[sender])
	s.Require().NoError(err)
	contractsCounterPerSender[sender]++
	return address
}

func (s *IntegrationTestSuite) instantiate2Wasm(ctx context.Context, c *chain, valIdx int, sender, codeID,
	msg, salt, label string,
) string {
	storeCmd := []string{
		gaiadBinary,
		txCommand,
		"wasm",
		"instantiate2",
		codeID,
		msg,
		salt,
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
	s.T().Log("successfully sent instantiate2 wasm tx")
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[c.id][0].GetHostPort("1317/tcp"))
	address, err := queryWasmContractAddress(chainEndpoint, sender, contractsCounterPerSender[sender])
	s.Require().NoError(err)
	contractsCounterPerSender[sender]++
	return address
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

func (s *IntegrationTestSuite) queryBuildAddress(ctx context.Context, c *chain, valIdx int, codeHash, creatorAddress, saltHexEncoded string,
) (res string) {
	cmd := []string{
		gaiadBinary,
		queryCommand,
		"wasm",
		"build-address",
		codeHash,
		creatorAddress,
		saltHexEncoded,
	}

	s.executeGaiaTxCommand(ctx, c, cmd, valIdx, func(stdOut []byte, stdErr []byte) bool {
		s.Require().NoError(yaml.Unmarshal(stdOut, &res))
		return true
	})
	return res
}
