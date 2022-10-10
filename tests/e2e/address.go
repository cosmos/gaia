package e2e

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// HDPath generates an HD path based on the wallet index
func HDPath(index int) string {
	return fmt.Sprintf("m/44'/118'/0'/0/%d", index)
}

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
