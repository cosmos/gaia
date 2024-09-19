package chainsuite

import (
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
)

func StrToSDKInt(s string) (sdkmath.Int, error) {
	s, _, _ = strings.Cut(s, ".")
	i, ok := sdkmath.NewIntFromString(s)
	if !ok {
		return sdkmath.Int{}, fmt.Errorf("s: %s", s)
	}
	return i, nil
}
