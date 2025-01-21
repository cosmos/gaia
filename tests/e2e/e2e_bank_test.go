package e2e

import (
	"fmt"
	"path/filepath"
	"time"

	"cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authTx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	extensiontypes "github.com/cosmos/gaia/v23/x/metaprotocols/types"
)

func (s *IntegrationTestSuite) testBankTokenTransfer() {
	s.Run("send_tokens_between_accounts", func() {
		var (
			err           error
			valIdx        = 0
			c             = s.chainA
			chainEndpoint = fmt.Sprintf("http://%s", s.valResources[c.id][valIdx].GetHostPort("1317/tcp"))
		)

		// define one sender and two recipient accounts
		alice, _ := c.genesisAccounts[1].keyInfo.GetAddress()
		bob, _ := c.genesisAccounts[2].keyInfo.GetAddress()
		charlie, _ := c.genesisAccounts[3].keyInfo.GetAddress()

		var beforeAliceUAtomBalance,
			beforeBobUAtomBalance,
			beforeCharlieUAtomBalance,
			afterAliceUAtomBalance,
			afterBobUAtomBalance,
			afterCharlieUAtomBalance sdk.Coin

		// get balances of sender and recipient accounts
		s.Require().Eventually(
			func() bool {
				beforeAliceUAtomBalance, err = getSpecificBalance(chainEndpoint, alice.String(), uatomDenom)
				s.Require().NoError(err)

				beforeBobUAtomBalance, err = getSpecificBalance(chainEndpoint, bob.String(), uatomDenom)
				s.Require().NoError(err)

				beforeCharlieUAtomBalance, err = getSpecificBalance(chainEndpoint, charlie.String(), uatomDenom)
				s.Require().NoError(err)

				return beforeAliceUAtomBalance.IsValid() && beforeBobUAtomBalance.IsValid() && beforeCharlieUAtomBalance.IsValid()
			},
			10*time.Second,
			5*time.Second,
		)

		// alice sends tokens to bob
		s.execBankSend(s.chainA, valIdx, alice.String(), bob.String(), tokenAmount.String(), standardFees.String(), false)

		// check that the transfer was successful
		s.Require().Eventually(
			func() bool {
				afterAliceUAtomBalance, err = getSpecificBalance(chainEndpoint, alice.String(), uatomDenom)
				s.Require().NoError(err)

				afterBobUAtomBalance, err = getSpecificBalance(chainEndpoint, bob.String(), uatomDenom)
				s.Require().NoError(err)

				decremented := beforeAliceUAtomBalance.Sub(tokenAmount).Sub(standardFees).IsEqual(afterAliceUAtomBalance)
				incremented := beforeBobUAtomBalance.Add(tokenAmount).IsEqual(afterBobUAtomBalance)

				return decremented && incremented
			},
			10*time.Second,
			5*time.Second,
		)

		// save the updated account balances of alice and bob
		beforeAliceUAtomBalance, beforeBobUAtomBalance = afterAliceUAtomBalance, afterBobUAtomBalance

		// alice sends tokens to bob and charlie, at once
		s.execBankMultiSend(s.chainA, valIdx, alice.String(), []string{bob.String(), charlie.String()}, tokenAmount.String(), standardFees.String(), false)

		s.Require().Eventually(
			func() bool {
				afterAliceUAtomBalance, err = getSpecificBalance(chainEndpoint, alice.String(), uatomDenom)
				s.Require().NoError(err)

				afterBobUAtomBalance, err = getSpecificBalance(chainEndpoint, bob.String(), uatomDenom)
				s.Require().NoError(err)

				afterCharlieUAtomBalance, err = getSpecificBalance(chainEndpoint, charlie.String(), uatomDenom)
				s.Require().NoError(err)

				decremented := beforeAliceUAtomBalance.Sub(tokenAmount).Sub(tokenAmount).Sub(standardFees).IsEqual(afterAliceUAtomBalance)
				incremented := beforeBobUAtomBalance.Add(tokenAmount).IsEqual(afterBobUAtomBalance) &&
					beforeCharlieUAtomBalance.Add(tokenAmount).IsEqual(afterCharlieUAtomBalance)

				return decremented && incremented
			},
			10*time.Second,
			5*time.Second,
		)
	})
}

// tests the bank send command with populated non_critical_extension_options field
// the Tx should succeed if the data can be properly encoded and decoded
// the tx is signed and broadcast using gaiad tx sign and broadcast commands
func (s *IntegrationTestSuite) bankSendWithNonCriticalExtensionOptions() {
	s.Run("transfer_with_non_critical_extension_options", func() {
		c := s.chainB

		submitterAccount := c.genesisAccounts[1]
		submitterAddress, err := submitterAccount.keyInfo.GetAddress()
		s.Require().NoError(err)
		sendMsg := banktypes.NewMsgSend(submitterAddress, submitterAddress, sdk.NewCoins(sdk.NewCoin(uatomDenom, math.NewInt(100))))

		// valid non-critical extension options
		ext := &extensiontypes.ExtensionData{
			ProtocolId:      "test-protocol",
			ProtocolVersion: "1",
			Data:            []byte("Hello Cosmos"),
		}

		extAny, err := codectypes.NewAnyWithValue(ext)
		s.Require().NoError(err)
		s.Require().NotNil(extAny)

		txBuilder := encodingConfig.TxConfig.NewTxBuilder()

		s.Require().NoError(txBuilder.SetMsgs(sendMsg))

		txBuilder.SetMemo("non-critical-ext-message-test")
		txBuilder.SetFeeAmount(sdk.NewCoins(standardFees))
		txBuilder.SetGasLimit(200000)

		// add extension options
		tx := txBuilder.GetTx()
		if etx, ok := tx.(authTx.ExtensionOptionsTxBuilder); ok {
			etx.SetNonCriticalExtensionOptions(extAny)
		}

		bz, err := encodingConfig.TxConfig.TxEncoder()(tx)
		s.Require().NoError(err)
		s.Require().NotNil(bz)

		txWithExt, err := decodeTx(bz)
		s.Require().NoError(err)
		s.Require().NotNil(txWithExt)

		rawTx, err := cdc.MarshalJSON(txWithExt)
		s.Require().NoError(err)
		s.Require().NotNil(rawTx)

		unsignedFname := "unsigned_non_critical_extension_option_tx.json"
		unsignedJSONFile := filepath.Join(c.validators[0].configDir(), unsignedFname)
		err = writeFile(unsignedJSONFile, rawTx)
		s.Require().NoError(err)

		signedTx, err := s.signTxFileOnline(c, 0, submitterAddress.String(), unsignedFname)
		s.Require().NoError(err)
		s.Require().NotNil(signedTx)

		signedFname := "signed_non_critical_extension_option_tx.json"
		signedJSONFile := filepath.Join(c.validators[0].configDir(), signedFname)
		err = writeFile(signedJSONFile, signedTx)
		s.Require().NoError(err)

		// if there's no errors the non_critical_extension_options field was properly encoded and decoded
		out, err := s.broadcastTxFile(c, 0, submitterAddress.String(), signedFname)
		s.Require().NoError(err)
		s.Require().NotNil(out)
	})
}

// tests the bank send command with invalid non_critical_extension_options field
// the tx should always fail to decode the extension options since no concrete type is registered for the provided extension field
func (s *IntegrationTestSuite) failedBankSendWithNonCriticalExtensionOptions() {
	s.Run("fail_encoding_invalid_non_critical_extension_options", func() {
		c := s.chainB

		submitterAccount := c.genesisAccounts[1]
		submitterAddress, err := submitterAccount.keyInfo.GetAddress()
		s.Require().NoError(err)
		sendMsg := banktypes.NewMsgSend(submitterAddress, submitterAddress, sdk.NewCoins(sdk.NewCoin(uatomDenom, math.NewInt(100))))

		// the message does not matter, as long as it is in the interface registry
		ext := &banktypes.MsgMultiSend{}

		extAny, err := codectypes.NewAnyWithValue(ext)
		s.Require().NoError(err)
		s.Require().NotNil(extAny)

		txBuilder := encodingConfig.TxConfig.NewTxBuilder()

		s.Require().NoError(txBuilder.SetMsgs(sendMsg))

		txBuilder.SetMemo("fail-non-critical-ext-message")
		txBuilder.SetFeeAmount(sdk.NewCoins(standardFees))
		txBuilder.SetGasLimit(200000)

		// add extension options
		tx := txBuilder.GetTx()
		if etx, ok := tx.(authTx.ExtensionOptionsTxBuilder); ok {
			etx.SetNonCriticalExtensionOptions(extAny)
		}

		bz, err := encodingConfig.TxConfig.TxEncoder()(tx)
		s.Require().NoError(err)
		s.Require().NotNil(bz)

		// decode fails because the provided extension option does not implement the correct TxExtensionOptionI interface
		txWithExt, err := decodeTx(bz)
		s.Require().Error(err)
		s.Require().ErrorContains(err, "failed to decode tx: no concrete type registered for type URL /cosmos.bank.v1beta1.MsgMultiSend against interface *tx.TxExtensionOptionI")
		s.Require().Nil(txWithExt)
	})
}
