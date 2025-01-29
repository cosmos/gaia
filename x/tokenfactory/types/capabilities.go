package types

const (
	EnableSetMetadata   = "enable_metadata"
	EnableForceTransfer = "enable_force_transfer"
	EnableBurnFrom      = "enable_burn_from"
	// Allows addresses of your choosing to mint tokens based on specific conditions.
	// via the IsSudoAdminFunc.
	// NOTE: with SudoMint enabled, the sudo admin can mint `any` token, not just tokenfactory tokens.
	// This is intended behavior as requested by other teams, rather than having its own module with very minor logic.
	// If you do not wish for this behavior, write your own and do not use this capability.
	EnableSudoMint = "enable_admin_sudo_mint"
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
