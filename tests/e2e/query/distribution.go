package query

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/cosmos/gaia/v23/tests/e2e/common"
)

func DelegatorWithdrawalAddress(endpoint string, delegatorAddr string) (types.QueryDelegatorWithdrawAddressResponse, error) {
	var res types.QueryDelegatorWithdrawAddressResponse

	body, err := common.HTTPGet(fmt.Sprintf("%s/cosmos/distribution/v1beta1/delegators/%s/withdraw_address", endpoint, delegatorAddr))
	if err != nil {
		return res, err
	}

	if err = common.Cdc.UnmarshalJSON(body, &res); err != nil {
		return res, err
	}
	return res, nil
}
