package e2e

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

/*
TestFeeGrant creates a test to ensure that Alice can grant the fees for bob.
Test Benchmarks:
1. Execute fee grant CLI command for Alice to pay bob fees
2. Send a transaction from bob with Alice as a fee granter
3. Check the bob balances if the fee was not deducted
4. Try to send a transaction from bob with Alice as a fee granter again. Should fail
because all amount granted was expended
*/
func (s *IntegrationTestSuite) TestFeeGrant() {
	s.Run("test fee grant module", func() {
		var (
			valIdx = 0
			chain  = s.chainA
			api    = fmt.Sprintf("http://%s", s.valResources[chain.id][valIdx].GetHostPort("1317/tcp"))
		)

		alice := chain.genesisAccounts[1].keyInfo.GetAddress()
		bob := chain.genesisAccounts[2].keyInfo.GetAddress()
		charlie := chain.genesisAccounts[3].keyInfo.GetAddress()

		// add fee grant from alice to bob
		s.execFeeGrant(
			chain,
			valIdx,
			alice.String(),
			bob.String(),
			standardFees.String(),
			withKeyValue(flagAllowedMessages, sdk.MsgTypeURL(&banktypes.MsgSend{})),
		)

		bobBalance, err := getSpecificBalance(api, bob.String(), uatomDenom)
		s.Require().NoError(err)

		// withdrawal all balance + fee + fee granter flag should succeed
		s.execBankSend(
			chain,
			valIdx,
			bob.String(),
			Address(),
			tokenAmount.String(),
			standardFees.String(),
			false,
			withKeyValue(flagFeeGranter, alice.String()),
		)

		// check if the bob balance was subtracted without the fees
		expectedBobBalance := bobBalance.Sub(tokenAmount)
		bobBalance, err = getSpecificBalance(api, bob.String(), uatomDenom)
		s.Require().NoError(err)
		s.Require().Equal(expectedBobBalance, bobBalance)

		// tx should fail after spend limit reach
		s.execBankSend(
			chain,
			valIdx,
			bob.String(),
			Address(),
			tokenAmount.String(),
			standardFees.String(),
			true,
			withKeyValue(flagFeeGranter, alice.String()),
		)

		// add fee grant from alice to charlie
		s.execFeeGrant(
			chain,
			valIdx,
			alice.String(),
			charlie.String(),
			standardFees.String(), // spend limit
			withKeyValue(flagAllowedMessages, sdk.MsgTypeURL(&banktypes.MsgSend{})),
		)

		// revoke fee grant from alice to charlie
		s.execFeeGrantRevoke(
			chain,
			valIdx,
			alice.String(),
			charlie.String(),
		)

		// tx should fail because the grant was revoked
		s.execBankSend(
			chain,
			valIdx,
			charlie.String(),
			Address(),
			tokenAmount.String(),
			standardFees.String(),
			true,
			withKeyValue(flagFeeGranter, alice.String()),
		)
	})
}
