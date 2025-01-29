package types

import (
	"testing"

	"github.com/stretchr/testify/suite"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CodecTestSuite struct {
	suite.Suite
}

func TestCodecSuite(t *testing.T) {
	suite.Run(t, new(CodecTestSuite))
}

func (suite *CodecTestSuite) TestRegisterInterfaces() {
	registry := codectypes.NewInterfaceRegistry()
	registry.RegisterInterface(sdk.MsgInterfaceProtoName, (*sdk.Msg)(nil))
	RegisterInterfaces(registry)

	impls := registry.ListImplementations(sdk.MsgInterfaceProtoName)
	suite.Require().Equal(7, len(impls))
	suite.Require().ElementsMatch([]string{
		"/gaia.tokenfactory.v1beta1.MsgCreateDenom",
		"/gaia.tokenfactory.v1beta1.MsgMint",
		"/gaia.tokenfactory.v1beta1.MsgBurn",
		"/gaia.tokenfactory.v1beta1.MsgChangeAdmin",
		"/gaia.tokenfactory.v1beta1.MsgSetDenomMetadata",
		"/gaia.tokenfactory.v1beta1.MsgForceTransfer",
		"/gaia.tokenfactory.v1beta1.MsgUpdateParams",
	}, impls)
}
