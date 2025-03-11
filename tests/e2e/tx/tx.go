package tx

import (
	"context"
	"cosmossdk.io/x/feegrant"
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	types7 "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	types3 "github.com/cosmos/cosmos-sdk/x/bank/types"
	types4 "github.com/cosmos/cosmos-sdk/x/distribution/types"
	types5 "github.com/cosmos/cosmos-sdk/x/gov/types"
	types2 "github.com/cosmos/cosmos-sdk/x/slashing/types"
	types6 "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gaia/v23/tests/e2e/common"
	"github.com/cosmos/gaia/v23/tests/e2e/query"
	"github.com/cosmos/gogoproto/proto"
	types8 "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/types"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v2"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Helper struct {
	Suite        *suite.Suite
	CommonHelper *common.Helper
}

func (h *Helper) StoreWasm(ctx context.Context, c *common.Chain, valIdx int, sender, wasmPath string) string {
	storeCmd := []string{
		common.GaiadBinary,
		common.TxCommand,
		"wasm",
		"store",
		wasmPath,
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, common.StandardFees.String()),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.Id),
		"--gas=5000000",
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}

	h.Suite.T().Logf("%s storing wasm on host chain %s", sender, h.CommonHelper.Resources.ChainB.Id)
	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, storeCmd, valIdx, h.CommonHelper.DefaultExecValidation(c, valIdx))
	h.Suite.T().Log("successfully sent store wasm tx")
	h.CommonHelper.TestCounters.ContractsCounter++
	return strconv.Itoa(h.CommonHelper.TestCounters.ContractsCounter)
}

func (h *Helper) InstantiateWasm(ctx context.Context, c *common.Chain, valIdx int, sender, codeID,
	msg, label string,
) string {
	storeCmd := []string{
		common.GaiadBinary,
		common.TxCommand,
		"wasm",
		"instantiate",
		codeID,
		msg,
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, common.StandardFees.String()),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.Id),
		fmt.Sprintf("--label=%s", label),
		"--no-admin",
		"--gas=500000",
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}

	h.Suite.T().Logf("%s instantiating wasm on host chain %s", sender, h.CommonHelper.Resources.ChainB.Id)
	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, storeCmd, valIdx, h.CommonHelper.DefaultExecValidation(c, valIdx))
	h.Suite.T().Log("successfully sent instantiate wasm tx")
	chainEndpoint := fmt.Sprintf("http://%s", h.CommonHelper.Resources.ValResources[c.Id][0].GetHostPort("1317/tcp"))
	address, err := query.QueryWasmContractAddress(chainEndpoint, sender, h.CommonHelper.TestCounters.ContractsCounterPerSender[sender])
	h.Suite.Require().NoError(err)
	h.CommonHelper.TestCounters.ContractsCounterPerSender[sender]++
	return address
}

func (h *Helper) Instantiate2Wasm(ctx context.Context, c *common.Chain, valIdx int, sender, codeID,
	msg, salt, label string,
) string {
	storeCmd := []string{
		common.GaiadBinary,
		common.TxCommand,
		"wasm",
		"instantiate2",
		codeID,
		msg,
		salt,
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, common.StandardFees.String()),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.Id),
		fmt.Sprintf("--label=%s", label),
		"--no-admin",
		"--gas=250000",
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}

	h.Suite.T().Logf("%s instantiating wasm on host chain %s", sender, h.CommonHelper.Resources.ChainB.Id)

	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, storeCmd, valIdx, h.CommonHelper.DefaultExecValidation(c, valIdx))
	h.Suite.T().Log("successfully sent instantiate2 wasm tx")
	chainEndpoint := fmt.Sprintf("http://%s", h.CommonHelper.Resources.ValResources[c.Id][0].GetHostPort("1317/tcp"))
	address, err := query.QueryWasmContractAddress(chainEndpoint, sender, h.CommonHelper.TestCounters.ContractsCounterPerSender[sender])
	h.Suite.Require().NoError(err)
	h.CommonHelper.TestCounters.ContractsCounterPerSender[sender]++
	return address
}

func (h *Helper) ExecuteWasm(ctx context.Context, c *common.Chain, valIdx int, sender, addr, msg string) {
	execCmd := []string{
		common.GaiadBinary,
		common.TxCommand,
		"wasm",
		"execute",
		addr,
		msg,
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, common.StandardFees.String()),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.Id),
		"--gas=250000",
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}
	h.Suite.T().Logf("%s executing wasm on host chain %s", sender, h.CommonHelper.Resources.ChainB.Id)
	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, execCmd, valIdx, h.CommonHelper.DefaultExecValidation(c, valIdx))
	h.Suite.T().Log("successfully sent execute wasm tx")
}

func (h *Helper) QueryBuildAddress(ctx context.Context, c *common.Chain, valIdx int, codeHash, creatorAddress, saltHexEncoded string,
) (res string) {
	cmd := []string{
		common.GaiadBinary,
		common.QueryCommand,
		"wasm",
		"build-address",
		codeHash,
		creatorAddress,
		saltHexEncoded,
	}

	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, cmd, valIdx, func(stdOut []byte, stdErr []byte) bool {
		h.Suite.Require().NoError(yaml.Unmarshal(stdOut, &res))
		return true
	})
	return res
}

func (h *Helper) ExecDecode(
	c *common.Chain,
	txPath string,
	opt ...common.FlagOption,
) string {
	opts := common.ApplyOptions(c.Id, opt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("%s - Executing gaiad decoding with %v", c.Id, txPath)
	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		"decode",
		txPath,
	}
	for flag, value := range opts {
		gaiaCommand = append(gaiaCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	var decoded string
	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, 0, func(stdOut []byte, stdErr []byte) bool {
		if stdErr != nil {
			return false
		}
		decoded = strings.TrimSuffix(string(stdOut), "\n")
		return true
	})
	h.Suite.T().Logf("successfully decode %v", txPath)
	return decoded
}

func (h *Helper) execVestingTx( //nolint:unused

	c *common.Chain,
	method string,
	args []string,
	opt ...common.FlagOption,
) {
	opts := common.ApplyOptions(c.Id, opt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("%s - Executing gaiad %s with %v", c.Id, method, args)
	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types.ModuleName,
		method,
		"-y",
	}
	gaiaCommand = append(gaiaCommand, args...)

	for flag, value := range opts {
		gaiaCommand = append(gaiaCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, 0, h.CommonHelper.DefaultExecValidation(c, 0))
	h.Suite.T().Logf("successfully %s with %v", method, args)
}

func (h *Helper) ExecCreatePeriodicVestingAccount( //nolint:unused

	c *common.Chain,
	address,
	jsonPath string,
	opt ...common.FlagOption,
) {
	h.Suite.T().Logf("Executing gaiad create periodic vesting account %s", c.Id)
	h.execVestingTx(c, "create-periodic-vesting-account", []string{address, jsonPath}, opt...)
	h.Suite.T().Logf("successfully created periodic vesting account %s with %s", address, jsonPath)
}

func (h *Helper) ExecUnjail(
	c *common.Chain,
	opt ...common.FlagOption,
) {
	opts := common.ApplyOptions(c.Id, opt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("Executing gaiad slashing unjail %s with options: %v", c.Id, opt)
	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types2.ModuleName,
		"unjail",
		"-y",
	}

	for flag, value := range opts {
		gaiaCommand = append(gaiaCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, 0, h.CommonHelper.DefaultExecValidation(c, 0))
	h.Suite.T().Logf("successfully unjail with options %v", opt)
}

func (h *Helper) ExecFeeGrant(c *common.Chain, valIdx int, granter, grantee, spendLimit string, opt ...common.FlagOption) {
	opt = append(opt, common.WithKeyValue(common.FlagFrom, granter))
	opt = append(opt, common.WithKeyValue(common.FlagSpendLimit, spendLimit))
	opts := common.ApplyOptions(c.Id, opt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("granting %s fee from %s on chain %s", grantee, granter, c.Id)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		feegrant.ModuleName,
		"grant",
		granter,
		grantee,
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.Id),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "300000"), // default 200000 isn't enough
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}
	for flag, value := range opts {
		gaiaCommand = append(gaiaCommand, fmt.Sprintf("--%s=%s", flag, value))
	}
	h.Suite.T().Logf("running feegrant on chain: %s - Tx %v", c.Id, gaiaCommand)

	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.CommonHelper.DefaultExecValidation(c, valIdx))
}

func (h *Helper) ExecFeeGrantRevoke(c *common.Chain, valIdx int, granter, grantee string, opt ...common.FlagOption) {
	opt = append(opt, common.WithKeyValue(common.FlagFrom, granter))
	opts := common.ApplyOptions(c.Id, opt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("revoking %s fee grant from %s on chain %s", grantee, granter, c.Id)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		feegrant.ModuleName,
		"revoke",
		granter,
		grantee,
		"-y",
	}
	for flag, value := range opts {
		gaiaCommand = append(gaiaCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.CommonHelper.DefaultExecValidation(c, valIdx))
}

func (h *Helper) ExecBankSend(
	c *common.Chain,
	valIdx int,
	from,
	to,
	amt,
	fees string,
	expectErr bool,
	opt ...common.FlagOption,
) {
	// TODO remove the hardcode opt after refactor, all methods should accept custom flags
	opt = append(opt, common.WithKeyValue(common.FlagFees, fees))
	opt = append(opt, common.WithKeyValue(common.FlagFrom, from))
	opts := common.ApplyOptions(c.Id, opt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("sending %s tokens from %s to %s on chain %s", amt, from, to, c.Id)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types3.ModuleName,
		"send",
		from,
		to,
		amt,
		"-y",
	}
	for flag, value := range opts {
		gaiaCommand = append(gaiaCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.expectErrExecValidation(c, valIdx, expectErr))
}

func (h *Helper) ExecBankMultiSend(
	c *common.Chain,
	valIdx int,
	from string,
	to []string,
	amt string,
	fees string,
	expectErr bool,
	opt ...common.FlagOption,
) {
	// TODO remove the hardcode opt after refactor, all methods should accept custom flags
	opt = append(opt, common.WithKeyValue(common.FlagFees, fees))
	opt = append(opt, common.WithKeyValue(common.FlagFrom, from))
	opts := common.ApplyOptions(c.Id, opt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("sending %s tokens from %s to %s on chain %s", amt, from, to, c.Id)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types3.ModuleName,
		"multi-send",
		from,
	}

	gaiaCommand = append(gaiaCommand, to...)
	gaiaCommand = append(gaiaCommand, amt, "-y")

	for flag, value := range opts {
		gaiaCommand = append(gaiaCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.expectErrExecValidation(c, valIdx, expectErr))
}

func (h *Helper) ExecDistributionFundCommunityPool(c *common.Chain, valIdx int, from, amt, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("Executing gaiad tx distribution fund-community-pool on chain %s", c.Id)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types4.ModuleName,
		"fund-community-pool",
		amt,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.Id),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fees),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.CommonHelper.DefaultExecValidation(c, valIdx))
	h.Suite.T().Logf("Successfully funded community pool")
}

func (h *Helper) RunGovExec(c *common.Chain, valIdx int, submitterAddr, govCommand string, proposalFlags []string, fees string, validationFunc func([]byte, []byte) bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	validateResponse := h.CommonHelper.DefaultExecValidation(c, valIdx)
	if validationFunc != nil {
		validateResponse = validationFunc
	}

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types5.ModuleName,
		govCommand,
	}

	generalFlags := []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, submitterAddr),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "50000000"), // default 200000 isn't enough
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.Id),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	gaiaCommand = common.ConcatFlags(gaiaCommand, proposalFlags, generalFlags)
	h.Suite.T().Logf("Executing gaiad tx gov %s on chain %s", govCommand, c.Id)
	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, validateResponse)
	h.Suite.T().Logf("Successfully executed %s", govCommand)
}

func (h *Helper) ExecDelegate(c *common.Chain, valIdx int, amount, valOperAddress, delegatorAddr, home, delegateFees string) { //nolint:unparam

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("Executing gaiad tx staking delegate %s", c.Id)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types6.ModuleName,
		"delegate",
		valOperAddress,
		amount,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.Id),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, delegateFees),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "250000"), // default 200_000 is not enough
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.CommonHelper.DefaultExecValidation(c, valIdx))
	h.Suite.T().Logf("%s successfully delegated %s to %s", delegatorAddr, amount, valOperAddress)
}

func (h *Helper) ExecUnbondDelegation(c *common.Chain, valIdx int, amount, valOperAddress, delegatorAddr, home, delegateFees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("Executing gaiad tx staking unbond %s", c.Id)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types6.ModuleName,
		"unbond",
		valOperAddress,
		amount,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.Id),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, delegateFees),
		"--gas=300000", // default 200_000 is not enough; gas fees are higher when unbonding is done after LSM operations
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.CommonHelper.DefaultExecValidation(c, valIdx))
	h.Suite.T().Logf("%s successfully undelegated %s to %s", delegatorAddr, amount, valOperAddress)
}

func (h *Helper) ExecCancelUnbondingDelegation(c *common.Chain, valIdx int, amount, valOperAddress, creationHeight, delegatorAddr, home, delegateFees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("Executing gaiad tx staking cancel-unbond %s", c.Id)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types6.ModuleName,
		"cancel-unbond",
		valOperAddress,
		amount,
		creationHeight,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.Id),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, delegateFees),
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.CommonHelper.DefaultExecValidation(c, valIdx))
	h.Suite.T().Logf("%s successfully canceled unbonding %s to %s", delegatorAddr, amount, valOperAddress)
}

func (h *Helper) ExecRedelegate(c *common.Chain, valIdx int, amount, originalValOperAddress,
	newValOperAddress, delegatorAddr, home, delegateFees string,
) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("Executing gaiad tx staking redelegate %s", c.Id)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types6.ModuleName,
		"redelegate",
		originalValOperAddress,
		newValOperAddress,
		amount,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.Id),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "350000"), // default 200000 isn't enough
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, delegateFees),
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.CommonHelper.DefaultExecValidation(c, valIdx))
	h.Suite.T().Logf("%s successfully redelegated %s from %s to %s", delegatorAddr, amount, originalValOperAddress, newValOperAddress)
}

func (h *Helper) ExecSetWithdrawAddress(
	c *common.Chain,
	valIdx int,
	fees,
	delegatorAddress,
	newWithdrawalAddress,
	homePath string,
) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("Setting distribution withdrawal address on chain %s for %s to %s", c.Id, delegatorAddress, newWithdrawalAddress)
	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types4.ModuleName,
		"set-withdraw-addr",
		newWithdrawalAddress,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddress),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.Id),
		fmt.Sprintf("--%s=%s", flags.FlagHome, homePath),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.CommonHelper.DefaultExecValidation(c, valIdx))
	h.Suite.T().Logf("Successfully set new distribution withdrawal address for %s to %s", delegatorAddress, newWithdrawalAddress)
}

func (h *Helper) ExecWithdrawReward(
	c *common.Chain,
	valIdx int,
	delegatorAddress,
	validatorAddress,
	homePath string,
) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("Withdrawing distribution rewards on chain %s for delegator %s from %s validator", c.Id, delegatorAddress, validatorAddress)
	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types4.ModuleName,
		"withdraw-rewards",
		validatorAddress,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddress),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, "300uatom"),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "auto"),
		fmt.Sprintf("--%s=%s", flags.FlagGasAdjustment, "1.5"),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.Id),
		fmt.Sprintf("--%s=%s", flags.FlagHome, homePath),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.CommonHelper.DefaultExecValidation(c, valIdx))
	h.Suite.T().Logf("Successfully withdrew distribution rewards for delegator %s from validator %s", delegatorAddress, validatorAddress)
}

func (h *Helper) expectErrExecValidation(chain *common.Chain, valIdx int, expectErr bool) func([]byte, []byte) bool {
	return func(stdOut []byte, stdErr []byte) bool {
		var txResp types7.TxResponse
		gotErr := common.Cdc.UnmarshalJSON(stdOut, &txResp) != nil
		if gotErr {
			h.Suite.Require().True(expectErr)
		}

		endpoint := fmt.Sprintf("http://%s", h.CommonHelper.Resources.ValResources[chain.Id][valIdx].GetHostPort("1317/tcp"))
		// wait for the tx to be committed on chain
		h.Suite.Require().Eventuallyf(
			func() bool {
				gotErr := common.QueryGaiaTx(endpoint, txResp.TxHash) != nil
				return gotErr == expectErr
			},
			time.Minute,
			5*time.Second,
			"stdOut: %s, stdErr: %s",
			string(stdOut), string(stdErr),
		)
		return true
	}
}

func (h *Helper) ExecEncode(
	c *common.Chain,
	txPath string,
	opt ...common.FlagOption,
) string {
	opts := common.ApplyOptions(c.Id, opt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("%s - Executing gaiad encoding with %v", c.Id, txPath)
	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		"encode",
		txPath,
	}
	for flag, value := range opts {
		gaiaCommand = append(gaiaCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	var encoded string
	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, 0, func(stdOut []byte, stdErr []byte) bool {
		if stdErr != nil {
			return false
		}
		encoded = strings.TrimSuffix(string(stdOut), "\n")
		return true
	})
	h.Suite.T().Logf("successfully encode with %v", txPath)
	return encoded
}

func (h *Helper) ExpectTxSubmitError(expectErrString string) func([]byte, []byte) bool {
	return func(stdOut []byte, stdErr []byte) bool {
		var txResp types7.TxResponse
		if err := common.Cdc.UnmarshalJSON(stdOut, &txResp); err != nil {
			return false
		}
		if strings.Contains(txResp.RawLog, expectErrString) {
			return true
		}
		return false
	}
}

func (h *Helper) ExecuteValidatorBond(c *common.Chain, valIdx int, valOperAddress, delegatorAddr, home, delegateFees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("Executing gaiad tx staking validator-bond %s", c.Id)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types6.ModuleName,
		"validator-bond",
		valOperAddress,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.Id),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, delegateFees),
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.CommonHelper.DefaultExecValidation(c, valIdx))
	h.Suite.T().Logf("%s successfully executed validator bond tx to %s", delegatorAddr, valOperAddress)
}

func (h *Helper) ExecuteTokenizeShares(c *common.Chain, valIdx int, amount, valOperAddress, delegatorAddr, home, delegateFees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("Executing gaiad tx staking tokenize-share %s", c.Id)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types6.ModuleName,
		"tokenize-share",
		valOperAddress,
		amount,
		delegatorAddr,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.Id),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, delegateFees),
		fmt.Sprintf("--%s=%d", flags.FlagGas, 1000000),
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.CommonHelper.DefaultExecValidation(c, valIdx))
	h.Suite.T().Logf("%s successfully executed tokenize share tx from %s", delegatorAddr, valOperAddress)
}

func (h *Helper) ExecuteRedeemShares(c *common.Chain, valIdx int, amount, delegatorAddr, home, delegateFees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("Executing gaiad tx staking redeem-tokens %s", c.Id)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types6.ModuleName,
		"redeem-tokens",
		amount,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.Id),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, delegateFees),
		fmt.Sprintf("--%s=%d", flags.FlagGas, 1000000),
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.CommonHelper.DefaultExecValidation(c, valIdx))
	h.Suite.T().Logf("%s successfully executed redeem share tx for %s", delegatorAddr, amount)
}

func (h *Helper) ExecuteTransferTokenizeShareRecord(c *common.Chain, valIdx int, recordID, owner, newOwner, home, txFees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	h.Suite.T().Logf("Executing gaiad tx staking transfer-tokenize-share-record %s", c.Id)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		types6.ModuleName,
		"transfer-tokenize-share-record",
		recordID,
		newOwner,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, owner),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.Id),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, txFees),
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, h.CommonHelper.DefaultExecValidation(c, valIdx))
	h.Suite.T().Logf("%s successfully executed transfer tokenize share record for %s", owner, recordID)
}

// SignTxFileOnline signs a transaction file using the gaiacli tx sign command
// the from flag is used to specify the keyring account to sign the transaction
// the from account must be registered in the keyring and exist on chain (have a balance or be a genesis account)
func (h *Helper) SignTxFileOnline(chain *common.Chain, valIdx int, from string, txFilePath string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		"sign",
		filepath.Join(common.GaiaHomePath, txFilePath),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, chain.Id),
		fmt.Sprintf("--%s=%s", flags.FlagHome, common.GaiaHomePath),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	var output []byte
	var erroutput []byte
	captureOutput := func(stdout []byte, stderr []byte) bool {
		output = stdout
		erroutput = stderr
		return true
	}

	h.CommonHelper.ExecuteGaiaTxCommand(ctx, chain, gaiaCommand, valIdx, captureOutput)
	if len(erroutput) > 0 {
		return nil, fmt.Errorf("failed to sign tx: %s", string(erroutput))
	}
	return output, nil
}

// BroadcastTxFile broadcasts a signed transaction file using the gaiacli tx broadcast command
// the from flag is used to specify the keyring account to sign the transaction
// the from account must be registered in the keyring and exist on chain (have a balance or be a genesis account)
func (h *Helper) BroadcastTxFile(chain *common.Chain, valIdx int, from string, txFilePath string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	broadcastTxCmd := []string{
		common.GaiadBinary,
		common.TxCommand,
		"broadcast",
		filepath.Join(common.GaiaHomePath, txFilePath),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, chain.Id),
		fmt.Sprintf("--%s=%s", flags.FlagHome, common.GaiaHomePath),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	var output []byte
	var erroutput []byte
	captureOutput := func(stdout []byte, stderr []byte) bool {
		output = stdout
		erroutput = stderr
		return true
	}

	h.CommonHelper.ExecuteGaiaTxCommand(ctx, chain, broadcastTxCmd, valIdx, captureOutput)
	if len(erroutput) > 0 {
		return nil, fmt.Errorf("failed to sign tx: %s", string(erroutput))
	}
	return output, nil
}

//nolint:unparam
func (h *Helper) SendIBC(c *common.Chain, valIdx int, sender, recipient, token, fees, note, channel string, absoluteTimeout *int64, expErr bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	ibcCmd := []string{
		common.GaiadBinary,
		common.TxCommand,
		"ibc-transfer",
		"transfer",
		"transfer",
		channel,
		recipient,
		token,
	}

	if absoluteTimeout != nil {
		ibcCmd = append(ibcCmd, "--absolute-timeouts")
		ibcCmd = append(ibcCmd, fmt.Sprintf("--packet-timeout-timestamp=%d", *absoluteTimeout))
	}

	ibcCmd = append(ibcCmd, []string{
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.Id),
		// fmt.Sprintf("--%s=%s", flags.FlagNote, note),
		fmt.Sprintf("--memo=%s", note),
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}...)

	h.Suite.T().Logf("sending %s from %s (%s) to %s (%s) with memo %s", token, h.CommonHelper.Resources.ChainA.Id, sender, h.CommonHelper.Resources.ChainB.Id, recipient, note)
	if expErr {
		h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, ibcCmd, valIdx, h.expectErrExecValidation(c, valIdx, true))
		h.Suite.T().Log("unsuccessfully sent IBC tokens")
	} else {
		h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, ibcCmd, valIdx, h.CommonHelper.DefaultExecValidation(c, valIdx))
		h.Suite.T().Log("successfully sent IBC tokens")
	}
}

func (h *Helper) RegisterICAAccount(c *common.Chain, valIdx int, sender, connectionID, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	version := string(types8.ModuleCdc.MustMarshalJSON(&types8.Metadata{
		Version:                types8.Version,
		ControllerConnectionId: connectionID,
		HostConnectionId:       connectionID,
		Encoding:               types8.EncodingProtobuf,
		TxType:                 types8.TxTypeSDKMultiMsg,
	}))

	icaCmd := []string{
		common.GaiadBinary,
		common.TxCommand,
		"interchain-accounts",
		"controller",
		"register",
		connectionID,
		fmt.Sprintf("--version=%s", version),
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.Id),
		"--gas=250000", // default 200_000 is not enough; gas fees increased after adding IBC fee middleware
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}
	h.Suite.T().Logf("%s registering ICA account on host chain %s", sender, h.CommonHelper.Resources.ChainB.Id)
	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, icaCmd, valIdx, h.CommonHelper.DefaultExecValidation(c, valIdx))
	h.Suite.T().Log("successfully sent register ICA account tx")
}

func (h *Helper) SendICATransaction(c *common.Chain, valIdx int, sender, connectionID, packetMsgPath, fees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	icaCmd := []string{
		common.GaiadBinary,
		common.TxCommand,
		"interchain-accounts",
		"controller",
		"send-tx",
		connectionID,
		packetMsgPath,
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.Id),
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}
	h.Suite.T().Logf("%s sending ICA transaction to the host chain %s", sender, h.CommonHelper.Resources.ChainB.Id)
	h.CommonHelper.ExecuteGaiaTxCommand(ctx, c, icaCmd, valIdx, h.CommonHelper.DefaultExecValidation(c, valIdx))
	h.Suite.T().Log("successfully sent ICA transaction")
}

func (h *Helper) BuildICASendTransactionFile(cdc codec.Codec, msgs []proto.Message, outputBaseDir string) {
	data, err := types8.SerializeCosmosTx(cdc, msgs, types8.EncodingProtobuf)
	h.Suite.Require().NoError(err)

	sendICATransaction := types8.InterchainAccountPacketData{
		Type: types8.EXECUTE_TX,
		Data: data,
	}

	sendICATransactionBody, err := json.MarshalIndent(sendICATransaction, "", " ")
	h.Suite.Require().NoError(err)

	outputPath := filepath.Join(outputBaseDir, "config", common.ICASendTransactionFileName)
	err = common.WriteFile(outputPath, sendICATransactionBody)
	h.Suite.Require().NoError(err)
}
