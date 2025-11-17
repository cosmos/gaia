package msg

import (
	"fmt"
	"path/filepath"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/gaia/v26/tests/e2e/common"
	"github.com/cosmos/gaia/v26/tests/e2e/data"
)

func WriteGovCommunitySpendProposal(c *common.Chain, amount sdk.Coin, recipient string) error {
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
	propMsgBody := fmt.Sprintf(template, common.GovModuleAddress, recipient, amount.Denom, amount.Amount.String())
	err := common.WriteFile(filepath.Join(c.Validators[0].ConfigDir(), "config", common.ProposalCommunitySpendFilename), []byte(propMsgBody))
	if err != nil {
		return err
	}
	return nil
}

func WriteSoftwareUpgradeProposal(c *common.Chain, height int64, name string) error {
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

	err := common.WriteFile(filepath.Join(c.Validators[0].ConfigDir(), "config", common.ProposalSoftwareUpgrade), []byte(propMsgBody))
	if err != nil {
		return err
	}
	return nil
}

func WriteCancelSoftwareUpgradeProposal(c *common.Chain) error {
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

	err := common.WriteFile(filepath.Join(c.Validators[0].ConfigDir(), "config", common.ProposalCancelSoftwareUpgrade), []byte(template))
	if err != nil {
		return err
	}
	return nil
}

func WriteLiquidStakingParamsUpdateProposal(c *common.Chain, global int64, val int64) error {
	template := `
	{
		"messages": [
		 {
		  "@type": "/gaia.liquid.v1beta1.MsgUpdateParams",
		  "authority": "%s",
		  "params": {
		   "global_liquid_staking_cap": "%s",
		   "validator_liquid_staking_cap": "%s"
		  }
		 }
		],
		"metadata": "ipfs://CID",
		"deposit": "100uatom",
		"title": "Update Liquid Params",
		"summary": "e2e-test updating Liquid staking params",
		"expedited": false
	   }`
	propMsgBody := fmt.Sprintf(template,
		common.GovAuthority,
		math.LegacyNewDecWithPrec(global, 2), // global_liquid_staking_cap
		math.LegacyNewDecWithPrec(val, 2),    // validator_liquid_staking_cap
	)

	return common.WriteFile(filepath.Join(c.Validators[0].ConfigDir(), "config", common.ProposalLiquidParamUpdateFilename),
		[]byte(propMsgBody))
}

// WriteGovParamChangeProposalBlocksPerEpoch writes a governance proposal JSON file to change the `BlocksPerEpoch`
// parameter to the provided `blocksPerEpoch`
func WriteGovParamChangeProposalBlocksPerEpoch(c *common.Chain, paramsJSON string) error {
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
		common.GovAuthority,
		paramsJSON,
	)

	err := common.WriteFile(filepath.Join(c.Validators[0].ConfigDir(), "config", common.ProposalBlocksPerEpochFilename), []byte(propMsgBody))
	if err != nil {
		return err
	}
	return nil
}

// WriteFailingExpeditedProposal writes a governance proposal JSON file.
// The proposal fails because only SoftwareUpgrade and CancelSoftwareUpgrade can be expedited.
func WriteFailingExpeditedProposal(c *common.Chain, blocksPerEpoch int64) error {
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
		common.GovAuthority,
		blocksPerEpoch,
	)

	err := common.WriteFile(filepath.Join(c.Validators[0].ConfigDir(), "config", common.ProposalFailExpedited), []byte(propMsgBody))
	if err != nil {
		return err
	}
	return nil
}

// MsgSoftwareUpgrade can be expedited and it can only be submitted using "tx gov submit-proposal" command.
func WriteExpeditedSoftwareUpgradeProp(c *common.Chain) error {
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

	err := common.WriteFile(filepath.Join(c.Validators[0].ConfigDir(), "config", common.ProposalExpeditedSoftwareUpgrade), []byte(body))
	if err != nil {
		return err
	}
	return nil
}

func WriteAddRateLimitAtomProposal(c *common.Chain, v2 bool) error {
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

	channel := common.TransferChannel
	if v2 {
		channel = common.V2TransferClient
	}
	propMsgBody := fmt.Sprintf(template,
		common.GovAuthority,
		common.UAtomDenom,       // denom: uatom
		channel,                 // channel_or_client_id: channel-0 / 08-wasm-1
		math.NewInt(1).String(), // max_percent_send: 1%
		math.NewInt(1).String(), // max_percent_recv: 1%
		24,                      // duration_hours: 24
	)

	err := common.WriteFile(filepath.Join(c.Validators[0].ConfigDir(), "config", common.ProposalAddRateLimitAtomFilename), []byte(propMsgBody))
	if err != nil {
		return err
	}
	return nil
}

func WriteAddRateLimitStakeProposal(c *common.Chain, v2 bool) error {
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

	channel := common.TransferChannel
	if v2 {
		channel = common.V2TransferClient
	}
	propMsgBody := fmt.Sprintf(template,
		common.GovAuthority,
		common.StakeDenom,        // denom: stake
		channel,                  // channel_or_client_id: channel-0 / 08-wasm-1
		math.NewInt(10).String(), // max_percent_send: 10%
		math.NewInt(5).String(),  // max_percent_recv: 5%
		6,                        // duration_hours: 6
	)

	err := common.WriteFile(filepath.Join(c.Validators[0].ConfigDir(), "config", common.ProposalAddRateLimitStakeFilename), []byte(propMsgBody))
	if err != nil {
		return err
	}
	return nil
}

func WriteUpdateRateLimitAtomProposal(c *common.Chain, v2 bool) error {
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

	channel := common.TransferChannel
	if v2 {
		channel = common.V2TransferClient
	}
	propMsgBody := fmt.Sprintf(template,
		common.GovAuthority,
		common.UAtomDenom,       // denom: uatom
		channel,                 // channel_or_client_id: channel-0 / 08-wasm-1
		math.NewInt(2).String(), // max_percent_send: 2%
		math.NewInt(1).String(), // max_percent_recv: 1%
		6,                       // duration_hours: 6
	)

	err := common.WriteFile(filepath.Join(c.Validators[0].ConfigDir(), "config", common.ProposalUpdateRateLimitAtomFilename), []byte(propMsgBody))
	if err != nil {
		return err
	}
	return nil
}

func WriteResetRateLimitAtomProposal(c *common.Chain, v2 bool) error {
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

	channel := common.TransferChannel
	if v2 {
		channel = common.V2TransferClient
	}
	propMsgBody := fmt.Sprintf(template,
		common.GovAuthority,
		common.UAtomDenom, // denom: uatom
		channel,           // channel_or_client_id: channel-0 / 08-wasm-1
	)

	err := common.WriteFile(filepath.Join(c.Validators[0].ConfigDir(), "config", common.ProposalResetRateLimitAtomFilename), []byte(propMsgBody))
	if err != nil {
		return err
	}
	return nil
}

func WriteRemoveRateLimitAtomProposal(c *common.Chain, v2 bool) error {
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

	channel := common.TransferChannel
	if v2 {
		channel = common.V2TransferClient
	}
	propMsgBody := fmt.Sprintf(template,
		common.GovAuthority,
		common.UAtomDenom, // denom: uatom
		channel,           // channel_or_client_id: channel-0 / 08-wasm-1
	)

	err := common.WriteFile(filepath.Join(c.Validators[0].ConfigDir(), "config", common.ProposalRemoveRateLimitAtomFilename), []byte(propMsgBody))
	if err != nil {
		return err
	}
	return nil
}

func WriteRemoveRateLimitStakeProposal(c *common.Chain, v2 bool) error {
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

	channel := common.TransferChannel
	if v2 {
		channel = common.V2TransferClient
	}
	propMsgBody := fmt.Sprintf(template,
		common.GovAuthority,
		common.StakeDenom, // denom: stake
		channel,           // channel_or_client_id: channel-0 / 08-wasm-1
	)

	err := common.WriteFile(filepath.Join(c.Validators[0].ConfigDir(), "config", common.ProposalRemoveRateLimitStakeFilename), []byte(propMsgBody))
	if err != nil {
		return err
	}
	return nil
}

func WriteStoreWasmLightClientProposal(c *common.Chain) error {
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
		common.GovAuthority,
		data.WasmDummyLightClient,
	)

	err := common.WriteFile(filepath.Join(c.Validators[0].ConfigDir(), "config", common.ProposalStoreWasmLightClientFilename), []byte(propMsgBody))
	if err != nil {
		return err
	}
	return nil
}
