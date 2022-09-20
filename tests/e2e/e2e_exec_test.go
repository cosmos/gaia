package e2e

import (
	"context"
	"fmt"
	"path/filepath"
	"time"
)

const (
	binaryName = "gaiad"

	flagFrom           = "from"
	flagHome           = "home"
	flagFees           = "fees"
	flagGas            = "gas"
	flagOutput         = "output"
	flagChainID        = "chain-id"
	flagBroadcastMode  = "broadcast-mode"
	flagKeyringBackend = "keyring-backend"
)

type flagOption func(map[string]interface{})

// withKeyValue add a new flag to command
func withKeyValue(key string, value interface{}) flagOption {
	return func(o map[string]interface{}) {
		o[key] = value
	}
}

func applyOptions(chainID, home string, options []flagOption) map[string]interface{} {
	opts := map[string]interface{}{
		flagKeyringBackend: "test",
		flagOutput:         "json",
		flagGas:            "auto",
		flagFrom:           "alice",
		flagBroadcastMode:  "sync",
		flagChainID:        chainID,
		flagHome:           home,
		flagFees:           fees.String(),
	}
	for _, apply := range options {
		apply(opts)
	}
	return opts
}

func (s *IntegrationTestSuite) execVestingTx(
	c *chain,
	home,
	method string,
	args []string,
	opt ...flagOption,
) {
	opts := applyOptions(c.id, home, opt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("%s - Executing gaiad %s with %v", c.id, method, args)
	gaiaCommand := []string{
		binaryName,
		"tx",
		"vesting",
		method,
		"-y",
	}
	gaiaCommand = append(gaiaCommand, args...)

	for flag, value := range opts {
		gaiaCommand = append(gaiaCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, 0, dockerUserRoot, "")
	s.T().Logf("successfully %s with %v", method, args)
}

func (s *IntegrationTestSuite) execCreatePermanentLockedAccount(
	c *chain,
	home,
	address,
	amount string,
	opt ...flagOption,
) {
	s.T().Logf("Executing gaiad create a permanent locked vesting account %s", c.id)
	s.execVestingTx(c, home, "create-permanent-locked-account", []string{address, amount}, opt...)
	s.T().Logf("successfully created permanent locked vesting account %s with %s", address, amount)
}

func (s *IntegrationTestSuite) execCreatePeriodicVestingAccount(
	c *chain,
	home,
	address string,
	opt ...flagOption,
) {
	jsonPath := filepath.Join(home, vestingPeriodFilePath)
	s.T().Logf("Executing gaiad create periodic vesting account %s", c.id)
	s.execVestingTx(c, home, "create-periodic-vesting-account", []string{address, jsonPath}, opt...)
	s.T().Logf("successfully created periodic vesting account %s with %s", address, jsonPath)
}
