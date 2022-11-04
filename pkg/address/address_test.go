package address

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertBech32Prefix(t *testing.T) {
	cases := []struct {
		name      string
		address   string
		prefix    string
		converted string
		err       error
	}{
		{
			name:      "Convert valid bech 32 address",
			address:   "akash1a6zlyvpnksx8wr6wz8wemur2xe8zyh0ytz6d88",
			converted: "cosmos1a6zlyvpnksx8wr6wz8wemur2xe8zyh0yxeh27a",
			prefix:    "cosmos",
		},
		{
			name:    "Convert invalid address",
			address: "invalidaddress",
			prefix:  "cosmos",
			err:     errors.New("cannot decode invalidaddress address: decoding bech32 failed: invalid separator index -1"),
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			convertedAddress, err := ConvertBech32Prefix(tt.address, tt.prefix)
			if tt.err != nil {
				require.ErrorContains(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.converted, convertedAddress)
		})
	}
}
