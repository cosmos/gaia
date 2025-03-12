package e2e

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/gaia/v23/tests/e2e/common"
	"github.com/cosmos/gaia/v23/tests/e2e/query"
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
	chainEndpoint := fmt.Sprintf("http://%s", s.Resources.ValResources[s.Resources.ChainA.ID][0].GetHostPort("1317/tcp"))

	valIdx := 0
	val := s.Resources.ChainA.Validators[valIdx]
	address, _ := val.KeyInfo.GetAddress()
	sender := address.String()
	dirName, err := os.Getwd()
	s.Require().NoError(err)

	// Copy file to container path and store the contract
	entryPointSrc := filepath.Join(dirName, "data/skip_go_entry_point.wasm")
	entryPointDst := filepath.Join(val.ConfigDir(), "config", "skip_go_entry_point.wasm")
	_, err = common.CopyFile(entryPointSrc, entryPointDst)
	s.Require().NoError(err)
	entryPointPath := configFile("skip_go_entry_point.wasm")
	entryPointCode := s.StoreWasm(ctx, s.Resources.ChainA, valIdx, sender, entryPointPath)

	adapterSrc := filepath.Join(dirName, "data/skip_go_ibc_adapter_ibc_callbacks.wasm")
	adapterDst := filepath.Join(val.ConfigDir(), "config", "skip_go_ibc_adapter_ibc_callbacks.wasm")
	_, err = common.CopyFile(adapterSrc, adapterDst)
	s.Require().NoError(err)
	adapterPath := configFile("skip_go_ibc_adapter_ibc_callbacks.wasm")
	adapterCode := s.StoreWasm(ctx, s.Resources.ChainA, valIdx, sender, adapterPath)

	entrypointPredictedAddress := s.QueryBuildAddress(ctx, s.Resources.ChainA, valIdx, Sha256SkipEntryPoint, sender, SaltHex)
	s.Require().NoError(err)

	instantiateAdapterJSON := fmt.Sprintf(`{"entry_point_contract_address":"%s"}`, entrypointPredictedAddress)
	adapterAddress := s.tx.InstantiateWasm(ctx, s.commonHelper.Resources.ChainA, valIdx, sender, adapterCode, instantiateAdapterJSON, "adapter")
	s.Require().NoError(err)

	instantiateEntrypointJSON := fmt.Sprintf(`{"swap_venues":[], "ibc_transfer_contract_address": "%s"}`, adapterAddress)
	entrypointAddress := s.tx.Instantiate2Wasm(ctx, s.commonHelper.Resources.ChainA, valIdx, sender, entryPointCode, instantiateEntrypointJSON, SaltHex, "entrypoint")
	s.Require().Equal(entrypointPredictedAddress, entrypointAddress)
	s.Require().NoError(err)

	s.T().Logf("Successfully deployed contracts: \nEntrypoint: %s\nAdapter:%s\n", entrypointAddress, adapterAddress)

	str := "transfer/channel-0/uatom"
	h := sha256.New()
	h.Write([]byte(str))
	bs := h.Sum(nil)

	recipientDenom := fmt.Sprintf("ibc/%X", bs)

	memo := buildCallbacksMemo(entrypointAddress, recipientDenom, adapterAddress)

	senderB, _ := s.commonHelper.Resources.ChainB.Validators[0].KeyInfo.GetAddress()
	s.tx.SendIBC(s.commonHelper.Resources.ChainB, 0, senderB.String(), adapterAddress, "1uatom", "3000000uatom", memo, common.TransferChannel, nil, false)
	s.commonHelper.HermesClearPacket(common.HermesConfigWithGasPrices, s.commonHelper.Resources.ChainB.ID, common.TransferPort, common.TransferChannel)

	balances, err := query.AllBalances(chainEndpoint, RecipientAddress)
	if err != nil {
		return
	}

	require.Equal(s.T(), balances[0].String(), "1"+recipientDenom)
}

func buildCallbacksMemo(entrypointAddress string, recipientDenom string, adapterAddress string) string {
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
	return memo
}
