package simulation

import (
	"fmt"
	"math/rand"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/cosmos/gaia/v22/x/lsm/keeper"
	"github.com/cosmos/gaia/v22/x/lsm/types"
)

// Simulation operation weight constants
const (
	DefaultWeightMsgTokenizeShares                       int = 25
	DefaultWeightMsgRedeemTokensforShares                int = 25
	DefaultWeightMsgTransferTokenizeShareRecord          int = 5
	DefaultWeightMsgEnableTokenizeShares                 int = 1
	DefaultWeightMsgDisableTokenizeShares                int = 1
	DefaultWeightMsgWithdrawAllTokenizeShareRecordReward int = 50

	OpWeightMsgTokenizeShares                       = "op_weight_msg_tokenize_shares"                           //nolint:gosec
	OpWeightMsgRedeemTokensforShares                = "op_weight_msg_redeem_tokens_for_shares"                  //nolint:gosec
	OpWeightMsgTransferTokenizeShareRecord          = "op_weight_msg_transfer_tokenize_share_record"            //nolint:gosec
	OpWeightMsgDisableTokenizeShares                = "op_weight_msg_disable_tokenize_shares"                   //nolint:gosec
	OpWeightMsgEnableTokenizeShares                 = "op_weight_msg_enable_tokenize_shares"                    //nolint:gosec
	OpWeightMsgWithdrawAllTokenizeShareRecordReward = "op_weight_msg_withdraw_all_tokenize_share_record_reward" //nolint:gosec
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams,
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	sk types.StakingKeeper,
	k *keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgTokenizeShares                       int
		weightMsgRedeemTokensforShares                int
		weightMsgTransferTokenizeShareRecord          int
		weightMsgDisableTokenizeShares                int
		weightMsgEnableTokenizeShares                 int
		weightMsgWithdrawAllTokenizeShareRecordReward int
	)

	appParams.GetOrGenerate(OpWeightMsgTokenizeShares, &weightMsgTokenizeShares, nil,
		func(_ *rand.Rand) {
			weightMsgTokenizeShares = DefaultWeightMsgTokenizeShares
		},
	)

	appParams.GetOrGenerate(OpWeightMsgRedeemTokensforShares, &weightMsgRedeemTokensforShares, nil,
		func(_ *rand.Rand) {
			weightMsgRedeemTokensforShares = DefaultWeightMsgRedeemTokensforShares
		},
	)

	appParams.GetOrGenerate(OpWeightMsgTransferTokenizeShareRecord, &weightMsgTransferTokenizeShareRecord, nil,
		func(_ *rand.Rand) {
			weightMsgTransferTokenizeShareRecord = DefaultWeightMsgTransferTokenizeShareRecord
		},
	)

	appParams.GetOrGenerate(OpWeightMsgDisableTokenizeShares, &weightMsgDisableTokenizeShares, nil,
		func(_ *rand.Rand) {
			weightMsgDisableTokenizeShares = DefaultWeightMsgDisableTokenizeShares
		},
	)

	appParams.GetOrGenerate(OpWeightMsgEnableTokenizeShares, &weightMsgEnableTokenizeShares, nil,
		func(_ *rand.Rand) {
			weightMsgEnableTokenizeShares = DefaultWeightMsgEnableTokenizeShares
		},
	)

	appParams.GetOrGenerate(OpWeightMsgWithdrawAllTokenizeShareRecordReward,
		&weightMsgWithdrawAllTokenizeShareRecordReward, nil, func(r *rand.Rand) {
			weightMsgWithdrawAllTokenizeShareRecordReward = DefaultWeightMsgWithdrawAllTokenizeShareRecordReward
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgTokenizeShares,
			SimulateMsgTokenizeShares(txGen, ak, bk, sk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgRedeemTokensforShares,
			SimulateMsgRedeemTokensforShares(txGen, ak, bk, sk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgTransferTokenizeShareRecord,
			SimulateMsgTransferTokenizeShareRecord(txGen, ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgDisableTokenizeShares,
			SimulateMsgDisableTokenizeShares(txGen, ak, bk, sk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgEnableTokenizeShares,
			SimulateMsgEnableTokenizeShares(txGen, ak, bk, sk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgWithdrawAllTokenizeShareRecordReward,
			SimulateMsgWithdrawAllTokenizeShareRecordReward(txGen, ak, bk, k),
		),
	}
}

// SimulateMsgTokenizeShares generates a MsgTokenizeShares with random values
func SimulateMsgTokenizeShares(txGen client.TxConfig, ak types.AccountKeeper, bk types.BankKeeper,
	sk types.StakingKeeper, k *keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msgType := sdk.MsgTypeURL(&types.MsgTokenizeShares{})

		vals, err := sk.GetAllValidators(ctx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to get validators"), nil, err
		}

		// get random validator
		validator, ok := testutil.RandSliceElem(r, vals)
		if !ok {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to pick validator"), nil, nil
		}

		valAddr, err := sk.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "error getting validator address bytes"), nil, err
		}

		delegations, err := sk.GetValidatorDelegations(ctx, valAddr)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "error getting validator delegations"), nil, nil
		}

		if delegations == nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "keeper does have any delegation entries"), nil, nil
		}

		// get random delegator from src validator
		delegation := delegations[r.Intn(len(delegations))]
		delAddr := delegation.GetDelegatorAddr()
		delAddrBz, err := ak.AddressCodec().StringToBytes(delAddr)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "error getting delegator address bytes"), nil, err
		}

		// make sure delegation is not a validator bond
		if delegation.ValidatorBond {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "can't tokenize a validator bond"), nil, nil
		}

		// make sure tokenizations are not disabled
		lockStatus, _ := k.GetTokenizeSharesLock(ctx, sdk.AccAddress(delAddrBz))
		if lockStatus != types.TOKENIZE_SHARE_LOCK_STATUS_UNLOCKED {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "tokenize shares disabled"), nil, nil
		}

		// Make sure that the delegator has no ongoing redelegations to the validator
		found, err := sk.HasReceivingRedelegation(ctx, delAddrBz, valAddr)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "error checking receiving redelegation"), nil, err
		}
		if found {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "delegator has redelegations in progress"), nil, nil
		}

		// get random destination validator
		totalBond := validator.TokensFromShares(delegation.GetShares()).TruncateInt()
		if !totalBond.IsPositive() {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "total bond is negative"), nil, nil
		}

		tokenizeShareAmt, err := simtypes.RandPositiveInt(r, totalBond)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to generate positive amount"), nil, err
		}

		if tokenizeShareAmt.IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "amount is zero"), nil, nil
		}

		bondDenom, err := sk.BondDenom(ctx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "failed to find bond denom"), nil, err
		}

		account := ak.GetAccount(ctx, sdk.AccAddress(delAddrBz))
		if account, ok := account.(vesting.VestingAccount); ok {
			if tokenizeShareAmt.GT(account.GetDelegatedFree().AmountOf(bondDenom)) {
				return simtypes.NoOpMsg(types.ModuleName, msgType, "account vests and amount exceeds free portion"), nil, nil
			}
		}

		// check if the shares truncate to zero
		shares, err := validator.SharesFromTokens(tokenizeShareAmt)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "invalid shares"), nil, err
		}

		if validator.TokensFromShares(shares).TruncateInt().IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "shares truncate to zero"), nil, nil // skip
		}

		// check that tokenization would not exceed global cap
		params, err := k.GetParams(ctx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "failed to get params"), nil, err
		}

		totalBondedTokens, err := sk.TotalBondedTokens(ctx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "failed to get total bonded errors"), nil, err
		}

		totalStaked := math.LegacyNewDecFromInt(totalBondedTokens)
		if totalStaked.IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "cannot happened - no validators bonded if stake is 0.0"), nil, nil // skip
		}
		totalLiquidStaked := math.LegacyNewDecFromInt(k.GetTotalLiquidStakedTokens(ctx).Add(tokenizeShareAmt))
		liquidStakedPercent := totalLiquidStaked.Quo(totalStaked)
		if liquidStakedPercent.GT(params.GlobalLiquidStakingCap) {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "global liquid staking cap exceeded"), nil, nil
		}

		// check that tokenization would not exceed validator liquid staking cap
		validatorTotalShares := validator.DelegatorShares
		validatorLiquidShares := validator.LiquidShares.Add(shares)
		validatorLiquidSharesPercent := validatorLiquidShares.Quo(validatorTotalShares)
		if validatorLiquidSharesPercent.GT(params.ValidatorLiquidStakingCap) {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "validator liquid staking cap exceeded"), nil, nil
		}

		// check that tokenization would not exceed validator bond cap
		maxValidatorLiquidShares := validator.ValidatorBondShares.Mul(params.ValidatorBondFactor)
		if validator.LiquidShares.Add(shares).GT(maxValidatorLiquidShares) {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "validator bond cap exceeded"), nil, nil
		}

		// need to retrieve the simulation account associated with delegation to retrieve PrivKey
		var simAccount simtypes.Account

		for _, simAcc := range accs {
			if simAcc.Address.Equals(sdk.AccAddress(delAddrBz)) {
				simAccount = simAcc
				break
			}
		}

		// if simaccount.PrivKey == nil, delegation address does not exist in accs. Return error
		if simAccount.PrivKey == nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "account private key is nil"), nil, nil
		}

		msg := &types.MsgTokenizeShares{
			DelegatorAddress:    delAddr,
			ValidatorAddress:    validator.GetOperator(),
			Amount:              sdk.NewCoin(bondDenom, tokenizeShareAmt),
			TokenizedShareOwner: delAddr,
		}

		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           txGen,
			Cdc:             nil,
			Msg:             msg,
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgRedeemTokensforShares generates a MsgRedeemTokensforShares with random values
func SimulateMsgRedeemTokensforShares(txGen client.TxConfig, ak types.AccountKeeper, bk types.BankKeeper,
	sk types.StakingKeeper, k *keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msgType := sdk.MsgTypeURL(&types.MsgRedeemTokensForShares{})

		redeemUser := simtypes.Account{}
		redeemCoin := sdk.Coin{}
		tokenizeShareRecord := types.TokenizeShareRecord{}

		records := k.GetAllTokenizeShareRecords(ctx)
		if len(records) > 0 {
			record := records[r.Intn(len(records))]
			for _, acc := range accs {
				balance := bk.GetBalance(ctx, acc.Address, record.GetShareTokenDenom())
				if balance.Amount.IsPositive() {
					redeemUser = acc
					redeemAmount, err := simtypes.RandPositiveInt(r, balance.Amount)
					if err == nil {
						redeemCoin = sdk.NewCoin(record.GetShareTokenDenom(), redeemAmount)
						tokenizeShareRecord = record
					}
					break
				}
			}
		}

		// if redeemUser.PrivKey == nil, redeem user does not exist in accs
		if redeemUser.PrivKey == nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "account private key is nil"), nil, nil
		}

		if redeemCoin.Amount.IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "empty balance in tokens"), nil, nil
		}

		valAddress, err := sk.ValidatorAddressCodec().StringToBytes(tokenizeShareRecord.Validator)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "invalid validator address"), nil, fmt.Errorf("invalid validator address")
		}
		validator, err := sk.GetValidator(ctx, valAddress)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "validator not found"), nil, fmt.Errorf("validator not found")
		}
		delegation, err := sk.GetDelegation(ctx, tokenizeShareRecord.GetModuleAddress(), valAddress)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "delegation not found"), nil, fmt.Errorf("delegation not found")
		}

		// prevent redemption that returns a 0 amount
		shares := math.LegacyNewDecFromInt(redeemCoin.Amount)
		if redeemCoin.Amount.Equal(delegation.Shares.TruncateInt()) {
			shares = delegation.Shares
		}

		if validator.TokensFromShares(shares).TruncateInt().IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "zero tokens returned"), nil, nil
		}

		account := ak.GetAccount(ctx, redeemUser.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		msg := &types.MsgRedeemTokensForShares{
			DelegatorAddress: redeemUser.Address.String(),
			Amount:           redeemCoin,
		}

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           txGen,
			Cdc:             nil,
			Msg:             msg,
			Context:         ctx,
			SimAccount:      redeemUser,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgTransferTokenizeShareRecord generates a MsgTransferTokenizeShareRecord with random values
func SimulateMsgTransferTokenizeShareRecord(txGen client.TxConfig, ak types.AccountKeeper, bk types.BankKeeper, k *keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msgType := sdk.MsgTypeURL(&types.MsgTransferTokenizeShareRecord{})

		simAccount, _ := simtypes.RandomAcc(r, accs)
		destAccount, _ := simtypes.RandomAcc(r, accs)
		transferRecord := types.TokenizeShareRecord{}

		records := k.GetAllTokenizeShareRecords(ctx)
		if len(records) > 0 {
			record := records[r.Intn(len(records))]
			for _, acc := range accs {
				if record.Owner == acc.Address.String() {
					simAccount = acc
					transferRecord = record
					break
				}
			}
		}

		// if simAccount.PrivKey == nil, record owner does not exist in accs
		if simAccount.PrivKey == nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "account private key is nil"), nil, nil
		}

		if transferRecord.Id == 0 {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "share record not found"), nil, nil
		}

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		msg := &types.MsgTransferTokenizeShareRecord{
			TokenizeShareRecordId: transferRecord.Id,
			Sender:                simAccount.Address.String(),
			NewOwner:              destAccount.Address.String(),
		}

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           txGen,
			Cdc:             nil,
			Msg:             msg,
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

func SimulateMsgDisableTokenizeShares(txGen client.TxConfig, ak types.AccountKeeper, bk types.BankKeeper,
	sk types.StakingKeeper, k *keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msgType := sdk.MsgTypeURL(&types.MsgDisableTokenizeShares{})
		simAccount, _ := simtypes.RandomAcc(r, accs)

		if simAccount.PrivKey == nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "account private key is nil"), nil, nil
		}

		denom, err := sk.BondDenom(ctx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "bond denom not found"), nil, err
		}

		balance := bk.GetBalance(ctx, simAccount.Address, denom).Amount
		if !balance.IsPositive() {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "balance is negative"), nil, nil
		}

		lockStatus, _ := k.GetTokenizeSharesLock(ctx, simAccount.Address)
		if lockStatus == types.TOKENIZE_SHARE_LOCK_STATUS_LOCKED {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "account already locked"), nil, nil
		}

		msg := &types.MsgDisableTokenizeShares{
			DelegatorAddress: simAccount.Address.String(),
		}

		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         txGen,
			Cdc:           nil,
			Msg:           msg,
			Context:       ctx,
			SimAccount:    simAccount,
			AccountKeeper: ak,
			Bankkeeper:    bk,
			ModuleName:    types.ModuleName,
		}
		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

func SimulateMsgEnableTokenizeShares(txGen client.TxConfig, ak types.AccountKeeper, bk types.BankKeeper,
	sk types.StakingKeeper, k *keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msgType := sdk.MsgTypeURL(&types.MsgEnableTokenizeShares{})
		simAccount, _ := simtypes.RandomAcc(r, accs)

		if simAccount.PrivKey == nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "account private key is nil"), nil, nil
		}

		denom, err := sk.BondDenom(ctx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "bond denom not found"), nil, err
		}

		balance := bk.GetBalance(ctx, simAccount.Address, denom).Amount
		if !balance.IsPositive() {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "balance is negative"), nil, nil
		}

		lockStatus, _ := k.GetTokenizeSharesLock(ctx, simAccount.Address)
		if lockStatus != types.TOKENIZE_SHARE_LOCK_STATUS_LOCKED {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "account is not locked"), nil, nil
		}

		msg := &types.MsgEnableTokenizeShares{
			DelegatorAddress: simAccount.Address.String(),
		}

		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         txGen,
			Cdc:           nil,
			Msg:           msg,
			Context:       ctx,
			SimAccount:    simAccount,
			AccountKeeper: ak,
			Bankkeeper:    bk,
			ModuleName:    types.ModuleName,
		}
		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgWithdrawAllTokenizeShareRecordReward simulates MsgWithdrawTokenizeShareRecordReward execution where
// a random account claim tokenize share record rewards.
func SimulateMsgWithdrawAllTokenizeShareRecordReward(txConfig client.TxConfig, ak types.AccountKeeper,
	bk types.BankKeeper, k *keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msgType := sdk.MsgTypeURL(&types.MsgWithdrawTokenizeShareRecordReward{})
		rewardOwner, _ := simtypes.RandomAcc(r, accs)

		records := k.GetAllTokenizeShareRecords(ctx)
		if len(records) > 0 {
			record := records[r.Intn(len(records))]
			for _, acc := range accs {
				if acc.Address.String() == record.Owner {
					rewardOwner = acc
					break
				}
			}
		}

		// if simaccount.PrivKey == nil, delegation address does not exist in accs. Return error
		if rewardOwner.PrivKey == nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "account private key is nil"), nil, nil
		}

		rewardOwnerAddr, err := ak.AddressCodec().BytesToString(rewardOwner.Address)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "error converting reward owner address"), nil, err
		}

		msg := &types.MsgWithdrawAllTokenizeShareRecordReward{
			OwnerAddress: rewardOwnerAddr,
		}

		account := ak.GetAccount(ctx, rewardOwner.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           txConfig,
			Cdc:             nil,
			Msg:             msg,
			Context:         ctx,
			SimAccount:      rewardOwner,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
