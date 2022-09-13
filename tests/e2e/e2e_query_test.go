package e2e

import (
	"fmt"
	"io"
	"net/http"

	sdk "github.com/cosmos/cosmos-sdk/types"
	globalfee "github.com/cosmos/gaia/v8/x/globalfee/types"
)

func queryAccount(endpoint string) (amt sdk.DecCoins, err error) {
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
