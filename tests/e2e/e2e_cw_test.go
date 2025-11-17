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

	"github.com/cosmos/gaia/v26/tests/e2e/common"
	query2 "github.com/cosmos/gaia/v26/tests/e2e/query"
)

func (s *IntegrationTestSuite) testCWCounter() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	valIdx := 0
	val := s.Resources.ChainA.Validators[valIdx]
	address, _ := val.KeyInfo.GetAddress()
	sender := address.String()
	dirName, err := os.Getwd()
	s.Require().NoError(err)

	// Copy file to container path and store the contract
	src := filepath.Join(dirName, "data/counter.wasm")
	dst := filepath.Join(val.ConfigDir(), "config", "counter.wasm")
	_, err = common.CopyFile(src, dst)
	s.Require().NoError(err)
	storeWasmPath := configFile("counter.wasm")
	s.StoreWasm(ctx, s.Resources.ChainA, valIdx, sender, storeWasmPath)

	// Instantiate the contract
	contractAddr := s.InstantiateWasm(ctx, s.Resources.ChainA, valIdx, sender, strconv.Itoa(s.TestCounters.ContractsCounter), "{\"count\":0}", "counter")
	chainEndpoint := fmt.Sprintf("http://%s", s.Resources.ValResources[s.Resources.ChainA.ID][0].GetHostPort("1317/tcp"))

	// Execute the contract
	s.ExecuteWasm(ctx, s.Resources.ChainA, valIdx, sender, contractAddr, "{\"increment\":{}}")

	// Validate count increment
	query := map[string]interface{}{
		"get_count": map[string]interface{}{},
	}
	queryJSON, err := json.Marshal(query)
	s.Require().NoError(err)
	queryMsg := base64.StdEncoding.EncodeToString(queryJSON)
	data, err := query2.WasmSmartContractState(chainEndpoint, contractAddr, queryMsg)
	s.Require().NoError(err)
	var counterResp map[string]int
	err = json.Unmarshal(data, &counterResp)
	s.Require().NoError(err)
	s.Require().Equal(1, counterResp["count"])
}
