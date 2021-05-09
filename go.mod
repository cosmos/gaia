module github.com/cosmos/gaia/v4

go 1.15

require (
	github.com/armon/go-metrics v0.3.7 // indirect
	github.com/cosmos/cosmos-sdk v0.43.0-alpha1.0.20210509051442-abd86777da6a
	github.com/cosmos/iavl v0.16.0 // indirect
	github.com/cosmos/ibc-go v1.0.0-alpha1
	github.com/gorilla/mux v1.8.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/common v0.23.0 // indirect
	github.com/rakyll/statik v0.1.7
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.10
	github.com/tendermint/tm-db v0.6.4
)

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
