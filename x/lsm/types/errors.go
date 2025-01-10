package types

import "cosmossdk.io/errors"

// x/lsm module sentinel errors
var (
	ErrRedelegationInProgress                  = errors.Register(ModuleName, 120, "delegator is not allowed to tokenize shares from validator with a redelegation in progress")
	ErrInsufficientShares                      = errors.Register(ModuleName, 22, "insufficient delegation shares")
	ErrTokenizeShareRecordNotExists            = errors.Register(ModuleName, 102, "tokenize share record not exists")
	ErrTokenizeShareRecordAlreadyExists        = errors.Register(ModuleName, 103, "tokenize share record already exists")
	ErrNotTokenizeShareRecordOwner             = errors.Register(ModuleName, 104, "not tokenize share record owner")
	ErrExceedingFreeVestingDelegations         = errors.Register(ModuleName, 105, "trying to exceed vested free delegation for vesting account")
	ErrOnlyBondDenomAllowdForTokenize          = errors.Register(ModuleName, 106, "only bond denom is allowed for tokenize")
	ErrInsufficientValidatorBondShares         = errors.Register(ModuleName, 107, "insufficient validator bond shares")
	ErrValidatorBondNotAllowedForTokenizeShare = errors.Register(ModuleName, 109, "validator bond delegation is not allowed to tokenize share")
	ErrGlobalLiquidStakingCapExceeded          = errors.Register(ModuleName, 111, "delegation or tokenization exceeds the global cap")
	ErrValidatorLiquidStakingCapExceeded       = errors.Register(ModuleName, 112, "delegation or tokenization exceeds the validator cap")
	ErrTokenizeSharesDisabledForAccount        = errors.Register(ModuleName, 113, "tokenize shares currently disabled for account")
	ErrTokenizeSharesAlreadyEnabledForAccount  = errors.Register(ModuleName, 115, "tokenize shares is already enabled for this account")
	ErrTokenizeSharesAlreadyDisabledForAccount = errors.Register(ModuleName, 116, "tokenize shares is already disabled for this account")
	ErrValidatorLiquidSharesUnderflow          = errors.Register(ModuleName, 117, "validator liquid shares underflow")
	ErrTotalLiquidStakedUnderflow              = errors.Register(ModuleName, 118, "total liquid staked underflow")
	ErrNotEnoughBalance                        = errors.Register(ModuleName, 101, "not enough balance")
	ErrTinyRedemptionAmount                    = errors.Register(ModuleName, 119, "too few tokens to redeem (truncates to zero tokens)")
	ErrNoDelegation                            = errors.Register(ModuleName, 19, "no delegation for (address, validator) tuple")
	ErrNoValidatorFound                        = errors.Register(ModuleName, 3, "validator does not exist")
)
