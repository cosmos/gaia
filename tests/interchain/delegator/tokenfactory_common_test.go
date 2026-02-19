package delegator_test

import (
	"fmt"

	"github.com/cosmos/gaia/v27/tests/interchain/delegator"
	"github.com/cosmos/interchaintest/v10/ibc"
)

// TokenFactoryBaseSuite is a base suite that provides common tokenfactory helpers.
// Test suites that need tokenfactory functionality should embed this instead
// of embedding *delegator.Suite directly.
type TokenFactoryBaseSuite struct {
	*delegator.Suite
}

// CreateDenom creates a tokenfactory denom and returns the full denom string.
// Returns an error if the creation fails, allowing the caller to provide context.
func (s *TokenFactoryBaseSuite) CreateDenom(wallet ibc.Wallet, subdenom string) (string, error) {
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		wallet.KeyName(),
		"tokenfactory", "create-denom", subdenom,
	)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("factory/%s/%s", wallet.FormattedAddress(), subdenom), nil
}

// Mint mints tokens for a tokenfactory denom.
// Returns an error if the mint fails, allowing the caller to provide context.
func (s *TokenFactoryBaseSuite) Mint(wallet ibc.Wallet, denom string, amount int64) error {
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		wallet.KeyName(),
		"tokenfactory", "mint",
		fmt.Sprintf("%d%s", amount, denom),
	)
	return err
}

// MintTo mints tokens for a tokenfactory denom directly to a recipient address.
// Returns an error if the mint fails, allowing the caller to provide context.
func (s *TokenFactoryBaseSuite) MintTo(wallet ibc.Wallet, denom string, amount int64, recipient string) error {
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		wallet.KeyName(),
		"tokenfactory", "mint-to",
		recipient,
		fmt.Sprintf("%d%s", amount, denom),
	)
	return err
}

// CreateAndMint creates a tokenfactory denom and mints tokens in one operation.
// Returns the denom string and any error that occurred.
func (s *TokenFactoryBaseSuite) CreateAndMint(wallet ibc.Wallet, subdenom string, amount int64) (string, error) {
	denom, err := s.CreateDenom(wallet, subdenom)
	if err != nil {
		return "", err
	}
	if err := s.Mint(wallet, denom, amount); err != nil {
		return "", err
	}
	return denom, nil
}
