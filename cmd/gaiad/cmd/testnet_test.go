package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCalculateIP(t *testing.T) {
	testCases := []struct {
		name     string
		startIP  string
		offset   int
		expected string
	}{
		{
			name:     "zero offset returns the starting address",
			startIP:  "192.168.0.1",
			offset:   0,
			expected: "192.168.0.1",
		},
		{
			name:     "simple increment within the last octet",
			startIP:  "192.168.0.1",
			offset:   5,
			expected: "192.168.0.6",
		},
		{
			name:     "carries into the third octet when the last octet overflows",
			startIP:  "192.168.0.250",
			offset:   10,
			expected: "192.168.1.4",
		},
		{
			name:     "carries when starting exactly on the boundary",
			startIP:  "10.0.0.255",
			offset:   1,
			expected: "10.0.1.0",
		},
		{
			name:     "carries across multiple octets",
			startIP:  "10.0.255.255",
			offset:   1,
			expected: "10.1.0.0",
		},
		{
			name:     "generates distinct IPs for a large validator set",
			startIP:  "192.168.0.200",
			offset:   100,
			expected: "192.168.1.44",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := calculateIP(tc.startIP, tc.offset)
			require.NoError(t, err)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestCalculateIPRejectsNonIPv4(t *testing.T) {
	_, err := calculateIP("not-an-ip", 1)
	require.Error(t, err)
}

// TestCalculateIPNoCollisions guards against the regression where bumping only
// the final octet wrapped back to 0, handing two validators the same IP.
func TestCalculateIPNoCollisions(t *testing.T) {
	const startIP = "192.168.0.200"
	seen := make(map[string]int, 300)

	for i := 0; i < 300; i++ {
		ip, err := calculateIP(startIP, i)
		require.NoError(t, err)
		if prev, ok := seen[ip]; ok {
			t.Fatalf("duplicate IP %s generated for offsets %d and %d", ip, prev, i)
		}
		seen[ip] = i
	}
}
