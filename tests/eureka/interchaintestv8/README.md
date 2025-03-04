# End to End Testing Suite with Interchaintest

The e2e tests are built using the [interchaintest](https://github.com/strangelove-ventures/interchaintest) library by Strangelove. It runs multiple docker container validators, and lets you test IBC enabled smart contracts.

These end to end tests are designed to run in the ci, but you can also run them locally.

## Running the tests locally

To run the tests locally, run the following commands from this directory:

```text
go test -v . -run=$TEST_SUITE_FN/$TEST_NAME
```

where `$TEST_NAME` is one of the test names of the `$TEST_SUITE_FN`. For example, to run the `TestDeploy` test, you would run:

```text
go test -v . -run=TestWithIbcEurekaTestSuite/TestDeploy_Groth16
```
