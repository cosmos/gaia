module github.com/cosmos/gaia/v6

go 1.16

require (
	github.com/aws/aws-sdk-go v1.27.0
	github.com/cosmos/cosmos-sdk v0.43.0
	github.com/cosmos/ibc-go v1.0.0
	github.com/gorilla/mux v1.8.0
	github.com/pkg/errors v0.9.1
	github.com/rakyll/statik v0.1.7
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
    github.com/tendermint/liquidity v1.2.6-0.20210513094606-6cd272e3814d // indirect
	github.com/tendermint/tendermint v0.34.12
	github.com/tendermint/tm-db v0.6.4
)

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/tendermint/tendermint => github.com/tendermint/tendermint v0.34.12

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
