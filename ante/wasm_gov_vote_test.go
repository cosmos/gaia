package ante_test

import (
	"testing"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/gaia/v26/ante"
	"github.com/cosmos/gaia/v26/app/helpers"
)

// mockMessenger is a mock implementation of wasmkeeper.Messenger for testing
type mockMessenger struct {
	dispatchMsgCalled bool
	returnEvents      []sdk.Event
	returnData        [][]byte
	returnMsgResp     [][]*codectypes.Any
	returnErr         error
}

func (m *mockMessenger) DispatchMsg(
	_ sdk.Context,
	_ sdk.AccAddress,
	_ string,
	_ wasmvmtypes.CosmosMsg,
) ([]sdk.Event, [][]byte, [][]*codectypes.Any, error) {
	m.dispatchMsgCalled = true
	return m.returnEvents, m.returnData, m.returnMsgResp, m.returnErr
}

// TestWasmGovVoteDecorator tests the GovVoteMessageHandler
func TestWasmGovVoteDecorator(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})
	stakingKeeper := gaiaApp.StakingKeeper

	// Get validator
	validators, err := stakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, validators)
	valAddr, err := stakingKeeper.ValidatorAddressCodec().StringToBytes(validators[0].GetOperator())
	require.NoError(t, err)
	valAddr = sdk.ValAddress(valAddr)

	// Create a second validator for testing multiple delegations
	pk := ed25519.GenPrivKeyFromSecret([]byte{uint8(42)}).PubKey()
	validator2, err := stakingtypes.NewValidator(
		sdk.ValAddress(pk.Address()).String(),
		pk,
		stakingtypes.Description{},
	)
	require.NoError(t, err)
	valAddr2, err := stakingKeeper.ValidatorAddressCodec().StringToBytes(validator2.GetOperator())
	require.NoError(t, err)
	valAddr2 = sdk.ValAddress(valAddr2)
	validator2.Status = stakingtypes.Bonded
	err = stakingKeeper.SetValidator(ctx, validator2)
	require.NoError(t, err)
	err = stakingKeeper.SetValidatorByConsAddr(ctx, validator2)
	require.NoError(t, err)
	err = stakingKeeper.SetNewValidatorByPowerIndex(ctx, validator2)
	require.NoError(t, err)
	err = stakingKeeper.Hooks().AfterValidatorCreated(ctx, valAddr2)
	require.NoError(t, err)

	// Create two mock "contract" addresses (simulated as regular accounts for staking)
	contractWithStake := sdk.AccAddress(ed25519.GenPrivKeyFromSecret([]byte{uint8(100)}).PubKey().Address())
	contractWithoutStake := sdk.AccAddress(ed25519.GenPrivKeyFromSecret([]byte{uint8(101)}).PubKey().Address())

	// Fund the contract with stake (mint coins from the mint module)
	fundCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10000000))) // 10 ATOM
	err = gaiaApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, fundCoins)
	require.NoError(t, err)
	err = gaiaApp.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, contractWithStake, fundCoins)
	require.NoError(t, err)

	// Delegate 1 ATOM to the contract with stake
	val, err := stakingKeeper.GetValidator(ctx, valAddr)
	require.NoError(t, err)
	_, err = stakingKeeper.Delegate(ctx, contractWithStake, math.NewInt(1000000), stakingtypes.Unbonded, val, true)
	require.NoError(t, err)

	// Create the decorator
	decorator := ante.NewGovVoteMessageDecorator(gaiaApp.StakingKeeper)
	mockMsg := &mockMessenger{}
	handler := decorator(mockMsg)

	tests := []struct {
		name         string
		contractAddr sdk.AccAddress
		msg          wasmvmtypes.CosmosMsg
		expectPass   bool
		expectCalled bool
	}{
		{
			name:         "contract with stake can vote via Gov.Vote",
			contractAddr: contractWithStake,
			msg: wasmvmtypes.CosmosMsg{
				Gov: &wasmvmtypes.GovMsg{
					Vote: &wasmvmtypes.VoteMsg{
						ProposalId: 1,
						Option:     wasmvmtypes.Yes,
					},
				},
			},
			expectPass:   true,
			expectCalled: true,
		},
		{
			name:         "contract without stake cannot vote via Gov.Vote",
			contractAddr: contractWithoutStake,
			msg: wasmvmtypes.CosmosMsg{
				Gov: &wasmvmtypes.GovMsg{
					Vote: &wasmvmtypes.VoteMsg{
						ProposalId: 1,
						Option:     wasmvmtypes.Yes,
					},
				},
			},
			expectPass:   false,
			expectCalled: false,
		},
		{
			name:         "contract with stake can vote via Gov.VoteWeighted",
			contractAddr: contractWithStake,
			msg: wasmvmtypes.CosmosMsg{
				Gov: &wasmvmtypes.GovMsg{
					VoteWeighted: &wasmvmtypes.VoteWeightedMsg{
						ProposalId: 1,
						Options: []wasmvmtypes.WeightedVoteOption{
							{Option: wasmvmtypes.Yes, Weight: "1.0"},
						},
					},
				},
			},
			expectPass:   true,
			expectCalled: true,
		},
		{
			name:         "contract without stake cannot vote via Gov.VoteWeighted",
			contractAddr: contractWithoutStake,
			msg: wasmvmtypes.CosmosMsg{
				Gov: &wasmvmtypes.GovMsg{
					VoteWeighted: &wasmvmtypes.VoteWeightedMsg{
						ProposalId: 1,
						Options: []wasmvmtypes.WeightedVoteOption{
							{Option: wasmvmtypes.Yes, Weight: "1.0"},
						},
					},
				},
			},
			expectPass:   false,
			expectCalled: false,
		},
		{
			name:         "contract with stake can vote via Any with MsgVote v1 type URL",
			contractAddr: contractWithStake,
			msg: wasmvmtypes.CosmosMsg{
				Any: &wasmvmtypes.AnyMsg{
					TypeURL: "/cosmos.gov.v1.MsgVote",
					Value:   []byte{}, // content doesn't matter for this test
				},
			},
			expectPass:   true,
			expectCalled: true,
		},
		{
			name:         "contract without stake cannot vote via Any with MsgVote v1 type URL",
			contractAddr: contractWithoutStake,
			msg: wasmvmtypes.CosmosMsg{
				Any: &wasmvmtypes.AnyMsg{
					TypeURL: "/cosmos.gov.v1.MsgVote",
					Value:   []byte{},
				},
			},
			expectPass:   false,
			expectCalled: false,
		},
		{
			name:         "contract with stake can vote via Any with MsgVote v1beta1 type URL",
			contractAddr: contractWithStake,
			msg: wasmvmtypes.CosmosMsg{
				Any: &wasmvmtypes.AnyMsg{
					TypeURL: "/cosmos.gov.v1beta1.MsgVote",
					Value:   []byte{},
				},
			},
			expectPass:   true,
			expectCalled: true,
		},
		{
			name:         "contract without stake cannot vote via Any with MsgVote v1beta1 type URL",
			contractAddr: contractWithoutStake,
			msg: wasmvmtypes.CosmosMsg{
				Any: &wasmvmtypes.AnyMsg{
					TypeURL: "/cosmos.gov.v1beta1.MsgVote",
					Value:   []byte{},
				},
			},
			expectPass:   false,
			expectCalled: false,
		},
		{
			name:         "contract with stake can vote via Any with MsgVoteWeighted v1 type URL",
			contractAddr: contractWithStake,
			msg: wasmvmtypes.CosmosMsg{
				Any: &wasmvmtypes.AnyMsg{
					TypeURL: "/cosmos.gov.v1.MsgVoteWeighted",
					Value:   []byte{},
				},
			},
			expectPass:   true,
			expectCalled: true,
		},
		{
			name:         "contract without stake cannot vote via Any with MsgVoteWeighted v1 type URL",
			contractAddr: contractWithoutStake,
			msg: wasmvmtypes.CosmosMsg{
				Any: &wasmvmtypes.AnyMsg{
					TypeURL: "/cosmos.gov.v1.MsgVoteWeighted",
					Value:   []byte{},
				},
			},
			expectPass:   false,
			expectCalled: false,
		},
		{
			name:         "contract with stake can vote via Any with MsgVoteWeighted v1beta1 type URL",
			contractAddr: contractWithStake,
			msg: wasmvmtypes.CosmosMsg{
				Any: &wasmvmtypes.AnyMsg{
					TypeURL: "/cosmos.gov.v1beta1.MsgVoteWeighted",
					Value:   []byte{},
				},
			},
			expectPass:   true,
			expectCalled: true,
		},
		{
			name:         "contract without stake cannot vote via Any with MsgVoteWeighted v1beta1 type URL",
			contractAddr: contractWithoutStake,
			msg: wasmvmtypes.CosmosMsg{
				Any: &wasmvmtypes.AnyMsg{
					TypeURL: "/cosmos.gov.v1beta1.MsgVoteWeighted",
					Value:   []byte{},
				},
			},
			expectPass:   false,
			expectCalled: false,
		},
		{
			name:         "non-vote message passes through for contract with stake",
			contractAddr: contractWithStake,
			msg: wasmvmtypes.CosmosMsg{
				Bank: &wasmvmtypes.BankMsg{
					Send: &wasmvmtypes.SendMsg{
						ToAddress: contractWithoutStake.String(),
						Amount:    []wasmvmtypes.Coin{{Denom: "uatom", Amount: "1000"}},
					},
				},
			},
			expectPass:   true,
			expectCalled: true,
		},
		{
			name:         "non-vote message passes through for contract without stake",
			contractAddr: contractWithoutStake,
			msg: wasmvmtypes.CosmosMsg{
				Bank: &wasmvmtypes.BankMsg{
					Send: &wasmvmtypes.SendMsg{
						ToAddress: contractWithStake.String(),
						Amount:    []wasmvmtypes.Coin{{Denom: "uatom", Amount: "1000"}},
					},
				},
			},
			expectPass:   true,
			expectCalled: true,
		},
		{
			name:         "non-gov Any message passes through for contract without stake",
			contractAddr: contractWithoutStake,
			msg: wasmvmtypes.CosmosMsg{
				Any: &wasmvmtypes.AnyMsg{
					TypeURL: "/cosmos.bank.v1beta1.MsgSend",
					Value:   []byte{},
				},
			},
			expectPass:   true,
			expectCalled: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset the mock
			mockMsg.dispatchMsgCalled = false

			_, _, _, err := handler.DispatchMsg(ctx, tc.contractAddr, "", tc.msg)
			if tc.expectPass {
				require.NoError(t, err, "expected %s to pass", tc.name)
			} else {
				require.Error(t, err, "expected %s to fail", tc.name)
				require.Contains(t, err.Error(), "insufficient stake")
			}
			require.Equal(t, tc.expectCalled, mockMsg.dispatchMsgCalled,
				"expected wrapped handler to be called: %v, got: %v", tc.expectCalled, mockMsg.dispatchMsgCalled)
		})
	}
}

// TestWasmGovVoteDecoratorWithPartialStake tests that a contract with less than 1 ATOM staked cannot vote
func TestWasmGovVoteDecoratorWithPartialStake(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})
	stakingKeeper := gaiaApp.StakingKeeper

	// Get validator
	validators, err := stakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, validators)
	valAddr, err := stakingKeeper.ValidatorAddressCodec().StringToBytes(validators[0].GetOperator())
	require.NoError(t, err)
	valAddr = sdk.ValAddress(valAddr)

	// Create a mock "contract" address
	contractAddr := sdk.AccAddress(ed25519.GenPrivKeyFromSecret([]byte{uint8(200)}).PubKey().Address())

	// Fund the contract
	fundCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10000000))) // 10 ATOM
	err = gaiaApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, fundCoins)
	require.NoError(t, err)
	err = gaiaApp.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, contractAddr, fundCoins)
	require.NoError(t, err)

	// Delegate only 0.5 ATOM (500000 uatom) - less than required 1 ATOM
	val, err := stakingKeeper.GetValidator(ctx, valAddr)
	require.NoError(t, err)
	_, err = stakingKeeper.Delegate(ctx, contractAddr, math.NewInt(500000), stakingtypes.Unbonded, val, true)
	require.NoError(t, err)

	// Create the decorator
	decorator := ante.NewGovVoteMessageDecorator(gaiaApp.StakingKeeper)
	mockMsg := &mockMessenger{}
	handler := decorator(mockMsg)

	// Try to vote - should fail
	msg := wasmvmtypes.CosmosMsg{
		Gov: &wasmvmtypes.GovMsg{
			Vote: &wasmvmtypes.VoteMsg{
				ProposalId: 1,
				Option:     wasmvmtypes.Yes,
			},
		},
	}

	_, _, _, err = handler.DispatchMsg(ctx, contractAddr, "", msg)
	require.Error(t, err, "contract with <1 ATOM should not be able to vote")
	require.Contains(t, err.Error(), "insufficient stake")
	require.False(t, mockMsg.dispatchMsgCalled, "wrapped handler should not be called")
}

// TestWasmGovVoteDecoratorWithMultipleValidators tests voting with stake spread across multiple validators
func TestWasmGovVoteDecoratorWithMultipleValidators(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})
	stakingKeeper := gaiaApp.StakingKeeper

	// Get validator
	validators, err := stakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, validators)
	valAddr1, err := stakingKeeper.ValidatorAddressCodec().StringToBytes(validators[0].GetOperator())
	require.NoError(t, err)
	valAddr1 = sdk.ValAddress(valAddr1)

	// Create a second validator
	pk := ed25519.GenPrivKeyFromSecret([]byte{uint8(50)}).PubKey()
	validator2, err := stakingtypes.NewValidator(
		sdk.ValAddress(pk.Address()).String(),
		pk,
		stakingtypes.Description{},
	)
	require.NoError(t, err)
	valAddr2, err := stakingKeeper.ValidatorAddressCodec().StringToBytes(validator2.GetOperator())
	require.NoError(t, err)
	valAddr2 = sdk.ValAddress(valAddr2)
	validator2.Status = stakingtypes.Bonded
	err = stakingKeeper.SetValidator(ctx, validator2)
	require.NoError(t, err)
	err = stakingKeeper.SetValidatorByConsAddr(ctx, validator2)
	require.NoError(t, err)
	err = stakingKeeper.SetNewValidatorByPowerIndex(ctx, validator2)
	require.NoError(t, err)
	err = stakingKeeper.Hooks().AfterValidatorCreated(ctx, valAddr2)
	require.NoError(t, err)

	// Create a mock "contract" address
	contractAddr := sdk.AccAddress(ed25519.GenPrivKeyFromSecret([]byte{uint8(201)}).PubKey().Address())

	// Fund the contract
	fundCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(10000000))) // 10 ATOM
	err = gaiaApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, fundCoins)
	require.NoError(t, err)
	err = gaiaApp.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, contractAddr, fundCoins)
	require.NoError(t, err)

	// Delegate 0.5 ATOM to each validator (total 1 ATOM)
	val1, err := stakingKeeper.GetValidator(ctx, valAddr1)
	require.NoError(t, err)
	_, err = stakingKeeper.Delegate(ctx, contractAddr, math.NewInt(500000), stakingtypes.Unbonded, val1, true)
	require.NoError(t, err)

	val2, err := stakingKeeper.GetValidator(ctx, valAddr2)
	require.NoError(t, err)
	_, err = stakingKeeper.Delegate(ctx, contractAddr, math.NewInt(500000), stakingtypes.Unbonded, val2, true)
	require.NoError(t, err)

	// Create the decorator
	decorator := ante.NewGovVoteMessageDecorator(gaiaApp.StakingKeeper)
	mockMsg := &mockMessenger{}
	handler := decorator(mockMsg)

	// Try to vote - should pass (total stake >= 1 ATOM)
	msg := wasmvmtypes.CosmosMsg{
		Gov: &wasmvmtypes.GovMsg{
			Vote: &wasmvmtypes.VoteMsg{
				ProposalId: 1,
				Option:     wasmvmtypes.Yes,
			},
		},
	}

	_, _, _, err = handler.DispatchMsg(ctx, contractAddr, "", msg)
	require.NoError(t, err, "contract with 1 ATOM total across validators should be able to vote")
	require.True(t, mockMsg.dispatchMsgCalled, "wrapped handler should be called")
}
