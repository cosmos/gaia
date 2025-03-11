package query

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/gaia/v23/tests/e2e/common"
	"io"
	"net/http"
)

func queryAccount(endpoint, address string) (acc types.AccountI, err error) {
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

func QueryDelayedVestingAccount(endpoint, address string) (vestingtypes.DelayedVestingAccount, error) {
	baseAcc, err := queryAccount(endpoint, address)
	if err != nil {
		return vestingtypes.DelayedVestingAccount{}, err
	}
	acc, ok := baseAcc.(*vestingtypes.DelayedVestingAccount)
	if !ok {
		return vestingtypes.DelayedVestingAccount{},
			fmt.Errorf("cannot cast %v to DelayedVestingAccount", baseAcc)
	}
	return *acc, nil
}

func QueryContinuousVestingAccount(endpoint, address string) (vestingtypes.ContinuousVestingAccount, error) {
	baseAcc, err := queryAccount(endpoint, address)
	if err != nil {
		return vestingtypes.ContinuousVestingAccount{}, err
	}
	acc, ok := baseAcc.(*vestingtypes.ContinuousVestingAccount)
	if !ok {
		return vestingtypes.ContinuousVestingAccount{},
			fmt.Errorf("cannot cast %v to ContinuousVestingAccount", baseAcc)
	}
	return *acc, nil
}

func QueryPeriodicVestingAccount(endpoint, address string) (vestingtypes.PeriodicVestingAccount, error) { //nolint:unused // this is called during e2e tests
	baseAcc, err := queryAccount(endpoint, address)
	if err != nil {
		return vestingtypes.PeriodicVestingAccount{}, err
	}
	acc, ok := baseAcc.(*vestingtypes.PeriodicVestingAccount)
	if !ok {
		return vestingtypes.PeriodicVestingAccount{},
			fmt.Errorf("cannot cast %v to PeriodicVestingAccount", baseAcc)
	}
	return *acc, nil
}
