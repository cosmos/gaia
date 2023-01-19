package types

import "crypto/md5"

const (
	// ModuleName is the name of the module
	ModuleName = "microtx"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey is the module name router key
	RouterKey = ModuleName

	// QuerierRoute to be used for querierer msgs
	QuerierRoute = ModuleName
)

var (
	// PrefixKey is an example prefix for store keys, items under this key would have keys like
	// Prefix00001, PrefixAxE10034ADF0018547, serialized to bytes. This is great for implicit
	// ordering or nesting like "Prefix[Sub-prefix][Item-identifier]"
	PrefixKey = HashString("Prefix")
)

// GetPrefixKey is an example function to return items under the key PrefixKey
// e.g. it would return Prefix00001 like in the comment above
func GetPrefixKey(subPrefix string) []byte {
	return AppendBytes(PrefixKey, []byte(subPrefix))
}

// Hashing string using cryptographic MD5 function
// returns 128bit(16byte) value
func HashString(input string) []byte {
	md5 := md5.New()
	md5.Write([]byte(input))
	return md5.Sum(nil)
}

func AppendBytes(args ...[]byte) []byte {
	length := 0
	for _, v := range args {
		length += len(v)
	}

	res := make([]byte, length)

	length = 0
	for _, v := range args {
		copy(res[length:length+len(v)], v)
		length += len(v)
	}

	return res
}
