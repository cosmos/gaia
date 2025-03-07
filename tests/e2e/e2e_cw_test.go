package e2e

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func (s *IntegrationTestSuite) testCWCounter() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	valIdx := 0
	val := s.chainA.validators[valIdx]
	address, _ := val.keyInfo.GetAddress()
	sender := address.String()
	dirName, err := os.Getwd()
	s.Require().NoError(err)

	// Copy file to container path and store the contract
	src := filepath.Join(dirName, "data/counter.wasm")
	dst := filepath.Join(val.configDir(), "config", "counter.wasm")
	_, err = copyFile(src, dst)
	s.Require().NoError(err)
	storeWasmPath := configFile("counter.wasm")
	s.storeWasm(ctx, s.chainA, valIdx, sender, storeWasmPath)

	// Instantiate the contract
	contractAddr := s.instantiateWasm(ctx, s.chainA, valIdx, sender, strconv.Itoa(s.testCounters.contractsCounter), "{\"count\":0}", "counter")
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

	// Execute the contract
	s.executeWasm(ctx, s.chainA, valIdx, sender, contractAddr, "{\"increment\":{}}")

	// Validate count increment
	query := map[string]interface{}{
		"get_count": map[string]interface{}{},
	}
	queryJSON, err := json.Marshal(query)
	s.Require().NoError(err)
	queryMsg := base64.StdEncoding.EncodeToString(queryJSON)
	data, err := queryWasmSmartContractState(chainEndpoint, contractAddr, queryMsg)
	s.Require().NoError(err)
	var counterResp map[string]int
	err = json.Unmarshal(data, &counterResp)
	s.Require().NoError(err)
	s.Require().Equal(1, counterResp["count"])
}
