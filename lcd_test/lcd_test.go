//nolint:bodyclose
package lcdtest

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	distrrest "github.com/cosmos/cosmos-sdk/x/distribution/client/rest"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/slashing"
)

const (
	name1 = "test1"
	memo  = "LCD test tx"
)

var fees = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}

func init() {
	crypto.BcryptSecurityParameter = 1
	version.Version = os.Getenv("VERSION")
}

func newKeybase() (keyring.Keyring, error) {
	return keyring.New(
		sdk.KeyringServiceName(),
		viper.GetString(flags.FlagKeyringBackend),
		InitClientHome(""),
		nil,
	)
}

// nolint: errcheck
func TestMain(m *testing.M) {
	viper.Set(flags.FlagKeyringBackend, "test")
	os.Exit(m.Run())
}

func TestNodeStatus(t *testing.T) {
	cleanup, _, _, port, err := InitializeLCD(1, []sdk.AccAddress{}, true)
	require.NoError(t, err)
	defer cleanup()
	getNodeInfo(t, port)
	getSyncStatus(t, port, false)
}

func TestBlock(t *testing.T) {
	cleanup, _, _, port, err := InitializeLCD(1, []sdk.AccAddress{}, true)
	require.NoError(t, err)
	defer cleanup()
	getBlock(t, port, -1, false)
	getBlock(t, port, 2, false)
	getBlock(t, port, 100000000, true)
}

func TestValidators(t *testing.T) {
	cleanup, _, _, port, err := InitializeLCD(1, []sdk.AccAddress{}, true)
	require.NoError(t, err)
	defer cleanup()
	resultVals := getValidatorSets(t, port, -1, false)
	require.Contains(t, resultVals.Validators[0].Address.String(), "cosmosvalcons")
	require.Contains(t, resultVals.Validators[0].PubKey, "cosmosvalconspub")
	getValidatorSets(t, port, 2, false)
	getValidatorSets(t, port, 10000000, true)
}

func TestCoinSend(t *testing.T) {
	kb, err := newKeybase()
	require.NoError(t, err)
	addr, _, err := CreateAddr(name1, kb)
	require.NoError(t, err)
	cleanup, _, _, port, err := InitializeLCD(1, []sdk.AccAddress{addr}, true)
	require.NoError(t, err)
	defer cleanup()

	bz, err := hex.DecodeString("8FA6AB57AD6870F6B5B2E57735F38F2F30E73CB6")
	require.NoError(t, err)
	someFakeAddr := sdk.AccAddress(bz)

	// query empty
	res, body := Request(t, port, "GET", fmt.Sprintf("/auth/accounts/%s", someFakeAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	initialBalance := getBalances(t, port, addr)

	// create TX
	receiveAddr, resultTx := doTransfer(t, port, name1, memo, addr, fees, kb)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	require.Equal(t, uint32(0), resultTx.Code)

	// query sender
	coins := getBalances(t, port, addr)
	expectedBalance := initialBalance[0].Sub(fees[0])

	require.Equal(t, sdk.DefaultBondDenom, coins[0].Denom)
	require.Equal(t, expectedBalance.Amount.SubRaw(1), coins[0].Amount)
	expectedBalance = coins[0]

	// query receiver
	coins2 := getBalances(t, port, receiveAddr)
	require.Equal(t, sdk.DefaultBondDenom, coins2[0].Denom)
	require.Equal(t, int64(1), coins2[0].Amount.Int64())

	// test failure with too little gas
	res, body, _ = doTransferWithGas(t, port, name1, memo, addr, "100", 0, false, true, fees, kb)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.Nil(t, err)

	// test failure with negative gas
	res, body, _ = doTransferWithGas(t, port, name1, memo, addr, "-200", 0, false, false, fees, kb)
	require.Equal(t, http.StatusBadRequest, res.StatusCode, body)

	// test failure with negative adjustment
	res, body, _ = doTransferWithGas(t, port, name1, memo, addr, "10000", -0.1, true, false, fees, kb)
	require.Equal(t, http.StatusBadRequest, res.StatusCode, body)

	// test failure with 0 gas
	res, body, _ = doTransferWithGas(t, port, name1, memo, addr, "0", 0, false, true, fees, kb)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	// test failure with wrong adjustment
	res, body, _ = doTransferWithGas(t, port, name1, memo, addr, flags.GasFlagAuto, 0.1, false, true, fees, kb)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	// run simulation and test success with estimated gas
	res, body, _ = doTransferWithGas(t, port, name1, memo, addr, "10000", 1.0, true, false, fees, kb)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var gasEstResp rest.GasEstimateResponse
	require.Nil(t, cdc.UnmarshalJSON([]byte(body), &gasEstResp))
	require.NotZero(t, gasEstResp.GasEstimate)

	balances := getBalances(t, port, addr)
	require.Equal(t, expectedBalance.Amount, balances.AmountOf(sdk.DefaultBondDenom))

	// run successful tx
	gas := fmt.Sprintf("%d", gasEstResp.GasEstimate)
	res, body, _ = doTransferWithGas(t, port, name1, memo, addr, gas, 1.0, false, true, fees, kb)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err = cdc.UnmarshalJSON([]byte(body), &resultTx)
	require.Nil(t, err)

	tests.WaitForHeight(resultTx.Height+1, port)
	require.Equal(t, uint32(0), resultTx.Code)

	balances = getBalances(t, port, addr)
	expectedBalance = expectedBalance.Sub(fees[0])
	require.Equal(t, expectedBalance.Amount.SubRaw(1), balances.AmountOf(sdk.DefaultBondDenom))
}

func TestCoinSendAccAuto(t *testing.T) {
	kb, err := newKeybase()
	require.NoError(t, err)
	addr, _, err := CreateAddr(name1, kb)
	require.NoError(t, err)
	cleanup, _, _, port, err := InitializeLCD(1, []sdk.AccAddress{addr}, true)
	require.NoError(t, err)
	defer cleanup()

	initialBalance := getBalances(t, port, addr)

	// send a transfer tx without specifying account number and sequence
	res, body, _ := doTransferWithGasAccAuto(
		t, port, name1, memo, addr, "200000", 1.0, false, true, fees, kb,
	)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	// query sender
	coins := getBalances(t, port, addr)
	expectedBalance := initialBalance[0].Sub(fees[0])

	require.Equal(t, sdk.DefaultBondDenom, coins[0].Denom)
	require.Equal(t, expectedBalance.Amount.SubRaw(1), coins[0].Amount)
}

func TestCoinMultiSendGenerateOnly(t *testing.T) {
	kb, err := newKeybase()
	require.NoError(t, err)
	addr, _, err := CreateAddr(name1, kb)
	require.NoError(t, err)
	cleanup, _, _, port, err := InitializeLCD(1, []sdk.AccAddress{addr}, true)
	require.NoError(t, err)
	defer cleanup()

	// generate only
	res, body, _ := doTransferWithGas(t, port, "", memo, addr, "200000", 1, false, false, fees, kb)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var stdTx auth.StdTx
	require.Nil(t, cdc.UnmarshalJSON([]byte(body), &stdTx))
	require.Equal(t, len(stdTx.Msgs), 1)
	require.Equal(t, stdTx.GetMsgs()[0].Route(), bank.RouterKey)
	require.Equal(t, stdTx.GetMsgs()[0].GetSigners(), []sdk.AccAddress{addr})
	require.Equal(t, 0, len(stdTx.Signatures))
	require.Equal(t, memo, stdTx.Memo)
	require.NotZero(t, stdTx.Fee.Gas)
	require.IsType(t, stdTx.GetMsgs()[0], bank.MsgSend{})
	require.Equal(t, addr, stdTx.GetMsgs()[0].(bank.MsgSend).FromAddress)
}

func TestCoinSendGenerateSignAndBroadcast(t *testing.T) {
	kb, err := newKeybase()
	require.NoError(t, err)
	addr, _, err := CreateAddr(name1, kb)
	require.NoError(t, err)
	cleanup, _, _, port, err := InitializeLCD(1, []sdk.AccAddress{addr}, true)
	require.NoError(t, err)
	defer cleanup()
	acc := getAccount(t, port, addr)

	// simulate tx
	res, body, _ := doTransferWithGas(t, port, name1, memo, addr, flags.GasFlagAuto, 1.0, true, false, fees, kb)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var gasEstResp rest.GasEstimateResponse
	require.Nil(t, cdc.UnmarshalJSON([]byte(body), &gasEstResp))
	require.NotZero(t, gasEstResp.GasEstimate)

	// generate tx
	gas := fmt.Sprintf("%d", gasEstResp.GasEstimate)
	res, body, _ = doTransferWithGas(t, port, name1, memo, addr, gas, 1, false, false, fees, kb)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var tx auth.StdTx
	require.Nil(t, cdc.UnmarshalJSON([]byte(body), &tx))
	require.Equal(t, len(tx.Msgs), 1)
	require.Equal(t, tx.Msgs[0].Route(), bank.RouterKey)
	require.Equal(t, tx.Msgs[0].GetSigners(), []sdk.AccAddress{addr})
	require.Equal(t, 0, len(tx.Signatures))
	require.Equal(t, memo, tx.Memo)
	require.NotZero(t, tx.Fee.Gas)

	gasEstimate := int64(tx.Fee.Gas)
	_, body = signAndBroadcastGenTx(t, port, name1, body, acc, 1.0, false, kb)

	// check if tx was committed
	var txResp sdk.TxResponse
	require.Nil(t, cdc.UnmarshalJSON([]byte(body), &txResp))
	require.Equal(t, uint32(0), txResp.Code)
	require.Equal(t, gasEstimate, txResp.GasWanted)
}

func TestEncodeTx(t *testing.T) {
	kb, err := newKeybase()
	require.NoError(t, err)
	addr, _, err := CreateAddr(name1, kb)
	require.NoError(t, err)
	cleanup, _, _, port, err := InitializeLCD(1, []sdk.AccAddress{addr}, true)
	require.NoError(t, err)
	defer cleanup()

	_, body, _ := doTransferWithGas(t, port, name1, memo, addr, "2", 1, false, false, fees, kb)
	var tx auth.StdTx
	require.Nil(t, cdc.UnmarshalJSON([]byte(body), &tx))

	encodedJSON, err := cdc.MarshalJSON(tx)
	require.NoError(t, err)
	res, body := Request(t, port, "POST", "/txs/encode", encodedJSON)

	// Make sure it came back ok, and that we can decode it back to the transaction
	// 200 response.
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	encodeResp := struct {
		Tx string `json:"tx"`
	}{}

	require.Nil(t, cdc.UnmarshalJSON([]byte(body), &encodeResp))

	// verify that the base64 decodes
	decodedBytes, err := base64.StdEncoding.DecodeString(encodeResp.Tx)
	require.Nil(t, err)

	// check that the transaction decodes as expected
	var decodedTx auth.StdTx
	require.Nil(t, cdc.UnmarshalBinaryBare(decodedBytes, &decodedTx))
	require.Equal(t, memo, decodedTx.Memo)
}

func TestTxs(t *testing.T) {
	kb, err := newKeybase()
	require.NoError(t, err)
	addr, _, err := CreateAddr(name1, kb)
	require.NoError(t, err)
	cleanup, _, _, port, err := InitializeLCD(1, []sdk.AccAddress{addr}, true)
	require.NoError(t, err)
	defer cleanup()

	var emptyTxs []sdk.TxResponse
	txResult := getTransactions(t, port)
	require.Equal(t, emptyTxs, txResult.Txs)

	// query empty
	txResult = getTransactions(t, port, fmt.Sprintf("message.sender=%s", addr.String()))
	require.Equal(t, emptyTxs, txResult.Txs)

	// also tests url decoding
	txResult = getTransactions(t, port, fmt.Sprintf("message.sender=%s", addr.String()))
	require.Equal(t, emptyTxs, txResult.Txs)

	txResult = getTransactions(t, port, fmt.Sprintf("message.action=submit_proposal&message.sender=%s", addr.String()))
	require.Equal(t, emptyTxs, txResult.Txs)

	// create tx
	receiveAddr, resultTx := doTransfer(t, port, name1, memo, addr, fees, kb)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx is queryable
	tx := getTransaction(t, port, resultTx.TxHash)
	require.Equal(t, resultTx.TxHash, tx.TxHash)

	// query sender
	txResult = getTransactions(t, port, fmt.Sprintf("message.sender=%s", addr.String()))
	require.Len(t, txResult.Txs, 1)
	require.Equal(t, resultTx.Height, txResult.Txs[0].Height)

	// query recipient
	txResult = getTransactions(t, port, fmt.Sprintf("transfer.recipient=%s", receiveAddr.String()))
	require.Len(t, txResult.Txs, 1)
	require.Equal(t, resultTx.Height, txResult.Txs[0].Height)

	// query transaction that doesn't exist
	validTxHash := "9ADBECAAD8DACBEC3F4F535704E7CF715C765BDCEDBEF086AFEAD31BA664FB0B"
	res, body := getTransactionRequest(t, port, validTxHash)
	require.True(t, strings.Contains(body, validTxHash))
	require.Equal(t, http.StatusNotFound, res.StatusCode)

	// bad query string
	res, body = getTransactionRequest(t, port, "badtxhash")
	require.True(t, strings.Contains(body, "encoding/hex"))
	require.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestValidatorsQuery(t *testing.T) {
	cleanup, valPubKeys, operAddrs, port, err := InitializeLCD(1, []sdk.AccAddress{}, true)
	require.NoError(t, err)
	defer cleanup()

	require.Equal(t, 1, len(valPubKeys))
	require.Equal(t, 1, len(operAddrs))

	validators := getValidators(t, port)
	require.Equal(t, 1, len(validators), fmt.Sprintf("%+v", validators))

	// make sure all the validators were found (order unknown because sorted by operator addr)
	foundVal := false

	if validators[0].ConsensusPubkey == sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, valPubKeys[0]) {
		foundVal = true
	}

	require.True(t, foundVal, "pk %v, operator %v", operAddrs[0], validators[0].OperatorAddress)
}

func TestValidatorQuery(t *testing.T) {
	cleanup, valPubKeys, operAddrs, port, err := InitializeLCD(1, []sdk.AccAddress{}, true)
	require.NoError(t, err)
	defer cleanup()
	require.Equal(t, 1, len(valPubKeys))
	require.Equal(t, 1, len(operAddrs))

	validator := getValidator(t, port, operAddrs[0])
	require.Equal(t, validator.OperatorAddress, operAddrs[0], "The returned validator does not hold the correct data")
}

func TestBonding(t *testing.T) {
	kb, err := newKeybase()
	require.NoError(t, err)
	addr, _, err := CreateAddr(name1, kb)
	require.NoError(t, err)

	cleanup, valPubKeys, operAddrs, port, err := InitializeLCD(2, []sdk.AccAddress{addr}, false)
	require.NoError(t, err)
	tests.WaitForHeight(1, port)
	defer cleanup()

	require.Equal(t, 2, len(valPubKeys))
	require.Equal(t, 2, len(operAddrs))

	amt := sdk.TokensFromConsensusPower(60)
	amtDec := amt.ToDec()
	validator := getValidator(t, port, operAddrs[0])

	initialBalance := getBalances(t, port, addr)

	// create bond TX
	delTokens := sdk.TokensFromConsensusPower(60)
	resultTx := doDelegate(t, port, name1, addr, operAddrs[0], delTokens, fees, kb)
	tests.WaitForHeight(resultTx.Height+1, port)

	require.Equal(t, uint32(0), resultTx.Code)

	// query tx
	txResult := getTransactions(t, port,
		fmt.Sprintf("message.action=delegate&message.sender=%s", addr),
		fmt.Sprintf("delegate.validator=%s", operAddrs[0]),
	)
	require.Len(t, txResult.Txs, 1)
	require.Equal(t, resultTx.Height, txResult.Txs[0].Height)

	// verify balance
	coins := getBalances(t, port, addr)
	expectedBalance := initialBalance[0].Sub(fees[0])
	require.Equal(t, expectedBalance.Amount.Sub(delTokens).String(), coins.AmountOf(sdk.DefaultBondDenom).String())
	expectedBalance = coins[0]

	// query delegation
	bond := getDelegation(t, port, addr, operAddrs[0])
	require.Equal(t, amtDec, bond.Shares)

	delegatorDels := getDelegatorDelegations(t, port, addr)
	require.Len(t, delegatorDels, 1)
	require.Equal(t, amtDec, delegatorDels[0].Shares)

	// query all delegations to validator
	bonds := getValidatorDelegations(t, port, operAddrs[0])
	require.Len(t, bonds, 2)

	bondedValidators := getDelegatorValidators(t, port, addr)
	require.Len(t, bondedValidators, 1)
	require.Equal(t, operAddrs[0], bondedValidators[0].OperatorAddress)
	require.Equal(t, validator.DelegatorShares.Add(amtDec).String(), bondedValidators[0].DelegatorShares.String())

	bondedValidator := getDelegatorValidator(t, port, addr, operAddrs[0])
	require.Equal(t, operAddrs[0], bondedValidator.OperatorAddress)

	// testing unbonding
	unbondingTokens := sdk.TokensFromConsensusPower(30)
	resultTx = doUndelegate(t, port, name1, addr, operAddrs[0], unbondingTokens, fees, kb)
	tests.WaitForHeight(resultTx.Height+1, port)

	require.Equal(t, uint32(0), resultTx.Code)

	// sender should have not received any coins as the unbonding has only just begun
	coins = getBalances(t, port, addr)
	expectedBalance = expectedBalance.Sub(fees[0])
	require.True(t,
		expectedBalance.Amount.LT(coins.AmountOf(sdk.DefaultBondDenom)) ||
			expectedBalance.Amount.Equal(coins.AmountOf(sdk.DefaultBondDenom)),
		"should get tokens back from automatic withdrawal after an unbonding delegation",
	)
	expectedBalance = coins[0]

	// query tx
	txResult = getTransactions(t, port,
		fmt.Sprintf("message.action=begin_unbonding&message.sender=%s", addr),
		fmt.Sprintf("unbond.validator=%s", operAddrs[0]),
	)
	require.Len(t, txResult.Txs, 1)
	require.Equal(t, resultTx.Height, txResult.Txs[0].Height)

	ubd := getUnbondingDelegation(t, port, addr, operAddrs[0])
	require.Len(t, ubd.Entries, 1)
	require.Equal(t, delTokens.QuoRaw(2), ubd.Entries[0].Balance)

	// test redelegation
	rdTokens := sdk.TokensFromConsensusPower(30)
	resultTx = doBeginRedelegation(t, port, name1, addr, operAddrs[0], operAddrs[1], rdTokens, fees, kb)
	require.Equal(t, uint32(0), resultTx.Code)
	tests.WaitForHeight(resultTx.Height+1, port)

	// query delegations, unbondings and redelegations from validator and delegator
	delegatorDels = getDelegatorDelegations(t, port, addr)
	require.Len(t, delegatorDels, 1)
	require.Equal(t, operAddrs[1], delegatorDels[0].ValidatorAddress)

	// TODO uncomment once all validators actually sign in the lcd tests
	//validator2 := getValidator(t, port, operAddrs[1])
	//delTokensAfterRedelegation := validator2.ShareTokens(delegatorDels[0].GetShares())
	//require.Equal(t, rdTokens.ToDec(), delTokensAfterRedelegation)

	// verify balance after paying fees
	expectedBalance = expectedBalance.Sub(fees[0])
	require.True(t,
		expectedBalance.Amount.LT(coins.AmountOf(sdk.DefaultBondDenom)) ||
			expectedBalance.Amount.Equal(coins.AmountOf(sdk.DefaultBondDenom)),
		"should get tokens back from automatic withdrawal after an unbonding delegation",
	)

	// query tx
	txResult = getTransactions(t, port,
		fmt.Sprintf("message.action=begin_redelegate&message.sender=%s", addr),
		fmt.Sprintf("redelegate.source_validator=%s", operAddrs[0]),
		fmt.Sprintf("redelegate.destination_validator=%s", operAddrs[1]),
	)
	require.Len(t, txResult.Txs, 1)
	require.Equal(t, resultTx.Height, txResult.Txs[0].Height)

	redelegation := getRedelegations(t, port, addr, operAddrs[0], operAddrs[1])
	require.Len(t, redelegation, 1)
	require.Len(t, redelegation[0].Entries, 1)

	delegatorUbds := getDelegatorUnbondingDelegations(t, port, addr)
	require.Len(t, delegatorUbds, 1)
	require.Len(t, delegatorUbds[0].Entries, 1)
	require.Equal(t, rdTokens, delegatorUbds[0].Entries[0].Balance)

	delegatorReds := getRedelegations(t, port, addr, nil, nil)
	require.Len(t, delegatorReds, 1)
	require.Len(t, delegatorReds[0].Entries, 1)

	validatorUbds := getValidatorUnbondingDelegations(t, port, operAddrs[0])
	require.Len(t, validatorUbds, 1)
	require.Len(t, validatorUbds[0].Entries, 1)
	require.Equal(t, rdTokens, validatorUbds[0].Entries[0].Balance)

	validatorReds := getRedelegations(t, port, nil, operAddrs[0], nil)
	require.Len(t, validatorReds, 1)
	require.Len(t, validatorReds[0].Entries, 1)

	// TODO Undonding status not currently implemented
	// require.Equal(t, sdk.Unbonding, bondedValidators[0].Status)

	// query txs
	txs := getBondingTxs(t, port, addr, "")
	require.Len(t, txs, 3, "All Txs found")

	txs = getBondingTxs(t, port, addr, "bond")
	require.Len(t, txs, 1, "All bonding txs found")

	txs = getBondingTxs(t, port, addr, "unbond")
	require.Len(t, txs, 1, "All unbonding txs found")

	txs = getBondingTxs(t, port, addr, "redelegate")
	require.Len(t, txs, 1, "All redelegation txs found")
}

func TestSubmitProposal(t *testing.T) {
	kb, err := newKeybase()
	require.NoError(t, err)
	addr, _, err := CreateAddr(name1, kb)
	require.NoError(t, err)
	cleanup, _, _, port, err := InitializeLCD(1, []sdk.AccAddress{addr}, true)
	require.NoError(t, err)
	defer cleanup()

	initialBalance := getBalances(t, port, addr)

	// create SubmitProposal TX
	proposalTokens := sdk.TokensFromConsensusPower(5)
	resultTx := doSubmitProposal(t, port, name1, addr, proposalTokens, fees, kb)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	require.Equal(t, uint32(0), resultTx.Code)

	bz, err := hex.DecodeString(resultTx.Data)
	require.NoError(t, err)
	proposalID := gov.GetProposalIDFromBytes(bz)

	// verify balance
	balances := getBalances(t, port, addr)
	expectedBalance := initialBalance[0].Sub(fees[0])
	require.Equal(t, expectedBalance.Amount.Sub(proposalTokens), balances.AmountOf(sdk.DefaultBondDenom))

	// query proposal
	proposal := getProposal(t, port, proposalID)
	require.Equal(t, "Test", proposal.GetTitle())

	proposer := getProposer(t, port, proposalID)
	require.Equal(t, addr.String(), proposer.Proposer)
	require.Equal(t, proposalID, proposer.ProposalID)
}

func TestSubmitCommunityPoolSpendProposal(t *testing.T) {
	// TODO: fix this test. Currently it is broken because we create a fault by injecting the tokens
	// directly into genesis
	t.Skip()
	kb, err := newKeybase()
	require.NoError(t, err)
	addr, _, err := CreateAddr(name1, kb)
	require.NoError(t, err)
	cleanup, _, _, port, err := InitializeLCD(1, []sdk.AccAddress{addr}, true)
	require.NoError(t, err)
	defer cleanup()

	initialBalance := getBalances(t, port, addr)

	// create proposal tx
	proposalTokens := sdk.TokensFromConsensusPower(5)
	resultTx := doSubmitCommunityPoolSpendProposal(t, port, name1, addr, proposalTokens, fees, kb)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	require.Equal(t, uint32(0), resultTx.Code)

	bz, err := hex.DecodeString(resultTx.Data)
	require.NoError(t, err)
	proposalID := gov.GetProposalIDFromBytes(bz)

	// verify balance
	balances := getBalances(t, port, addr)
	expectedBalance := initialBalance[0].Sub(fees[0])
	require.Equal(t, expectedBalance.Amount.Sub(proposalTokens), balances.AmountOf(sdk.DefaultBondDenom))

	// query proposal
	proposal := getProposal(t, port, proposalID)
	require.Equal(t, "Test", proposal.GetTitle())

	proposer := getProposer(t, port, proposalID)
	require.Equal(t, addr.String(), proposer.Proposer)
	require.Equal(t, proposalID, proposer.ProposalID)
}

func TestSubmitParamChangeProposal(t *testing.T) {
	kb, err := newKeybase()
	require.NoError(t, err)
	addr, _, err := CreateAddr(name1, kb)
	require.NoError(t, err)
	cleanup, _, _, port, err := InitializeLCD(1, []sdk.AccAddress{addr}, true)
	require.NoError(t, err)
	defer cleanup()

	initialBalance := getBalances(t, port, addr)

	// create proposal tx
	proposalTokens := sdk.TokensFromConsensusPower(5)
	resultTx := doSubmitParamChangeProposal(t, port, name1, addr, proposalTokens, fees, kb)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	require.Equal(t, uint32(0), resultTx.Code)

	bz, err := hex.DecodeString(resultTx.Data)
	require.NoError(t, err)
	proposalID := gov.GetProposalIDFromBytes(bz)

	// verify balance
	balances := getBalances(t, port, addr)
	expectedBalance := initialBalance[0].Sub(fees[0])
	require.Equal(t, expectedBalance.Amount.Sub(proposalTokens), balances.AmountOf(sdk.DefaultBondDenom))

	// query proposal
	proposal := getProposal(t, port, proposalID)
	require.Equal(t, "Test", proposal.GetTitle())

	proposer := getProposer(t, port, proposalID)
	require.Equal(t, addr.String(), proposer.Proposer)
	require.Equal(t, proposalID, proposer.ProposalID)
}

func TestDeposit(t *testing.T) {
	kb, err := newKeybase()
	require.NoError(t, err)
	addr, _, err := CreateAddr(name1, kb)
	require.NoError(t, err)
	cleanup, _, _, port, err := InitializeLCD(1, []sdk.AccAddress{addr}, true)
	require.NoError(t, err)
	defer cleanup()

	initialBalance := getBalances(t, port, addr)

	// create SubmitProposal TX
	proposalTokens := sdk.TokensFromConsensusPower(5)
	resultTx := doSubmitProposal(t, port, name1, addr, proposalTokens, fees, kb)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	require.Equal(t, uint32(0), resultTx.Code)

	bz, err := hex.DecodeString(resultTx.Data)
	require.NoError(t, err)
	proposalID := gov.GetProposalIDFromBytes(bz)

	// verify balance
	coins := getBalances(t, port, addr)
	expectedBalance := initialBalance[0].Sub(fees[0])
	require.Equal(t, expectedBalance.Amount.Sub(proposalTokens), coins.AmountOf(sdk.DefaultBondDenom))
	expectedBalance = coins[0]

	// query proposal
	proposal := getProposal(t, port, proposalID)
	require.Equal(t, "Test", proposal.GetTitle())

	// create SubmitProposal TX
	depositTokens := sdk.TokensFromConsensusPower(5)
	resultTx = doDeposit(t, port, name1, addr, proposalID, depositTokens, fees, kb)
	tests.WaitForHeight(resultTx.Height+1, port)

	// verify balance after deposit and fee
	balances := getBalances(t, port, addr)
	expectedBalance = expectedBalance.Sub(fees[0])
	require.Equal(t, expectedBalance.Amount.Sub(depositTokens), balances.AmountOf(sdk.DefaultBondDenom))

	// query tx
	txResult := getTransactions(t, port, fmt.Sprintf("message.action=deposit&message.sender=%s", addr))
	require.Len(t, txResult.Txs, 1)
	require.Equal(t, resultTx.Height, txResult.Txs[0].Height)

	// query proposal
	totalCoins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(10))}
	proposal = getProposal(t, port, proposalID)
	require.True(t, proposal.TotalDeposit.IsEqual(totalCoins))

	// query deposit
	deposit := getDeposit(t, port, proposalID, addr)
	require.True(t, deposit.Amount.IsEqual(totalCoins))
}

func TestVote(t *testing.T) {
	kb, err := newKeybase()
	require.NoError(t, err)
	addr, _, err := CreateAddr(name1, kb)
	require.NoError(t, err)
	cleanup, _, operAddrs, port, err := InitializeLCD(1, []sdk.AccAddress{addr}, true)
	require.NoError(t, err)
	defer cleanup()

	initialBalance := getBalances(t, port, addr)

	// create SubmitProposal TX
	proposalTokens := sdk.TokensFromConsensusPower(10)
	resultTx := doSubmitProposal(t, port, name1, addr, proposalTokens, fees, kb)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check if tx was committed
	require.Equal(t, uint32(0), resultTx.Code)

	bz, err := hex.DecodeString(resultTx.Data)
	require.NoError(t, err)
	proposalID := gov.GetProposalIDFromBytes(bz)

	// verify balance
	coins := getBalances(t, port, addr)
	expectedBalance := initialBalance[0].Sub(fees[0])
	require.Equal(t, expectedBalance.Amount.Sub(proposalTokens), coins.AmountOf(sdk.DefaultBondDenom))
	expectedBalance = coins[0]

	// query proposal
	proposal := getProposal(t, port, proposalID)
	require.Equal(t, "Test", proposal.GetTitle())
	require.Equal(t, gov.StatusVotingPeriod, proposal.Status)

	// vote
	resultTx = doVote(t, port, name1, addr, proposalID, "Yes", fees, kb)
	tests.WaitForHeight(resultTx.Height+1, port)

	// verify balance after vote and fee
	coins = getBalances(t, port, addr)
	expectedBalance = expectedBalance.Sub(fees[0])
	require.Equal(t, expectedBalance.Amount, coins.AmountOf(sdk.DefaultBondDenom))
	expectedBalance = coins[0]

	// query tx
	txResult := getTransactions(t, port, fmt.Sprintf("message.action=vote&message.sender=%s", addr))
	require.Len(t, txResult.Txs, 1)
	require.Equal(t, resultTx.Height, txResult.Txs[0].Height)

	vote := getVote(t, port, proposalID, addr)
	require.Equal(t, proposalID, vote.ProposalID)
	require.Equal(t, gov.OptionYes, vote.Option)

	tally := getTally(t, port, proposalID)
	require.Equal(t, sdk.ZeroInt(), tally.Yes, "tally should be 0 as the address is not bonded")

	// create bond TX
	delTokens := sdk.TokensFromConsensusPower(60)
	resultTx = doDelegate(t, port, name1, addr, operAddrs[0], delTokens, fees, kb)
	tests.WaitForHeight(resultTx.Height+1, port)

	// verify balance
	coins = getBalances(t, port, addr)
	expectedBalance = expectedBalance.Sub(fees[0])
	require.Equal(t, expectedBalance.Amount.Sub(delTokens), coins.AmountOf(sdk.DefaultBondDenom))
	expectedBalance = coins[0]

	tally = getTally(t, port, proposalID)
	require.Equal(t, delTokens, tally.Yes, "tally should be equal to the amount delegated")

	// change vote option
	resultTx = doVote(t, port, name1, addr, proposalID, "No", fees, kb)
	tests.WaitForHeight(resultTx.Height+1, port)

	// verify balance
	balances := getBalances(t, port, addr)
	expectedBalance = expectedBalance.Sub(fees[0])
	require.Equal(t, expectedBalance.Amount, balances.AmountOf(sdk.DefaultBondDenom))

	tally = getTally(t, port, proposalID)
	require.Equal(t, sdk.ZeroInt(), tally.Yes, "tally should be 0 the user changed the option")
	require.Equal(t, delTokens, tally.No, "tally should be equal to the amount delegated")
}

func TestUnjail(t *testing.T) {
	kb, err := newKeybase()
	require.NoError(t, err)
	addr, _, err := CreateAddr(name1, kb)
	require.NoError(t, err)
	cleanup, valPubKeys, _, port, err := InitializeLCD(1, []sdk.AccAddress{addr}, true)
	require.NoError(t, err)
	defer cleanup()

	// NOTE: any less than this and it fails
	tests.WaitForHeight(3, port)
	pkString, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, valPubKeys[0])
	require.NoError(t, err)
	signingInfo := getSigningInfo(t, port, pkString)
	tests.WaitForHeight(4, port)
	require.Equal(t, true, signingInfo.IndexOffset > 0)
	require.Equal(t, time.Unix(0, 0).UTC(), signingInfo.JailedUntil)
	require.Equal(t, true, signingInfo.MissedBlocksCounter == 0)
	signingInfoList := getSigningInfoList(t, port)
	require.NotZero(t, len(signingInfoList))
}

func TestProposalsQuery(t *testing.T) {
	kb, err := newKeybase()
	require.NoError(t, err)
	addrs, _, names, errors := CreateAddrs(kb, 2)
	require.Empty(t, errors)

	cleanup, _, _, port, err := InitializeLCD(1, []sdk.AccAddress{addrs[0], addrs[1]}, true)
	require.NoError(t, err)
	defer cleanup()

	depositParam := getDepositParam(t, port)
	halfMinDeposit := depositParam.MinDeposit.AmountOf(sdk.DefaultBondDenom).QuoRaw(2)
	getVotingParam(t, port)
	getTallyingParam(t, port)

	// Addr1 proposes (and deposits) proposals #1 and #2
	resultTx := doSubmitProposal(t, port, names[0], addrs[0], halfMinDeposit, fees, kb)
	bz, err := hex.DecodeString(resultTx.Data)
	require.NoError(t, err)

	proposalID1 := gov.GetProposalIDFromBytes(bz)
	tests.WaitForHeight(resultTx.Height+1, port)

	resultTx = doSubmitProposal(t, port, names[0], addrs[0], halfMinDeposit, fees, kb)
	bz, err = hex.DecodeString(resultTx.Data)
	require.NoError(t, err)

	proposalID2 := gov.GetProposalIDFromBytes(bz)
	tests.WaitForHeight(resultTx.Height+1, port)

	// Addr2 proposes (and deposits) proposals #3
	resultTx = doSubmitProposal(t, port, names[1], addrs[1], halfMinDeposit, fees, kb)
	bz, err = hex.DecodeString(resultTx.Data)
	require.NoError(t, err)

	proposalID3 := gov.GetProposalIDFromBytes(bz)
	tests.WaitForHeight(resultTx.Height+1, port)

	// Addr2 deposits on proposals #2 & #3
	resultTx = doDeposit(t, port, names[1], addrs[1], proposalID2, halfMinDeposit, fees, kb)
	tests.WaitForHeight(resultTx.Height+1, port)

	resultTx = doDeposit(t, port, names[1], addrs[1], proposalID3, halfMinDeposit, fees, kb)
	tests.WaitForHeight(resultTx.Height+1, port)

	// check deposits match proposal and individual deposits
	deposits := getDeposits(t, port, proposalID1)
	require.Len(t, deposits, 1)
	deposit := getDeposit(t, port, proposalID1, addrs[0])
	require.Equal(t, deposit, deposits[0])

	deposits = getDeposits(t, port, proposalID2)
	require.Len(t, deposits, 2)
	deposit = getDeposit(t, port, proposalID2, addrs[0])
	require.True(t, deposit.Equal(deposits[0]))
	deposit = getDeposit(t, port, proposalID2, addrs[1])
	require.True(t, deposit.Equal(deposits[1]))

	deposits = getDeposits(t, port, proposalID3)
	require.Len(t, deposits, 1)
	deposit = getDeposit(t, port, proposalID3, addrs[1])
	require.Equal(t, deposit, deposits[0])

	// increasing the amount of the deposit should update the existing one
	depositTokens := sdk.TokensFromConsensusPower(1)
	resultTx = doDeposit(t, port, names[0], addrs[0], proposalID1, depositTokens, fees, kb)
	tests.WaitForHeight(resultTx.Height+1, port)

	deposits = getDeposits(t, port, proposalID1)
	require.Len(t, deposits, 1)

	// Only proposals #1 should be in Deposit Period
	proposals := getProposalsFilterStatus(t, port, gov.StatusDepositPeriod)
	require.Len(t, proposals, 1)
	require.Equal(t, proposalID1, proposals[0].ProposalID)

	// Only proposals #2 and #3 should be in Voting Period
	proposals = getProposalsFilterStatus(t, port, gov.StatusVotingPeriod)
	require.Len(t, proposals, 2)
	require.Equal(t, proposalID2, proposals[0].ProposalID)
	require.Equal(t, proposalID3, proposals[1].ProposalID)

	// Addr1 votes on proposals #2 & #3
	resultTx = doVote(t, port, names[0], addrs[0], proposalID2, "Yes", fees, kb)
	tests.WaitForHeight(resultTx.Height+1, port)
	resultTx = doVote(t, port, names[0], addrs[0], proposalID3, "Yes", fees, kb)
	tests.WaitForHeight(resultTx.Height+1, port)

	// Addr2 votes on proposal #3
	resultTx = doVote(t, port, names[1], addrs[1], proposalID3, "Yes", fees, kb)
	tests.WaitForHeight(resultTx.Height+1, port)

	// Test query all proposals
	proposals = getProposalsAll(t, port)
	require.Equal(t, proposalID1, (proposals[0]).ProposalID)
	require.Equal(t, proposalID2, (proposals[1]).ProposalID)
	require.Equal(t, proposalID3, (proposals[2]).ProposalID)

	// Test query deposited by addr1
	proposals = getProposalsFilterDepositor(t, port, addrs[0])
	require.Equal(t, proposalID1, (proposals[0]).ProposalID)

	// Test query deposited by addr2
	proposals = getProposalsFilterDepositor(t, port, addrs[1])
	require.Equal(t, proposalID2, (proposals[0]).ProposalID)
	require.Equal(t, proposalID3, (proposals[1]).ProposalID)

	// Test query voted by addr1
	proposals = getProposalsFilterVoter(t, port, addrs[0])
	require.Equal(t, proposalID2, (proposals[0]).ProposalID)
	require.Equal(t, proposalID3, (proposals[1]).ProposalID)

	// Test query voted by addr2
	proposals = getProposalsFilterVoter(t, port, addrs[1])
	require.Equal(t, proposalID3, (proposals[0]).ProposalID)

	// Test query voted and deposited by addr1
	proposals = getProposalsFilterVoterDepositor(t, port, addrs[0], addrs[0])
	require.Equal(t, proposalID2, (proposals[0]).ProposalID)

	// Test query votes on Proposal 2
	votes := getVotes(t, port, proposalID2)
	require.Len(t, votes, 1)
	require.Equal(t, addrs[0], votes[0].Voter)

	// Test query votes on Proposal 3
	votes = getVotes(t, port, proposalID3)
	require.Len(t, votes, 2)
	require.True(t, addrs[0].String() == votes[0].Voter.String() || addrs[0].String() == votes[1].Voter.String())
	require.True(t, addrs[1].String() == votes[0].Voter.String() || addrs[1].String() == votes[1].Voter.String())
}

func TestSlashingGetParams(t *testing.T) {
	cleanup, _, _, port, err := InitializeLCD(1, []sdk.AccAddress{}, true)
	require.NoError(t, err)
	defer cleanup()

	res, body := Request(t, port, "GET", "/slashing/parameters", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var params slashing.Params
	err = cdc.UnmarshalJSON([]byte(body), &params)
	require.NoError(t, err)
}

func TestDistributionGetParams(t *testing.T) {
	cleanup, _, _, port, err := InitializeLCD(1, []sdk.AccAddress{}, true)
	require.NoError(t, err)
	defer cleanup()

	res, body := Request(t, port, "GET", "/distribution/parameters", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.NoError(t, cdc.UnmarshalJSON([]byte(body), &disttypes.Params{}))
}

func TestDistributionFlow(t *testing.T) {
	kb, err := newKeybase()
	require.NoError(t, err)
	addr, _, err := CreateAddr(name1, kb)
	require.NoError(t, err)
	cleanup, _, valAddrs, port, err := InitializeLCD(1, []sdk.AccAddress{addr}, true)
	require.NoError(t, err)
	defer cleanup()

	valAddr := valAddrs[0]
	operAddr := sdk.AccAddress(valAddr)

	var outstanding disttypes.ValidatorOutstandingRewards
	res, body := Request(t, port, "GET", fmt.Sprintf("/distribution/validators/%s/outstanding_rewards", valAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.NoError(t, cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &outstanding))

	var valDistInfo distrrest.ValidatorDistInfo
	res, body = Request(t, port, "GET", "/distribution/validators/"+valAddr.String(), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.NoError(t, cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &valDistInfo))
	require.Equal(t, valDistInfo.OperatorAddress.String(), sdk.AccAddress(valAddr).String())

	// Delegate some coins
	delTokens := sdk.TokensFromConsensusPower(60)
	resultTx := doDelegate(t, port, name1, addr, valAddr, delTokens, fees, kb)
	tests.WaitForHeight(resultTx.Height+1, port)
	require.Equal(t, uint32(0), resultTx.Code)

	// send some coins
	_, resultTx = doTransfer(t, port, name1, memo, addr, fees, kb)
	tests.WaitForHeight(resultTx.Height+5, port)
	require.Equal(t, uint32(0), resultTx.Code)

	// Query outstanding rewards changed
	res, body = Request(t, port, "GET", fmt.Sprintf("/distribution/validators/%s/outstanding_rewards", valAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.NoError(t, cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &outstanding))

	// Query validator distribution info
	res, body = Request(t, port, "GET", "/distribution/validators/"+valAddr.String(), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	require.NoError(t, cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &valDistInfo))

	// Query validator's rewards
	var rewards sdk.DecCoins

	res, body = Request(t, port, "GET", fmt.Sprintf("/distribution/validators/%s/rewards", valAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.NoError(t, cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &rewards))

	// Query self-delegation
	res, body = Request(t, port, "GET", fmt.Sprintf("/distribution/delegators/%s/rewards/%s", operAddr, valAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.NoError(t, cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &rewards))

	// Query delegation
	res, body = Request(t, port, "GET", fmt.Sprintf("/distribution/delegators/%s/rewards/%s", addr, valAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.NoError(t, cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &rewards))

	// Query delegator's rewards total
	var delRewards disttypes.QueryDelegatorTotalRewardsResponse
	res, body = Request(t, port, "GET", fmt.Sprintf("/distribution/delegators/%s/rewards", operAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.NoError(t, json.Unmarshal(extractResultFromResponse(t, []byte(body)), &delRewards))

	// Query delegator's withdrawal address
	var withdrawAddr string
	res, body = Request(t, port, "GET", fmt.Sprintf("/distribution/delegators/%s/withdraw_address", operAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.NoError(t, cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &withdrawAddr))
	require.Equal(t, operAddr.String(), withdrawAddr)

	// Withdraw delegator's rewards
	resultTx = doWithdrawDelegatorAllRewards(t, port, name1, addr, fees)
	require.Equal(t, uint32(0), resultTx.Code)
}

func TestMintingQueries(t *testing.T) {
	kb, err := newKeybase()
	require.NoError(t, err)
	addr, _, err := CreateAddr(name1, kb)
	require.NoError(t, err)
	cleanup, _, _, port, err := InitializeLCD(1, []sdk.AccAddress{addr}, true)
	require.NoError(t, err)
	defer cleanup()

	res, body := Request(t, port, "GET", "/minting/parameters", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var params mint.Params
	require.NoError(t, cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &params))

	res, body = Request(t, port, "GET", "/minting/inflation", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var inflation sdk.Dec
	require.NoError(t, cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &inflation))

	res, body = Request(t, port, "GET", "/minting/annual-provisions", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var annualProvisions sdk.Dec
	require.NoError(t, cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &annualProvisions))
}

func TestAccountBalanceQuery(t *testing.T) {
	kb, err := newKeybase()
	require.NoError(t, err)
	addr, _, err := CreateAddr(name1, kb)
	require.NoError(t, err)
	cleanup, _, _, port, err := InitializeLCD(1, []sdk.AccAddress{addr}, true)
	require.NoError(t, err)
	defer cleanup()

	bz, err := hex.DecodeString("8FA6AB57AD6870F6B5B2E57735F38F2F30E73CB6")
	require.NoError(t, err)
	someFakeAddr := sdk.AccAddress(bz)

	// empty account
	res, body := Request(t, port, "GET", fmt.Sprintf("/auth/accounts/%s", someFakeAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.Contains(t, body, `"type":"cosmos-sdk/Account"`)

	// empty account balance
	res, body = Request(t, port, "GET", fmt.Sprintf("/bank/balances/%s", someFakeAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.Contains(t, body, "[]")

}
