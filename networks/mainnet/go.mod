module github.com/cosmos/mainnet

go 1.19

require (
	github.com/cosmos/cosmos-sdk v0.45.12 // indirect
	github.com/cosmos/gaia/v8 v8.0.0 // indirect
	github.com/tendermint/tendermint v0.34.24 // indirect
)

replace (
	// Use cosmos keyring
	github.com/99designs/keyring => github.com/cosmos/keyring v1.2.0

	// dgrijalva/jwt-go is deprecated and doesn't receive security updates.
	// TODO: remove it: https://github.com/cosmos/cosmos-sdk/issues/13134
	github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt/v4 v4.4.2

	// Fix upstream GHSA-h395-qcrw-5vmq vulnerability.
	// TODO Remove it: https://github.com/cosmos/cosmos-sdk/issues/10409
	github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.8.1

	// Use regen gogoproto fork
	// This for is replaced by cosmos/gogoproto in future versions
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

	github.com/jhump/protoreflect => github.com/jhump/protoreflect v1.9.0
	// use informal system fork of tendermint
	github.com/tendermint/tendermint => github.com/informalsystems/tendermint v0.34.24

	// latest grpc doesn't work with with our modified proto compiler, so we need to enforce
	// the following version across all dependencies.
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)
