module github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v11

go 1.25.9

require (
	cosmossdk.io/api v1.0.0
	cosmossdk.io/core v1.1.0
	cosmossdk.io/errors v1.1.0
	cosmossdk.io/log/v2 v2.1.0
	cosmossdk.io/math v1.5.3
	github.com/cometbft/cometbft v0.39.1
	github.com/cosmos/cosmos-db v1.1.3
	github.com/cosmos/cosmos-sdk v0.54.2
	github.com/cosmos/gogoproto v1.7.2
	github.com/cosmos/ibc-go/v11 v11.0.0
	github.com/golang/mock v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/hashicorp/go-metrics v0.5.4
	github.com/iancoleman/orderedmap v0.3.0
	github.com/spf13/cast v1.10.0
	github.com/spf13/cobra v1.10.2
	github.com/spf13/viper v1.21.0
	github.com/stretchr/testify v1.11.1
)

replace (
	github.com/99designs/keyring => github.com/cosmos/keyring v1.2.0
	github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt/v4 v4.4.2
	github.com/syndtr/goleveldb => github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
)
