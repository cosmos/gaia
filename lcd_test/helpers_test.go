package lcdtest

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	crkeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	authcutils "github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	bankrest "github.com/cosmos/cosmos-sdk/x/bank/client/rest"
	distrrest "github.com/cosmos/cosmos-sdk/x/distribution/client/rest"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
	gcutils "github.com/cosmos/cosmos-sdk/x/gov/client/utils"
	paramscutils "github.com/cosmos/cosmos-sdk/x/params/client/utils"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingrest "github.com/cosmos/cosmos-sdk/x/slashing/client/rest"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingrest "github.com/cosmos/cosmos-sdk/x/staking/client/rest"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// Request makes a test LCD test request. It returns a response object and a
// stringified response body.
func Request(t *testing.T, port, method, path string, payload []byte) (*http.Response, string) {
	var (
		err error
		res *http.Response
	)
	url := fmt.Sprintf("http://localhost:%v%v", port, path)
	fmt.Printf("REQUEST %s %s\n", method, url)

	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	require.Nil(t, err)

	res, err = http.DefaultClient.Do(req)
	require.Nil(t, err)

	output, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	require.Nil(t, err)

	return res, string(output)
}

// ----------------------------------------------------------------------
// ICS 0 - Tendermint
// ----------------------------------------------------------------------
// GET /node_info The properties of the connected node
func getNodeInfo(t *testing.T, port string) rpc.NodeInfoResponse {
	res, body := Request(t, port, "GET", "/node_info", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var nodeInfo rpc.NodeInfoResponse
	err := cdc.UnmarshalJSON([]byte(body), &nodeInfo)
	require.Nil(t, err, "failed parse node info")

	require.NotEqual(t, rpc.NodeInfoResponse{}, nodeInfo, "res: %v", res)
	return nodeInfo
}

// GET /syncing Syncing state of node
func getSyncStatus(t *testing.T, port string, syncing bool) {
	res, body := Request(t, port, "GET", "/syncing", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var syncResp rpc.SyncingResponse
	err := cdc.UnmarshalJSON([]byte(body), &syncResp)
	require.Nil(t, err, "failed parse syncing info")

	require.Equal(t, syncResp.Syncing, syncing)
}

// GET /blocks/latest Get the latest block
// GET /blocks/{height} Get a block at a certain height
func getBlock(t *testing.T, port string, height int, expectFail bool) ctypes.ResultBlock {
	var url string
	if height > 0 {
		url = fmt.Sprintf("/blocks/%d", height)
	} else {
		url = "/blocks/latest"
	}
	var resultBlock ctypes.ResultBlock

	res, body := Request(t, port, "GET", url, nil)
	if expectFail {
		require.Equal(t, http.StatusNotFound, res.StatusCode, body)
		return resultBlock
	}
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err := cdc.UnmarshalJSON([]byte(body), &resultBlock)
	require.Nil(t, err, "Couldn't parse block")

	require.NotEqual(t, ctypes.ResultBlock{}, resultBlock)
	return resultBlock
}

func extractResultFromResponse(t *testing.T, body []byte) []byte {
	var resp rest.ResponseWithHeight
	require.NoError(t, cdc.UnmarshalJSON([]byte(body), &resp))

	return resp.Result
}

// GET /validatorsets/{height} Get a validator set a certain height
// GET /validatorsets/latest Get the latest validator set
func getValidatorSets(t *testing.T, port string, height int, expectFail bool) rpc.ResultValidatorsOutput {
	var url string
	if height > 0 {
		url = fmt.Sprintf("/validatorsets/%d", height)
	} else {
		url = "/validatorsets/latest"
	}

	var resultVals rpc.ResultValidatorsOutput
	res, body := Request(t, port, "GET", url, nil)

	if expectFail {
		require.Equal(t, http.StatusNotFound, res.StatusCode, body)
		return resultVals
	}

	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &resultVals)
	require.Nil(t, err, "Couldn't parse validator set")

	require.NotEqual(t, rpc.ResultValidatorsOutput{}, resultVals)
	return resultVals
}

// GET /txs/{hash} get tx by hash
func getTransaction(t *testing.T, port string, hash string) sdk.TxResponse {
	var tx sdk.TxResponse
	res, body := getTransactionRequest(t, port, hash)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err := cdc.UnmarshalJSON([]byte(body), &tx)
	require.NoError(t, err)
	return tx
}

func getTransactionRequest(t *testing.T, port, hash string) (*http.Response, string) {
	return Request(t, port, "GET", fmt.Sprintf("/txs/%s", hash), nil)
}

// POST /txs broadcast txs

// GET /txs search transactions
func getTransactions(t *testing.T, port string, tags ...string) *sdk.SearchTxsResult {
	var txs []sdk.TxResponse
	result := sdk.NewSearchTxsResult(0, 0, 1, 30, txs)
	if len(tags) == 0 {
		return &result
	}
	queryStr := strings.Join(tags, "&")
	res, body := Request(t, port, "GET", fmt.Sprintf("/txs?%s", queryStr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err := cdc.UnmarshalJSON([]byte(body), &result)
	require.NoError(t, err)
	return &result
}

// ----------------------------------------------------------------------
// ICS 1 - Keys
// ----------------------------------------------------------------------
// GET /keys List of accounts stored locally
func getKeys(t *testing.T, port string) []keys.KeyOutput {
	res, body := Request(t, port, "GET", "/keys", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var m []keys.KeyOutput
	err := cdc.UnmarshalJSON([]byte(body), &m)
	require.Nil(t, err)
	return m
}

// POST /keys Create a new account locally
func doKeysPost(t *testing.T, port, name, password, mnemonic string, account int, index int) keys.KeyOutput {
	pk := clientkeys.NewAddNewKey(name, password, mnemonic, account, index)
	req, err := cdc.MarshalJSON(pk)
	require.NoError(t, err)

	res, body := Request(t, port, "POST", "/keys", req)

	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var resp keys.KeyOutput
	err = cdc.UnmarshalJSON([]byte(body), &resp)
	require.Nil(t, err, body)
	return resp
}

// GET /keys/seed Create a new seed to create a new account defaultValidFor
func getKeysSeed(t *testing.T, port string) string {
	res, body := Request(t, port, "GET", "/keys/seed", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	reg, err := regexp.Compile(`([a-z]+ ){12}`)
	require.Nil(t, err)
	match := reg.MatchString(body)
	require.True(t, match, "Returned seed has wrong format", body)
	return body
}

// POST /keys/{name}/recove Recover a account from a seed
func doRecoverKey(t *testing.T, port, recoverName, recoverPassword, mnemonic string, account uint32, index uint32) {
	pk := clientkeys.NewRecoverKey(recoverPassword, mnemonic, int(account), int(index))
	req, err := cdc.MarshalJSON(pk)
	require.NoError(t, err)

	res, body := Request(t, port, "POST", fmt.Sprintf("/keys/%s/recover", recoverName), req)

	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var resp keys.KeyOutput
	err = codec.Cdc.UnmarshalJSON([]byte(body), &resp)
	require.Nil(t, err, body)

	addr1Bech32 := resp.Address
	_, err = sdk.AccAddressFromBech32(addr1Bech32)
	require.NoError(t, err, "Failed to return a correct bech32 address")
}

// GET /keys/{name} Get a certain locally stored account
func getKey(t *testing.T, port, name string) keys.KeyOutput {
	res, body := Request(t, port, "GET", fmt.Sprintf("/keys/%s", name), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var resp keys.KeyOutput
	err := cdc.UnmarshalJSON([]byte(body), &resp)
	require.Nil(t, err)
	return resp
}

// PUT /keys/{name} Update the password for this account in the KMS
func updateKey(t *testing.T, port, name, oldPassword, newPassword string, fail bool) {
	kr := clientkeys.NewUpdateKeyReq(oldPassword, newPassword)
	req, err := cdc.MarshalJSON(kr)
	require.NoError(t, err)
	keyEndpoint := fmt.Sprintf("/keys/%s", name)
	res, body := Request(t, port, "PUT", keyEndpoint, req)
	if fail {
		require.Equal(t, http.StatusUnauthorized, res.StatusCode, body)
		return
	}
	require.Equal(t, http.StatusOK, res.StatusCode, body)
}

// GET /auth/accounts/{address} Get the account information on blockchain
func getAccount(t *testing.T, port string, addr sdk.AccAddress) (acc authexported.Account) {
	res, body := Request(t, port, "GET", fmt.Sprintf("/auth/accounts/%s", addr.String()), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.Nil(t, cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &acc))

	return acc
}

// ----------------------------------------------------------------------
// ICS 20 - Tokens
// ----------------------------------------------------------------------

// POST /tx/broadcast Send a signed Tx
func doBroadcast(t *testing.T, port string, tx auth.StdTx) (*http.Response, string) {
	txReq := authrest.BroadcastReq{Tx: tx, Mode: "block"}

	req, err := cdc.MarshalJSON(txReq)
	require.Nil(t, err)

	return Request(t, port, "POST", "/txs", req)
}

// doTransfer performs a balance transfer with auto gas calculation. It also signs
// the tx and broadcasts it.
func doTransfer(
	t *testing.T, port, seed, name, memo, pwd string, addr sdk.AccAddress, fees sdk.Coins,
) (sdk.AccAddress, sdk.TxResponse) {

	resp, body, recvAddr := doTransferWithGas(
		t, port, seed, name, memo, pwd, addr, "", 1.0, false, true, fees,
	)
	require.Equal(t, http.StatusOK, resp.StatusCode, body)

	var txResp sdk.TxResponse
	err := cdc.UnmarshalJSON([]byte(body), &txResp)
	require.NoError(t, err)

	return recvAddr, txResp
}

// doTransferWithGas performs a balance transfer with a specified gas value. The
// broadcast parameter determines if the tx should only be generated or also
// signed and broadcasted. The sending account's number and sequence are
// determined prior to generating the tx.
func doTransferWithGas(
	t *testing.T, port, seed, name, memo, pwd string, addr sdk.AccAddress,
	gas string, gasAdjustment float64, simulate, broadcast bool, fees sdk.Coins,
) (resp *http.Response, body string, receiveAddr sdk.AccAddress) {

	// create receive address
	kb := crkeys.NewInMemory()

	receiveInfo, _, err := kb.CreateMnemonic(
		"receive_address", crkeys.English, client.DefaultKeyPass, crkeys.SigningAlgo("secp256k1"),
	)
	require.Nil(t, err)

	receiveAddr = sdk.AccAddress(receiveInfo.GetPubKey().Address())
	acc := getAccount(t, port, addr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()
	chainID := viper.GetString(client.FlagChainID)

	from := addr.String()
	baseReq := rest.NewBaseReq(
		from, memo, chainID, gas, fmt.Sprintf("%f", gasAdjustment), accnum, sequence, fees, nil, simulate,
	)

	sr := bankrest.SendReq{
		Amount:  sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)},
		BaseReq: baseReq,
	}

	req, err := cdc.MarshalJSON(sr)
	require.NoError(t, err)

	// generate tx
	resp, body = Request(t, port, "POST", fmt.Sprintf("/bank/accounts/%s/transfers", receiveAddr), req)
	if !broadcast {
		return resp, body, receiveAddr
	}

	// sign and broadcast
	resp, body = signAndBroadcastGenTx(t, port, name, pwd, body, acc, gasAdjustment, simulate)
	return resp, body, receiveAddr
}

// doTransferWithGasAccAuto is similar to doTransferWithGas except that it
// automatically determines the account's number and sequence when generating the
// tx.
func doTransferWithGasAccAuto(
	t *testing.T, port, seed, name, memo, pwd string, addr sdk.AccAddress,
	gas string, gasAdjustment float64, simulate, broadcast bool, fees sdk.Coins,
) (resp *http.Response, body string, receiveAddr sdk.AccAddress) {

	// create receive address
	kb := crkeys.NewInMemory()
	acc := getAccount(t, port, addr)

	receiveInfo, _, err := kb.CreateMnemonic(
		"receive_address", crkeys.English, client.DefaultKeyPass, crkeys.SigningAlgo("secp256k1"),
	)
	require.Nil(t, err)

	receiveAddr = sdk.AccAddress(receiveInfo.GetPubKey().Address())
	chainID := viper.GetString(client.FlagChainID)

	from := addr.String()
	baseReq := rest.NewBaseReq(
		from, memo, chainID, gas, fmt.Sprintf("%f", gasAdjustment), 0, 0, fees, nil, simulate,
	)

	sr := bankrest.SendReq{
		Amount:  sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)},
		BaseReq: baseReq,
	}

	req, err := cdc.MarshalJSON(sr)
	require.NoError(t, err)

	resp, body = Request(t, port, "POST", fmt.Sprintf("/bank/accounts/%s/transfers", receiveAddr), req)
	if !broadcast {
		return resp, body, receiveAddr
	}

	// sign and broadcast
	resp, body = signAndBroadcastGenTx(t, port, name, pwd, body, acc, gasAdjustment, simulate)
	return resp, body, receiveAddr
}

// signAndBroadcastGenTx accepts a successfully generated unsigned tx, signs it,
// and broadcasts it.
func signAndBroadcastGenTx(
	t *testing.T, port, name, pwd, genTx string, acc authexported.Account, gasAdjustment float64, simulate bool,
) (resp *http.Response, body string) {

	chainID := viper.GetString(client.FlagChainID)

	var tx auth.StdTx
	err := cdc.UnmarshalJSON([]byte(genTx), &tx)
	require.Nil(t, err)

	txbldr := auth.NewTxBuilder(
		authcutils.GetTxEncoder(cdc),
		acc.GetAccountNumber(),
		acc.GetSequence(),
		tx.Fee.Gas,
		gasAdjustment,
		simulate,
		chainID,
		tx.Memo,
		tx.Fee.Amount,
		nil,
	)

	signedTx, err := txbldr.SignStdTx(name, pwd, tx, false)
	require.NoError(t, err)

	return doBroadcast(t, port, signedTx)
}

// ----------------------------------------------------------------------
// ICS 21 - Stake
// ----------------------------------------------------------------------

// POST /staking/delegators/{delegatorAddr}/delegations Submit delegation
func doDelegate(
	t *testing.T, port, name, pwd string, delAddr sdk.AccAddress,
	valAddr sdk.ValAddress, amount sdk.Int, fees sdk.Coins,
) sdk.TxResponse {

	acc := getAccount(t, port, delAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()
	chainID := viper.GetString(client.FlagChainID)
	from := acc.GetAddress().String()

	baseReq := rest.NewBaseReq(from, "", chainID, "", "", accnum, sequence, fees, nil, false)
	msg := stakingrest.DelegateRequest{
		BaseReq:          baseReq,
		DelegatorAddress: delAddr,
		ValidatorAddress: valAddr,
		Amount:           sdk.NewCoin(sdk.DefaultBondDenom, amount),
	}

	req, err := cdc.MarshalJSON(msg)
	require.NoError(t, err)

	resp, body := Request(t, port, "POST", fmt.Sprintf("/staking/delegators/%s/delegations", delAddr.String()), req)
	require.Equal(t, http.StatusOK, resp.StatusCode, body)

	// sign and broadcast
	resp, body = signAndBroadcastGenTx(t, port, name, pwd, body, acc, client.DefaultGasAdjustment, false)
	require.Equal(t, http.StatusOK, resp.StatusCode, body)

	var txResp sdk.TxResponse
	err = cdc.UnmarshalJSON([]byte(body), &txResp)
	require.NoError(t, err)

	return txResp
}

// POST /staking/delegators/{delegatorAddr}/delegations Submit delegation
func doUndelegate(
	t *testing.T, port, name, pwd string, delAddr sdk.AccAddress,
	valAddr sdk.ValAddress, amount sdk.Int, fees sdk.Coins,
) sdk.TxResponse {

	acc := getAccount(t, port, delAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()
	chainID := viper.GetString(client.FlagChainID)
	from := acc.GetAddress().String()

	baseReq := rest.NewBaseReq(from, "", chainID, "", "", accnum, sequence, fees, nil, false)
	msg := stakingrest.UndelegateRequest{
		BaseReq:          baseReq,
		DelegatorAddress: delAddr,
		ValidatorAddress: valAddr,
		Amount:           sdk.NewCoin(sdk.DefaultBondDenom, amount),
	}

	req, err := cdc.MarshalJSON(msg)
	require.NoError(t, err)

	resp, body := Request(t, port, "POST", fmt.Sprintf("/staking/delegators/%s/unbonding_delegations", delAddr), req)
	require.Equal(t, http.StatusOK, resp.StatusCode, body)

	resp, body = signAndBroadcastGenTx(t, port, name, pwd, body, acc, client.DefaultGasAdjustment, false)
	require.Equal(t, http.StatusOK, resp.StatusCode, body)

	var txResp sdk.TxResponse
	err = cdc.UnmarshalJSON([]byte(body), &txResp)
	require.NoError(t, err)

	return txResp
}

// POST /staking/delegators/{delegatorAddr}/delegations Submit delegation
func doBeginRedelegation(
	t *testing.T, port, name, pwd string, delAddr sdk.AccAddress, valSrcAddr,
	valDstAddr sdk.ValAddress, amount sdk.Int, fees sdk.Coins,
) sdk.TxResponse {

	acc := getAccount(t, port, delAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()
	chainID := viper.GetString(client.FlagChainID)
	from := acc.GetAddress().String()

	baseReq := rest.NewBaseReq(from, "", chainID, "", "", accnum, sequence, fees, nil, false)
	msg := stakingrest.RedelegateRequest{
		BaseReq:             baseReq,
		DelegatorAddress:    delAddr,
		ValidatorSrcAddress: valSrcAddr,
		ValidatorDstAddress: valDstAddr,
		Amount:              sdk.NewCoin(sdk.DefaultBondDenom, amount),
	}

	req, err := cdc.MarshalJSON(msg)
	require.NoError(t, err)

	resp, body := Request(t, port, "POST", fmt.Sprintf("/staking/delegators/%s/redelegations", delAddr), req)
	require.Equal(t, http.StatusOK, resp.StatusCode, body)

	resp, body = signAndBroadcastGenTx(t, port, name, pwd, body, acc, client.DefaultGasAdjustment, false)
	require.Equal(t, http.StatusOK, resp.StatusCode, body)

	var txResp sdk.TxResponse
	err = cdc.UnmarshalJSON([]byte(body), &txResp)
	require.NoError(t, err)

	return txResp
}

// GET /staking/delegators/{delegatorAddr}/delegations Get all delegations from a delegator
func getDelegatorDelegations(t *testing.T, port string, delegatorAddr sdk.AccAddress) staking.DelegationResponses {
	res, body := Request(t, port, "GET", fmt.Sprintf("/staking/delegators/%s/delegations", delegatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var dels staking.DelegationResponses

	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &dels)
	require.Nil(t, err)

	return dels
}

// GET /staking/delegators/{delegatorAddr}/unbonding_delegations Get all unbonding delegations from a delegator
func getDelegatorUnbondingDelegations(t *testing.T, port string, delegatorAddr sdk.AccAddress) []staking.UnbondingDelegation {
	res, body := Request(t, port, "GET", fmt.Sprintf("/staking/delegators/%s/unbonding_delegations", delegatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var ubds []staking.UnbondingDelegation

	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &ubds)
	require.Nil(t, err)

	return ubds
}

// GET /staking/redelegations?delegator=0xdeadbeef&validator_from=0xdeadbeef&validator_to=0xdeadbeef& Get redelegations filters by params passed in
func getRedelegations(t *testing.T, port string, delegatorAddr sdk.AccAddress, srcValidatorAddr sdk.ValAddress, dstValidatorAddr sdk.ValAddress) staking.RedelegationResponses {
	var res *http.Response
	var body string
	endpoint := "/staking/redelegations?"
	if !delegatorAddr.Empty() {
		endpoint += fmt.Sprintf("delegator=%s&", delegatorAddr)
	}
	if !srcValidatorAddr.Empty() {
		endpoint += fmt.Sprintf("validator_from=%s&", srcValidatorAddr)
	}
	if !dstValidatorAddr.Empty() {
		endpoint += fmt.Sprintf("validator_to=%s&", dstValidatorAddr)
	}

	res, body = Request(t, port, "GET", endpoint, nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var redels staking.RedelegationResponses
	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &redels)
	require.Nil(t, err)

	return redels
}

// GET /staking/delegators/{delegatorAddr}/validators Query all validators that a delegator is bonded to
func getDelegatorValidators(t *testing.T, port string, delegatorAddr sdk.AccAddress) []staking.Validator {
	res, body := Request(t, port, "GET", fmt.Sprintf("/staking/delegators/%s/validators", delegatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var bondedValidators []staking.Validator

	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &bondedValidators)
	require.Nil(t, err)

	return bondedValidators
}

// GET /staking/delegators/{delegatorAddr}/validators/{validatorAddr} Query a validator that a delegator is bonded to
func getDelegatorValidator(t *testing.T, port string, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) staking.Validator {
	res, body := Request(t, port, "GET", fmt.Sprintf("/staking/delegators/%s/validators/%s", delegatorAddr, validatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var bondedValidator staking.Validator
	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &bondedValidator)
	require.Nil(t, err)

	return bondedValidator
}

// GET /staking/delegators/{delegatorAddr}/txs Get all staking txs (i.e msgs) from a delegator
func getBondingTxs(t *testing.T, port string, delegatorAddr sdk.AccAddress, query string) []sdk.TxResponse {
	var res *http.Response
	var body string

	if len(query) > 0 {
		res, body = Request(t, port, "GET", fmt.Sprintf("/staking/delegators/%s/txs?type=%s", delegatorAddr, query), nil)
	} else {
		res, body = Request(t, port, "GET", fmt.Sprintf("/staking/delegators/%s/txs", delegatorAddr), nil)
	}
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var txs []sdk.TxResponse

	err := cdc.UnmarshalJSON([]byte(body), &txs)
	require.Nil(t, err)

	return txs
}

// GET /staking/delegators/{delegatorAddr}/delegations/{validatorAddr} Query the current delegation between a delegator and a validator
func getDelegation(t *testing.T, port string, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) staking.DelegationResponse {
	res, body := Request(t, port, "GET", fmt.Sprintf("/staking/delegators/%s/delegations/%s", delegatorAddr, validatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var bond staking.DelegationResponse
	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &bond)
	require.Nil(t, err)

	return bond
}

// GET /staking/delegators/{delegatorAddr}/unbonding_delegations/{validatorAddr} Query all unbonding delegations between a delegator and a validator
func getUnbondingDelegation(t *testing.T, port string, delegatorAddr sdk.AccAddress,
	validatorAddr sdk.ValAddress) staking.UnbondingDelegation {

	res, body := Request(t, port, "GET",
		fmt.Sprintf("/staking/delegators/%s/unbonding_delegations/%s",
			delegatorAddr, validatorAddr), nil)

	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var unbond staking.UnbondingDelegation
	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &unbond)
	require.Nil(t, err)

	return unbond
}

// GET /staking/validators Get all validator candidates
func getValidators(t *testing.T, port string) []staking.Validator {
	res, body := Request(t, port, "GET", "/staking/validators", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var validators []staking.Validator
	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &validators)
	require.Nil(t, err)

	return validators
}

// GET /staking/validators/{validatorAddr} Query the information from a single validator
func getValidator(t *testing.T, port string, validatorAddr sdk.ValAddress) staking.Validator {
	res, body := Request(t, port, "GET", fmt.Sprintf("/staking/validators/%s", validatorAddr.String()), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var validator staking.Validator
	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &validator)
	require.Nil(t, err)

	return validator
}

// GET /staking/validators/{validatorAddr}/delegations Get all delegations from a validator
func getValidatorDelegations(t *testing.T, port string, validatorAddr sdk.ValAddress) []staking.Delegation {
	res, body := Request(t, port, "GET", fmt.Sprintf("/staking/validators/%s/delegations", validatorAddr.String()), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var delegations []staking.Delegation
	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &delegations)
	require.Nil(t, err)

	return delegations
}

// GET /staking/validators/{validatorAddr}/unbonding_delegations Get all unbonding delegations from a validator
func getValidatorUnbondingDelegations(t *testing.T, port string, validatorAddr sdk.ValAddress) []staking.UnbondingDelegation {
	res, body := Request(t, port, "GET", fmt.Sprintf("/staking/validators/%s/unbonding_delegations", validatorAddr.String()), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var ubds []staking.UnbondingDelegation
	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &ubds)
	require.Nil(t, err)

	return ubds
}

// ----------------------------------------------------------------------
// ICS 22 - Gov
// ----------------------------------------------------------------------
// POST /gov/proposals Submit a proposal
func doSubmitProposal(
	t *testing.T, port, seed, name, pwd string, proposerAddr sdk.AccAddress,
	amount sdk.Int, fees sdk.Coins,
) sdk.TxResponse {

	acc := getAccount(t, port, proposerAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()
	chainID := viper.GetString(client.FlagChainID)
	from := acc.GetAddress().String()

	baseReq := rest.NewBaseReq(from, "", chainID, "", "", accnum, sequence, fees, nil, false)
	pr := govrest.PostProposalReq{
		Title:          "Test",
		Description:    "test",
		ProposalType:   "Text",
		Proposer:       proposerAddr,
		InitialDeposit: sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, amount)},
		BaseReq:        baseReq,
	}

	req, err := cdc.MarshalJSON(pr)
	require.NoError(t, err)

	// submitproposal
	resp, body := Request(t, port, "POST", "/gov/proposals", req)
	require.Equal(t, http.StatusOK, resp.StatusCode, body)

	resp, body = signAndBroadcastGenTx(t, port, name, pwd, body, acc, client.DefaultGasAdjustment, false)
	require.Equal(t, http.StatusOK, resp.StatusCode, body)

	var txResp sdk.TxResponse
	err = cdc.UnmarshalJSON([]byte(body), &txResp)
	require.NoError(t, err)

	return txResp
}

func doSubmitParamChangeProposal(
	t *testing.T, port, seed, name, pwd string, proposerAddr sdk.AccAddress,
	amount sdk.Int, fees sdk.Coins,
) sdk.TxResponse {

	acc := getAccount(t, port, proposerAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()
	chainID := viper.GetString(client.FlagChainID)
	from := acc.GetAddress().String()

	baseReq := rest.NewBaseReq(from, "", chainID, "", "", accnum, sequence, fees, nil, false)
	pr := paramscutils.ParamChangeProposalReq{
		BaseReq:     baseReq,
		Title:       "Test",
		Description: "test",
		Proposer:    proposerAddr,
		Deposit:     sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, amount)},
		Changes: paramscutils.ParamChangesJSON{
			paramscutils.NewParamChangeJSON("staking", "MaxValidators", "", []byte(`105`)),
		},
	}

	req, err := cdc.MarshalJSON(pr)
	require.NoError(t, err)

	resp, body := Request(t, port, "POST", "/gov/proposals/param_change", req)
	require.Equal(t, http.StatusOK, resp.StatusCode, body)

	resp, body = signAndBroadcastGenTx(t, port, name, pwd, body, acc, client.DefaultGasAdjustment, false)
	require.Equal(t, http.StatusOK, resp.StatusCode, body)

	var txResp sdk.TxResponse
	err = cdc.UnmarshalJSON([]byte(body), &txResp)
	require.NoError(t, err)

	return txResp
}

func doSubmitCommunityPoolSpendProposal(
	t *testing.T, port, seed, name, pwd string, proposerAddr sdk.AccAddress,
	amount sdk.Int, fees sdk.Coins,
) sdk.TxResponse {

	acc := getAccount(t, port, proposerAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()
	chainID := viper.GetString(client.FlagChainID)
	from := acc.GetAddress().String()

	baseReq := rest.NewBaseReq(from, "", chainID, "", "", accnum, sequence, fees, nil, false)
	pr := distrrest.CommunityPoolSpendProposalReq{
		BaseReq:     baseReq,
		Title:       "Test",
		Description: "test",
		Proposer:    proposerAddr,
		Recipient:   proposerAddr,
		Deposit:     sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, amount)},
		Amount:      sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(5))},
	}

	req, err := cdc.MarshalJSON(pr)
	require.NoError(t, err)

	resp, body := Request(t, port, "POST", "/gov/proposals/community_pool_spend", req)
	require.Equal(t, http.StatusOK, resp.StatusCode, body)

	resp, body = signAndBroadcastGenTx(t, port, name, pwd, body, acc, client.DefaultGasAdjustment, false)
	require.Equal(t, http.StatusOK, resp.StatusCode, body)

	var txResp sdk.TxResponse
	err = cdc.UnmarshalJSON([]byte(body), &txResp)
	require.NoError(t, err)

	return txResp
}

// GET /gov/proposals Query proposals
func getProposalsAll(t *testing.T, port string) []gov.Proposal {
	res, body := Request(t, port, "GET", "/gov/proposals", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var proposals []gov.Proposal
	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &proposals)
	require.Nil(t, err)
	return proposals
}

// GET /gov/proposals?depositor=%s Query proposals
func getProposalsFilterDepositor(t *testing.T, port string, depositorAddr sdk.AccAddress) []gov.Proposal {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals?depositor=%s", depositorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var proposals []gov.Proposal
	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &proposals)
	require.Nil(t, err)
	return proposals
}

// GET /gov/proposals?voter=%s Query proposals
func getProposalsFilterVoter(t *testing.T, port string, voterAddr sdk.AccAddress) []gov.Proposal {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals?voter=%s", voterAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var proposals []gov.Proposal
	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &proposals)
	require.Nil(t, err)
	return proposals
}

// GET /gov/proposals?depositor=%s&voter=%s Query proposals
func getProposalsFilterVoterDepositor(t *testing.T, port string, voterAddr, depositorAddr sdk.AccAddress) []gov.Proposal {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals?depositor=%s&voter=%s", depositorAddr, voterAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var proposals []gov.Proposal
	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &proposals)
	require.Nil(t, err)
	return proposals
}

// GET /gov/proposals?status=%s Query proposals
func getProposalsFilterStatus(t *testing.T, port string, status gov.ProposalStatus) []gov.Proposal {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals?status=%s", status), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var proposals []gov.Proposal
	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &proposals)
	require.Nil(t, err)
	return proposals
}

// POST /gov/proposals/{proposalId}/deposits Deposit tokens to a proposal
func doDeposit(
	t *testing.T, port, seed, name, pwd string, proposerAddr sdk.AccAddress,
	proposalID uint64, amount sdk.Int, fees sdk.Coins,
) sdk.TxResponse {

	acc := getAccount(t, port, proposerAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()
	chainID := viper.GetString(client.FlagChainID)
	from := acc.GetAddress().String()

	baseReq := rest.NewBaseReq(from, "", chainID, "", "", accnum, sequence, fees, nil, false)
	dr := govrest.DepositReq{
		Depositor: proposerAddr,
		Amount:    sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, amount)},
		BaseReq:   baseReq,
	}

	req, err := cdc.MarshalJSON(dr)
	require.NoError(t, err)

	resp, body := Request(t, port, "POST", fmt.Sprintf("/gov/proposals/%d/deposits", proposalID), req)
	require.Equal(t, http.StatusOK, resp.StatusCode, body)

	resp, body = signAndBroadcastGenTx(t, port, name, pwd, body, acc, client.DefaultGasAdjustment, false)
	require.Equal(t, http.StatusOK, resp.StatusCode, body)

	var txResp sdk.TxResponse
	err = cdc.UnmarshalJSON([]byte(body), &txResp)
	require.NoError(t, err)

	return txResp
}

// GET /gov/proposals/{proposalId}/deposits Query deposits
func getDeposits(t *testing.T, port string, proposalID uint64) []gov.Deposit {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d/deposits", proposalID), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var deposits []gov.Deposit
	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &deposits)
	require.Nil(t, err)
	return deposits
}

// GET /gov/proposals/{proposalId}/tally Get a proposal's tally result at the current time
func getTally(t *testing.T, port string, proposalID uint64) gov.TallyResult {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d/tally", proposalID), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var tally gov.TallyResult
	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &tally)
	require.Nil(t, err)

	return tally
}

// POST /gov/proposals/{proposalId}/votes Vote a proposal
func doVote(
	t *testing.T, port, seed, name, pwd string, proposerAddr sdk.AccAddress,
	proposalID uint64, option string, fees sdk.Coins,
) sdk.TxResponse {

	// get the account to get the sequence
	acc := getAccount(t, port, proposerAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()
	chainID := viper.GetString(client.FlagChainID)
	from := acc.GetAddress().String()

	baseReq := rest.NewBaseReq(from, "", chainID, "", "", accnum, sequence, fees, nil, false)
	vr := govrest.VoteReq{
		Voter:   proposerAddr,
		Option:  option,
		BaseReq: baseReq,
	}

	req, err := cdc.MarshalJSON(vr)
	require.NoError(t, err)

	resp, body := Request(t, port, "POST", fmt.Sprintf("/gov/proposals/%d/votes", proposalID), req)
	require.Equal(t, http.StatusOK, resp.StatusCode, body)

	resp, body = signAndBroadcastGenTx(t, port, name, pwd, body, acc, client.DefaultGasAdjustment, false)
	require.Equal(t, http.StatusOK, resp.StatusCode, body)

	var txResp sdk.TxResponse
	err = cdc.UnmarshalJSON([]byte(body), &txResp)
	require.NoError(t, err)

	return txResp
}

// GET /gov/proposals/{proposalId}/votes Query voters
func getVotes(t *testing.T, port string, proposalID uint64) []gov.Vote {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d/votes", proposalID), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var votes []gov.Vote
	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &votes)
	require.Nil(t, err)
	return votes
}

// GET /gov/proposals/{proposalId} Query a proposal
func getProposal(t *testing.T, port string, proposalID uint64) gov.Proposal {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d", proposalID), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var proposal gov.Proposal
	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &proposal)
	require.Nil(t, err)

	return proposal
}

// GET /gov/proposals/{proposalId}/deposits/{depositor} Query deposit
func getDeposit(t *testing.T, port string, proposalID uint64, depositorAddr sdk.AccAddress) gov.Deposit {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d/deposits/%s", proposalID, depositorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var deposit gov.Deposit
	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &deposit)
	require.Nil(t, err)

	return deposit
}

// GET /gov/proposals/{proposalId}/votes/{voter} Query vote
func getVote(t *testing.T, port string, proposalID uint64, voterAddr sdk.AccAddress) gov.Vote {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d/votes/%s", proposalID, voterAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var vote gov.Vote
	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &vote)
	require.Nil(t, err)

	return vote
}

// GET /gov/proposals/{proposalId}/proposer
func getProposer(t *testing.T, port string, proposalID uint64) gcutils.Proposer {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d/proposer", proposalID), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var proposer gcutils.Proposer
	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &proposer)

	require.Nil(t, err)
	return proposer
}

// GET /gov/parameters/deposit Query governance deposit parameters
func getDepositParam(t *testing.T, port string) gov.DepositParams {
	res, body := Request(t, port, "GET", "/gov/parameters/deposit", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var depositParams gov.DepositParams
	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &depositParams)
	require.Nil(t, err)

	return depositParams
}

// GET /gov/parameters/tallying Query governance tally parameters
func getTallyingParam(t *testing.T, port string) gov.TallyParams {
	res, body := Request(t, port, "GET", "/gov/parameters/tallying", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var tallyParams gov.TallyParams
	err := cdc.UnmarshalJSON([]byte(body), &tallyParams)
	require.Nil(t, err)
	return tallyParams
}

// GET /gov/parameters/voting Query governance voting parameters
func getVotingParam(t *testing.T, port string) gov.VotingParams {
	res, body := Request(t, port, "GET", "/gov/parameters/voting", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var votingParams gov.VotingParams
	err := cdc.UnmarshalJSON([]byte(body), &votingParams)
	require.Nil(t, err)
	return votingParams
}

// ----------------------------------------------------------------------
// ICS 23 - Slashing
// ----------------------------------------------------------------------
// GET /slashing/validators/{validatorPubKey}/signing_info Get sign info of given validator
func getSigningInfo(t *testing.T, port string, validatorPubKey string) slashing.ValidatorSigningInfo {
	res, body := Request(t, port, "GET", fmt.Sprintf("/slashing/validators/%s/signing_info", validatorPubKey), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var signingInfo slashing.ValidatorSigningInfo
	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &signingInfo)
	require.Nil(t, err)

	return signingInfo
}

// ----------------------------------------------------------------------
// ICS 23 - SlashingList
// ----------------------------------------------------------------------
// GET /slashing/signing_infos Get sign info of all validators with pagination
func getSigningInfoList(t *testing.T, port string) []slashing.ValidatorSigningInfo {
	res, body := Request(t, port, "GET", "/slashing/signing_infos?page=1&limit=1", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var signingInfo []slashing.ValidatorSigningInfo
	err := cdc.UnmarshalJSON(extractResultFromResponse(t, []byte(body)), &signingInfo)
	require.Nil(t, err)

	return signingInfo
}

// TODO: Test this functionality, it is not currently in any of the tests
// POST /slashing/validators/{validatorAddr}/unjail Unjail a jailed validator
func doUnjail(
	t *testing.T, port, seed, name, pwd string, valAddr sdk.ValAddress, fees sdk.Coins,
) sdk.TxResponse {

	acc := getAccount(t, port, sdk.AccAddress(valAddr.Bytes()))
	from := acc.GetAddress().String()
	chainID := viper.GetString(client.FlagChainID)

	baseReq := rest.NewBaseReq(from, "", chainID, "", "", 1, 1, fees, nil, false)
	ur := slashingrest.UnjailReq{
		BaseReq: baseReq,
	}
	req, err := cdc.MarshalJSON(ur)
	require.NoError(t, err)

	resp, body := Request(t, port, "POST", fmt.Sprintf("/slashing/validators/%s/unjail", valAddr.String()), req)
	require.Equal(t, http.StatusOK, resp.StatusCode, body)

	resp, body = signAndBroadcastGenTx(t, port, name, pwd, body, acc, client.DefaultGasAdjustment, false)
	require.Equal(t, http.StatusOK, resp.StatusCode, body)

	var txResp sdk.TxResponse
	err = cdc.UnmarshalJSON([]byte(body), &txResp)
	require.NoError(t, err)

	return txResp
}

type unjailReq struct {
	BaseReq rest.BaseReq `json:"base_req"`
}

// ICS24 - fee distribution

// POST /distribution/delegators/{delgatorAddr}/rewards Withdraw delegator rewards
func doWithdrawDelegatorAllRewards(
	t *testing.T, port, seed, name, pwd string, delegatorAddr sdk.AccAddress, fees sdk.Coins,
) sdk.TxResponse {
	// get the account to get the sequence
	acc := getAccount(t, port, delegatorAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()
	chainID := viper.GetString(client.FlagChainID)
	from := acc.GetAddress().String()

	baseReq := rest.NewBaseReq(from, "", chainID, "", "", accnum, sequence, fees, nil, false)
	wr := struct {
		BaseReq rest.BaseReq `json:"base_req"`
	}{BaseReq: baseReq}

	req := cdc.MustMarshalJSON(wr)

	resp, body := Request(t, port, "POST", fmt.Sprintf("/distribution/delegators/%s/rewards", delegatorAddr), req)
	require.Equal(t, http.StatusOK, resp.StatusCode, body)

	resp, body = signAndBroadcastGenTx(t, port, name, pwd, body, acc, client.DefaultGasAdjustment, false)
	require.Equal(t, http.StatusOK, resp.StatusCode, body)

	var txResp sdk.TxResponse
	err := cdc.UnmarshalJSON([]byte(body), &txResp)
	require.NoError(t, err)

	return txResp
}

func mustParseDecCoins(dcstring string) sdk.DecCoins {
	dcoins, err := sdk.ParseDecCoins(dcstring)
	if err != nil {
		panic(err)
	}
	return dcoins
}
