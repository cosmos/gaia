package gaia

import (
	"runtime/debug"
	"testing"

	"github.com/google/gofuzz"
)

func TestFuzzGenesisAccountValidate(t *testing.T) {
	if testing.Short() {
		t.Skip("running in -short mode")
	}

	t.Parallel()

	acct := new(SimGenesisAccount)
	i := 0
	defer func() {
		r := recover()
		if r == nil {
			return
		}

		// Otherwise report on the configuration and iteration.
		t.Fatalf("Failed SimGenesisAccount on iteration #%d: %#v\n\n%s\n\n%s", i, acct, r, debug.Stack())
	}()

	f := fuzz.New()
	for i = 0; i < 1e5; i++ {
		acct = new(SimGenesisAccount)
		f.Fuzz(acct)
		acct.Validate() //nolint:errcheck
	}
}
