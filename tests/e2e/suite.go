package e2e

import (
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/gaia/v23/tests/e2e/common"
	"github.com/cosmos/gaia/v23/tests/e2e/tx"
)

type IntegrationTestSuite struct {
	suite.Suite
	commonHelper common.Helper
	tx           tx.Helper
}
