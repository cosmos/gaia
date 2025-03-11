package query

import (
	tendermintv1beta1 "cosmossdk.io/api/cosmos/base/tendermint/v1beta1"
	"fmt"
	"github.com/cosmos/gaia/v23/tests/e2e/common"
	ratelimittypes "github.com/cosmos/ibc-apps/modules/rate-limiting/v10/types"
	wasmclienttypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/types"
	providertypes "github.com/cosmos/interchain-security/v7/x/ccv/provider/types"
	"io"
	"net/http"
	"strings"

	evidencetypes "cosmossdk.io/x/evidence/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

// if coin is zero, return empty coin.
func GetSpecificBalance(endpoint, addr, denom string) (amt sdk.Coin, err error) {
	balances, err := QueryGaiaAllBalances(endpoint, addr)
	if err != nil {
		return amt, err
	}
	for _, c := range balances {
		if strings.Contains(c.Denom, denom) {
			amt = c
			break
		}
	}
	return amt, nil
}

func QueryGaiaAllBalances(endpoint, addr string) (sdk.Coins, error) {
	body, err := common.HttpGet(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", endpoint, addr))
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	var balancesResp banktypes.QueryAllBalancesResponse
	if err := common.Cdc.UnmarshalJSON(body, &balancesResp); err != nil {
		return nil, err
	}

	return balancesResp.Balances, nil
}

func QuerySupplyOf(endpoint, denom string) (sdk.Coin, error) {
	body, err := common.HttpGet(fmt.Sprintf("%s/cosmos/bank/v1beta1/supply/by_denom?denom=%s", endpoint, denom))
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	var supplyOfResp banktypes.QuerySupplyOfResponse
	if err := common.Cdc.UnmarshalJSON(body, &supplyOfResp); err != nil {
		return sdk.Coin{}, err
	}

	return supplyOfResp.Amount, nil
}

func QueryStakingParams(endpoint string) (stakingtypes.QueryParamsResponse, error) {
	body, err := common.HttpGet(fmt.Sprintf("%s/cosmos/staking/v1beta1/params", endpoint))
	if err != nil {
		return stakingtypes.QueryParamsResponse{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	var params stakingtypes.QueryParamsResponse
	if err := common.Cdc.UnmarshalJSON(body, &params); err != nil {
		return stakingtypes.QueryParamsResponse{}, err
	}

	return params, nil
}

func QueryDelegation(endpoint string, validatorAddr string, delegatorAddr string) (stakingtypes.QueryDelegationResponse, error) {
	var res stakingtypes.QueryDelegationResponse

	body, err := common.HttpGet(fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations/%s", endpoint, validatorAddr, delegatorAddr))
	if err != nil {
		return res, err
	}

	if err = common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

func QueryUnbondingDelegation(endpoint string, validatorAddr string, delegatorAddr string) (stakingtypes.QueryUnbondingDelegationResponse, error) {
	var res stakingtypes.QueryUnbondingDelegationResponse
	body, err := common.HttpGet(fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations/%s/unbonding_delegation", endpoint, validatorAddr, delegatorAddr))
	if err != nil {
		return res, err
	}

	if err = common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

func QueryDelegatorWithdrawalAddress(endpoint string, delegatorAddr string) (disttypes.QueryDelegatorWithdrawAddressResponse, error) {
	var res disttypes.QueryDelegatorWithdrawAddressResponse

	body, err := common.HttpGet(fmt.Sprintf("%s/cosmos/distribution/v1beta1/delegators/%s/withdraw_address", endpoint, delegatorAddr))
	if err != nil {
		return res, err
	}

	if err = common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

func QueryGovProposal(endpoint string, proposalID int) (govtypesv1beta1.QueryProposalResponse, error) {
	var govProposalResp govtypesv1beta1.QueryProposalResponse

	path := fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals/%d", endpoint, proposalID)

	body, err := common.HttpGet(path)
	if err != nil {
		return govProposalResp, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	if err := common.Cdc.UnmarshalJSON(body, &govProposalResp); err != nil {
		return govProposalResp, err
	}

	return govProposalResp, nil
}

func QueryGovProposalV1(endpoint string, proposalID int) (govtypesv1.QueryProposalResponse, error) {
	var govProposalResp govtypesv1.QueryProposalResponse

	path := fmt.Sprintf("%s/cosmos/gov/v1/proposals/%d", endpoint, proposalID)

	body, err := common.HttpGet(path)
	if err != nil {
		return govProposalResp, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	if err := common.Cdc.UnmarshalJSON(body, &govProposalResp); err != nil {
		return govProposalResp, err
	}

	return govProposalResp, nil
}

func queryAccount(endpoint, address string) (acc sdk.AccountI, err error) {
	var res authtypes.QueryAccountResponse
	resp, err := http.Get(fmt.Sprintf("%s/cosmos/auth/v1beta1/accounts/%s", endpoint, address))
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := common.Cdc.UnmarshalJSON(bz, &res); err != nil {
		return nil, err
	}
	return acc, common.Cdc.UnpackAny(res.Account, &acc)
}

func QueryDelayedVestingAccount(endpoint, address string) (authvesting.DelayedVestingAccount, error) {
	baseAcc, err := queryAccount(endpoint, address)
	if err != nil {
		return authvesting.DelayedVestingAccount{}, err
	}
	acc, ok := baseAcc.(*authvesting.DelayedVestingAccount)
	if !ok {
		return authvesting.DelayedVestingAccount{},
			fmt.Errorf("cannot cast %v to DelayedVestingAccount", baseAcc)
	}
	return *acc, nil
}

func QueryContinuousVestingAccount(endpoint, address string) (authvesting.ContinuousVestingAccount, error) {
	baseAcc, err := queryAccount(endpoint, address)
	if err != nil {
		return authvesting.ContinuousVestingAccount{}, err
	}
	acc, ok := baseAcc.(*authvesting.ContinuousVestingAccount)
	if !ok {
		return authvesting.ContinuousVestingAccount{},
			fmt.Errorf("cannot cast %v to ContinuousVestingAccount", baseAcc)
	}
	return *acc, nil
}

func QueryPeriodicVestingAccount(endpoint, address string) (authvesting.PeriodicVestingAccount, error) { //nolint:unused // this is called during e2e tests
	baseAcc, err := queryAccount(endpoint, address)
	if err != nil {
		return authvesting.PeriodicVestingAccount{}, err
	}
	acc, ok := baseAcc.(*authvesting.PeriodicVestingAccount)
	if !ok {
		return authvesting.PeriodicVestingAccount{},
			fmt.Errorf("cannot cast %v to PeriodicVestingAccount", baseAcc)
	}
	return *acc, nil
}

func QueryValidator(endpoint, address string) (stakingtypes.Validator, error) {
	var res stakingtypes.QueryValidatorResponse

	body, err := common.HttpGet(fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s", endpoint, address))
	if err != nil {
		return stakingtypes.Validator{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	if err := common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return stakingtypes.Validator{}, err
	}
	return res.Validator, nil
}

func QueryValidators(endpoint string) (stakingtypes.Validators, error) {
	var res stakingtypes.QueryValidatorsResponse
	body, err := common.HttpGet(fmt.Sprintf("%s/cosmos/staking/v1beta1/validators", endpoint))
	if err != nil {
		return stakingtypes.Validators{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	if err := common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return stakingtypes.Validators{}, err
	}

	return stakingtypes.Validators{Validators: res.Validators}, nil
}

func queryEvidence(endpoint, hash string) (evidencetypes.QueryEvidenceResponse, error) { //nolint:unused // this is called during e2e tests
	var res evidencetypes.QueryEvidenceResponse
	body, err := common.HttpGet(fmt.Sprintf("%s/cosmos/evidence/v1beta1/evidence/%s", endpoint, hash))
	if err != nil {
		return res, err
	}

	if err = common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

func QueryAllEvidence(endpoint string) (evidencetypes.QueryAllEvidenceResponse, error) {
	var res evidencetypes.QueryAllEvidenceResponse
	body, err := common.HttpGet(fmt.Sprintf("%s/cosmos/evidence/v1beta1/evidence", endpoint))
	if err != nil {
		return res, err
	}

	if err = common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}

func QueryTokenizeShareRecordByID(endpoint string, recordID int) (stakingtypes.TokenizeShareRecord, error) {
	var res stakingtypes.QueryTokenizeShareRecordByIdResponse

	body, err := common.HttpGet(fmt.Sprintf("%s/cosmos/staking/v1beta1/tokenize_share_record_by_id/%d", endpoint, recordID))
	if err != nil {
		return stakingtypes.TokenizeShareRecord{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	if err := common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return stakingtypes.TokenizeShareRecord{}, err
	}
	return res.Record, nil
}

func QueryAllRateLimits(endpoint string) ([]ratelimittypes.RateLimit, error) {
	var res ratelimittypes.QueryAllRateLimitsResponse

	body, err := common.HttpGet(fmt.Sprintf("%s/Stride-Labs/ibc-rate-limiting/ratelimit/ratelimits", endpoint))
	if err != nil {
		return []ratelimittypes.RateLimit{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	if err := common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return []ratelimittypes.RateLimit{}, err
	}
	return res.RateLimits, nil
}

func QueryRateLimit(endpoint, channelID, denom string) (ratelimittypes.QueryRateLimitResponse, error) {
	var res ratelimittypes.QueryRateLimitResponse

	body, err := common.HttpGet(fmt.Sprintf("%s/Stride-Labs/ibc-rate-limiting/ratelimit/ratelimit/%s/by_denom?denom=%s", endpoint, channelID, denom))
	if err != nil {
		return ratelimittypes.QueryRateLimitResponse{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	if err := common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return ratelimittypes.QueryRateLimitResponse{}, err
	}
	return res, nil
}

func QueryRateLimitsByChainID(endpoint, channelID string) ([]ratelimittypes.RateLimit, error) {
	var res ratelimittypes.QueryRateLimitsByChainIdResponse

	body, err := common.HttpGet(fmt.Sprintf("%s/Stride-Labs/ibc-rate-limiting/ratelimit/ratelimits/%s", endpoint, channelID))
	if err != nil {
		return []ratelimittypes.RateLimit{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	if err := common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return []ratelimittypes.RateLimit{}, err
	}
	return res.RateLimits, nil
}

func QueryICAAccountAddress(endpoint, owner, connectionID string) (string, error) {
	body, err := common.HttpGet(fmt.Sprintf("%s/ibc/apps/interchain_accounts/controller/v1/owners/%s/connections/%s", endpoint, owner, connectionID))
	if err != nil {
		return "", fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	var icaAccountResp icacontrollertypes.QueryInterchainAccountResponse
	if err := common.Cdc.UnmarshalJSON(body, &icaAccountResp); err != nil {
		return "", err
	}

	return icaAccountResp.Address, nil
}

func QueryBlocksPerEpoch(endpoint string) (int64, error) {
	body, err := common.HttpGet(fmt.Sprintf("%s/interchain_security/ccv/provider/params", endpoint))
	if err != nil {
		return 0, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	var response providertypes.QueryParamsResponse
	if err = common.Cdc.UnmarshalJSON(body, &response); err != nil {
		return 0, err
	}

	return response.Params.BlocksPerEpoch, nil
}

func QueryWasmContractAddress(endpoint, creator string, idx uint64) (string, error) {
	body, err := common.HttpGet(fmt.Sprintf("%s/cosmwasm/wasm/v1/contracts/creator/%s", endpoint, creator))
	if err != nil {
		return "", fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	var response wasmtypes.QueryContractsByCreatorResponse
	if err = common.Cdc.UnmarshalJSON(body, &response); err != nil {
		return "", err
	}

	return response.ContractAddresses[idx], nil
}

func QueryWasmSmartContractState(endpoint, address, msg string) ([]byte, error) {
	body, err := common.HttpGet(fmt.Sprintf("%s/cosmwasm/wasm/v1/contract/%s/smart/%s", endpoint, address, msg))
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	var response wasmtypes.QuerySmartContractStateResponse
	if err = common.Cdc.UnmarshalJSON(body, &response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

func QueryIbcWasmChecksums(endpoint string) ([]string, error) {
	body, err := common.HttpGet(fmt.Sprintf("%s/ibc/lightclients/wasm/v1/checksums", endpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	var response wasmclienttypes.QueryChecksumsResponse
	if err = common.Cdc.UnmarshalJSON(body, &response); err != nil {
		return nil, err
	}

	return response.Checksums, nil
}

func GetLatestBlockHeight(endpoint string) (int, error) {
	body, err := common.HttpGet(fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/blocks/latest", endpoint))
	if err != nil {
		return 0, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	var response tendermintv1beta1.GetLatestBlockResponse
	if err = common.Cdc.UnmarshalJSON(body, &response); err != nil {
		return 0, err
	}
	return int(response.GetBlock().GetLastCommit().Height), nil
}

func ExecQueryEvidence(endpoint, hash string) (evidencetypes.Equivocation, error) {
	_, err := common.HttpGet(fmt.Sprintf("%s/cosmos/evidence/v1beta1/evidence/%s", endpoint, hash))
	if err != nil {
		return evidencetypes.Equivocation{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	//var response evidencetypes.QueryEvidenceResponse
	//if err = common.Cdc.UnmarshalJSON(body, &response); err != nil {
	//	return evidencetypes.Equivocation{}, err
	//}

	var evidence evidencetypes.Equivocation
	//err = common.Cdc.UnpackAny(response.Evidence, &evidence)
	//if err != nil {
	//	return evidencetypes.Equivocation{}, err
	//}

	return evidence, nil
}
