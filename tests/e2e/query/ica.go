package query

import (
	"fmt"

	"github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/types"

	"github.com/cosmos/gaia/v25/tests/e2e/common"
)

func ICAAccountAddress(endpoint, owner, connectionID string) (string, error) {
	body, err := common.HTTPGet(fmt.Sprintf("%s/ibc/apps/interchain_accounts/controller/v1/owners/%s/connections/%s", endpoint, owner, connectionID))
	if err != nil {
		return "", fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	var icaAccountResp types.QueryInterchainAccountResponse
	if err := common.Cdc.UnmarshalJSON(body, &icaAccountResp); err != nil {
		return "", err
	}

	return icaAccountResp.Address, nil
}
