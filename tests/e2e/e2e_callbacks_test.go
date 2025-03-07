package e2e

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/stretchr/testify/assert/yaml"
	"github.com/stretchr/testify/require"
)

const (
	Sha256SkipEntryPoint          = "4ee07a1474cb1429cfbdba98fb52ca2efc2fe8602f8e1978dbc3f45b71511ca9"
	Sha256SkipAdapterIBCCallbacks = "21c375f75e09197478cd345b0a6376824a75471d8e22577dc36b74739277f027"
	SaltHex                       = "74657374696e67" // "testing" hex encoded
	RecipientAddress              = "cosmos1hrgj37s5dcqrte6srj9p2uqul3nxpmmqfhqp67"
)

func (s *IntegrationTestSuite) testCallbacksCWSkipGo() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

	valIdx := 0
	val := s.chainA.validators[valIdx]
	address, _ := val.keyInfo.GetAddress()
	sender := address.String()
	dirName, err := os.Getwd()
	s.Require().NoError(err)

	// Copy file to container path and store the contract
	entryPointSrc := filepath.Join(dirName, "data/skip_go_entry_point.wasm")
	entryPointDst := filepath.Join(val.configDir(), "config", "skip_go_entry_point.wasm")
	_, err = copyFile(entryPointSrc, entryPointDst)
	s.Require().NoError(err)
	storeWasmPath := configFile("skip_go_entry_point.wasm")
	s.storeWasm(ctx, s.chainA, valIdx, sender, storeWasmPath)

	adapterSrc := filepath.Join(dirName, "data/skip_go_ibc_adapter_ibc_callbacks.wasm")
	adapterDst := filepath.Join(val.configDir(), "config", "skip_go_ibc_adapter_ibc_callbacks.wasm")
	_, err = copyFile(adapterSrc, adapterDst)
	s.Require().NoError(err)
	storeWasmPath = configFile("skip_go_ibc_adapter_ibc_callbacks.wasm")
	s.storeWasm(ctx, s.chainA, valIdx, sender, storeWasmPath)

	entrypointPredictedAddress := s.queryBuildAddress(ctx, s.chainA, valIdx, Sha256SkipEntryPoint, sender, SaltHex)
	s.Require().NoError(err)

	instantiateAdapterJSON := fmt.Sprintf(`{"entry_point_contract_address":"%s"}`, entrypointPredictedAddress)
	s.instantiateWasm(ctx, s.chainA, valIdx, sender, "3", instantiateAdapterJSON, "adapter")
	adapterAddress, err := queryWasmContractAddress(chainEndpoint, address.String(), 1)
	s.Require().NoError(err)

	instantiateEntrypointJSON := fmt.Sprintf(`{"swap_venues":[], "ibc_transfer_contract_address": "%s"}`, adapterAddress)
	s.instantiate2Wasm(ctx, s.chainA, valIdx, sender, "2", instantiateEntrypointJSON, SaltHex, "entrypoint")
	entrypointAddress, err := queryWasmContractAddress(chainEndpoint, address.String(), 2)
	s.Require().Equal(entrypointPredictedAddress, entrypointAddress)
	s.Require().NoError(err)

	s.T().Logf("Successfully deployed contracts: \nEntrypoint: %s\nAdapter:%s\n", entrypointAddress, adapterAddress)

	str := "transfer/channel-0/uatom"
	h := sha256.New()
	h.Write([]byte(str))
	bs := h.Sum(nil)

	recipientDenom := fmt.Sprintf("ibc/%X", bs)

	ibcHooksData := fmt.Sprintf(`"wasm": {
						"contract": "%s",
						"msg": {
						  "action": {
							"sent_asset": {
							  "native": {
								"denom":"%s",
								"amount":"1"
							  }
							},
							"exact_out": false,
							"timeout_timestamp": %d,
							"action": {
							  "transfer":{
								"to_address": "%s"
							  }
							}
						  }
						}
					  }`, entrypointAddress, recipientDenom, time.Now().Add(time.Minute).UnixNano(), RecipientAddress)
	destCallbackData := fmt.Sprintf(`"dest_callback": {
					"address": "%s",
					"gas_limit": "%d"
				  }`, adapterAddress, 10_000_000)

	memo := fmt.Sprintf("{%s,%s}", destCallbackData, ibcHooksData)

	senderB, _ := s.chainB.validators[0].keyInfo.GetAddress()
	s.sendIBC(s.chainB, 0, senderB.String(), adapterAddress, "1uatom", "3000000uatom", memo, transferChannel, nil, false)
	s.hermesClearPacket(hermesConfigWithGasPrices, s.chainB.id, transferPort, transferChannel)

	balances, err := queryGaiaAllBalances(chainEndpoint, RecipientAddress)
	if err != nil {
		return
	}

	require.Equal(s.T(), balances[0].String(), "1"+recipientDenom)
}

func (s *IntegrationTestSuite) queryBuildAddress(ctx context.Context, c *chain, valIdx int, codeHash, creatorAddress, saltHexEncoded string,
) (res string) {
	cmd := []string{
		gaiadBinary,
		queryCommand,
		"wasm",
		"build-address",
		codeHash,
		creatorAddress,
		saltHexEncoded,
	}

	s.executeGaiaTxCommand(ctx, c, cmd, valIdx, func(stdOut []byte, stdErr []byte) bool {
		s.Require().NoError(yaml.Unmarshal(stdOut, &res))
		return true
	})
	return res
}
