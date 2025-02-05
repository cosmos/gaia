package e2e

import (
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"os"
	"path/filepath"
	"time"
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
	src := filepath.Join(dirName, "data/counter.wasm")
	dst := filepath.Join(val.configDir(), "config", "counter.wasm")
	_, err = copyFile(src, dst)
	s.Require().NoError(err)
	storeWasmPath := configFile("counter.wasm")
	s.storeWasm(ctx, s.chainA, valIdx, sender, storeWasmPath)
	s.instantiateWasm(ctx, s.chainA, valIdx, sender, "1", "{\"count\":0}", "counter")
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

func (s *IntegrationTestSuite) instantiateWasm(ctx context.Context, c *chain, valIdx int, sender, codeId,
	msg, label string) {
	storeCmd := []string{
		gaiadBinary,
		txCommand,
		"wasm",
		"instantiate",
		codeId,
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
