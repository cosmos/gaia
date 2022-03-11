// package e2e defines an integration testing suite used for full end-to-end
// testing functionality.
//
// The file e2e_suite_test.go defines the testing suite and contains the core
// bootrapping logic that creates a testing environment via Docker containers.
// A testing network is created dynamically and contains multiple Docker
// containers:
//
// 1. Two independent Gaia networks
// 3. A hermes relayer connecting the two Gaia networks over IBC
//
// The file e2e_test.go contains the actual end-to-end integration tests that
// utilize the testing suite.
package e2e
