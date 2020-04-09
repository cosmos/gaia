package app

import (
	"fmt"

	"github.com/tendermint/tendermint/crypto/secp256k1"

	codecstd "github.com/cosmos/cosmos-sdk/codec/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"

	"github.com/cosmos/gaia/app"
)

// This will fail half the time with the second output being 173
// This is due to secp256k1 signatures not being constant size.
// nolint: vet
func ExampleTxSendSize() {
	cdc := codecstd.MakeCodec(app.ModuleBasics)

	var gas uint64 = 1

	priv1 := secp256k1.GenPrivKeySecp256k1([]byte{0})
	addr1 := sdk.AccAddress(priv1.PubKey().Address())
	priv2 := secp256k1.GenPrivKeySecp256k1([]byte{1})
	addr2 := sdk.AccAddress(priv2.PubKey().Address())
	coins := sdk.Coins{sdk.NewCoin("denom", sdk.NewInt(10))}
	msg1 := bank.MsgMultiSend{
		Inputs:  []bank.Input{bank.NewInput(addr1, coins)},
		Outputs: []bank.Output{bank.NewOutput(addr2, coins)},
	}
	fee := auth.NewStdFee(gas, coins)
	signBytes := auth.StdSignBytes("example-chain-ID",
		1, 1, fee, []sdk.Msg{msg1}, "")
	sig, err := priv1.Sign(signBytes)
	if err != nil {
		return
	}
	sigs := []auth.StdSignature{{nil, sig}}
	tx := auth.NewStdTx([]sdk.Msg{msg1}, fee, sigs, "")
	fmt.Println(len(cdc.MustMarshalBinaryBare([]sdk.Msg{msg1})))
	fmt.Println(len(cdc.MustMarshalBinaryBare(tx)))
	// output: 80
	// 169
}
