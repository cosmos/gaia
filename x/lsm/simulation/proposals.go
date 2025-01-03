package simulation

import (
	"math/rand"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/cosmos/gaia/v22/x/lsm/types"
)

// Simulation operation weights constants
const (
	DefaultWeightMsgUpdateParams int = 100

	OpWeightMsgUpdateParams = "op_weight_msg_update_params" //nolint:gosec
)

// ProposalMsgs defines the module weighted proposals' contents
func ProposalMsgs() []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{
		simulation.NewWeightedProposalMsg(
			OpWeightMsgUpdateParams,
			DefaultWeightMsgUpdateParams,
			SimulateMsgUpdateParams,
		),
	}
}

// SimulateMsgUpdateParams returns a random MsgUpdateParams
func SimulateMsgUpdateParams(r *rand.Rand, _ sdk.Context, _ []simtypes.Account) sdk.Msg {
	// use the default gov module account address as authority
	var authority sdk.AccAddress = address.Module("gov")

	params := types.DefaultParams()
	params.GlobalLiquidStakingCap = simtypes.RandomDecAmount(r, math.LegacyOneDec())
	params.ValidatorLiquidStakingCap = simtypes.RandomDecAmount(r, math.LegacyOneDec())
	randSeed := simtypes.RandIntBetween(r, -1, 10000)
	if randSeed != 0 {
		params.ValidatorBondFactor = math.LegacyNewDecFromInt(math.NewInt(int64(randSeed)))
	}

	return &types.MsgUpdateParams{
		Authority: authority.String(),
		Params:    params,
	}
}
