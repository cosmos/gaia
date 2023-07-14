package e2e

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *IntegrationTestSuite) testLSM() {
	// TODO: Set parameters (global liquid staking cap, validator liquid staking cap, validator bond factor)
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

	validatorA := s.chainA.validators[0]
	// validatorB := s.chainA.validators[1]
	validatorAAddr := validatorA.keyInfo.GetAddress()
	// validatorBAddr := validatorB.keyInfo.GetAddress()

	validatorAddressA := sdk.ValAddress(validatorAAddr).String()
	// validatorAddressB := sdk.ValAddress(validatorBAddr).String()

	delegatorAddress := s.chainA.genesisAccounts[2].keyInfo.GetAddress().String()

	fees := sdk.NewCoin(uatomDenom, sdk.NewInt(1))

	// Validator bond
	s.executeValidatorBond(s.chainA, 0, validatorAddressA, validatorAAddr.String(), gaiaHomePath, fees.String())

	// Validate validator bond successful
	s.Require().Eventually(
		func() bool {
			res, err := queryDelegation(chainEndpoint, validatorAddressA, validatorAAddr.String())
			isValidatorBond := res.GetDelegationResponse().GetDelegation().ValidatorBond
			s.Require().NoError(err)

			return isValidatorBond == true
		},
		20*time.Second,
		5*time.Second,
	)

	delegationAmount := sdk.NewInt(500000000)
	delegation := sdk.NewCoin(uatomDenom, delegationAmount) // 500 atom

	// Alice delegate uatom to Validator A
	s.executeDelegate(s.chainA, 0, delegation.String(), validatorAddressA, delegatorAddress, gaiaHomePath, fees.String())

	// Validate delegation successful
	s.Require().Eventually(
		func() bool {
			res, err := queryDelegation(chainEndpoint, validatorAddressA, delegatorAddress)
			amt := res.GetDelegationResponse().GetDelegation().GetShares()
			s.Require().NoError(err)

			return amt.Equal(sdk.NewDecFromInt(delegationAmount))
		},
		20*time.Second,
		5*time.Second,
	)

	// Tokenize shares
	tokenizeAmount := sdk.NewInt(200000000)
	tokenize := sdk.NewCoin(uatomDenom, tokenizeAmount) // 200 atom
	s.executeTokenizeShares(s.chainA, 0, tokenize.String(), validatorAddressA, delegatorAddress, gaiaHomePath, fees.String())

	// Validate delegation reduced
	s.Require().Eventually(
		func() bool {
			res, err := queryDelegation(chainEndpoint, validatorAddressA, delegatorAddress)
			amt := res.GetDelegationResponse().GetDelegation().GetShares()
			s.Require().NoError(err)

			return amt.Equal(sdk.NewDecFromInt(delegationAmount.Sub(tokenizeAmount)))
		},
		20*time.Second,
		5*time.Second,
	)

	// Validate balance increased
	shareDenom := fmt.Sprintf("%s/%s", strings.ToLower(validatorAddressA), strconv.Itoa(1))
	s.Require().Eventually(
		func() bool {
			res, err := queryGaiaAllBalances(chainEndpoint, delegatorAddress)
			s.Require().NoError(err)
			shareAmount := res.AmountOf(shareDenom)
			return shareAmount.Equal(tokenizeAmount)
		},
		20*time.Second,
		5*time.Second,
	)

	// TODO: Bank send LSM token
	// TODO: TransferTokenizeShareRecord (transfer reward ownership)
	// TODO: IBC transfer LSM token
	// TODO: Redeem tokens for shares
}
