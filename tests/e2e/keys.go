package e2e

import (
	"os/exec"
	
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
)

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

func createMemoryKey() (mnemonic string, info *keyring.Record, err error) {
	mnemonic, err = createMnemonic()
	if err != nil {
		return "", nil, err
	}

	account, err := createMemoryKeyFromMnemonic(mnemonic)
	if err != nil {
		return "", nil, err
	}

	return mnemonic, account, nil
}

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

// createRandomAccount create a random account into key store and return the address
func createRandomAccount(configDir, name string) (string, error) {
	kb, err := keyring.New(keyringAppName, keyring.BackendTest, configDir, nil, cdc)
	if err != nil {
		return "", err
	}

	keyringAlgos, _ := kb.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
	if err != nil {
		return "", err
	}

	mnemonic, err := createMnemonic()
	if err != nil {
		return "", err
	}

	account, err := kb.NewAccount(name, mnemonic, "", sdk.FullFundraiserPath, algo)
	if err != nil {
		return "", err
	}
	accAddr, err := account.GetAddress()
	if err != nil {
		return "", err
	}

	// TODO find a better way to add accounts on demand without giving folder permissions every time
	return accAddr.String(), exec.Command("chmod", "-R", "0777", configDir).Run()
}
