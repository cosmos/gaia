package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewQueryInterchainAccountRequest(t *testing.T) {
	var (
		connectionID = "connection-0"
		owner        = "cosmos1a6zlyvpnksx8wr6wz8wemur2xe8zyh0yxeh27a"
		got          = NewQueryInterchainAccountRequest(connectionID, owner)
	)
	require.Equal(t, connectionID, got.ConnectionId)
	require.Equal(t, owner, got.Owner)
}

func TestNewQueryInterchainAccountResponse(t *testing.T) {
	var (
		interchainAccAddr = "cosmos1a6zlyvpnksx8wr6wz8wemur2xe8zyh0yxeh27a"
		got               = NewQueryInterchainAccountResponse(interchainAccAddr)
	)
	require.Equal(t, interchainAccAddr, got.InterchainAccountAddress)
}
