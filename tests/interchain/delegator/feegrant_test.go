package delegator_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/gaia/v23/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v23/tests/interchain/delegator"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/suite"
)

type FeegrantSuite struct {
	*delegator.Suite
}

func (s *FeegrantSuite) TestFeegrant() {
	var (
		granter       = s.DelegatorWallet
		grantee       = s.DelegatorWallet2
		fundsReceiver = s.DelegatorWallet3
	)

	tests := []struct {
		name   string
		revoke func(expireTime time.Time)
	}{
		{
			name: "revoke",
			revoke: func(_ time.Time) {
				_, err := s.Chain.GetNode().ExecTx(
					s.GetContext(),
					granter.FormattedAddress(),
					"feegrant", "revoke", granter.FormattedAddress(), grantee.FormattedAddress(),
				)
				s.Require().NoError(err)
			},
		},
		{
			name: "expire",
			revoke: func(expire time.Time) {
				<-time.After(time.Until(expire))
				err := testutil.WaitForBlocks(s.GetContext(), 1, s.Chain)
				s.Require().NoError(err)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		s.Run(tt.name, func() {
			expire := time.Now().Add(20 * chainsuite.CommitTimeout)
			_, err := s.Chain.GetNode().ExecTx(
				s.GetContext(),
				granter.FormattedAddress(),
				"feegrant", "grant", granter.FormattedAddress(), grantee.FormattedAddress(),
				"--expiration", expire.Format(time.RFC3339),
			)
			s.Require().NoError(err)

			granterBalanceBefore, err := s.Chain.GetBalance(s.GetContext(), granter.FormattedAddress(), chainsuite.Uatom)
			s.Require().NoError(err)
			granteeBalanceBefore, err := s.Chain.GetBalance(s.GetContext(), grantee.FormattedAddress(), chainsuite.Uatom)
			s.Require().NoError(err)

			_, err = s.Chain.GetNode().ExecTx(s.GetContext(), grantee.FormattedAddress(),
				"bank", "send", grantee.FormattedAddress(), fundsReceiver.FormattedAddress(), txAmountUatom(),
				"--fee-granter", granter.FormattedAddress(),
			)
			s.Require().NoError(err)

			granteeBalanceAfter, err := s.Chain.GetBalance(s.GetContext(), grantee.FormattedAddress(), chainsuite.Uatom)
			s.Require().NoError(err)
			granterBalanceAfter, err := s.Chain.GetBalance(s.GetContext(), granter.FormattedAddress(), chainsuite.Uatom)
			s.Require().NoError(err)

			s.Require().True(granterBalanceAfter.LT(granterBalanceBefore), "granterBalanceBefore: %s, granterBalanceAfter: %s", granterBalanceBefore, granterBalanceAfter)
			s.Require().True(granteeBalanceAfter.Equal(granteeBalanceBefore.Sub(sdkmath.NewInt(txAmount))), "granteeBalanceBefore: %s, granteeBalanceAfter: %s", granteeBalanceBefore, granteeBalanceAfter)

			tt.revoke(expire)

			_, err = s.Chain.GetNode().ExecTx(s.GetContext(), grantee.FormattedAddress(),
				"bank", "send", grantee.FormattedAddress(), fundsReceiver.FormattedAddress(), txAmountUatom(),
				"--fee-granter", granter.FormattedAddress(),
			)
			s.Require().Error(err)
		})
	}
}

func TestFeegrant(t *testing.T) {
	s := &FeegrantSuite{Suite: &delegator.Suite{Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
		UpgradeOnSetup: true,
	})}}
	suite.Run(t, s)
}
