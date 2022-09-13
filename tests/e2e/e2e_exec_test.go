package e2e

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

const (
	binaryName = "gaiad"

	flagFrom             = "from"
	flagFees             = "fees"
	flagGas              = "gas"
	flagOutput           = "output"
	flagKeyringBackend   = "keyring-backend"
	flagVestingAmount    = "vesting-amount"
	flagVestingStartTime = "vesting-start-time"
	flagVestingEndTime   = "vesting-end-time"
)

type flagOption func(map[string]interface{})

// withKeyValue add a new flag to command
func withKeyValue(key string, value interface{}) flagOption {
	return func(o map[string]interface{}) {
		o[key] = value
	}
}

func applyOptions(options []flagOption) map[string]interface{} {
	opts := map[string]interface{}{
		flagKeyringBackend: "test",
		flagOutput:         "json",
		flagGas:            "auto",
		flagFrom:           "alice",
		flagFees:           fees.String(),
	}
	for _, apply := range options {
		apply(opts)
	}
	return opts
}

func (s *IntegrationTestSuite) exec(
	c *chain,
	method string,
	args []string,
	opt ...flagOption,
) {
	opts := applyOptions(opt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("%s - Executing gaiad %s with %v", c.id, method, args)
	gaiaCommand := []string{
		binaryName,
		method,
	}
	gaiaCommand = append(gaiaCommand, args...)

	for flag, value := range opts {
		gaiaCommand = append(gaiaCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	s.executeGaiaTxCommand(ctx, c, gaiaCommand, 0, "")
	s.T().Logf("successfully %s with %v", method, args)
}

func (s *IntegrationTestSuite) execCreateVestingAccount(
	c *chain,
	address,
	amount string,
	endtime int64,
	opt ...flagOption,
) {
	s.T().Logf("Executing gaiad add genesis account %s", c.id)
	s.exec(c, "create-vesting-account", []string{address, amount, strconv.Itoa(int(endtime))}, opt...)
	s.T().Logf("successfully created genesis account %s with %s", address, amount)
}

func (s *IntegrationTestSuite) execCreatePermanentLockedAccount(
	c *chain,
	address,
	amount string,
	opt ...flagOption,
) {
	s.T().Logf("Executing gaiad create a permanent vesting account %s", c.id)
	s.exec(c, "create-permanent-locked-account", []string{address, amount}, opt...)
	s.T().Logf("successfully created permanent vesting account %s with %s", address, amount)
}

func (s *IntegrationTestSuite) execCreatePeriodicVestingAccount(
	c *chain,
	address,
	periodFilepath string,
	opt ...flagOption,
) {
	s.T().Logf("Executing gaiad create periodic vesting account %s", c.id)
	s.exec(c, "create-periodic-vesting-account", []string{address, periodFilepath}, opt...)
	s.T().Logf("successfully created periodic vesting account %s with %s", address, periodFilepath)
}
