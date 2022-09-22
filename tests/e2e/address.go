package e2e

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/rand"
)

// PubKey returns a sample account PubKey
func PubKey() crypto.PubKey {
	seed := []byte(strconv.Itoa(rand.Int()))
	return ed25519.GenPrivKeyFromSecret(seed).PubKey()
}

// AccAddress returns a sample account address
func AccAddress() sdk.AccAddress {
	addr := PubKey().Address()
	return sdk.AccAddress(addr)
}

// Address returns a sample string account address
func Address() string {
	return AccAddress().String()
}

// ConsAddress returns a sample consensus address
func ConsAddress() sdk.ConsAddress {
	return sdk.ConsAddress(PubKey().Address())
}
