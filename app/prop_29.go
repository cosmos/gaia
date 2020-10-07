package app

import (
	"bytes"
	"encoding/base64"
	"log"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/cosmos/cosmos-sdk/types"
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

func BTCDonor1() FundRecoveryMessage {

	sourceAddress, _ := types.AccAddressFromBech32("cosmos17s3zkmz6d42rgvtxj3mxaqqtl6dtn4c75lwl4k")
	destAddress, _ := types.AccAddressFromBech32("cosmos1hq5jdspaysmpt5cgg3w4g2u84th3qklr4e5jlq")

	return FundRecoveryMessage{`{ "Message":"To this day, I can’t understand how I managed to loose my fundraiser seed. I followed Cosmos all along, even considered founding a company to become a Validator myself at one point. Logged on to vote for a validator when Mainnet launched, and … no seed. A painful moment.Now there’s hope thanks to this proposal and thanks to me still being able to sign the Bitcoin address I used for donating to the fundraiser. I’d be truly happy if my fellow Cosmonauts (alas, in spirit only at the moment), and the Validators would help us use the flexibility of Cosmos to recover our Genesis coins.", "Txid":"https://btc1.trezor.io/tx/331db408685ff221193e88e0d548493920ab785e2d62f3898ce807463afc11e3", "Contribution":"0.4 BTC", "Contributing Addresses":[ { "address":"1JsiFGmKZvr3iR4KejpMswmapCoB2oWsog" }, { "address":"1E1dPcQwyPBxYRUfYBt64BpZnvsgSam6SS" } ], "Genesis Cosmos Address":"cosmos17s3zkmz6d42rgvtxj3mxaqqtl6dtn4c75lwl4k", "Recovery Cosmos Address":"cosmos1hq5jdspaysmpt5cgg3w4g2u84th3qklr4e5jlq" }`, "IDM3egFNpQRboMrnxMAlIp2XAI0dylAEAysQeNTVLodBP3DU0aY/co0K5ngafdBPgIzGqxA2+ZH/c36tyB+PdWs=", sourceAddress, destAddress, types.NewInt64Coin("uatom", 4180221000)}
}

func BTCDonor2() FundRecoveryMessage {

	sourceAddress, _ := types.AccAddressFromBech32("cosmos1fhejq5a3z8kehknrrjupt2wdl4l0wh9y2vayhr")
	destAddress, _ := types.AccAddressFromBech32("cosmos12h5m3cuupza33psgzegjdvvsrngadjjmgnql34")

	return FundRecoveryMessage{`{ "Message":" I took a screenshot of the seed and due to my stupidity the only copy of the phrase. In Jan 2019 , I was installing Ubuntu on an external disk , by using my macbook as the medium, where I overrode the OS on my macbook itself instead of writing it to the external disk. The wipe replaced the entire OS and the Data Recovery company I gave it to were not able to retrieve the data. I lost atoms and some other crypto whose wallets were only on that macbook.", "Txid":"https://btc1.trezor.io/tx/ed6598ff038f1beaf1e1fe1c32aa50f770eaf16e6b7b86f45ed565837e546f05", "Contribution":"3 BTC", "Contributing Addresses":[ { "address":"16zuFhEMVTggK1rkMxwtsbBnSRH5NZ55nP" }, { "address":"1AC8nqvxTh8PHaMU5Fn8ZBhbodRasaRSRd" }, { "address":"1Gyi2AbGL2v5dbHogeezfi8om47c5q6EM8" } ], "Genesis Cosmos Address":"cosmos1fhejq5a3z8kehknrrjupt2wdl4l0wh9y2vayhr", "Recovery Cosmos Address":"cosmos12h5m3cuupza33psgzegjdvvsrngadjjmgnql34" }
	`, "IItkbg76ZjfFmMlyAmFAMz7CCt1FIw7VluWQbDvNhGdFZ6Y/FohkqlvXdv7aon8MGm6LbntbL3RxzR780PLYHY8=", sourceAddress, destAddress, types.NewInt64Coin("uatom", 31406121000)}
}

func BTCDonor3() FundRecoveryMessage {

	sourceAddress, _ := types.AccAddressFromBech32("cosmos1fhejq5a3z8kehknrrjupt2wdl4l0wh9y2vayhr")
	destAddress, _ := types.AccAddressFromBech32("cosmos12h5m3cuupza33psgzegjdvvsrngadjjmgnql34")

	return FundRecoveryMessage{`{ "Message":" I took a screenshot of the seed and due to my stupidity the only copy of the phrase. In Jan 2019 , I was installing Ubuntu on an external disk , by using my macbook as the medium, where I overrode the OS on my macbook itself instead of writing it to the external disk. The wipe replaced the entire OS and the Data Recovery company I gave it to were not able to retrieve the data. I lost atoms and some other crypto whose wallets were only on that macbook.", "Txid":"https://btc1.trezor.io/tx/ed6598ff038f1beaf1e1fe1c32aa50f770eaf16e6b7b86f45ed565837e546f05", "Contribution":"3 BTC", "Contributing Addresses":[ { "address":"16zuFhEMVTggK1rkMxwtsbBnSRH5NZ55nP" }, { "address":"1AC8nqvxTh8PHaMU5Fn8ZBhbodRasaRSRd" }, { "address":"1Gyi2AbGL2v5dbHogeezfi8om47c5q6EM8" } ], "Genesis Cosmos Address":"cosmos1fhejq5a3z8kehknrrjupt2wdl4l0wh9y2vayhr", "Recovery Cosmos Address":"cosmos12h5m3cuupza33psgzegjdvvsrngadjjmgnql34" }
	`, "IItkbg76ZjfFmMlyAmFAMz7CCt1FIw7VluWQbDvNhGdFZ6Y/FohkqlvXdv7aon8MGm6LbntbL3RxzR780PLYHY8=", sourceAddress, destAddress, types.NewInt64Coin("uatom", 31406121000)}
}

func verifyEthereumSignature(sig, msg, addr string) {
	sigBytes, _ := hexutil.Decode(sig)
	msgBytes, _ := hexutil.Decode(msg)
	addrBytes, _ := hexutil.Decode(addr)

	hash := crypto.Keccak256Hash(msgBytes)

	sigPublicKey, err := crypto.Ecrecover(hash.Bytes(), sigBytes)

	pubkey, _ := crypto.DecompressPubkey(sigPublicKey)

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
