set dotenv-load

# Build the contracts using `forge build`
build-contracts: clean
	forge build

# Build the relayer using `cargo build`
build-relayer:
	cargo build --bin relayer --release --locked

# Build the operator using `cargo build`
build-operator:
	cargo build --bin operator --release --locked

# Build riscv elf files using `~/.sp1/bin/cargo-prove`
build-sp1-programs:
  @echo "Building SP1 programs in 'target/elf-compilation/riscv32im-succinct-zkvm-elf/release/'"
  ~/.sp1/bin/cargo-prove prove build -p sp1-ics07-tendermint-update-client --locked
  ~/.sp1/bin/cargo-prove prove build -p sp1-ics07-tendermint-membership --locked
  ~/.sp1/bin/cargo-prove prove build -p sp1-ics07-tendermint-uc-and-membership --locked
  ~/.sp1/bin/cargo-prove prove build -p sp1-ics07-tendermint-misbehaviour --locked

# Build and optimize the eth wasm light client using `cosmwasm/optimizer`. Requires `docker` and `gzip`
build-cw-ics08-wasm-eth:
	docker run --rm -v "$(pwd)":/code --mount type=volume,source="$(basename "$(pwd)")_cache",target=/target --mount type=volume,source=registry_cache,target=/usr/local/cargo/registry cosmwasm/optimizer:0.16.1 ./programs/cw-ics08-wasm-eth
	cp artifacts/cw_ics08_wasm_eth.wasm e2e/interchaintestv8/wasm 
	gzip e2e/interchaintestv8/wasm/cw_ics08_wasm_eth.wasm -f

# Build the relayer docker image
# Only for linux/amd64 since sp1 doesn't have an arm image built
build-relayer-image:
    docker build -t eureka-relayer:latest --platform linux/amd64 .

# Clean up the cache and out directories
clean:
	@echo "Cleaning up cache and out directories"
	-rm -rf cache out broadcast # ignore errors

# Run the foundry tests
test-foundry testname=".\\*":
	forge test -vvv --show-progress --match-test ^{{testname}}\(.\*\)\$

# Run the benchmark tests
# Run with `just test-benchmark Plonk"` to run only Plonk benchmarks
# Run with `just test-benchmark Groth16"` to run only Groth16 benchmarks
test-benchmark testname=".\\*":
	forge test -vvv --show-progress --gas-report --match-path test/solidity-ibc/BenchmarkTest.t.sol --match-test {{testname}}

# Run the cargo tests
test-cargo:
	cargo test --all --locked

# Run the tests in abigen
test-abigen:
	@echo "Running abigen tests..."
	cd abigen && go test -v ./...

# Run forge fmt and bun solhint
lint:
	@echo "Linting the Solidity code..."
	forge fmt --check && bun solhint -w 0 '{scripts,contracts,test}/**/*.sol'
	@echo "Linting the Go code..."
	cd e2e/interchaintestv8 && golangci-lint run
	cd abigen && golangci-lint run
	@echo "Linting the Rust code..."
	cargo fmt --all -- --check && cargo clippy --all-targets --all-features -- -D warnings
	@echo "Linting the Protobuf files..."
	buf lint

# Generate the ABI files for the contracts
generate-abi: build-contracts
	jq '.abi' out/ICS26Router.sol/ICS26Router.json > abi/ICS26Router.json
	jq '.abi' out/ICS20Transfer.sol/ICS20Transfer.json > abi/ICS20Transfer.json
	jq '.abi' ./out/SP1ICS07Tendermint.sol/SP1ICS07Tendermint.json > abi/SP1ICS07Tendermint.json
	jq '.abi' out/ERC20.sol/ERC20.json > abi/ERC20.json
	jq '.abi' out/IBCERC20.sol/IBCERC20.json > abi/IBCERC20.json
	abigen --abi abi/ERC20.json --pkg erc20 --type Contract --out e2e/interchaintestv8/types/erc20/contract.go
	abigen --abi abi/SP1ICS07Tendermint.json --pkg sp1ics07tendermint --type Contract --out abigen/sp1ics07tendermint/contract.go
	abigen --abi abi/ICS20Transfer.json --pkg ics20transfer --type Contract --out abigen/ics20transfer/contract.go
	abigen --abi abi/ICS26Router.json --pkg ics26router --type Contract --out abigen/ics26router/contract.go
	abigen --abi abi/IBCERC20.json --pkg ibcerc20 --type Contract --out abigen/ibcerc20/contract.go

# Generate go types for the e2e tests from the etheruem light client code
generate-ethereum-types:
	cargo run --bin generate_json_schema --features test-utils
	quicktype --src-lang schema --lang go --just-types-and-package --package ethereum --src ethereum_types_schema.json --out e2e/interchaintestv8/types/ethereum/types.gen.go --top-level GeneratedTypes
	rm ethereum_types_schema.json
	sed -i.bak 's/int64/uint64/g' e2e/interchaintestv8/types/ethereum/types.gen.go # quicktype generates int64 instead of uint64 :(
	rm -f e2e/interchaintestv8/types/ethereum/types.gen.go.bak # this is to be linux and mac compatible (coming from the sed command)
	cd e2e/interchaintestv8 && golangci-lint run --fix types/ethereum/types.gen.go

# Run the e2e tests
# Run any e2e test in the interchaintestv8 test suite using the test's full name
# For example, `just test-e2e TestWithIbcEurekaTestSuite/TestDeploy_Groth16`
test-e2e testname: clean
	@echo "Running {{testname}} test..."
	cd tests/eureka/interchaintestv8 && go test -v -run '^{{testname}}$' -timeout 40m

# Run any e2e test in the IbcEurekaTestSuite using the test's name
# For example, `just test-e2e-eureka TestDeploy_Groth16`
test-e2e-eureka testname: clean
	@echo "Running {{testname}} test..."
	just test-e2e TestWithIbcEurekaTestSuite/{{testname}}

# Run any e2e test in the RelayerTestSuite using the test's name
# For example, `just test-e2e-relayer TestRelayerInfo`
test-e2e-relayer testname: clean
	@echo "Running {{testname}} test..."
	just test-e2e TestWithRelayerTestSuite/{{testname}}

# Run any e2e test in the CosmosRelayerTestSuite using the test's name
# For example, `just test-e2e-cosmos-relayer TestRelayerInfo`
test-e2e-cosmos-relayer testname: clean
	@echo "Running {{testname}} test..."
	just test-e2e TestWithCosmosRelayerTestSuite/{{testname}}

# Run anu e2e test in the SP1ICS07TendermintTestSuite using the test's name
# For example, `just test-e2e-sp1-ics07 TestDeploy_Groth16`
test-e2e-sp1-ics07 testname: clean
	@echo "Running {{testname}} test..."
	just test-e2e TestWithSP1ICS07TendermintTestSuite/{{testname}}

# Run any e2e test in the MultichainTestSuite using the test's name
# For example, `just test-e2e-multichain TestDeploy_Groth16`
test-e2e-multichain testname: clean
	@echo "Running {{testname}} test..."
	just test-e2e TestWithMultichainTestSuite/{{testname}}

# Install the sp1-ics07-tendermint operator for use in the e2e tests
install-operator:
	cargo install --bin operator --path programs/operator --locked

# Install the relayer using `cargo install`
install-relayer:
	cargo install --bin relayer --path programs/relayer --locked

# Generate the `genesis.json` file using $TENDERMINT_RPC_URL in the `.env` file
# Note that the `scripts/genesis.json` file is ignored in the `.gitignore` file
genesis-sp1-ics07: build-sp1-programs
  @echo "Generating the genesis file..."
  RUST_LOG=info cargo run --bin operator --release -- genesis -o scripts/genesis.json

# Deploy the SP1ICS07Tendermint contract to the Eth Sepolia testnet if the `.env` file is present
deploy-sp1-ics07: genesis-sp1-ics07
  @echo "Deploying the SP1ICS07Tendermint contract"
  forge install
  forge script scripts/SP1ICS07Tendermint.s.sol --rpc-url $RPC_URL --private-key $PRIVATE_KEY --broadcast

# Generate the fixtures for the Solidity tests using the e2e tests
generate-fixtures-solidity: clean install-operator install-relayer
	@echo "Generating fixtures... This may take a while."
	@echo "Generating recvPacket and acknowledgePacket groth16 fixtures..."
	cd e2e/interchaintestv8 && GENERATE_SOLIDITY_FIXTURES=true SP1_PROVER=network go test -v -run '^TestWithIbcEurekaTestSuite/TestICS20TransferERC20TokenfromEthereumToCosmosAndBack_Groth16$' -timeout 40m
	@echo "Generating recvPacket and acknowledgePacket plonk fixtures..."
	cd e2e/interchaintestv8 && GENERATE_SOLIDITY_FIXTURES=true SP1_PROVER=network go test -v -run '^TestWithIbcEurekaTestSuite/TestICS20TransferERC20TokenfromEthereumToCosmosAndBack_Plonk$' -timeout 40m
	@echo "Generating recvPacket and acknowledgePacket groth16 fixtures for 25 packets..."
	cd e2e/interchaintestv8 && GENERATE_SOLIDITY_FIXTURES=true SP1_PROVER=network go test -v -run '^TestWithIbcEurekaTestSuite/Test_25_ICS20TransferERC20TokenfromEthereumToCosmosAndBack_Groth16$' -timeout 40m
	@echo "Generating recvPacket and acknowledgePacket groth16 fixtures for 50 packets..."
	cd e2e/interchaintestv8 && GENERATE_SOLIDITY_FIXTURES=true SP1_PROVER=network go test -v -run '^TestWithIbcEurekaTestSuite/Test_50_ICS20TransferERC20TokenfromEthereumToCosmosAndBack_Groth16$' -timeout 40m
	@echo "Generating recvPacket and acknowledgePacket plonk fixtures for 50 packets..."
	cd e2e/interchaintestv8 && GENERATE_SOLIDITY_FIXTURES=true SP1_PROVER=network go test -v -run '^TestWithIbcEurekaTestSuite/Test_50_ICS20TransferERC20TokenfromEthereumToCosmosAndBack_Plonk$' -timeout 40m
	@echo "Generating native SdkCoin recvPacket groth16 fixtures..."
	cd e2e/interchaintestv8 && GENERATE_SOLIDITY_FIXTURES=true SP1_PROVER=network go test -v -run '^TestWithIbcEurekaTestSuite/TestICS20TransferNativeCosmosCoinsToEthereumAndBack_Groth16$' -timeout 40m
	@echo "Generating native SdkCoin recvPacket plonk fixtures..."
	cd e2e/interchaintestv8 && GENERATE_SOLIDITY_FIXTURES=true SP1_PROVER=network go test -v -run '^TestWithIbcEurekaTestSuite/TestICS20TransferNativeCosmosCoinsToEthereumAndBack_Plonk$' -timeout 40m
	@echo "Generating timeoutPacket groth16 fixtures..."
	cd e2e/interchaintestv8 && GENERATE_SOLIDITY_FIXTURES=true SP1_PROVER=network go test -v -run '^TestWithIbcEurekaTestSuite/TestTimeoutPacketFromEth_Groth16$' -timeout 40m
	@echo "Generating timeoutPacket plonk fixtures..."
	cd e2e/interchaintestv8 && GENERATE_SOLIDITY_FIXTURES=true SP1_PROVER=network go test -v -run '^TestWithIbcEurekaTestSuite/TestTimeoutPacketFromEth_Plonk$' -timeout 40m

# Generate the fixture files for the Celestia Mocha testnet using the prover parameter.
# The prover parameter should be one of: ["mock", "network", "local"]
# This generates the fixtures for all programs in parallel using GNU parallel.
# If prover is set to network, this command requires the `NETWORK_PRIVATE_KEY` environment variable to be set.
generate-fixtures-sp1-ics07: install-operator
  @echo "Generating fixtures... This may take a while (up to 20 minutes)"
  TENDERMINT_RPC_URL="${TENDERMINT_RPC_URL%/}" && \
  CURRENT_HEIGHT=$(curl "$TENDERMINT_RPC_URL"/block | jq -r ".result.block.header.height") && \
  TRUSTED_HEIGHT=$(($CURRENT_HEIGHT-100)) && \
  TARGET_HEIGHT=$(($CURRENT_HEIGHT-10)) && \
  echo "For celestia fixtures, trusted block: $TRUSTED_HEIGHT, target block: $TARGET_HEIGHT, from $TENDERMINT_RPC_URL" && \
  parallel --progress --shebang --ungroup -j 6 ::: \
    "RUST_LOG=info SP1_PROVER=network ./target/release/operator fixtures update-client --trusted-block $TRUSTED_HEIGHT --target-block $TARGET_HEIGHT -o 'test/sp1-ics07/fixtures/update_client_fixture-plonk.json'" \
    "sleep 20 && RUST_LOG=info SP1_PROVER=network ./target/release/operator fixtures update-client --trusted-block $TRUSTED_HEIGHT --target-block $TARGET_HEIGHT -p groth16 -o 'test/sp1-ics07/fixtures/update_client_fixture-groth16.json'" \
    "sleep 40 && RUST_LOG=info SP1_PROVER=network ./target/release/operator fixtures update-client-and-membership --key-paths clients/07-tendermint-0/clientState,clients/07-tendermint-001/clientState --trusted-block $TRUSTED_HEIGHT --target-block $TARGET_HEIGHT -o 'test/sp1-ics07/fixtures/uc_and_memberships_fixture-plonk.json'" \
    "sleep 60 && RUST_LOG=info SP1_PROVER=network ./target/release/operator fixtures update-client-and-membership --key-paths clients/07-tendermint-0/clientState,clients/07-tendermint-001/clientState --trusted-block $TRUSTED_HEIGHT --target-block $TARGET_HEIGHT -p groth16 -o 'test/sp1-ics07/fixtures/uc_and_memberships_fixture-groth16.json'" \
    "sleep 80 && RUST_LOG=info SP1_PROVER=network ./target/release/operator fixtures membership --key-paths clients/07-tendermint-0/clientState,clients/07-tendermint-001/clientState --trusted-block $TRUSTED_HEIGHT -o 'test/sp1-ics07/fixtures/memberships_fixture-plonk.json'" \
    "sleep 100 && RUST_LOG=info SP1_PROVER=network ./target/release/operator fixtures membership --key-paths clients/07-tendermint-0/clientState,clients/07-tendermint-001/clientState --trusted-block $TRUSTED_HEIGHT -p groth16 -o 'test/sp1-ics07/fixtures/memberships_fixture-groth16.json'"
  cd e2e/interchaintestv8 && RUST_LOG=info SP1_PROVER=network GENERATE_SOLIDITY_FIXTURES=true go test -v -run '^TestWithSP1ICS07TendermintTestSuite/TestDoubleSignMisbehaviour_Plonk$' -timeout 40m
  cd e2e/interchaintestv8 && RUST_LOG=info SP1_PROVER=network GENERATE_SOLIDITY_FIXTURES=true go test -v -run '^TestWithSP1ICS07TendermintTestSuite/TestBreakingTimeMonotonicityMisbehaviour_Groth16' -timeout 40m
  cd e2e/interchaintestv8 && RUST_LOG=info SP1_PROVER=network GENERATE_SOLIDITY_FIXTURES=true go test -v -run '^TestWithSP1ICS07TendermintTestSuite/Test100Membership_Groth16' -timeout 40m
  cd e2e/interchaintestv8 && RUST_LOG=info SP1_PROVER=network GENERATE_SOLIDITY_FIXTURES=true go test -v -run '^TestWithSP1ICS07TendermintTestSuite/Test25Membership_Plonk' -timeout 40m
  @echo "Fixtures generated at 'test/sp1-ics07/fixtures'"

# Generate the relayer proto files
relayer-proto-gen:
    @echo "Generating Protobuf files for relayer"
    buf generate --template buf.gen.yaml
