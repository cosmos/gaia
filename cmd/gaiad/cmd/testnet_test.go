package cmd

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnsureTestnetDirDoesNotExist(t *testing.T) {
	testCases := []struct {
		name        string
		baseDir     func(t *testing.T) string
		errContains string
	}{
		{
			name: "missing directory",
			baseDir: func(t *testing.T) string {
				t.Helper()
				return filepath.Join(t.TempDir(), "missing")
			},
		},
		{
			name: "existing directory",
			baseDir: func(t *testing.T) string {
				t.Helper()
				return t.TempDir()
			},
			errContains: "testnets directory already exists for chain-id 'test-chain'",
		},
		{
			name: "stat error",
			baseDir: func(t *testing.T) string {
				t.Helper()
				return filepath.Join(t.TempDir(), "invalid\x00path")
			},
			errContains: "failed to check testnet directory",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ensureTestnetDirDoesNotExist(tc.baseDir(t), "test-chain")
			if tc.errContains == "" {
				require.NoError(t, err)
				return
			}

			require.ErrorContains(t, err, tc.errContains)
		})
	}
}
