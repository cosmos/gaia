package simulation

import (
	"github.com/cosmos/cosmos-sdk/types/kv"
)

type ResultUnmarshaler interface{}

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding DenomTrace type.
func NewDecodeStore(cdc ResultUnmarshaler) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		return "TODO"
		// switch {
		// case bytes.Equal(kvA.Key[:1], types.PortKey):
		// 	return fmt.Sprintf("Port A: %s\nPort B: %s", string(kvA.Value), string(kvB.Value))

		// case bytes.Equal(kvA.Key[:1], types.DenomTraceKey):
		// 	denomTraceA := cdc.MustUnmarshalDenomTrace(kvA.Value)
		// 	denomTraceB := cdc.MustUnmarshalDenomTrace(kvB.Value)
		// 	return fmt.Sprintf("DenomTrace A: %s\nDenomTrace B: %s", denomTraceA.IBCDenom(), denomTraceB.IBCDenom())

		// default:
		// 	panic(fmt.Sprintf("invalid %s key prefix %X", types.ModuleName, kvA.Key[:1]))
		// }
	}
}
