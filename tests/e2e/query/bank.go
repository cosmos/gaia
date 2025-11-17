package query

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/cosmos/gaia/v26/tests/e2e/common"
)

// if coin is zero, return empty coin.
func SpecificBalance(endpoint, addr, denom string) (amt types.Coin, err error) {
	balances, err := AllBalances(endpoint, addr)
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

func AllBalances(endpoint, addr string) (types.Coins, error) {
	body, err := common.HTTPGet(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", endpoint, addr))
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	var balancesResp banktypes.QueryAllBalancesResponse
	if err := common.Cdc.UnmarshalJSON(body, &balancesResp); err != nil {
		return nil, err
	}

	return balancesResp.Balances, nil
}

func SupplyOf(endpoint, denom string) (types.Coin, error) {
	body, err := common.HTTPGet(fmt.Sprintf("%s/cosmos/bank/v1beta1/supply/by_denom?denom=%s", endpoint, denom))
	if err != nil {
		return types.Coin{}, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	var supplyOfResp banktypes.QuerySupplyOfResponse
	if err := common.Cdc.UnmarshalJSON(body, &supplyOfResp); err != nil {
		return types.Coin{}, err
	}

	return supplyOfResp.Amount, nil
}
