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

// createMemoryKey creates a key into keystore with a random mnemonic
func createMemoryKey() (info *keyring.Record, err error) {
	mnemonic, err := createMnemonic()
	if err != nil {
		return nil, err
	}

	account, err := createMemoryKeyFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// createMemoryKeyFromMnemonic creates a key into keystore from a mnemonic
func createMemoryKeyFromMnemonic(mnemonic string) (*keyring.Record, error) {
	kb, err := keyring.New(keyringAppName, keyring.BackendMemory, "", nil, cdc)
	if err != nil {
		return nil, err
	}

	keyringAlgos, _ := kb.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
	if err != nil {
		return nil, err
	}

	account, err := kb.NewAccount("", mnemonic, "", sdk.FullFundraiserPath, algo)
	if err != nil {
		return nil, err
	}

	return account, nil
}
