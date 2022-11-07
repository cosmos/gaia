package address

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/bech32"
)

// ConvertBech32Prefix convert bech32 address to specified prefix.
func ConvertBech32Prefix(address, prefix string) (string, error) {
	_, bz, err := bech32.DecodeAndConvert(address)
	if err != nil {
		return "", fmt.Errorf("cannot decode %s address: %s", address, err)
	}

	convertedAddress, err := bech32.ConvertAndEncode(prefix, bz)
	if err != nil {
		return "", fmt.Errorf("cannot convert %s address: %s", address, err)
	}

	return convertedAddress, nil
}
