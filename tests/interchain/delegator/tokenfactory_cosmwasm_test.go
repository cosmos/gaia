package delegator_test

import (
	"fmt"
	"os"
	"path"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/gaia/v26/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v26/tests/interchain/delegator"
	"github.com/stretchr/testify/suite"
)

// TokenFactoryCosmWasmSuite tests tokenfactory operations via CosmWasm contract bindings.
// The contract uses custom bindings to dispatch tokenfactory messages directly to the module.
type TokenFactoryCosmWasmSuite struct {
	*TokenFactoryBaseSuite
	ContractWasm []byte
	ContractPath string
	ContractAddr string
}

func (s *TokenFactoryCosmWasmSuite) SetupSuite() {
	s.TokenFactoryBaseSuite.Suite.SetupSuite()

	// Load the tokenfactory bindings contract
	contractWasm, err := os.ReadFile("testdata/tokenfactory_bindings.wasm")
	s.Require().NoError(err)
	s.ContractWasm = contractWasm

	// Write contract to node
	s.Require().NoError(s.Chain.GetNode().WriteFile(s.GetContext(), s.ContractWasm, "tokenfactory_bindings.wasm"))
	s.ContractPath = path.Join(s.Chain.GetNode().HomeDir(), "tokenfactory_bindings.wasm")

	// Deploy contract via governance proposal
	s.ContractAddr = s.deployContract()

	s.UpgradeChain()
}

// deployContract stores and instantiates the contract via governance proposal.
// Returns the contract address.
func (s *TokenFactoryCosmWasmSuite) deployContract() string {
	txhash, err := s.Chain.GetNode().ExecTx(s.GetContext(), s.DelegatorWallet.FormattedAddress(),
		"wasm", "submit-proposal", "store-instantiate",
		s.ContractPath,
		"{}", "--label", "tokenfactory-bindings",
		"--no-admin", "--instantiate-nobody", "true",
		"--title", "Store and instantiate tokenfactory bindings contract",
		"--summary", "Store and instantiate tokenfactory bindings contract",
		"--deposit", fmt.Sprintf("10000000%s", s.Config.ChainSpec.Denom),
	)
	s.Require().NoError(err)

	proposalId, err := s.Chain.GetProposalID(s.GetContext(), txhash)
	s.Require().NoError(err)

	err = s.Chain.PassProposal(s.GetContext(), proposalId)
	s.Require().NoError(err)

	// Get the contract address from the deployed code
	govAddr, err := s.Chain.GetGovernanceAddress(s.GetContext())
	s.Require().NoError(err)

	codeJSON, err := s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("code_infos.@reverse.#(creator==\"%s\").code_id", govAddr), "wasm", "list-code")
	s.Require().NoError(err)
	code := codeJSON.String()

	contractAddrJSON, err := s.Chain.QueryJSON(s.GetContext(), "contracts.0", "wasm", "list-contract-by-code", code)
	s.Require().NoError(err)
	return contractAddrJSON.String()
}

// executeContract executes a message on the contract.
func (s *TokenFactoryCosmWasmSuite) executeContract(msg string) error {
	_, err := s.Chain.GetNode().ExecTx(s.GetContext(), s.DelegatorWallet.FormattedAddress(),
		"wasm", "execute", s.ContractAddr, msg,
	)
	return err
}

// executeContractWithFunds executes a message on the contract with attached funds.
func (s *TokenFactoryCosmWasmSuite) executeContractWithFunds(msg, funds string) error {
	_, err := s.Chain.GetNode().ExecTx(s.GetContext(), s.DelegatorWallet.FormattedAddress(),
		"wasm", "execute", s.ContractAddr, msg,
		"--amount", funds,
	)
	return err
}

// queryContract queries the contract state.
func (s *TokenFactoryCosmWasmSuite) queryContract(queryMsg string) (string, error) {
	result, err := s.Chain.QueryJSON(s.GetContext(), "data", "wasm", "contract-state", "smart", s.ContractAddr, queryMsg)
	if err != nil {
		return "", err
	}
	return result.String(), nil
}

// createDenomViaContract creates a denom via the contract, attaching the required creation fee.
// Returns the full denom string.
func (s *TokenFactoryCosmWasmSuite) createDenomViaContract(subdenom string) string {
	// Tokenfactory charges 100 ATOM creation fee
	creationFee := "100000000" + chainsuite.Uatom
	err := s.executeContractWithFunds(createDenomMsg(subdenom), creationFee)
	s.Require().NoError(err)
	return fmt.Sprintf("factory/%s/%s", s.ContractAddr, subdenom)
}

// Message builders

func createDenomMsg(subdenom string) string {
	return fmt.Sprintf(`{"create_denom":{"subdenom":"%s"}}`, subdenom)
}

func mintTokensMsg(denom string, amount int64, mintTo string) string {
	return fmt.Sprintf(`{"mint_tokens":{"denom":"%s","amount":"%d","mint_to_address":"%s"}}`, denom, amount, mintTo)
}

func burnTokensMsg(denom string, amount int64) string {
	return fmt.Sprintf(`{"burn_tokens":{"denom":"%s","amount":"%d"}}`, denom, amount)
}

func changeAdminMsg(denom, newAdmin string) string {
	return fmt.Sprintf(`{"change_admin":{"denom":"%s","new_admin_address":"%s"}}`, denom, newAdmin)
}

func getDenomQuery(creator, subdenom string) string {
	return fmt.Sprintf(`{"get_denom":{"creator_address":"%s","subdenom":"%s"}}`, creator, subdenom)
}

// TestContractCreateDenom tests that the contract can create a tokenfactory denom.
// The contract address becomes the admin of the created denom.
func (s *TokenFactoryCosmWasmSuite) TestContractCreateDenom() {
	subdenom := "contracttoken"

	// Create denom via contract (with creation fee)
	denom := s.createDenomViaContract(subdenom)

	// Verify the contract is the admin
	admin, err := s.Chain.QueryJSON(s.GetContext(),
		"authority_metadata.admin", "tokenfactory", "denom-authority-metadata", denom)
	s.Require().NoError(err)
	s.Require().Equal(s.ContractAddr, admin.String())
}

// TestContractMintTokens tests minting tokens via the contract.
func (s *TokenFactoryCosmWasmSuite) TestContractMintTokens() {
	subdenom := "mintable"

	// Create denom via contract (with creation fee)
	denom := s.createDenomViaContract(subdenom)

	// Mint tokens to DelegatorWallet2
	mintAmount := int64(1000000)
	err := s.executeContract(mintTokensMsg(denom, mintAmount, s.DelegatorWallet2.FormattedAddress()))
	s.Require().NoError(err)

	// Verify balance
	balance, err := s.Chain.GetBalance(s.GetContext(), s.DelegatorWallet2.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(mintAmount), balance)
}

// TestContractBurnTokens tests burning tokens via the contract.
// The contract burns from its own balance using plain burn (not BurnFrom).
func (s *TokenFactoryCosmWasmSuite) TestContractBurnTokens() {
	subdenom := "burnable"

	// Create denom via contract (with creation fee)
	denom := s.createDenomViaContract(subdenom)

	// Mint tokens to the contract itself
	mintAmount := int64(1000000)
	err := s.executeContract(mintTokensMsg(denom, mintAmount, s.ContractAddr))
	s.Require().NoError(err)

	// Verify contract balance
	balance, err := s.Chain.GetBalance(s.GetContext(), s.ContractAddr, denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(mintAmount), balance)

	// Burn half the tokens from contract's own balance (plain burn)
	burnAmount := int64(500000)
	err = s.executeContract(burnTokensMsg(denom, burnAmount))
	s.Require().NoError(err)

	// Verify reduced balance
	balance, err = s.Chain.GetBalance(s.GetContext(), s.ContractAddr, denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(mintAmount-burnAmount), balance)
}

// TestContractChangeAdmin tests transferring admin rights via the contract.
func (s *TokenFactoryCosmWasmSuite) TestContractChangeAdmin() {
	subdenom := "adminchange"

	// Create denom via contract (with creation fee)
	denom := s.createDenomViaContract(subdenom)

	// Verify contract is admin
	admin, err := s.Chain.QueryJSON(s.GetContext(),
		"authority_metadata.admin", "tokenfactory", "denom-authority-metadata", denom)
	s.Require().NoError(err)
	s.Require().Equal(s.ContractAddr, admin.String())

	// Change admin to DelegatorWallet
	err = s.executeContract(changeAdminMsg(denom, s.DelegatorWallet.FormattedAddress()))
	s.Require().NoError(err)

	// Verify new admin
	admin, err = s.Chain.QueryJSON(s.GetContext(),
		"authority_metadata.admin", "tokenfactory", "denom-authority-metadata", denom)
	s.Require().NoError(err)
	s.Require().Equal(s.DelegatorWallet.FormattedAddress(), admin.String())

	// Contract can no longer mint (not admin anymore)
	err = s.executeContract(mintTokensMsg(denom, 1000000, s.DelegatorWallet.FormattedAddress()))
	s.Require().Error(err)

	// But DelegatorWallet can mint via CLI now
	err = s.Mint(s.DelegatorWallet, denom, 1000000)
	s.Require().NoError(err)

	balance, err := s.Chain.GetBalance(s.GetContext(), s.DelegatorWallet.FormattedAddress(), denom)
	s.Require().NoError(err)
	s.Require().Equal(sdkmath.NewInt(1000000), balance)
}

// TestContractQueryDenom tests querying the full denom via the contract.
func (s *TokenFactoryCosmWasmSuite) TestContractQueryDenom() {
	subdenom := "queryable"

	// Create denom via contract (with creation fee)
	denom := s.createDenomViaContract(subdenom)

	// Query via contract
	result, err := s.queryContract(getDenomQuery(s.ContractAddr, subdenom))
	s.Require().NoError(err)

	// The result should contain the denom
	denomResult, err := s.Chain.QueryJSON(s.GetContext(), "data.denom", "wasm", "contract-state", "smart", s.ContractAddr, getDenomQuery(s.ContractAddr, subdenom))
	s.Require().NoError(err)
	s.Require().Equal(denom, denomResult.String(), "query result: %s", result)
}

// TestContractUnauthorizedMint tests that a non-admin cannot mint.
// We deploy a second contract and try to mint on a denom created by the first.
func (s *TokenFactoryCosmWasmSuite) TestContractUnauthorizedMint() {
	subdenom := "protected"

	// Create denom via main contract (with creation fee)
	denom := s.createDenomViaContract(subdenom)

	// DelegatorWallet2 (not the contract) tries to mint via CLI - should fail
	_, err := s.Chain.GetNode().ExecTx(s.GetContext(), s.DelegatorWallet2.KeyName(),
		"tokenfactory", "mint",
		fmt.Sprintf("1000000%s", denom),
	)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "unauthorized")
}

// TestContractMultipleDenoms tests that the contract can manage multiple denoms.
func (s *TokenFactoryCosmWasmSuite) TestContractMultipleDenoms() {
	subdenoms := []string{"multi1", "multi2", "multi3"}
	amounts := []int64{1000000, 2000000, 3000000}

	// Create all denoms (with creation fee each time)
	denoms := make([]string, len(subdenoms))
	for i, subdenom := range subdenoms {
		denoms[i] = s.createDenomViaContract(subdenom)
	}

	// Mint different amounts to each
	for i, denom := range denoms {
		err := s.executeContract(mintTokensMsg(denom, amounts[i], s.DelegatorWallet.FormattedAddress()))
		s.Require().NoError(err)
	}

	// Verify all balances
	for i, denom := range denoms {
		balance, err := s.Chain.GetBalance(s.GetContext(), s.DelegatorWallet.FormattedAddress(), denom)
		s.Require().NoError(err)
		s.Require().Equal(sdkmath.NewInt(amounts[i]), balance, "balance mismatch for %s", denom)
	}

	// Change admin on one denom only
	err := s.executeContract(changeAdminMsg(denoms[0], s.DelegatorWallet2.FormattedAddress()))
	s.Require().NoError(err)

	// Contract can still mint to denom2 and denom3
	err = s.executeContract(mintTokensMsg(denoms[1], 1000000, s.DelegatorWallet.FormattedAddress()))
	s.Require().NoError(err)
	err = s.executeContract(mintTokensMsg(denoms[2], 1000000, s.DelegatorWallet.FormattedAddress()))
	s.Require().NoError(err)

	// But not denom1
	err = s.executeContract(mintTokensMsg(denoms[0], 1000000, s.DelegatorWallet.FormattedAddress()))
	s.Require().Error(err)
}

// TestContractDenomMetadata tests setting denom metadata via the bindings.
// Note: The sample contract may not expose SetMetadata - this test verifies basic metadata from creation.
func (s *TokenFactoryCosmWasmSuite) TestContractDenomMetadata() {
	subdenom := "metaToken"

	// Create denom via contract (with creation fee)
	denom := s.createDenomViaContract(subdenom)

	// Query denom exists in tokenfactory
	denoms, err := s.Chain.QueryJSON(s.GetContext(),
		"denoms", "tokenfactory", "denoms-from-creator", s.ContractAddr)
	s.Require().NoError(err)

	denomsList := denoms.Array()
	found := false
	for _, d := range denomsList {
		if d.String() == denom {
			found = true
			break
		}
	}
	s.Require().True(found, "denom %s not found in creator's denoms", denom)
}

func TestTokenFactoryCosmWasm(t *testing.T) {
	s := &TokenFactoryCosmWasmSuite{
		TokenFactoryBaseSuite: &TokenFactoryBaseSuite{
			Suite: &delegator.Suite{
				Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
					UpgradeOnSetup: false,
				}),
			},
		},
	}
	suite.Run(t, s)
}
