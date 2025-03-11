package e2e

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/cosmos/gaia/v23/tests/e2e/common"
	"github.com/cosmos/gaia/v23/tests/e2e/query"
)

// /*
// TestFeeGrant creates a test to ensure that Alice can grant the fees for bob.
// Test Benchmarks:
// 1. Execute fee grant CLI command for Alice to pay bob fees
// 2. Send a transaction from bob with Alice as a fee granter
// 3. Check the bob balances if the fee was not deducted
// 4. Try to send a transaction from bob with Alice as a fee granter again. Should fail
// because all amount granted was expended
//
//	*/
func (s *IntegrationTestSuite) testFeeGrant() {
	s.Run("test fee grant module", func() {
		var (
			valIdx = 0
			c      = s.commonHelper.Resources.ChainA
			api    = fmt.Sprintf("http://%s", s.commonHelper.Resources.ValResources[c.ID][valIdx].GetHostPort("1317/tcp"))
		)

		alice, _ := c.GenesisAccounts[1].KeyInfo.GetAddress()
		bob, _ := c.GenesisAccounts[2].KeyInfo.GetAddress()
		charlie, _ := c.GenesisAccounts[3].KeyInfo.GetAddress()

		// add fee grant from alice to bob
		s.tx.ExecFeeGrant(
			c,
			valIdx,
			alice.String(),
			bob.String(),
			common.StandardFees.String(),
			common.WithKeyValue(common.FlagAllowedMessages, sdk.MsgTypeURL(&banktypes.MsgSend{})),
		)

		bobBalance, err := query.SpecificBalance(api, bob.String(), common.UAtomDenom)
		s.Require().NoError(err)

		// withdrawal all balance + fee + fee granter flag should succeed
		s.tx.ExecBankSend(
			c,
			valIdx,
			bob.String(),
			common.Address(),
			common.TokenAmount.String(),
			common.StandardFees.String(),
			false,
			common.WithKeyValue(common.FlagFeeGranter, alice.String()),
		)

		// check if the bob balance was subtracted without the fees
		expectedBobBalance := bobBalance.Sub(common.TokenAmount)
		bobBalance, err = query.SpecificBalance(api, bob.String(), common.UAtomDenom)
		s.Require().NoError(err)
		s.Require().Equal(expectedBobBalance, bobBalance)

		// tx should fail after spend limit reach
		s.tx.ExecBankSend(
			c,
			valIdx,
			bob.String(),
			common.Address(),
			common.TokenAmount.String(),
			common.StandardFees.String(),
			true,
			common.WithKeyValue(common.FlagFeeGranter, alice.String()),
		)

		// add fee grant from alice to charlie
		s.tx.ExecFeeGrant(
			c,
			valIdx,
			alice.String(),
			charlie.String(),
			common.StandardFees.String(), // spend limit
			common.WithKeyValue(common.FlagAllowedMessages, sdk.MsgTypeURL(&banktypes.MsgSend{})),
		)

		// revoke fee grant from alice to charlie
		s.tx.ExecFeeGrantRevoke(
			c,
			valIdx,
			alice.String(),
			charlie.String(),
		)

		// tx should fail because the grant was revoked
		s.tx.ExecBankSend(
			c,
			valIdx,
			charlie.String(),
			common.Address(),
			common.TokenAmount.String(),
			common.StandardFees.String(),
			true,
			common.WithKeyValue(common.FlagFeeGranter, alice.String()),
		)
	})
}
