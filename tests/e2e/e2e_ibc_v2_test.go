package e2e

import (
	"fmt"
	"github.com/cosmos/gaia/v23/tests/e2e/common"
	"github.com/cosmos/gaia/v23/tests/e2e/msg"
	"github.com/cosmos/gaia/v23/tests/e2e/query"
	"path/filepath"
	"time"
)

func (s *IntegrationTestSuite) TestV2RecvPacket() {
	chain := s.commonHelper.Resources.ChainA

	submitterAccount := chain.GenesisAccounts[1]
	submitterAddress, err := submitterAccount.KeyInfo.GetAddress()
	s.Require().NoError(err)

	endpoint := fmt.Sprintf("http://%s", s.commonHelper.Resources.ValResources[chain.ID][0].GetHostPort("1317/tcp"))

	timeoutTimestamp := uint64(time.Now().Add(time.Minute * 5).Unix())

	rawTx, err := s.tx.CreateIBCV2RecvPacketTx(timeoutTimestamp, "1", submitterAddress.String(), RecipientAddress, "")
	s.Require().NoError(err)
	s.Require().NotNil(rawTx)

	unsignedFname := "unsigned_recv_tx.json"
	unsignedJSONFile := filepath.Join(chain.Validators[0].ConfigDir(), unsignedFname)
	err = common.WriteFile(unsignedJSONFile, rawTx)
	s.Require().NoError(err)

	signedTx, err := s.tx.SignTxFileOnline(chain, 0, submitterAddress.String(), unsignedFname)
	s.Require().NoError(err)
	s.Require().NotNil(signedTx)

	signedFname := "signed_recv_tx.json"
	signedJSONFile := filepath.Join(chain.Validators[0].ConfigDir(), signedFname)
	err = common.WriteFile(signedJSONFile, signedTx)
	s.Require().NoError(err)

	// if there's no errors the non_critical_extension_options field was properly encoded and decoded
	out, err := s.tx.BroadcastTxFile(chain, 0, submitterAddress.String(), signedFname)
	s.Require().NoError(err)
	s.Require().NotNil(out)

	s.Require().Eventually(
		func() bool {
			balances, err := query.AllBalances(endpoint, RecipientAddress)
			s.Require().NoError(err)
			return balances[0].String() == "1ibc/1FBF3660E6387150C8BBDAA82EF8CE3C0AADE1F1BD921AE7529D892A53321A74"
		},
		10*time.Second,
		5*time.Second,
	)
}

func (s *IntegrationTestSuite) TestV2Callback() {
	chain := s.commonHelper.Resources.ChainA

	submitterAccount := chain.GenesisAccounts[1]
	submitterAddress, err := submitterAccount.KeyInfo.GetAddress()
	s.Require().NoError(err)

	endpoint := fmt.Sprintf("http://%s", s.commonHelper.Resources.ValResources[chain.ID][0].GetHostPort("1317/tcp"))

	timeoutTimestamp := uint64(time.Now().Add(time.Minute * 5).Unix())

	s.Require().NotEmpty(common.EntrypointAddress)
	s.Require().NotEmpty(common.AdapterAddress)

	memo := msg.BuildCallbacksMemo(common.EntrypointAddress, "ibc/1FBF3660E6387150C8BBDAA82EF8CE3C0AADE1F1BD921AE7529D892A53321A74", common.AdapterAddress, RecipientAddress)

	rawTx, err := s.tx.CreateIBCV2RecvPacketTx(timeoutTimestamp, "1", submitterAddress.String(), common.AdapterAddress, memo)
	s.Require().NoError(err)

	unsignedFname := "unsigned_recv_callback_tx.json"
	unsignedJSONFile := filepath.Join(chain.Validators[0].ConfigDir(), unsignedFname)
	err = common.WriteFile(unsignedJSONFile, rawTx)
	s.Require().NoError(err)

	signedTx, err := s.tx.SignTxFileOnline(chain, 0, submitterAddress.String(), unsignedFname)
	s.Require().NoError(err)
	s.Require().NotNil(signedTx)

	signedFname := "signed_recv_callback_tx.json"
	signedJSONFile := filepath.Join(chain.Validators[0].ConfigDir(), signedFname)
	err = common.WriteFile(signedJSONFile, signedTx)
	s.Require().NoError(err)

	// if there's no errors the non_critical_extension_options field was properly encoded and decoded
	out, err := s.tx.BroadcastTxFile(chain, 0, submitterAddress.String(), signedFname)
	s.Require().NoError(err)
	s.Require().NotNil(out)

	s.Require().Eventually(
		func() bool {
			balances, err := query.AllBalances(endpoint, RecipientAddress)
			s.Require().NoError(err)
			return balances[0].String() == "2ibc/1FBF3660E6387150C8BBDAA82EF8CE3C0AADE1F1BD921AE7529D892A53321A74"
		},
		10*time.Second,
		5*time.Second,
	)
}
