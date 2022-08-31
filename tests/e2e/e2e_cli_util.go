package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	staketypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	globalfee "github.com/cosmos/gaia/v8/x/globalfee/types"
	icamauth "github.com/cosmos/gaia/v8/x/icamauth/types"
)

func queryGaiaTx(endpoint, txHash string) error {
	resp, err := http.Get(fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/%s", endpoint, txHash))
	if err != nil {
		return fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("tx query returned non-200 status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	txResp := result["tx_response"].(map[string]interface{})
	if v := txResp["code"]; v.(float64) != 0 {
		return fmt.Errorf("tx %s failed with status code %v", txHash, v)
	}

	return nil
}

// if coin is zero, return empty coin.
func getSpecificBalance(endpoint, addr, denom string) (amt sdk.Coin, err error) {
	balances, err := queryGaiaAllBalances(endpoint, addr)
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

func queryGaiaAllBalances(endpoint, addr string) (sdk.Coins, error) {
	resp, err := http.Get(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", endpoint, addr))
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	defer resp.Body.Close()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var balancesResp banktypes.QueryAllBalancesResponse
	if err := cdc.UnmarshalJSON(bz, &balancesResp); err != nil {
		return nil, err
	}

	return balancesResp.Balances, nil
}

func queryGaiaDenomBalance(endpoint, addr, denom string) (sdk.Coin, error) {
	var zeroCoin sdk.Coin

	path := fmt.Sprintf(
		"%s/cosmos/bank/v1beta1/balances/%s/by_denom?denom=%s",
		endpoint, addr, denom,
	)

	resp, err := http.Get(path) //nolint:gosec // this is used as a part of the e2e suite.
	if err != nil {
		return zeroCoin, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	defer resp.Body.Close()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return zeroCoin, err
	}

	var balanceResp banktypes.QueryBalanceResponse
	if err := cdc.UnmarshalJSON(bz, &balanceResp); err != nil {
		return zeroCoin, err
	}

	return *balanceResp.Balance, nil
}

func queryGovProposal(endpoint string, proposalID int) (govv1beta1.QueryProposalResponse, error) {
	var govProposalResp govv1beta1.QueryProposalResponse

	path := fmt.Sprintf("%s/cosmos/gov/v1beta1/proposals/%d", endpoint, proposalID)

	resp, err := http.Get(path) //nolint:gosec // this is only used during tests
	if err != nil {
		return govProposalResp, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return govProposalResp, err
	}

	if err := cdc.UnmarshalJSON(body, &govProposalResp); err != nil {
		return govProposalResp, err
	}

	return govProposalResp, nil
}

func queryGlobalFees(endpoint string) (amt sdk.DecCoins, err error) {
	resp, err := http.Get(fmt.Sprintf("%s/gaia/globalfee/v1beta1/minimum_gas_prices", endpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var fees globalfee.QueryMinimumGasPricesResponse
	if err := cdc.UnmarshalJSON(bz, &fees); err != nil {
		return sdk.DecCoins{}, err
	}

	return fees.MinimumGasPrices, nil
}

func queryICAaddr(endpoint, owner, connectionID string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/gaia/icamauth/v1beta1/interchain_account/owner/%s/connection/%s", endpoint, owner, connectionID))
	if err != nil {
		return "", fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("tx query returned non-200 status: %d", resp.StatusCode)
	}

	icaAddrResp := icamauth.QueryInterchainAccountFromAddressResponse{}
	if err = cdc.UnmarshalJSON(bz, &icaAddrResp); err != nil {
		return "", err
	}

	return icaAddrResp.GetInterchainAccountAddress(), nil
}

func queryDelegation(endpoint string, validatorAddr string, delegatorAddr string) (staketypes.QueryDelegationResponse, error) {
	var delegationRes staketypes.QueryDelegationResponse

	resp, err := http.Get(fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations/%s", endpoint, validatorAddr, delegatorAddr)) //nolint:gosec // this is only used during tests
	if err != nil {
		return delegationRes, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return delegationRes, err
	}

	if err := cdc.UnmarshalJSON(body, &delegationRes); err != nil {
		return delegationRes, err
	}

	return delegationRes, nil
}
