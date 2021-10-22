module github.com/cosmos/gaia/v6

go 1.17

require (
	github.com/cosmos/cosmos-sdk v0.44.2
	github.com/cosmos/ibc-go v1.2.1
	github.com/gorilla/mux v1.8.0
	github.com/gravity-devs/liquidity v1.4.0
	github.com/rakyll/statik v0.1.7
	github.com/spf13/cast v1.4.1
	github.com/spf13/cobra v1.2.1
	github.com/strangelove-ventures/packet-forward-middleware v0.0.0-20211012183028-c3c7e62b2d93
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.13
	github.com/tendermint/tm-db v0.6.4
)

replace (
	github.com/cosmos/cosmos-sdk => github.com/cosmos/cosmos-sdk v0.44.2
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)
