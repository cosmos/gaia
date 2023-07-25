package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestDefaultParams(t *testing.T) {
	p := DefaultParams()
	require.EqualValues(t, p.MinimumGasPrices, sdk.DecCoins{})
	require.EqualValues(t, p.BypassMinFeeMsgTypes, DefaultBypassMinFeeMsgTypes)
	require.EqualValues(t, p.MaxTotalBypassMinFeeMsgGasUsage, DefaultmaxTotalBypassMinFeeMsgGasUsage)
}

func Test_validateMinGasPrices(t *testing.T) {
	tests := map[string]struct {
		coins     interface{}
		expectErr bool
	}{
		"DefaultParams, pass": {
			DefaultParams().MinimumGasPrices,
			false,
		},
		"DecCoins conversion fails, fail": {
			sdk.Coins{sdk.NewCoin("photon", sdk.OneInt())},
			true,
		},
		"coins amounts are zero, pass": {
			sdk.DecCoins{
				sdk.NewDecCoin("atom", sdk.ZeroInt()),
				sdk.NewDecCoin("photon", sdk.ZeroInt()),
			},
			false,
		},
		"duplicate coins denoms, fail": {
			sdk.DecCoins{
				sdk.NewDecCoin("photon", sdk.OneInt()),
				sdk.NewDecCoin("photon", sdk.OneInt()),
			},
			true,
		},
		"coins are not sorted by denom alphabetically, fail": {
			sdk.DecCoins{
				sdk.NewDecCoin("photon", sdk.OneInt()),
				sdk.NewDecCoin("atom", sdk.OneInt()),
			},
			true,
		},
		"negative amount, fail": {
			sdk.DecCoins{
				sdk.DecCoin{Denom: "photon", Amount: sdk.OneDec().Neg()},
			},
			true,
		},
		"invalid denom, fail": {
			sdk.DecCoins{
				sdk.DecCoin{Denom: "photon!", Amount: sdk.OneDec().Neg()},
			},
			true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := validateMinimumGasPrices(test.coins)
			if test.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func Test_validateBypassMinFeeMsgTypes(t *testing.T) {
	tests := map[string]struct {
		msgTypes  interface{}
		expectErr bool
	}{
		"DefaultParams, pass": {
			DefaultParams().BypassMinFeeMsgTypes,
			false,
		},
		"wrong msg type should make conversion fail, fail": {
			[]int{0, 1, 2, 3},
			true,
		},
		"empty msg types, pass": {
			[]string{},
			false,
		},
		"empty msg type, fail": {
			[]string{""},
			true,
		},
		"invalid msg type name, fail": {
			[]string{"ibc.core.channel.v1.MsgRecvPacket"},
			true,
		},
		"mixed valid and invalid msgs, fail": {
			[]string{
				"/ibc.core.channel.v1.MsgRecvPacket",
				"",
			},
			true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := validateBypassMinFeeMsgTypes(test.msgTypes)
			if test.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func Test_validateMaxTotalBypassMinFeeMsgGasUsage(t *testing.T) {
	tests := map[string]struct {
		msgTypes  interface{}
		expectErr bool
	}{
		"DefaultParams, pass": {
			DefaultParams().MaxTotalBypassMinFeeMsgGasUsage,
			false,
		},
		"zero value, pass": {
			uint64(0),
			false,
		},
		"negative value, fail": {
			-1,
			true,
		},
		"invalid type, fail": {
			"5",
			true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := validateMaxTotalBypassMinFeeMsgGasUsage(test.msgTypes)
			if test.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
