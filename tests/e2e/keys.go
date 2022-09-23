package e2e

import (
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
)

// createMnemonic creates a random string mnemonic
func createMnemonic() (string, error) {
	entropySeed, err := bip39.NewEntropy(256)
	if err != nil {
		return "", err
	}

	mnemonic, err := bip39.NewMnemonic(entropySeed)
	if err != nil {
		return "", err
	}

	return mnemonic, nil
}

// createMemoryKeyFromMnemonic creates a key into keystore from a mnemonic
func createMemoryKeyFromMnemonic(name, mnemonic, configDir string) (*keyring.Record, error) {
	kb, err := keyring.New(keyringAppName, keyring.BackendTest, configDir, nil, cdc)
	if err != nil {
		return nil, err
	}

	keyringAlgos, _ := kb.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
	if err != nil {
		return nil, err
	}

	return kb.NewAccount(name, mnemonic, "", sdk.FullFundraiserPath, algo)
}
