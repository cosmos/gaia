package e2e

import (
	"fmt"
	"strconv"
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (s *IntegrationTestSuite) testStaking() {
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

	validatorA := s.chainA.validators[0]
	validatorB := s.chainA.validators[1]
	validatorAAddr, _ := validatorA.keyInfo.GetAddress()
	validatorBAddr, _ := validatorB.keyInfo.GetAddress()

	validatorAddressA := sdk.ValAddress(validatorAAddr).String()
	validatorAddressB := sdk.ValAddress(validatorBAddr).String()

	delegatorAddress, _ := s.chainA.genesisAccounts[2].keyInfo.GetAddress()

	fees := sdk.NewCoin(uatomDenom, math.NewInt(1))

	existingDelegation := math.LegacyZeroDec()
	res, err := queryDelegation(chainEndpoint, validatorAddressA, delegatorAddress.String())
	if err == nil {
		existingDelegation = res.GetDelegationResponse().GetDelegation().GetShares()
	}

	delegationAmount := math.NewInt(500000000)
	delegation := sdk.NewCoin(uatomDenom, delegationAmount) // 500 atom

	// Alice delegate uatom to Validator A
	s.execDelegate(s.chainA, 0, delegation.String(), validatorAddressA, delegatorAddress.String(), gaiaHomePath, fees.String())

	// Validate delegation successful
	s.Require().Eventually(
		func() bool {
			res, err := queryDelegation(chainEndpoint, validatorAddressA, delegatorAddress.String())
			amt := res.GetDelegationResponse().GetDelegation().GetShares()
			s.Require().NoError(err)

			return amt.Equal(existingDelegation.Add(math.LegacyNewDecFromInt(delegationAmount)))
		},
		20*time.Second,
		5*time.Second,
	)

	redelegationAmount := delegationAmount.Quo(math.NewInt(2))
	redelegation := sdk.NewCoin(uatomDenom, redelegationAmount) // 250 atom

	// Alice re-delegate half of her uatom delegation from Validator A to Validator B
	s.execRedelegate(s.chainA, 0, redelegation.String(), validatorAddressA, validatorAddressB, delegatorAddress.String(), gaiaHomePath, fees.String())

	// Validate re-delegation successful
	s.Require().Eventually(
		func() bool {
			res, err := queryDelegation(chainEndpoint, validatorAddressB, delegatorAddress.String())
			amt := res.GetDelegationResponse().GetDelegation().GetShares()
			s.Require().NoError(err)

			return amt.Equal(math.LegacyNewDecFromInt(redelegationAmount))
		},
		20*time.Second,
		5*time.Second,
	)

	var (
		currDelegation       sdk.Coin
		currDelegationAmount math.Int
	)

	// query alice's current delegation from validator A
	s.Require().Eventually(
		func() bool {
			res, err := queryDelegation(chainEndpoint, validatorAddressA, delegatorAddress.String())
			amt := res.GetDelegationResponse().GetDelegation().GetShares()
			s.Require().NoError(err)

			currDelegationAmount = amt.TruncateInt()
			currDelegation = sdk.NewCoin(uatomDenom, currDelegationAmount)

			return currDelegation.IsValid()
		},
		20*time.Second,
		5*time.Second,
	)

	// Alice unbonds all her uatom delegation from Validator A
	s.execUnbondDelegation(s.chainA, 0, currDelegation.String(), validatorAddressA, delegatorAddress.String(), gaiaHomePath, fees.String())

	var ubdDelegationEntry types.UnbondingDelegationEntry

	// validate unbonding delegations
	s.Require().Eventually(
		func() bool {
			res, err := queryUnbondingDelegation(chainEndpoint, validatorAddressA, delegatorAddress.String())
			s.Require().NoError(err)

			s.Require().Len(res.GetUnbond().Entries, 1)
			ubdDelegationEntry = res.GetUnbond().Entries[0]

			return ubdDelegationEntry.Balance.Equal(currDelegationAmount)
		},
		20*time.Second,
		5*time.Second,
	)

	// cancel the full amount of unbonding delegations from Validator A
	s.execCancelUnbondingDelegation(
		s.chainA,
		0,
		currDelegation.String(),
		validatorAddressA,
		strconv.Itoa(int(ubdDelegationEntry.CreationHeight)),
		delegatorAddress.String(),
		gaiaHomePath,
		fees.String(),
	)

	// validate that unbonding delegation was successfully canceled
	s.Require().Eventually(
		func() bool {
			resDel, err := queryDelegation(chainEndpoint, validatorAddressA, delegatorAddress.String())
			amt := resDel.GetDelegationResponse().GetDelegation().GetShares()
			s.Require().NoError(err)

			// expect that no unbonding delegations are found for validator A
			_, err = queryUnbondingDelegation(chainEndpoint, validatorAddressA, delegatorAddress.String())
			s.Require().Error(err)

			// expect to get the delegation back
			return amt.Equal(math.LegacyNewDecFromInt(currDelegationAmount))
		},
		20*time.Second,
		5*time.Second,
	)
}
