package e2e

import (
	"github.com/cosmos/gaia/v23/tests/e2e/common"
	"github.com/cosmos/gaia/v23/tests/e2e/msg"
	"github.com/cosmos/gaia/v23/tests/e2e/tx"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
	TestCounters common.TestCounters
	commonHelper common.Helper
	tx           tx.Helper
	msg          msg.Helper
}
