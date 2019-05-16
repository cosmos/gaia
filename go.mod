module github.com/cosmos/gaia

go 1.12

require (
	github.com/cosmos/cosmos-sdk v0.34.4
	github.com/otiai10/copy v1.0.1
	github.com/rakyll/statik v0.1.6
	github.com/spf13/cobra v0.0.3
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.3.0
	github.com/tendermint/go-amino v0.15.0
	github.com/tendermint/tendermint v0.31.5
)

replace github.com/cosmos/cosmos-sdk => /home/alessio/work/cosmos-sdk

replace golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5
