package params

import (
	"strings"

	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
)

var (
	// BypassMinFeeMsgTypesKey defines the configuration key for the
	// BypassMinFeeMsgTypes value.
	//nolint: gosec
	BypassMinFeeMsgTypesKey = "bypass-min-fee-msg-types"

	// customGaiaConfigTemplate defines Gaia's custom application configuration TOML template.
	customGaiaConfigTemplate = `
###############################################################################
###                        Custom Gaia Configuration                        ###
###############################################################################
# bypass-min-fee-msg-types defines custom message types the operator may set that
# will bypass minimum fee checks during CheckTx.
# NOTE:
# bypass-min-fee-msg-types = [] will deactivate the bypass - no messages will be allowed to bypass the minimum fee check
# bypass-min-fee-msg-types = [<MsgType>...] will allow messages of specified type to bypass the minimum fee check
# removing bypass-min-fee-msg-types from the config file will apply the default values:
# ["/ibc.core.channel.v1.MsgRecvPacket", "/ibc.core.channel.v1.MsgAcknowledgement", "/ibc.core.client.v1.MsgUpdateClient"]
#
# Example:
# bypass-min-fee-msg-types = ["/ibc.core.channel.v1.MsgRecvPacket", "/ibc.core.channel.v1.MsgAcknowledgement", "/ibc.core.client.v1.MsgUpdateClient"]
bypass-min-fee-msg-types = [{{ range .BypassMinFeeMsgTypes }}{{ printf "%q, " . }}{{end}}]
`
)

// CustomConfigTemplate defines Gaia's custom application configuration TOML
// template. It extends the core SDK template.
func CustomConfigTemplate() string {
	config := serverconfig.DefaultConfigTemplate
	lines := strings.Split(config, "\n")
	// add the Gaia config at the second line of the file
	lines[2] += customGaiaConfigTemplate
	return strings.Join(lines, "\n")
}

// CustomAppConfig defines Gaia's custom application configuration.
type CustomAppConfig struct {
	serverconfig.Config

	// BypassMinFeeMsgTypes defines custom message types the operator may set that
	// will bypass minimum fee checks during CheckTx.
	// NOTE:
	// bypass-min-fee-msg-types = [] will deactivate the bypass - no messages will be allowed to bypass the minimum fee check
	// bypass-min-fee-msg-types = [<some_msg_type>] will allow messages of specified type to bypass the minimum fee check
	// omitting bypass-min-fee-msg-types from the config file will use the default values: ["/ibc.core.channel.v1.MsgRecvPacket", "/ibc.core.channel.v1.MsgAcknowledgement", "/ibc.core.client.v1.MsgUpdateClient"]
	BypassMinFeeMsgTypes []string `mapstructure:"bypass-min-fee-msg-types"`
}
