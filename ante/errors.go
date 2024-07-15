package ante

import (
	"fmt"
)

func ErrNeitherNativeDenom(coinDenom, denom string) error {
	return fmt.Errorf("neither of coin.Denom %s and denom %s is the native denom of the chain", coinDenom, denom)
}

func ErrDenomNotRegistered(denom string) error {
	return fmt.Errorf("denom %s not registered in host zone", denom)
}

func ErrExpectedOneCoin(count int) error {
	return fmt.Errorf("expected exactly one native coin, got %d", count)
}
