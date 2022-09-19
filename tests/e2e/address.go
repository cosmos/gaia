package e2e

import (
	"math/rand"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

const (
	// HDPathZero is the parts of the BIP44 HD path that are fixed by
	// what we used during the ATOM fundraiser.
	HDPathZero = "m/44'/118'/0'/0/0"
	// HDPathOne is the second wallet from the BIP44 HD path.
	HDPathOne = "m/44'/118'/0'/0/1"
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
