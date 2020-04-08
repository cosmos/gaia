package types

import (
	"encoding/binary"

	bandoracle "github.com/bandprotocol/chain/x/oracle/types"
)

const (
	// ModuleName is the name of the module
	ModuleName = "consuming"
	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey is the message route for consuming
	RouterKey = ModuleName

	// QuerierRoute is the querier route for consuming
	QuerierRoute = ModuleName

	Version = "ics20-1"
)

var (
	// ResultStoreKeyPrefix is a prefix for storing result
	ResultStoreKeyPrefix = []byte{0xff}
)

// ResultStoreKey is a function to generate key for each result in store
func ResultStoreKey(requestID bandoracle.RequestID) []byte {
	return append(ResultStoreKeyPrefix, int64ToBytes(int64(requestID))...)
}

func int64ToBytes(num int64) []byte {
	result := make([]byte, 8)
	binary.BigEndian.PutUint64(result, uint64(num))
	return result
}
