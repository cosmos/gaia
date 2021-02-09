package gaia

/// This file implements Proposal 29 as passed by the Cosmos Hub. During the migration from
/// cosmoshub-3 to cosmoshub-4, the migration function Validates the signed messages and then
/// moves balances from their genesis accounts to new accounts specified in the signed messages.

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type recoveryMessages []FundRecoveryMessage

func (r *recoveryMessages) IsSourceAddress(addr string) (int, bool) {
	for i, m := range *r {
		if m.sourceAddress.String() == addr {
			return i, true
		}
	}
	return -1, false
}

func (r *recoveryMessages) IsDestAddress(addr string) (int, bool) {
	for i, m := range *r {
		if m.destAddress.String() == addr {
			return i, true
		}
	}
	return -1, false
}

func (r *recoveryMessages) GetRemainingBalances() (balances []banktypes.Balance) {
	zeroBalance := sdk.NewInt64Coin("uatom", 0)
	for _, m := range *r {
		if !m.destBalance.IsEqual(zeroBalance) {
			balances = append(balances, banktypes.Balance{
				Address: m.destAddress.String(),
				Coins:   []sdk.Coin{m.destBalance},
			})
		}

	}
	return
}

// GetRemainingAccounts returns the recovery destination addresses with positive balance
func (r *recoveryMessages) GetRemainingAccounts() (addresses []sdk.Address) {
	zeroBalance := sdk.NewInt64Coin("uatom", 0)
	for _, m := range *r {
		if !m.destBalance.IsEqual(zeroBalance) {
			addresses = append(addresses, m.destAddress)
		}

	}
	return addresses
}

// FundRecoveryMessage were signed messages provided by fundraiser particpants who could not access their ATOM to facilate this process for recovering access to their funds.
type FundRecoveryMessage struct {
	signedMessage string
	signature     string
	sourceAddress sdk.Address
	destAddress   sdk.Address
	destBalance   sdk.Coin
}

func (f *FundRecoveryMessage) unMarshalSignedJSON() (signedJSON, error) {
	newLineStripped := strings.ReplaceAll(f.signedMessage, `\n`, "")
	return unmarshalSignedJSON([]byte(strings.ReplaceAll(newLineStripped, "\\", "")))
}

// nolint: unparam
func (f *FundRecoveryMessage) verifyBitcoinSignature(addrIdx int) (signedJSON, error) {
	msgJSON, err := f.unMarshalSignedJSON()
	if err != nil {
		return signedJSON{}, err
	}

	err = verifyBitcoinSignature(f.signature, f.signedMessage, msgJSON.ContributingAddresses[addrIdx].Address)
	if err != nil {
		return signedJSON{}, err
	}

	return msgJSON, nil
}

// nolint: unparam
func (f *FundRecoveryMessage) verifyEthereumSignature(addrIdx int) (signedJSON, error) {
	msgJSON, err := f.unMarshalSignedJSON()
	if err != nil {
		return signedJSON{}, err
	}

	err = verifyEthereumSignature(f.signature, f.signedMessage, msgJSON.ContributingAddresses[addrIdx].Address)
	if err != nil {
		return signedJSON{}, err
	}

	return msgJSON, nil
}

func unmarshalSignedJSON(data []byte) (signedJSON, error) {
	var r signedJSON
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *signedJSON) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type signedJSON struct {
	Message               string                `json:"Message"`
	Txid                  string                `json:"Txid"`
	Contribution          string                `json:"Contribution"`
	ContributingAddresses []contributingAddress `json:"Contributing Addresses"`
	GenesisCosmosAddress  string                `json:"Genesis Cosmos Address"`
	RecoveryCosmosAddress string                `json:"Recovery Cosmos Address"`
}

type contributingAddress struct {
	Address string `json:"address"`
}

func btcDonor1() FundRecoveryMessage {

	sourceAddress, _ := sdk.AccAddressFromBech32("cosmos17s3zkmz6d42rgvtxj3mxaqqtl6dtn4c75lwl4k") // nolint: errcheck
	destAddress, _ := sdk.AccAddressFromBech32("cosmos1hq5jdspaysmpt5cgg3w4g2u84th3qklr4e5jlq")   // nolint: errcheck

	return FundRecoveryMessage{
		`{ "Message":"To this day, I can’t understand how I managed to loose my fundraiser seed. I followed Cosmos all along, even considered founding a company to become a Validator myself at one point. Logged on to vote for a validator when Mainnet launched, and … no seed. A painful moment.Now there’s hope thanks to this proposal and thanks to me still being able to sign the Bitcoin address I used for donating to the fundraiser. I’d be truly happy if my fellow Cosmonauts (alas, in spirit only at the moment), and the Validators would help us use the flexibility of Cosmos to recover our Genesis coins.", "Txid":"https://btc1.trezor.io/tx/331db408685ff221193e88e0d548493920ab785e2d62f3898ce807463afc11e3", "Contribution":"0.4 BTC", "Contributing Addresses":[ { "address":"1JsiFGmKZvr3iR4KejpMswmapCoB2oWsog" }, { "address":"1E1dPcQwyPBxYRUfYBt64BpZnvsgSam6SS" } ], "Genesis Cosmos Address":"cosmos17s3zkmz6d42rgvtxj3mxaqqtl6dtn4c75lwl4k", "Recovery Cosmos Address":"cosmos1hq5jdspaysmpt5cgg3w4g2u84th3qklr4e5jlq" }`,
		"IDM3egFNpQRboMrnxMAlIp2XAI0dylAEAysQeNTVLodBP3DU0aY/co0K5ngafdBPgIzGqxA2+ZH/c36tyB+PdWs=",
		sourceAddress,
		destAddress,
		sdk.NewInt64Coin("uatom", 4180221000)}
}

func btcDonor2() FundRecoveryMessage {

	sourceAddress, _ := sdk.AccAddressFromBech32("cosmos1fhejq5a3z8kehknrrjupt2wdl4l0wh9y2vayhr") // nolint: errcheck
	destAddress, _ := sdk.AccAddressFromBech32("cosmos12h5m3cuupza33psgzegjdvvsrngadjjmgnql34")   // nolint: errcheck

	return FundRecoveryMessage{
		`{ "Message":" I took a screenshot of the seed and due to my stupidity the only copy of the phrase. In Jan 2019 , I was installing Ubuntu on an external disk , by using my macbook as the medium, where I overrode the OS on my macbook itself instead of writing it to the external disk. The wipe replaced the entire OS and the Data Recovery company I gave it to were not able to retrieve the data. I lost atoms and some other crypto whose wallets were only on that macbook.", "Txid":"https://btc1.trezor.io/tx/ed6598ff038f1beaf1e1fe1c32aa50f770eaf16e6b7b86f45ed565837e546f05", "Contribution":"3 BTC", "Contributing Addresses":[ { "address":"16zuFhEMVTggK1rkMxwtsbBnSRH5NZ55nP" }, { "address":"1AC8nqvxTh8PHaMU5Fn8ZBhbodRasaRSRd" }, { "address":"1Gyi2AbGL2v5dbHogeezfi8om47c5q6EM8" } ], "Genesis Cosmos Address":"cosmos1fhejq5a3z8kehknrrjupt2wdl4l0wh9y2vayhr", "Recovery Cosmos Address":"cosmos12h5m3cuupza33psgzegjdvvsrngadjjmgnql34" }`,
		"IItkbg76ZjfFmMlyAmFAMz7CCt1FIw7VluWQbDvNhGdFZ6Y/FohkqlvXdv7aon8MGm6LbntbL3RxzR780PLYHY8=",
		sourceAddress,
		destAddress,
		sdk.NewInt64Coin("uatom", 31406121000)}
}

func btcDonor3() FundRecoveryMessage {

	sourceAddress, _ := sdk.AccAddressFromBech32("cosmos14qyqets0c94u9hjmvrm4n8s2v5pgnk9kjh93ay") // nolint: errcheck
	destAddress, _ := sdk.AccAddressFromBech32("cosmos12sptngkpvc3alssd9wcgr9sn5zh2rdg8gt27x0")   // nolint: errcheck

	return FundRecoveryMessage{
		`{ "Message":"My investment partner Calvin  was going to handle the donation for us. He was up early and waiting for it to start. When the sale started, he informed me that a BTC address appeared on his screen and asked  me on the phone if he should go ahead and send the BTC there. He was never presented with a flow to obtain a seedphrase, or to confirm it. In prior ICOs you would just send Bitcoin to a wallet address, I assumed the atom sale was similarly designed so I let him know that we should just go ahead and send our BTC. He sent 22 BTC in two tx's", "Txid":"https://btc1.trezor.io/tx/f1033af3d9b7bc9e5675acf74940d982acda6848a5e0e9234ead0f6043f2fc65", "Contribution":"12 BTC", "Contributing Addresses":[ { "address":"1LXHU6XHzsCjWhuiLe24CRLLBwVQVeZimg" },{ "address":"1HsGtAmmHRA29g6UgikqX7hyPog9vaWyE9" } ], "Genesis Cosmos Address":"cosmos14qyqets0c94u9hjmvrm4n8s2v5pgnk9kjh93ay", "Recovery Cosmos Address":"cosmos12sptngkpvc3alssd9wcgr9sn5zh2rdg8gt27x0" }`,
		"H4+JhJH23leWopflTc5lIS4Wz089YWrUBCPGo9AwLktfUf5+Kpzo/nWS121/qyMxBF1G2eFUe+gwUhuRdZ4FWPA=",
		sourceAddress,
		destAddress,
		sdk.NewInt64Coin("uatom", 125649621000)}
}

/// BTC Donor 4 should have the same message as Donor 3 because they signed they same message from two difference addresses that were both donation sources.

func btcDonor4() FundRecoveryMessage {

	sourceAddress, _ := sdk.AccAddressFromBech32("cosmos138a2ulzndl7gezsd6symywvdpzes4awj9eypkr") // nolint: errcheck
	destAddress, _ := sdk.AccAddressFromBech32("cosmos1nc8s85wnax2jn2zch4cuyqumfqml53cumhwk82")   // nolint: errcheck

	return FundRecoveryMessage{
		`{ "Message":"My investment partner Calvin  was going to handle the donation for us. He was up early and waiting for it to start. When the sale started, he informed me that a BTC address appeared on his screen and asked  me on the phone if he should go ahead and send the BTC there. He was never presented with a flow to obtain a seedphrase, or to confirm it. In prior ICOs you would just send Bitcoin to a wallet address, I assumed the atom sale was similarly designed so I let him know that we should just go ahead and send our BTC. He sent 22 BTC in two tx's", "Txid":"https://btc1.trezor.io/tx/1a51fbb14d14ffb6270ddf551408069e22726ca2fbd8fd49ab5301f1157ff631", "Contribution":"10 BTC", "Contributing Addresses":[ { "address":"14VcHxn6YLAf1iBFVJtn3n8x3zf18ifybZ" }, { "address":"1L6V9E2LaP3eEQiWnzWyzGxzHUsRqVSxCQ" }, { "address":"13GKeu4BNbe1mfjDpwpVaZ2Hn34cHDrQPn" } ], "Genesis Cosmos Address":"cosmos138a2ulzndl7gezsd6symywvdpzes4awj9eypkr", "Recovery Cosmos Address":"cosmos1nc8s85wnax2jn2zch4cuyqumfqml53cumhwk82" }`,
		"IFb7u28IBBzn9UuXrOJnzO0kL8yItojE8X3GAcD+bk2JSXRHgsJLj0wYdXn9P8HqlDtl7jF5a1Bxu/f3Lxul+Xk=",
		sourceAddress,
		destAddress,
		sdk.NewInt64Coin("uatom", 104706621000)}
}

func ethDonor1() FundRecoveryMessage {
	sourceAddress, _ := sdk.AccAddressFromBech32("cosmos1m06fehere8q09sqcj08dcc0xyj3fjlzc2x24y4") // nolint: errcheck
	destAddress, _ := sdk.AccAddressFromBech32("cosmos17cg9xxpjnammafyqwfryr3hn32h6vjmh9x0y6j")   // nolint: errcheck

	return FundRecoveryMessage{
		`{ "Message":"I remember After I successfully sent my 16 Eth to the Cosmos fundraising address 3 years ago,I wrote down those seed words and had a picture.But when we could claim,I found my seed phase does not work,I've tried every means I could,but all failed.Honestly,I still don’t know the problem,maybe I wrote wrong letters,maybe something wrong with that webpage,or maybe I mixed something,only God knows.", "Txid":"https://etherscan.io/tx/0x42d0f860e1cd484f51647f34479843008b69d8f1158c94ad44ae30df33fdc080", "Contribution":"16 ETH", "Contributing Addresses":[ { "address":"0x53ad4398f76a453a2d4dac4470f0b81cd1d72715" } ], "Genesis Cosmos Address":"cosmos1m06fehere8q09sqcj08dcc0xyj3fjlzc2x24y4", "Recovery Cosmos Address":"cosmos17cg9xxpjnammafyqwfryr3hn32h6vjmh9x0y6j" }`,
		"0xceed630f7e8d102b125a22d9ec06ced12a016f376b408b9832a63a9b4b5f352b4b67c94e9cfa35e4f2676f0f92643f8b56caf82f6c6eabd765b787d3f6af77fb1b",
		sourceAddress,
		destAddress,
		sdk.NewInt64Coin("uatom", 6512400000)}
}

func ethDonor2() FundRecoveryMessage {
	sourceAddress, _ := sdk.AccAddressFromBech32("cosmos1gzuqry88awndjjsa5exzx4gwnmktcpdrxgdcf6") // nolint: errcheck
	destAddress, _ := sdk.AccAddressFromBech32("cosmos1teux7wdnnq03r7r277yu762mq3cket5mg4xd3e")   // nolint: errcheck

	return FundRecoveryMessage{
		"{\n   \"Message\":\"In April 2017, I participated in 24 ETH at Cosmos ico. In about two years when Cosmos was listed on exchange, the cell phone that had taken the seed phrase picture was destroyed. I tried to go to a data recovery company, but I couldn't recover the picture, so there was no way to get ATOM. However, I still have the private key and keystore file of the Ethereum wallet that participated in Cosmos ico.\",\n   \"Txid\":\"https://etherscan.io/tx/0xe2bb8c832c237b9ed898d4616649347e84931d56b8942cb409cafd6b01e1913d\",\n   \"Contribution\":\"24 ETH\",\n   \"Contributing Addresses\":[\n      {\n         \"address\":\"0xff3fa81a59f31bd563d2554401438a1678d43593\"\n      }\n   ],\n   \"Genesis Cosmos Address\":\"cosmos1gzuqry88awndjjsa5exzx4gwnmktcpdrxgdcf6\",\n   \"Recovery Cosmos Address\":\"cosmos1teux7wdnnq03r7r277yu762mq3cket5mg4xd3e\"\n}\n",
		"0x38de4018152de5f42d24b1150c04f5010dc3edce9a3436a57d318beae5e6955228a5e2c1255591c0324e4e9f1bbd13806e51bbdb11259c9e2aeddbdbc91bc11a1b",
		sourceAddress,
		destAddress,
		sdk.NewInt64Coin("uatom", 9769500000)}
}

func ethDonor3() FundRecoveryMessage {
	sourceAddress, _ := sdk.AccAddressFromBech32("cosmos1r3xvguuhwvlk34esxclvrh3g7ycmcqqc2kcn9v") // nolint: errcheck
	destAddress, _ := sdk.AccAddressFromBech32("cosmos18qjynuyrfja9qugzs4zjcs6dh0qyprqa2vwktp")   // nolint: errcheck

	return FundRecoveryMessage{
		"{ \"Message\":\"One of the early Ethereum alternatives is Cosmos. This is one of the reasons ICONOMI that is a collective of 1500 small donors , donated to  Cosmos fundraiser about  2,222 ETH similarly to numerous ICO where we participated. Due to the possible bug  in the code, we did not get the seed phrases. During the subscription process using the Brave browser, following all the steps the seed phrases simply did not show-up. Iconomi will be grateful to the community for the recovery of the atoms.\", \"Txid\":\"https://etherscan.io/tx/0x0c85c7cc2b66840357c3f293ae2010f0c79f2cc9f4b1220028afe780fdfdb426\", \"Contribution\":\"2222 ETH\", \"Contributing Addresses\":[ { \"address\":\"0xb4dc54df11d2dcecd046e5c7318fb241a73ee370\" } ], \"Genesis Cosmos Address\":\"cosmos1r3xvguuhwvlk34esxclvrh3g7ycmcqqc2kcn9v\", \"Recovery Cosmos Address\":\"cosmos18qjynuyrfja9qugzs4zjcs6dh0qyprqa2vwktp\" }",
		"0x4c29f9d74a070a8c475553597a1bd461137af0ba9120c183a1cfe3dc8c729f367dcf76ed9a384eea18c920b7ba7613ddb5632da9642fdb42cc183ff5ea74614e1b",
		sourceAddress,
		destAddress,
		sdk.NewInt64Coin("uatom", 904509000000)}
}

func validateFundRecovery() recoveryMessages {

	//Bitcoin Donors
	bDonor1 := btcDonor1()
	bDonor2 := btcDonor2()
	bDonor3 := btcDonor3()
	bDonor4 := btcDonor4()

	//Bitcoin Verification
	_, err := bDonor1.verifyBitcoinSignature(0)
	if err != nil {
		log.Fatalf("BTC1 %e", err)
	}
	_, err = bDonor2.verifyBitcoinSignature(0)
	if err != nil {
		log.Fatalf("BTC2 %e", err)
	}
	_, err = bDonor3.verifyBitcoinSignature(1)
	if err != nil {
		log.Fatalf("BTC3 %e", err)
	}
	_, err = bDonor4.verifyBitcoinSignature(0)
	if err != nil {
		log.Fatalf("BTC4 %e", err)
	}

	//Ethereum Donors
	eDonor1 := ethDonor1()
	eDonor2 := ethDonor2()
	eDonor3 := ethDonor3()

	//Ethereum Verification
	_, err = eDonor1.verifyEthereumSignature(0)
	if err != nil {
		log.Fatalf("Eth1 %e", err)
	}

	_, err = eDonor2.verifyEthereumSignature(0)
	if err != nil {
		log.Fatalf("Eth2 %e", err)
	}

	_, err = eDonor3.verifyEthereumSignature(0)
	if err != nil {
		log.Fatalf("Eth3 %e", err)
	}

	return []FundRecoveryMessage{bDonor1, bDonor2, bDonor3, bDonor4, eDonor1, eDonor2, eDonor3}
}

// EcRecover returns the address for the account that was used to create the signature.
// Note, this function is compatible with eth_sign and personal_sign. As such it recovers
// the address of:
// hash = keccak256("\x19Ethereum Signed Message:\n"${message length}${message})
// addr = ecrecover(hash, signature)
//
// Note, the signature must conform to the secp256k1 curve R, S and V values, where
// the V value must be 27 or 28 for legacy reasons.
//
// https://github.com/ethereum/go-ethereum/wiki/Management-APIs#personal_ecRecover
func ecRecover(sig, msg hexutil.Bytes) (common.Address, error) {
	if len(sig) != crypto.SignatureLength {
		return common.Address{}, fmt.Errorf("signature must be %d bytes long", crypto.SignatureLength)
	}
	if sig[crypto.RecoveryIDOffset] != 27 && sig[crypto.RecoveryIDOffset] != 28 {
		return common.Address{}, fmt.Errorf("invalid Ethereum signature (V is not 27 or 28)")
	}
	sig[crypto.RecoveryIDOffset] -= 27 // Transform yellow paper V from 27/28 to 0/1

	rpk, err := crypto.SigToPub(accounts.TextHash(msg), sig)
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*rpk), nil
}

// Eth_sign verifier for MEW signatures.
func verifyEthereumSignature(sig, msg, addr string) error {
	var sigBz hexutil.Bytes

	if err := sigBz.UnmarshalText([]byte(sig)); err != nil {
		return fmt.Errorf("eth sig unmarshal error: %w", err)
	}

	addrRecovered, err := ecRecover(sigBz, []byte(msg))
	if err != nil {
		return fmt.Errorf("ecRecover error: %w", err)
	}

	address := common.HexToAddress(addr)

	if !bytes.Equal(addrRecovered.Bytes(), address.Bytes()) {
		log.Fatalf("invalid signature, given address %s, recovered address %s", address.String(), addrRecovered.String())
	}

	return nil
}

const messageSignatureHeader = "Bitcoin Signed Message:\n"

// Modified Bitcoin signature key recovery and address verification script that verifies
// signed messages against simple pay to pub key hash addresses.
// Should panic on failure.
func verifyBitcoinSignature(sig, msg, addr string) error {
	var buf bytes.Buffer
	err := wire.WriteVarString(&buf, 0, messageSignatureHeader)
	if err != nil {
		return fmt.Errorf("msg signature header serialization failed: %w", err)
	}

	if err := wire.WriteVarString(&buf, 0, msg); err != nil {
		return fmt.Errorf("msg serialization failed: %w", err)
	}

	expectedMessageHash := chainhash.DoubleHashB(buf.Bytes())

	sigBytes, err := base64.StdEncoding.DecodeString(sig)
	if err != nil {
		return fmt.Errorf("base64 signature decoding failed: %w", err)
	}

	pk, wasCompressed, err := btcec.RecoverCompact(btcec.S256(), sigBytes, expectedMessageHash)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	var serializedPK []byte
	if wasCompressed {
		serializedPK = pk.SerializeCompressed()
	} else {
		serializedPK = pk.SerializeUncompressed()
	}

	address, err := btcutil.NewAddressPubKey(serializedPK, &chaincfg.MainNetParams)
	if err != nil {
		return fmt.Errorf("address recovery from pubkey failed: %w", err)
	}

	if address.EncodeAddress() != addr {
		return fmt.Errorf("address mismatch, expected %s, got %s", addr, address.EncodeAddress())
	}

	return nil
}

func Prop29Migration(authGenesis *authtypes.GenesisState, bankGenesis *banktypes.GenesisState, distrGenesis *distr.GenesisState) (authtypes.GenesisState, banktypes.GenesisState, distr.GenesisState) {
	fundRecovery := validateFundRecovery()

	recoveryAccounting := sdk.NewInt64Coin("uatom", 0)
	emptyCoins := []sdk.Coin{}
	distModuleAccount := authtypes.NewModuleAddress(distr.ModuleName)

	// zero out all source addresses balances and accumulate the total funds being moved
	for i, balance := range bankGenesis.Balances {
		_, isSourceAddress := fundRecovery.IsSourceAddress(balance.Address)
		if !isSourceAddress {
			continue
		}

		if len(balance.Coins) > 1 {
			log.Fatal("expected all balances to contain only 1 denom during the migration")
		}
		// accumulate all the coins removed from the balances into the pool
		recoveryAccounting = recoveryAccounting.Add(balance.Coins[0])
		// Empty the source address balance
		bankGenesis.Balances[i].Coins = emptyCoins
	}

	// migrate the balances to the the destination addresses
	for i, balance := range bankGenesis.Balances {
		index, isDestAddress := fundRecovery.IsDestAddress(balance.Address)
		if !isDestAddress {
			continue
		}

		// transfer coins from the atom recovery pool to the dest address
		recoveryAccounting = recoveryAccounting.Sub(fundRecovery[index].destBalance)
		bankGenesis.Balances[i].Coins = bankGenesis.Balances[i].Coins.Add(fundRecovery[index].destBalance)
		fundRecovery[index].destBalance = sdk.NewInt64Coin("uatom", 0)
	}

	// add the balances to the bank genesis
	bankGenesis.Balances = append(bankGenesis.Balances, fundRecovery.GetRemainingBalances()...)

	accs, err := authtypes.UnpackAccounts(authGenesis.Accounts)
	if err != nil {
		log.Fatalf("could not unpack genesis accounts: %s", err.Error())
	}

	// add the accounts with positive balance to the genesis accounts
	for _, addr := range fundRecovery.GetRemainingAccounts() {
		recoveryAccount := authtypes.NewBaseAccount(sdk.AccAddress(addr.Bytes()), nil, 0, 0)
		accs = append(accs, recoveryAccount)
	}

	genAccs, err := authtypes.PackAccounts(accs)
	if err != nil {
		log.Fatalf("could not pack genesis accounts: %s", err.Error())
	}

	authGenesis.Accounts = genAccs

	// subtract the coins from the addresses with positive balances from the recovery accounting
	for _, balance := range fundRecovery.GetRemainingBalances() {
		recoveryAccounting = recoveryAccounting.Sub(balance.Coins[0])
	}

	// Add the remaining ATOMs to the fee pool by adding them to distribution modules
	for i, balance := range bankGenesis.Balances {
		if distModuleAccount.String() != balance.Address {
			continue
		}

		// add coins to the community pool bank and distr balances
		bankGenesis.Balances[i].Coins = bankGenesis.Balances[i].Coins.Add(recoveryAccounting)
		distrGenesis.FeePool.CommunityPool = distrGenesis.FeePool.CommunityPool.Add(sdk.NewDecCoinFromCoin(recoveryAccounting))
		break
	}

	return *authGenesis, *bankGenesis, *distrGenesis

}
