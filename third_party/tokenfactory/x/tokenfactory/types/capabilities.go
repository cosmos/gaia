package types

const (
	EnableSetMetadata   = "enable_metadata"
	EnableForceTransfer = "enable_force_transfer"
	EnableBurnFrom      = "enable_burn_from"
	// EnableCommunityPoolFeeFunding sends tokens to the community pool when a new fee is charged (if one is set in params).
	// This is useful for ICS chains, or networks who wish to just have the fee tokens burned (not gas fees, just the extra on top).
	EnableCommunityPoolFeeFunding = "enable_community_pool_fee_funding"
)

func IsCapabilityEnabled(enabledCapabilities []string, capability string) bool {
	if len(enabledCapabilities) == 0 {
		return false
	}

	for _, v := range enabledCapabilities {
		if v == capability {
			return true
		}
	}

	return false
}
