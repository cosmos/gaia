package e2e

import (
	"cosmossdk.io/math"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gaia/v23/tests/e2e/data"
	"path/filepath"
)

func (s *IntegrationTestSuite) writeGovCommunitySpendProposal(c *chain, amount sdk.Coin, recipient string) {
	template := `
	{
		"messages":[
		  {
			"@type": "/cosmos.distribution.v1beta1.MsgCommunityPoolSpend",
			"authority": "%s",
			"recipient": "%s",
			"amount": [{
				"denom": "%s",
				"amount": "%s"
			}]
		  }
		],
		"deposit": "100uatom",
		"proposer": "Proposing validator address",
		"metadata": "Community Pool Spend",
		"title": "Fund Team!",
		"summary": "summary",
		"expedited": false
	}
	`
	propMsgBody := fmt.Sprintf(template, govModuleAddress, recipient, amount.Denom, amount.Amount.String())
	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalCommunitySpendFilename), []byte(propMsgBody))
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) writeSoftwareUpgradeProposal(c *chain, height int64, name string) {
	body := `{
		"messages": [
		 {
		  "@type": "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
		  "authority": "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
		  "plan": {
		   "name": "%s",
		   "height": "%d",
		   "info": "test",
		   "upgraded_client_state": null
		  }
		 }
		],
		"metadata": "ipfs://CID",
		"deposit": "100uatom",
		"title": "title",
		"summary": "test"
	   }`

	propMsgBody := fmt.Sprintf(body, name, height)

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalSoftwareUpgrade), []byte(propMsgBody))
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) writeCancelSoftwareUpgradeProposal(c *chain) {
	template := `{
		"messages": [
		 {
		  "@type": "/cosmos.upgrade.v1beta1.MsgCancelUpgrade",
		  "authority": "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"
		 }
		],
		"metadata": "ipfs://CID",
		"deposit": "100uatom",
		"title": "title",
		"summary": "test"
	   }`

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalCancelSoftwareUpgrade), []byte(template))
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) writeLiquidStakingParamsUpdateProposal(c *chain, oldParams stakingtypes.Params) {
	template := `
	{
		"messages": [
		 {
		  "@type": "/cosmos.staking.v1beta1.MsgUpdateParams",
		  "authority": "%s",
		  "params": {
		   "unbonding_time": "%s",
		   "max_validators": %d,
		   "max_entries": %d,
		   "historical_entries": %d,
		   "bond_denom": "%s",
		   "min_commission_rate": "%s",
		   "validator_bond_factor": "%s",
		   "global_liquid_staking_cap": "%s",
		   "validator_liquid_staking_cap": "%s"
		  }
		 }
		],
		"metadata": "ipfs://CID",
		"deposit": "100uatom",
		"title": "Update LSM Params",
		"summary": "e2e-test updating LSM staking params",
		"expedited": false
	   }`
	propMsgBody := fmt.Sprintf(template,
		govAuthority,
		oldParams.UnbondingTime,
		oldParams.MaxValidators,
		oldParams.MaxEntries,
		oldParams.HistoricalEntries,
		oldParams.BondDenom,
		oldParams.MinCommissionRate,
		math.LegacyNewDec(250),           // validator bond factor
		math.LegacyNewDecWithPrec(25, 2), // 25 global_liquid_staking_cap
		math.LegacyNewDecWithPrec(50, 2), // 50 validator_liquid_staking_cap
	)

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalLSMParamUpdateFilename), []byte(propMsgBody))
	s.Require().NoError(err)
}

// writeGovParamChangeProposalBlocksPerEpoch writes a governance proposal JSON file to change the `BlocksPerEpoch`
// parameter to the provided `blocksPerEpoch`
func (s *IntegrationTestSuite) writeGovParamChangeProposalBlocksPerEpoch(c *chain, paramsJSON string) {
	template := `
	{
		"messages":[
		  {
			"@type": "/interchain_security.ccv.provider.v1.MsgUpdateParams",
   			"authority": "%s",
			"params": %s
		  }
		],
		"deposit": "100uatom",
		"proposer": "sample proposer",
		"metadata": "sample metadata",
		"title": "blocks per epoch title",
		"summary": "blocks per epoch summary",
		"expedited": false
	}`

	propMsgBody := fmt.Sprintf(template,
		govAuthority,
		paramsJSON,
	)

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalBlocksPerEpochFilename), []byte(propMsgBody))
	s.Require().NoError(err)
}

// writeFailingExpeditedProposal writes a governance proposal JSON file.
// The proposal fails because only SoftwareUpgrade and CancelSoftwareUpgrade can be expedited.
func (s *IntegrationTestSuite) writeFailingExpeditedProposal(c *chain, blocksPerEpoch int64) {
	template := `
	{
		"messages":[
		  {
			"@type": "/cosmos.gov.v1.MsgExecLegacyContent",
			"authority": "%s",
			"content": {
				"@type": "/cosmos.params.v1beta1.ParameterChangeProposal",
				"title": "BlocksPerEpoch",
				"description": "change blocks per epoch",
				"changes": [{
				  "subspace": "provider",
				  "key": "BlocksPerEpoch",
				  "value": "\"%d\""
				}]
			}
		  }
		],
		"deposit": "100uatom",
		"proposer": "sample proposer",
		"metadata": "sample metadata",
		"title": "blocks per epoch title",
		"summary": "blocks per epoch summary",
		"expedited": true
	}`

	propMsgBody := fmt.Sprintf(template,
		govAuthority,
		blocksPerEpoch,
	)

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalFailExpedited), []byte(propMsgBody))
	s.Require().NoError(err)
}

// MsgSoftwareUpgrade can be expedited and it can only be submitted using "tx gov submit-proposal" command.
func (s *IntegrationTestSuite) writeExpeditedSoftwareUpgradeProp(c *chain) {
	body := `{
 "messages": [
  {
   "@type": "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
   "authority": "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
   "plan": {
    "name": "test-expedited-upgrade",
    "height": "123456789",
    "info": "test",
    "upgraded_client_state": null
   }
  }
 ],
 "metadata": "ipfs://CID",
 "deposit": "100uatom",
 "title": "title",
 "summary": "test",
 "expedited": true
}`

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalExpeditedSoftwareUpgrade), []byte(body))
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) writeAddRateLimitAtomProposal(c *chain, v2 bool) {
	template := `
	{
		"messages": [
		 {
		  "@type": "/ratelimit.v1.MsgAddRateLimit",
		  "authority": "%s",
		  "denom": "%s",
		  "channel_or_client_id": "%s",
		  "max_percent_send": "%s",
		  "max_percent_recv": "%s",
		  "duration_hours": "%d"
		 }
		],
		"metadata": "ipfs://CID",
		"deposit": "100uatom",
		"title": "Add Rate Limit on (channel-0, uatom)",
		"summary": "e2e-test adding an IBC rate limit"
	   }`

	channel := transferChannel
	if v2 {
		channel = v2TransferClient
	}
	propMsgBody := fmt.Sprintf(template,
		govAuthority,
		uatomDenom,              // denom: uatom
		channel,                 // channel_or_client_id: channel-0 / 08-wasm-1
		math.NewInt(1).String(), // max_percent_send: 1%
		math.NewInt(1).String(), // max_percent_recv: 1%
		24,                      // duration_hours: 24
	)

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalAddRateLimitAtomFilename), []byte(propMsgBody))
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) writeAddRateLimitStakeProposal(c *chain, v2 bool) {
	template := `
	{
		"messages": [
		 {
		  "@type": "/ratelimit.v1.MsgAddRateLimit",
		  "authority": "%s",
		  "denom": "%s",
		  "channel_or_client_id": "%s",
		  "max_percent_send": "%s",
		  "max_percent_recv": "%s",
		  "duration_hours": "%d"
		 }
		],
		"metadata": "ipfs://CID",
		"deposit": "100uatom",
		"title": "Add Rate Limit on (channel-0, stake)",
		"summary": "e2e-test adding an IBC rate limit"
	   }`

	channel := transferChannel
	if v2 {
		channel = v2TransferClient
	}
	propMsgBody := fmt.Sprintf(template,
		govAuthority,
		stakeDenom,               // denom: stake
		channel,                  // channel_or_client_id: channel-0 / 08-wasm-1
		math.NewInt(10).String(), // max_percent_send: 10%
		math.NewInt(5).String(),  // max_percent_recv: 5%
		6,                        // duration_hours: 6
	)

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalAddRateLimitStakeFilename), []byte(propMsgBody))
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) writeUpdateRateLimitAtomProposal(c *chain, v2 bool) {
	template := `
	{
		"messages": [
		 {
		  "@type": "/ratelimit.v1.MsgUpdateRateLimit",
		  "authority": "%s",
		  "denom": "%s",
		  "channel_or_client_id": "%s",
		  "max_percent_send": "%s",
		  "max_percent_recv": "%s",
		  "duration_hours": "%d"
		 }
		],
		"metadata": "ipfs://CID",
		"deposit": "100uatom",
		"title": "Update Rate Limit on (channel-0, uatom)",
		"summary": "e2e-test updating an IBC rate limit"
	   }`

	channel := transferChannel
	if v2 {
		channel = v2TransferClient
	}
	propMsgBody := fmt.Sprintf(template,
		govAuthority,
		uatomDenom,              // denom: uatom
		channel,                 // channel_or_client_id: channel-0 / 08-wasm-1
		math.NewInt(2).String(), // max_percent_send: 2%
		math.NewInt(1).String(), // max_percent_recv: 1%
		6,                       // duration_hours: 6
	)

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalUpdateRateLimitAtomFilename), []byte(propMsgBody))
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) writeResetRateLimitAtomProposal(c *chain, v2 bool) {
	template := `
	{
		"messages": [
		 {
		  "@type": "/ratelimit.v1.MsgResetRateLimit",
		  "authority": "%s",
		  "denom": "%s",
		  "channel_or_client_id": "%s"
		 }
		],
		"metadata": "ipfs://CID",
		"deposit": "100uatom",
		"title": "Reset Rate Limit on (channel-0, uatom)",
		"summary": "e2e-test resetting an IBC rate limit"
	   }`

	channel := transferChannel
	if v2 {
		channel = v2TransferClient
	}
	propMsgBody := fmt.Sprintf(template,
		govAuthority,
		uatomDenom, // denom: uatom
		channel,    // channel_or_client_id: channel-0 / 08-wasm-1
	)

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalResetRateLimitAtomFilename), []byte(propMsgBody))
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) writeRemoveRateLimitAtomProposal(c *chain, v2 bool) {
	template := `
	{
		"messages": [
		 {
		  "@type": "/ratelimit.v1.MsgRemoveRateLimit",
		  "authority": "%s",
		  "denom": "%s",
		  "channel_or_client_id": "%s"
		 }
		],
		"metadata": "ipfs://CID",
		"deposit": "100uatom",
		"title": "Remove Rate Limit (channel-0, uatom)",
		"summary": "e2e-test removing an IBC rate limit"
	   }`

	channel := transferChannel
	if v2 {
		channel = v2TransferClient
	}
	propMsgBody := fmt.Sprintf(template,
		govAuthority,
		uatomDenom, // denom: uatom
		channel,    // channel_or_client_id: channel-0 / 08-wasm-1
	)

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalRemoveRateLimitAtomFilename), []byte(propMsgBody))
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) writeRemoveRateLimitStakeProposal(c *chain, v2 bool) {
	template := `
	{
		"messages": [
		 {
		  "@type": "/ratelimit.v1.MsgRemoveRateLimit",
		  "authority": "%s",
		  "denom": "%s",
		  "channel_or_client_id": "%s"
		 }
		],
		"metadata": "ipfs://CID",
		"deposit": "100uatom",
		"title": "Remove Rate Limit (channel-0, stake)",
		"summary": "e2e-test removing an IBC rate limit"
	   }`

	channel := transferChannel
	if v2 {
		channel = v2TransferClient
	}
	propMsgBody := fmt.Sprintf(template,
		govAuthority,
		stakeDenom, // denom: stake
		channel,    // channel_or_client_id: channel-0 / 08-wasm-1
	)

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalRemoveRateLimitStakeFilename), []byte(propMsgBody))
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) writeStoreWasmLightClientProposal(c *chain) {
	template := `
	{
		"messages": [
			{
			"@type": "/ibc.lightclients.wasm.v1.MsgStoreCode",
			"signer": "%s",
			"wasm_byte_code": "%s"
			}
		],
		"metadata": "AQ==",
		"deposit": "100uatom",
		"title": "Store wasm light client code",
		"summary": "e2e-test storing wasm light client code"
	   }`
	propMsgBody := fmt.Sprintf(template,
		govAuthority,
		data.WasmDummyLightClient,
	)

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalStoreWasmLightClientFilename), []byte(propMsgBody))
	s.Require().NoError(err)
}
