package antetest

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/gaia/v9/x/globalfee/ante"
)

type feeUtilsTestSuite struct {
	suite.Suite
}

func TestFeeUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(feeUtilsTestSuite))
}

func (s *feeUtilsTestSuite) TestContainZeroCoins() {
	zeroCoin1 := sdk.NewCoin("photon", sdk.ZeroInt())
	zeroCoin2 := sdk.NewCoin("stake", sdk.ZeroInt())
	coin1 := sdk.NewCoin("photon", sdk.NewInt(1))
	coin2 := sdk.NewCoin("stake", sdk.NewInt(2))
	coin3 := sdk.NewCoin("quark", sdk.NewInt(3))
	// coins must be valid !!!
	coinsEmpty := sdk.Coins{}
	coinsNonEmpty := sdk.Coins{coin1, coin2}
	coinsCointainZero := sdk.Coins{coin1, zeroCoin2}
	coinsCointainTwoZero := sdk.Coins{zeroCoin1, zeroCoin2, coin3}
	coinsAllZero := sdk.Coins{zeroCoin1, zeroCoin2}

	tests := []struct {
		c  sdk.Coins
		ok bool
	}{
		{
			coinsEmpty,
			true,
		},
		{
			coinsNonEmpty,
			false,
		},
		{
			coinsCointainZero,
			true,
		},
		{
			coinsCointainTwoZero,
			true,
		},
		{
			coinsAllZero,
			true,
		},
	}

	for _, test := range tests {
		ok := ante.ContainZeroCoins(test.c)
		s.Require().Equal(test.ok, ok)
	}
}

func (s *feeUtilsTestSuite) TestCombinedFeeRequirement() {
	zeroCoin1 := sdk.NewCoin("photon", sdk.ZeroInt())
	zeroCoin2 := sdk.NewCoin("stake", sdk.ZeroInt())
	zeroCoin3 := sdk.NewCoin("quark", sdk.ZeroInt())
	coin1 := sdk.NewCoin("photon", sdk.NewInt(1))
	coin2 := sdk.NewCoin("stake", sdk.NewInt(2))
	coin1High := sdk.NewCoin("photon", sdk.NewInt(10))
	coin2High := sdk.NewCoin("stake", sdk.NewInt(20))
	coinNewDenom1 := sdk.NewCoin("Newphoton", sdk.NewInt(1))
	coinNewDenom2 := sdk.NewCoin("Newstake", sdk.NewInt(1))
	// coins must be valid !!! and sorted!!!
	coinsEmpty := sdk.Coins{}
	coinsNonEmpty := sdk.Coins{coin1, coin2}.Sort()
	coinsNonEmptyHigh := sdk.Coins{coin1High, coin2High}.Sort()
	coinsNonEmptyOneHigh := sdk.Coins{coin1High, coin2}.Sort()
	coinsNewDenom := sdk.Coins{coinNewDenom1, coinNewDenom2}.Sort()
	coinsNewOldDenom := sdk.Coins{coin1, coinNewDenom1}.Sort()
	coinsNewOldDenomHigh := sdk.Coins{coin1High, coinNewDenom1}.Sort()
	coinsCointainZero := sdk.Coins{coin1, zeroCoin2}.Sort()
	coinsCointainZeroNewDenom := sdk.Coins{coin1, zeroCoin3}.Sort()
	coinsAllZero := sdk.Coins{zeroCoin1, zeroCoin2}.Sort()
	tests := map[string]struct {
		cGlobal  sdk.Coins
		c        sdk.Coins
		combined sdk.Coins
	}{
		"global fee empty, min fee empty, combined fee empty": {
			cGlobal:  coinsEmpty,
			c:        coinsEmpty,
			combined: coinsEmpty,
		},
		"global fee empty, min fee nonempty, combined fee empty": {
			cGlobal:  coinsEmpty,
			c:        coinsNonEmpty,
			combined: coinsEmpty,
		},
		"global fee nonempty, min fee empty, combined fee = global fee": {
			cGlobal:  coinsNonEmpty,
			c:        coinsNonEmpty,
			combined: coinsNonEmpty,
		},
		"global fee and min fee have overlapping denom, min fees amounts are all higher": {
			cGlobal:  coinsNonEmpty,
			c:        coinsNonEmptyHigh,
			combined: coinsNonEmptyHigh,
		},
		"global fee and min fee have overlapping denom, one of min fees amounts is higher": {
			cGlobal:  coinsNonEmpty,
			c:        coinsNonEmptyOneHigh,
			combined: coinsNonEmptyOneHigh,
		},
		"global fee and min fee have no overlapping denom, combined fee = global fee": {
			cGlobal:  coinsNonEmpty,
			c:        coinsNewDenom,
			combined: coinsNonEmpty,
		},
		"global fees and min fees have partial overlapping denom, min fee amount <= global fee amount, combined fees = global fees": {
			cGlobal:  coinsNonEmpty,
			c:        coinsNewOldDenom,
			combined: coinsNonEmpty,
		},
		"global fees and min fees have partial overlapping denom, one min fee amount > global fee amount, combined fee = overlapping highest": {
			cGlobal:  coinsNonEmpty,
			c:        coinsNewOldDenomHigh,
			combined: sdk.Coins{coin1High, coin2},
		},
		"global fees have zero fees, min fees have overlapping non-zero fees, combined fees = overlapping highest": {
			cGlobal:  coinsCointainZero,
			c:        coinsNonEmpty,
			combined: sdk.Coins{coin1, coin2},
		},
		"global fees have zero fees, min fees have overlapping zero fees": {
			cGlobal:  coinsCointainZero,
			c:        coinsCointainZero,
			combined: coinsCointainZero,
		},
		"global fees have zero fees, min fees have non-overlapping zero fees": {
			cGlobal:  coinsCointainZero,
			c:        coinsCointainZeroNewDenom,
			combined: coinsCointainZero,
		},
		"global fees are all zero fees, min fees have overlapping zero fees": {
			cGlobal:  coinsAllZero,
			c:        coinsAllZero,
			combined: coinsAllZero,
		},
		"global fees are all zero fees, min fees have overlapping non-zero fees, combined fee = overlapping highest": {
			cGlobal:  coinsAllZero,
			c:        coinsCointainZeroNewDenom,
			combined: sdk.Coins{coin1, zeroCoin2},
		},
		"global fees are all zero fees, fees have one overlapping non-zero fee": {
			cGlobal:  coinsAllZero,
			c:        coinsCointainZero,
			combined: coinsCointainZero,
		},
	}

	for name, test := range tests {
		s.Run(name, func() {
			allFees := ante.CombinedFeeRequirement(test.cGlobal, test.c)
			s.Require().Equal(test.combined, allFees)
		})
	}
}

func (s *feeUtilsTestSuite) TestDenomsSubsetOfIncludingZero() {
	emptyCoins := sdk.Coins{}

	zeroCoin1 := sdk.NewCoin("photon", sdk.ZeroInt())
	zeroCoin2 := sdk.NewCoin("stake", sdk.ZeroInt())
	zeroCoin3 := sdk.NewCoin("quark", sdk.ZeroInt())

	coin1 := sdk.NewCoin("photon", sdk.NewInt(1))
	coin2 := sdk.NewCoin("stake", sdk.NewInt(2))
	coin3 := sdk.NewCoin("quark", sdk.NewInt(3))

	coinNewDenom1 := sdk.NewCoin("newphoton", sdk.NewInt(1))
	coinNewDenom2 := sdk.NewCoin("newstake", sdk.NewInt(2))
	coinNewDenom3 := sdk.NewCoin("newquark", sdk.NewInt(3))
	coinNewDenom1Zero := sdk.NewCoin("newphoton", sdk.ZeroInt())
	// coins must be valid !!! and sorted!!!
	coinsAllZero := sdk.Coins{zeroCoin1, zeroCoin2, zeroCoin3}.Sort()
	coinsAllZeroShort := sdk.Coins{zeroCoin1, zeroCoin2}.Sort()
	coinsContainZero := sdk.Coins{zeroCoin1, zeroCoin2, coin3}.Sort()
	coinsContainZeroNewDenoms := sdk.Coins{zeroCoin1, zeroCoin2, coinNewDenom1Zero}.Sort()
	coins := sdk.Coins{coin1, coin2, coin3}.Sort()
	coinsShort := sdk.Coins{coin1, coin2}.Sort()
	coinsAllNewDenom := sdk.Coins{coinNewDenom1, coinNewDenom2, coinNewDenom3}.Sort()
	coinsOldNewDenom := sdk.Coins{coin1, coin2, coinNewDenom1}.Sort()

	tests := map[string]struct {
		superset sdk.Coins
		set      sdk.Coins
		subset   bool
	}{
		"empty coins is a DenomsSubsetOf empty coins": {
			superset: emptyCoins,
			set:      emptyCoins,
			subset:   true,
		},
		"nonempty coins is not a DenomsSubsetOf empty coins": {
			superset: emptyCoins,
			set:      coins,
			subset:   false,
		},
		"empty coins is not a DenomsSubsetOf nonempty, nonzero coins": {
			superset: emptyCoins,
			set:      coins,
			subset:   false,
		},
		"empty coins is a DenomsSubsetOf coins of all zeros": {
			superset: coinsAllZero,
			set:      emptyCoins,
			subset:   true,
		},
		"empty coins is a DenomsSubsetOf coinsContainZero": {
			superset: coinsContainZero,
			set:      emptyCoins,
			subset:   true,
		},
		"two sets no denoms overlapping, DenomsSubsetOf = false": {
			superset: coins,
			set:      coinsAllNewDenom,
			subset:   false,
		},
		"two sets have partially overlapping denoms, DenomsSubsetOf = false": {
			superset: coins,
			set:      coinsOldNewDenom,
			subset:   false,
		},
		"two sets are nonzero, set's denoms are all in superset, DenomsSubsetOf = true": {
			superset: coins,
			set:      coinsShort,
			subset:   true,
		},
		"supersets are zero coins, set's denoms are all in superset, DenomsSubsetOf = true": {
			superset: coinsAllZero,
			set:      coinsShort,
			subset:   true,
		},
		"supersets contains zero coins, set's denoms are all in superset, DenomsSubsetOf = true": {
			superset: coinsContainZero,
			set:      coinsShort,
			subset:   true,
		},
		"supersets contains zero coins, set's denoms contains zero coins, denoms are overlapping DenomsSubsetOf = true": {
			superset: coinsContainZero,
			set:      coinsContainZero,
			subset:   true,
		},
		"supersets contains zero coins, set's denoms contains zero coins, denoms are not overlapping DenomsSubsetOf = false": {
			superset: coinsContainZero,
			set:      coinsContainZeroNewDenoms,
			subset:   false,
		},
		"two sets of all zero coins, have the same denoms, DenomsSubsetOf = true": {
			superset: coinsAllZero,
			set:      coinsAllZeroShort,
			subset:   true,
		},
	}

	for name, test := range tests {
		s.Run(name, func() {
			subset := ante.DenomsSubsetOfIncludingZero(test.set, test.superset)
			s.Require().Equal(test.subset, subset)
		})
	}
}

func (s *feeUtilsTestSuite) TestIsAnyGTEIncludingZero() {
	emptyCoins := sdk.Coins{}

	zeroCoin1 := sdk.NewCoin("photon", sdk.ZeroInt())
	zeroCoin2 := sdk.NewCoin("stake", sdk.ZeroInt())
	zeroCoin3 := sdk.NewCoin("quark", sdk.ZeroInt())

	coin1 := sdk.NewCoin("photon", sdk.NewInt(10))
	coin1Low := sdk.NewCoin("photon", sdk.NewInt(1))
	coin1High := sdk.NewCoin("photon", sdk.NewInt(100))
	coin2 := sdk.NewCoin("stake", sdk.NewInt(20))
	coin2Low := sdk.NewCoin("stake", sdk.NewInt(2))
	coin2High := sdk.NewCoin("stake", sdk.NewInt(200))
	coin3 := sdk.NewCoin("quark", sdk.NewInt(30))

	coinNewDenom1 := sdk.NewCoin("newphoton", sdk.NewInt(10))
	coinNewDenom2 := sdk.NewCoin("newstake", sdk.NewInt(20))
	coinNewDenom3 := sdk.NewCoin("newquark", sdk.NewInt(30))
	zeroCoinNewDenom1 := sdk.NewCoin("newphoton", sdk.NewInt(10))
	zeroCoinNewDenom2 := sdk.NewCoin("newstake", sdk.NewInt(20))
	zeroCoinNewDenom3 := sdk.NewCoin("newquark", sdk.NewInt(30))
	// coins must be valid !!! and sorted!!!
	coinsAllZero := sdk.Coins{zeroCoin1, zeroCoin2, zeroCoin3}.Sort()
	coinsAllNewDenomAllZero := sdk.Coins{zeroCoinNewDenom1, zeroCoinNewDenom2, zeroCoinNewDenom3}.Sort()
	coinsAllZeroShort := sdk.Coins{zeroCoin1, zeroCoin2}.Sort()
	coinsContainZero := sdk.Coins{zeroCoin1, zeroCoin2, coin3}.Sort()

	coins := sdk.Coins{coin1, coin2, coin3}.Sort()
	coinsHighHigh := sdk.Coins{coin1High, coin2High}
	coinsHighLow := sdk.Coins{coin1High, coin2Low}.Sort()
	coinsLowLow := sdk.Coins{coin1Low, coin2Low}.Sort()
	// coinsShort := sdk.Coins{coin1, coin2}.Sort()
	coinsAllNewDenom := sdk.Coins{coinNewDenom1, coinNewDenom2, coinNewDenom3}.Sort()
	coinsOldNewDenom := sdk.Coins{coin1, coinNewDenom1, coinNewDenom2}.Sort()
	coinsOldLowNewDenom := sdk.Coins{coin1Low, coinNewDenom1, coinNewDenom2}.Sort()
	tests := map[string]struct {
		c1  sdk.Coins
		c2  sdk.Coins
		gte bool // greater or equal
	}{
		"zero coins are GTE zero coins": {
			c1:  coinsAllZero,
			c2:  coinsAllZero,
			gte: true,
		},
		"zero coins(short) are GTE zero coins": {
			c1:  coinsAllZero,
			c2:  coinsAllZeroShort,
			gte: true,
		},
		"zero coins are GTE zero coins(short)": {
			c1:  coinsAllZeroShort,
			c2:  coinsAllZero,
			gte: true,
		},
		"c2 is all zero coins, with different denoms from c1 which are all zero coins too": {
			c1:  coinsAllZero,
			c2:  coinsAllNewDenomAllZero,
			gte: false,
		},
		"empty coins are GTE empty coins": {
			c1:  emptyCoins,
			c2:  emptyCoins,
			gte: true,
		},
		"empty coins are GTE zero coins": {
			c1:  coinsAllZero,
			c2:  emptyCoins,
			gte: true,
		},
		"empty coins are GTE coins that contain zero denom": {
			c1:  coinsContainZero,
			c2:  emptyCoins,
			gte: true,
		},
		"zero coins are not GTE empty coins": {
			c1:  emptyCoins,
			c2:  coinsAllZero,
			gte: false,
		},
		"empty coins are not GTE nonzero coins": {
			c1:  coins,
			c2:  emptyCoins,
			gte: false,
		},
		// special case, not the opposite result of the above case
		"nonzero coins are not GTE empty coins": {
			c1:  emptyCoins,
			c2:  coins,
			gte: false,
		},
		"nonzero coins are GTE zero coins, has overlapping denom": {
			c1:  coinsAllZero,
			c2:  coins,
			gte: true,
		},
		"nonzero coins are GTE coins contain zero coins, zero coin is overlapping denom": {
			c1:  coinsContainZero,
			c2:  coins,
			gte: true,
		},
		"one denom amount higher, one denom amount lower": {
			c1:  coins,
			c2:  coinsHighLow,
			gte: true,
		},
		"all coins amounts are lower, denom overlapping": {
			c1:  coins,
			c2:  coinsLowLow,
			gte: false,
		},
		"all coins amounts are higher, denom overlapping": {
			c1:  coins,
			c2:  coinsHighHigh,
			gte: true,
		},
		"denoms are all not overlapping": {
			c1:  coins,
			c2:  coinsAllNewDenom,
			gte: false,
		},
		"denom not all overlapping, one overlapping denom is gte": {
			c1:  coins,
			c2:  coinsOldNewDenom,
			gte: true,
		},
		"denom not all overlapping, the only one overlapping denom is smaller": {
			c1:  coins,
			c2:  coinsOldLowNewDenom,
			gte: false,
		},
	}

	for name, test := range tests {
		s.Run(name, func() {
			gte := ante.IsAnyGTEIncludingZero(test.c2, test.c1)
			s.Require().Equal(test.gte, gte)
		})
	}
}
