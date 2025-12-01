package delegator_test

import (
	"context"
	"fmt"
	"path"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/gaia/v26/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v26/tests/interchain/delegator"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/stretchr/testify/suite"
)

type TokenFactoryAuthzSuite struct {
	*delegator.Suite
}

// createDenom creates a tokenfactory denom and returns the full denom string
func (s *TokenFactoryAuthzSuite) createDenom(subdenom string) string {
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "create-denom", subdenom,
	)
	s.Require().NoError(err)
	return fmt.Sprintf("factory/%s/%s", s.DelegatorWallet.FormattedAddress(), subdenom)
}

// mint mints tokens for a given denom
func (s *TokenFactoryAuthzSuite) mint(denom string, amount int64) {
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		s.DelegatorWallet.KeyName(),
		"tokenfactory", "mint",
		fmt.Sprintf("%d%s", amount, denom),
	)
	s.Require().NoError(err)
}

// authzGenExec generates a transaction and executes it via authz exec
func (s *TokenFactoryAuthzSuite) authzGenExec(ctx context.Context, grantee ibc.Wallet, command ...string) error {
	txjson, err := s.Chain.GenerateTx(ctx, 1, command...)
	s.Require().NoError(err)

	err = s.Chain.GetNode().WriteFile(ctx, []byte(txjson), "tx.json")
	s.Require().NoError(err)

	_, err = s.Chain.GetNode().ExecTx(
		ctx,
		grantee.FormattedAddress(),
		"authz", "exec", path.Join(s.Chain.Validators[1].HomeDir(), "tx.json"),
	)
	return err
}

// grantGenericAuthz grants generic authorization for a message type
func (s *TokenFactoryAuthzSuite) grantGenericAuthz(granter, grantee ibc.Wallet, msgType string) {
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		granter.FormattedAddress(),
		"authz", "grant", grantee.FormattedAddress(), "generic",
		"--msg-type", msgType,
	)
	s.Require().NoError(err)
}

// revokeAuthz revokes authorization
func (s *TokenFactoryAuthzSuite) revokeAuthz(granter, grantee ibc.Wallet, msgType string) {
	_, err := s.Chain.GetNode().ExecTx(
		s.GetContext(),
		granter.FormattedAddress(),
		"authz", "revoke", grantee.FormattedAddress(), msgType,
	)
	s.Require().NoError(err)
}

// TestAuthzMint verifies delegate minting capability via authz
func (s *TokenFactoryAuthzSuite) TestAuthzMint() {
	ctx := s.GetContext()

	// Create denom with DelegatorWallet (admin)
	denom := s.createDenom("authzmint")

	// Grant MsgMint authorization to DelegatorWallet2
	s.grantGenericAuthz(s.DelegatorWallet, s.DelegatorWallet2, "/osmosis.tokenfactory.v1beta1.MsgMint")

	// DelegatorWallet2 executes mint on behalf of DelegatorWallet
	mintAmount := int64(1000000)
	err := s.authzGenExec(ctx, s.DelegatorWallet2,
		"tokenfactory", "mint", fmt.Sprintf("%d%s", mintAmount, denom),
		"--from", s.DelegatorWallet.FormattedAddress())
	s.Require().NoError(err)

	// Verify tokens minted to granter's account
	balance, err := s.Chain.GetBalance(ctx, s.DelegatorWallet.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(mintAmount), balance)

	// Revoke authorization
	s.revokeAuthz(s.DelegatorWallet, s.DelegatorWallet2, "/osmosis.tokenfactory.v1beta1.MsgMint")

	// Attempt mint again (should fail)
	err = s.authzGenExec(ctx, s.DelegatorWallet2,
		"tokenfactory", "mint", fmt.Sprintf("%d%s", mintAmount, denom),
		"--from", s.DelegatorWallet.FormattedAddress())
	s.Require().Error(err)
}

// TestAuthzBurn verifies delegate burning capability
func (s *TokenFactoryAuthzSuite) TestAuthzBurn() {
	ctx := s.GetContext()

	// Create denom and mint tokens
	denom := s.createDenom("authzburn")
	mintAmount := int64(2000000)
	s.mint(denom, mintAmount)

	// Grant MsgBurn authorization
	s.grantGenericAuthz(s.DelegatorWallet, s.DelegatorWallet2, "/osmosis.tokenfactory.v1beta1.MsgBurn")

	// DelegatorWallet2 executes burn
	burnAmount := int64(500000)
	err := s.authzGenExec(ctx, s.DelegatorWallet2,
		"tokenfactory", "burn", fmt.Sprintf("%d%s", burnAmount, denom),
		"--from", s.DelegatorWallet.FormattedAddress())
	s.Require().NoError(err)

	// Verify balance decreased
	balance, err := s.Chain.GetBalance(ctx, s.DelegatorWallet.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(mintAmount-burnAmount), balance)

	// Revoke and verify burn fails
	s.revokeAuthz(s.DelegatorWallet, s.DelegatorWallet2, "/osmosis.tokenfactory.v1beta1.MsgBurn")

	err = s.authzGenExec(ctx, s.DelegatorWallet2,
		"tokenfactory", "burn", fmt.Sprintf("%d%s", 100, denom),
		"--from", s.DelegatorWallet.FormattedAddress())
	s.Require().Error(err)

	// Attempt burn exceeding balance (should fail)
	s.grantGenericAuthz(s.DelegatorWallet, s.DelegatorWallet2, "/osmosis.tokenfactory.v1beta1.MsgBurn")
	err = s.authzGenExec(ctx, s.DelegatorWallet2,
		"tokenfactory", "burn", fmt.Sprintf("%d%s", 10000000, denom),
		"--from", s.DelegatorWallet.FormattedAddress())
	s.Require().Error(err)
}

// TestAuthzCreateDenom verifies delegate denom creation capability
func (s *TokenFactoryAuthzSuite) TestAuthzCreateDenom() {
	ctx := s.GetContext()

	// Grant MsgCreateDenom authorization
	s.grantGenericAuthz(s.DelegatorWallet, s.DelegatorWallet2, "/osmosis.tokenfactory.v1beta1.MsgCreateDenom")

	// DelegatorWallet2 creates denom on behalf of DelegatorWallet
	subdenom := "authzcreate"
	err := s.authzGenExec(ctx, s.DelegatorWallet2,
		"tokenfactory", "create-denom", subdenom,
		"--from", s.DelegatorWallet.FormattedAddress())
	s.Require().NoError(err)

	// Verify denom created with granter as admin (not grantee)
	denom := fmt.Sprintf("factory/%s/%s", s.DelegatorWallet.FormattedAddress(), subdenom)
	admin, err := s.Chain.QueryJSON(ctx,
		"authority_metadata.admin", "tokenfactory", "denom-authority-metadata", denom)
	s.Require().NoError(err)
	s.Require().Equal(s.DelegatorWallet.FormattedAddress(), admin.String())

	// Verify granter can mint (is admin)
	_, err = s.Chain.GetNode().ExecTx(ctx, s.DelegatorWallet.KeyName(),
		"tokenfactory", "mint", fmt.Sprintf("%d%s", 1000, denom))
	s.Require().NoError(err)

	// Verify grantee cannot mint (not admin)
	_, err = s.Chain.GetNode().ExecTx(ctx, s.DelegatorWallet2.KeyName(),
		"tokenfactory", "mint", fmt.Sprintf("%d%s", 1000, denom))
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "unauthorized")
}

// TestAuthzChangeAdmin verifies delegate admin transfer capability
func (s *TokenFactoryAuthzSuite) TestAuthzChangeAdmin() {
	ctx := s.GetContext()

	// Create denom with DelegatorWallet
	denom := s.createDenom("authzadmin")

	// Grant MsgChangeAdmin authorization
	s.grantGenericAuthz(s.DelegatorWallet, s.DelegatorWallet2, "/osmosis.tokenfactory.v1beta1.MsgChangeAdmin")

	// DelegatorWallet2 executes change-admin to DelegatorWallet3
	err := s.authzGenExec(ctx, s.DelegatorWallet2,
		"tokenfactory", "change-admin", denom, s.DelegatorWallet3.FormattedAddress(),
		"--from", s.DelegatorWallet.FormattedAddress())
	s.Require().NoError(err)

	// Verify admin changed
	admin, err := s.Chain.QueryJSON(ctx,
		"authority_metadata.admin", "tokenfactory", "denom-authority-metadata", denom)
	s.Require().NoError(err)
	s.Require().Equal(s.DelegatorWallet3.FormattedAddress(), admin.String())

	// Verify old admin cannot mint
	_, err = s.Chain.GetNode().ExecTx(ctx, s.DelegatorWallet.KeyName(),
		"tokenfactory", "mint", fmt.Sprintf("%d%s", 1000, denom))
	s.Require().Error(err)

	// Verify new admin can mint
	_, err = s.Chain.GetNode().ExecTx(ctx, s.DelegatorWallet3.KeyName(),
		"tokenfactory", "mint", fmt.Sprintf("%d%s", 1000, denom))
	s.Require().NoError(err)
}

// TestAuthzModifyMetadata verifies delegate metadata modification
func (s *TokenFactoryAuthzSuite) TestAuthzModifyMetadata() {
	ctx := s.GetContext()

	// Create denom
	denom := s.createDenom("authzmeta")

	// Grant MsgSetDenomMetadata authorization
	s.grantGenericAuthz(s.DelegatorWallet, s.DelegatorWallet2, "/osmosis.tokenfactory.v1beta1.MsgSetDenomMetadata")

	// DelegatorWallet2 executes modify-metadata
	err := s.authzGenExec(ctx, s.DelegatorWallet2,
		"tokenfactory", "modify-metadata", denom, "AUTHZ", "Authz Test Token", "6",
		"--from", s.DelegatorWallet.FormattedAddress())
	s.Require().NoError(err)

	// Query and verify changes
	metadata, err := s.Chain.QueryJSON(ctx, "metadata", "bank", "denom-metadata", denom)
	s.Require().NoError(err)
	s.Require().Equal("AUTHZ", metadata.Get("symbol").String())

	// Revoke and verify cannot modify
	s.revokeAuthz(s.DelegatorWallet, s.DelegatorWallet2, "/osmosis.tokenfactory.v1beta1.MsgSetDenomMetadata")

	err = s.authzGenExec(ctx, s.DelegatorWallet2,
		"tokenfactory", "modify-metadata", denom, "FAIL", "Should Fail", "6",
		"--from", s.DelegatorWallet.FormattedAddress())
	s.Require().Error(err)
}

// TestAuthzMultipleOperations verifies multiple grants work independently
func (s *TokenFactoryAuthzSuite) TestAuthzMultipleOperations() {
	ctx := s.GetContext()

	// Create denom and mint
	denom := s.createDenom("authzmulti")
	s.mint(denom, 3000000)

	// Grant multiple authorizations
	s.grantGenericAuthz(s.DelegatorWallet, s.DelegatorWallet2, "/osmosis.tokenfactory.v1beta1.MsgMint")
	s.grantGenericAuthz(s.DelegatorWallet, s.DelegatorWallet2, "/osmosis.tokenfactory.v1beta1.MsgBurn")
	s.grantGenericAuthz(s.DelegatorWallet, s.DelegatorWallet2, "/osmosis.tokenfactory.v1beta1.MsgSetDenomMetadata")

	// DelegatorWallet2 performs all three operations
	err := s.authzGenExec(ctx, s.DelegatorWallet2,
		"tokenfactory", "mint", fmt.Sprintf("%d%s", 1000000, denom),
		"--from", s.DelegatorWallet.FormattedAddress())
	s.Require().NoError(err)

	err = s.authzGenExec(ctx, s.DelegatorWallet2,
		"tokenfactory", "burn", fmt.Sprintf("%d%s", 500000, denom),
		"--from", s.DelegatorWallet.FormattedAddress())
	s.Require().NoError(err)

	err = s.authzGenExec(ctx, s.DelegatorWallet2,
		"tokenfactory", "modify-metadata", denom, "MULTI", "Multi Test", "6",
		"--from", s.DelegatorWallet.FormattedAddress())
	s.Require().NoError(err)

	// Revoke only MsgMint
	s.revokeAuthz(s.DelegatorWallet, s.DelegatorWallet2, "/osmosis.tokenfactory.v1beta1.MsgMint")

	// Verify mint fails
	err = s.authzGenExec(ctx, s.DelegatorWallet2,
		"tokenfactory", "mint", fmt.Sprintf("%d%s", 1000, denom),
		"--from", s.DelegatorWallet.FormattedAddress())
	s.Require().Error(err)

	// Verify burn still works
	err = s.authzGenExec(ctx, s.DelegatorWallet2,
		"tokenfactory", "burn", fmt.Sprintf("%d%s", 100000, denom),
		"--from", s.DelegatorWallet.FormattedAddress())
	s.Require().NoError(err)

	// Verify metadata still works
	err = s.authzGenExec(ctx, s.DelegatorWallet2,
		"tokenfactory", "modify-metadata", denom, "MULTI2", "Multi Test 2", "6",
		"--from", s.DelegatorWallet.FormattedAddress())
	s.Require().NoError(err)

	// Revoke remaining
	s.revokeAuthz(s.DelegatorWallet, s.DelegatorWallet2, "/osmosis.tokenfactory.v1beta1.MsgBurn")
	s.revokeAuthz(s.DelegatorWallet, s.DelegatorWallet2, "/osmosis.tokenfactory.v1beta1.MsgSetDenomMetadata")

	// Verify all fail
	err = s.authzGenExec(ctx, s.DelegatorWallet2,
		"tokenfactory", "burn", fmt.Sprintf("%d%s", 1000, denom),
		"--from", s.DelegatorWallet.FormattedAddress())
	s.Require().Error(err)

	err = s.authzGenExec(ctx, s.DelegatorWallet2,
		"tokenfactory", "modify-metadata", denom, "FAIL", "Should Fail", "6",
		"--from", s.DelegatorWallet.FormattedAddress())
	s.Require().Error(err)
}

// TestAuthzChainedDelegation verifies authz chaining doesn't work (by design)
func (s *TokenFactoryAuthzSuite) TestAuthzChainedDelegation() {
	ctx := s.GetContext()

	// Create denom
	denom := s.createDenom("authzchain")

	// DelegatorWallet grants mint to DelegatorWallet2
	s.grantGenericAuthz(s.DelegatorWallet, s.DelegatorWallet2, "/osmosis.tokenfactory.v1beta1.MsgMint")

	// DelegatorWallet2 grants mint to DelegatorWallet3
	s.grantGenericAuthz(s.DelegatorWallet2, s.DelegatorWallet3, "/osmosis.tokenfactory.v1beta1.MsgMint")

	// DelegatorWallet3 attempts mint on behalf of DelegatorWallet (should fail - no chaining)
	err := s.authzGenExec(ctx, s.DelegatorWallet3,
		"tokenfactory", "mint", fmt.Sprintf("%d%s", 1000, denom),
		"--from", s.DelegatorWallet.FormattedAddress())
	s.Require().Error(err)
}

// TestAuthzUnauthorizedOperation verifies operations without grant fail properly
func (s *TokenFactoryAuthzSuite) TestAuthzUnauthorizedOperation() {
	ctx := s.GetContext()

	// Create denom
	denom := s.createDenom("authzunauth")

	// DelegatorWallet2 attempts mint WITHOUT grant (should fail)
	err := s.authzGenExec(ctx, s.DelegatorWallet2,
		"tokenfactory", "mint", fmt.Sprintf("%d%s", 1000, denom),
		"--from", s.DelegatorWallet.FormattedAddress())
	s.Require().Error(err)

	// Grant mint authorization
	s.grantGenericAuthz(s.DelegatorWallet, s.DelegatorWallet2, "/osmosis.tokenfactory.v1beta1.MsgMint")

	// Mint should now work
	err = s.authzGenExec(ctx, s.DelegatorWallet2,
		"tokenfactory", "mint", fmt.Sprintf("%d%s", 1000, denom),
		"--from", s.DelegatorWallet.FormattedAddress())
	s.Require().NoError(err)

	// DelegatorWallet2 attempts burn WITHOUT burn grant (should fail)
	err = s.authzGenExec(ctx, s.DelegatorWallet2,
		"tokenfactory", "burn", fmt.Sprintf("%d%s", 500, denom),
		"--from", s.DelegatorWallet.FormattedAddress())
	s.Require().Error(err)

	// Grant burn authorization
	s.grantGenericAuthz(s.DelegatorWallet, s.DelegatorWallet2, "/osmosis.tokenfactory.v1beta1.MsgBurn")

	// Burn should now work
	err = s.authzGenExec(ctx, s.DelegatorWallet2,
		"tokenfactory", "burn", fmt.Sprintf("%d%s", 500, denom),
		"--from", s.DelegatorWallet.FormattedAddress())
	s.Require().NoError(err)
}

func TestTokenFactoryAuthz(t *testing.T) {
	s := &TokenFactoryAuthzSuite{
		Suite: &delegator.Suite{
			Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
				UpgradeOnSetup: true,
			}),
		},
	}
	suite.Run(t, s)
}
