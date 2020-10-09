package app

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// FundRecoveryMessage were signed messages provided by fundraiser particpants who could not access their ATOM to facilate this process for recovering access to their funds.
type FundRecoveryMessage struct {
	signedMessage string
	signature     string
	sourceAddress types.Address
	destAddress   types.Address
	destBalance   types.Coin
}

func (f *FundRecoveryMessage) unmarshalSignedJson() (SignedJSON, error) {
	return UnmarshalSignedJSON([]byte(f.signedMessage))
}

func (f *FundRecoveryMessage) VerifyBitcoinSignature(addrIdx int) (SignedJSON, error) {
	msgJson, err := f.unmarshalSignedJson()
	if err != nil {
		return SignedJSON{}, err
	}
	//fmt.Println(msgJson)
	VerifyBitcoinSignature(f.signature, f.signedMessage, msgJson.ContributingAddresses[addrIdx].Address)
	return msgJson, nil
}

func (f *FundRecoveryMessage) VerifyEthereumSignature(addrIdx int) (SignedJSON, error) {
	msgJson, err := f.unmarshalSignedJson()
	if err != nil {
		return SignedJSON{}, err
	}
	//fmt.Println(msgJson)
	VerifyEthereumSignature(f.signature, f.signedMessage, msgJson.ContributingAddresses[addrIdx].Address)
	return msgJson, nil
}

func UnmarshalSignedJSON(data []byte) (SignedJSON, error) {
	var r SignedJSON
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *SignedJSON) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type SignedJSON struct {
	Message               string                `json:"Message"`
	Txid                  string                `json:"Txid"`
	Contribution          string                `json:"Contribution"`
	ContributingAddresses []ContributingAddress `json:"Contributing Addresses"`
	GenesisCosmosAddress  string                `json:"Genesis Cosmos Address"`
	RecoveryCosmosAddress string                `json:"Recovery Cosmos Address"`
}

type ContributingAddress struct {
	Address string `json:"address"`
}

func BTCDonor1() FundRecoveryMessage {

	sourceAddress, _ := types.AccAddressFromBech32("cosmos17s3zkmz6d42rgvtxj3mxaqqtl6dtn4c75lwl4k")
	destAddress, _ := types.AccAddressFromBech32("cosmos1hq5jdspaysmpt5cgg3w4g2u84th3qklr4e5jlq")

	return FundRecoveryMessage{
		`{ "Message":"To this day, I can’t understand how I managed to loose my fundraiser seed. I followed Cosmos all along, even considered founding a company to become a Validator myself at one point. Logged on to vote for a validator when Mainnet launched, and … no seed. A painful moment.Now there’s hope thanks to this proposal and thanks to me still being able to sign the Bitcoin address I used for donating to the fundraiser. I’d be truly happy if my fellow Cosmonauts (alas, in spirit only at the moment), and the Validators would help us use the flexibility of Cosmos to recover our Genesis coins.", "Txid":"https://btc1.trezor.io/tx/331db408685ff221193e88e0d548493920ab785e2d62f3898ce807463afc11e3", "Contribution":"0.4 BTC", "Contributing Addresses":[ { "address":"1JsiFGmKZvr3iR4KejpMswmapCoB2oWsog" }, { "address":"1E1dPcQwyPBxYRUfYBt64BpZnvsgSam6SS" } ], "Genesis Cosmos Address":"cosmos17s3zkmz6d42rgvtxj3mxaqqtl6dtn4c75lwl4k", "Recovery Cosmos Address":"cosmos1hq5jdspaysmpt5cgg3w4g2u84th3qklr4e5jlq" }`,
		"IDM3egFNpQRboMrnxMAlIp2XAI0dylAEAysQeNTVLodBP3DU0aY/co0K5ngafdBPgIzGqxA2+ZH/c36tyB+PdWs=",
		sourceAddress,
		destAddress,
		types.NewInt64Coin("uatom", 4180221000)}
}

func BTCDonor2() FundRecoveryMessage {

	sourceAddress, _ := types.AccAddressFromBech32("cosmos1fhejq5a3z8kehknrrjupt2wdl4l0wh9y2vayhr")
	destAddress, _ := types.AccAddressFromBech32("cosmos12h5m3cuupza33psgzegjdvvsrngadjjmgnql34")

	return FundRecoveryMessage{
		`{ "Message":" I took a screenshot of the seed and due to my stupidity the only copy of the phrase. In Jan 2019 , I was installing Ubuntu on an external disk , by using my macbook as the medium, where I overrode the OS on my macbook itself instead of writing it to the external disk. The wipe replaced the entire OS and the Data Recovery company I gave it to were not able to retrieve the data. I lost atoms and some other crypto whose wallets were only on that macbook.", "Txid":"https://btc1.trezor.io/tx/ed6598ff038f1beaf1e1fe1c32aa50f770eaf16e6b7b86f45ed565837e546f05", "Contribution":"3 BTC", "Contributing Addresses":[ { "address":"16zuFhEMVTggK1rkMxwtsbBnSRH5NZ55nP" }, { "address":"1AC8nqvxTh8PHaMU5Fn8ZBhbodRasaRSRd" }, { "address":"1Gyi2AbGL2v5dbHogeezfi8om47c5q6EM8" } ], "Genesis Cosmos Address":"cosmos1fhejq5a3z8kehknrrjupt2wdl4l0wh9y2vayhr", "Recovery Cosmos Address":"cosmos12h5m3cuupza33psgzegjdvvsrngadjjmgnql34" }`,
		"IItkbg76ZjfFmMlyAmFAMz7CCt1FIw7VluWQbDvNhGdFZ6Y/FohkqlvXdv7aon8MGm6LbntbL3RxzR780PLYHY8=",
		sourceAddress,
		destAddress,
		types.NewInt64Coin("uatom", 31406121000)}
}

func BTCDonor3() FundRecoveryMessage {

	sourceAddress, _ := types.AccAddressFromBech32("cosmos14qyqets0c94u9hjmvrm4n8s2v5pgnk9kjh93ay")
	destAddress, _ := types.AccAddressFromBech32("cosmos12sptngkpvc3alssd9wcgr9sn5zh2rdg8gt27x0")

	return FundRecoveryMessage{
		`{ "Message":"My investment partner Calvin  was going to handle the donation for us. He was up early and waiting for it to start. When the sale started, he informed me that a BTC address appeared on his screen and asked  me on the phone if he should go ahead and send the BTC there. He was never presented with a flow to obtain a seedphrase, or to confirm it. In prior ICOs you would just send Bitcoin to a wallet address, I assumed the atom sale was similarly designed so I let him know that we should just go ahead and send our BTC. He sent 22 BTC in two tx's", "Txid":"https://btc1.trezor.io/tx/f1033af3d9b7bc9e5675acf74940d982acda6848a5e0e9234ead0f6043f2fc65", "Contribution":"12 BTC", "Contributing Addresses":[ { "address":"1LXHU6XHzsCjWhuiLe24CRLLBwVQVeZimg" },{ "address":"1HsGtAmmHRA29g6UgikqX7hyPog9vaWyE9" } ], "Genesis Cosmos Address":"cosmos14qyqets0c94u9hjmvrm4n8s2v5pgnk9kjh93ay", "Recovery Cosmos Address":"cosmos12sptngkpvc3alssd9wcgr9sn5zh2rdg8gt27x0" }`,
		"H4+JhJH23leWopflTc5lIS4Wz089YWrUBCPGo9AwLktfUf5+Kpzo/nWS121/qyMxBF1G2eFUe+gwUhuRdZ4FWPA=",
		sourceAddress,
		destAddress,
		types.NewInt64Coin("uatom", 125649621000)}
}

func BTCDonor4() FundRecoveryMessage {

	sourceAddress, _ := types.AccAddressFromBech32("cosmos138a2ulzndl7gezsd6symywvdpzes4awj9eypkr")
	destAddress, _ := types.AccAddressFromBech32("cosmos1nc8s85wnax2jn2zch4cuyqumfqml53cumhwk82")

	return FundRecoveryMessage{
		`{ "Message":"My investment partner Calvin  was going to handle the donation for us. He was up early and waiting for it to start. When the sale started, he informed me that a BTC address appeared on his screen and asked  me on the phone if he should go ahead and send the BTC there. He was never presented with a flow to obtain a seedphrase, or to confirm it. In prior ICOs you would just send Bitcoin to a wallet address, I assumed the atom sale was similarly designed so I let him know that we should just go ahead and send our BTC. He sent 22 BTC in two tx's", "Txid":"https://btc1.trezor.io/tx/1a51fbb14d14ffb6270ddf551408069e22726ca2fbd8fd49ab5301f1157ff631", "Contribution":"10 BTC", "Contributing Addresses":[ { "address":"14VcHxn6YLAf1iBFVJtn3n8x3zf18ifybZ" }, { "address":"1L6V9E2LaP3eEQiWnzWyzGxzHUsRqVSxCQ" }, { "address":"13GKeu4BNbe1mfjDpwpVaZ2Hn34cHDrQPn" } ], "Genesis Cosmos Address":"cosmos138a2ulzndl7gezsd6symywvdpzes4awj9eypkr", "Recovery Cosmos Address":"cosmos1nc8s85wnax2jn2zch4cuyqumfqml53cumhwk82" }`,
		"IFb7u28IBBzn9UuXrOJnzO0kL8yItojE8X3GAcD+bk2JSXRHgsJLj0wYdXn9P8HqlDtl7jF5a1Bxu/f3Lxul+Xk=",
		sourceAddress,
		destAddress,
		types.NewInt64Coin("uatom", 104706621000)}
}

func EthDonor1() FundRecoveryMessage {
	sourceAddress, _ := types.AccAddressFromBech32("cosmos1m06fehere8q09sqcj08dcc0xyj3fjlzc2x24y4")
	destAddress, _ := types.AccAddressFromBech32("cosmos17cg9xxpjnammafyqwfryr3hn32h6vjmh9x0y6j")

	return FundRecoveryMessage{
		`{ \"Message\":\"I remember After I successfully sent my 16 Eth to the Cosmos fundraising address 3 years ago,I wrote down those seed words and had a picture.But when we could claim,I found my seed phase does not work,I've tried every means I could,but all failed.Honestly,I still don’t know the problem,maybe I wrote wrong letters,maybe something wrong with that webpage,or maybe I mixed something,only God knows.\", \"Txid\":\"https://etherscan.io/tx/0x42d0f860e1cd484f51647f34479843008b69d8f1158c94ad44ae30df33fdc080\", \"Contribution\":\"16 ETH\", \"Contributing Addresses\":[ { \"address\":\"0x53ad4398f76a453a2d4dac4470f0b81cd1d72715\" } ], \"Genesis Cosmos Address\":\"cosmos1m06fehere8q09sqcj08dcc0xyj3fjlzc2x24y4\", \"Recovery Cosmos Address\":\"cosmos17cg9xxpjnammafyqwfryr3hn32h6vjmh9x0y6j\" }`,
		"ceed630f7e8d102b125a22d9ec06ced12a016f376b408b9832a63a9b4b5f352b4b67c94e9cfa35e4f2676f0f92643f8b56caf82f6c6eabd765b787d3f6af77fb1b",
		sourceAddress,
		destAddress,
		types.NewInt64Coin("uatom", 6512400000)}
}

func VerifyEthereumSignature(sig, msg, addr string) {
	sigBytes, _ := hexutil.Decode(sig)
	addrBytes, _ := hexutil.Decode(addr)

	hash := accounts.TextHash([]byte(msg))

	sigPublicKey, err := crypto.Ecrecover(hash, sigBytes)

	pubkey, err := crypto.DecompressPubkey(sigPublicKey)

	if err != nil {
		log.Fatal(err)
	}

	addrRecovered := crypto.PubkeyToAddress(*pubkey)

	if bytes.Compare(addrRecovered.Bytes(), addrBytes) != 0 {
		log.Fatal("invalid signature")

	}

	if err != nil {
		log.Fatal(err)
	}

}

const messageSignatureHeader = "Bitcoin Signed Message:\n"

func VerifyBitcoinSignature(sig, msg, addr string) {

	var buf bytes.Buffer
	wire.WriteVarString(&buf, 0, messageSignatureHeader)
	wire.WriteVarString(&buf, 0, msg)
	expectedMessageHash := chainhash.DoubleHashB(buf.Bytes())

	sigBytes, _ := base64.StdEncoding.DecodeString(sig)

	pk, wasCompressed, _ := btcec.RecoverCompact(btcec.S256(), sigBytes,
		expectedMessageHash)

	var serializedPK []byte
	if wasCompressed {
		serializedPK = pk.SerializeCompressed()
	} else {
		serializedPK = pk.SerializeUncompressed()
	}

	address, err := btcutil.NewAddressPubKey(serializedPK, &chaincfg.MainNetParams)

	if err != nil {
		log.Fatal("Address recovery failed")

	}

	if address.EncodeAddress() != addr {
		log.Fatal("invalid signature")

	}

}
