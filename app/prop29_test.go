package app

import "testing"

func TestVerifySignatures(t *testing.T) {
	//Bitcoin Donors
	bDonor1 := BTCDonor1()
	bDonor2 := BTCDonor2()
	bDonor3 := BTCDonor3()
	bDonor4 := BTCDonor4()

	//Bitcoin Verification
	bDonor1.VerifyBitcoinSignature(0)
	bDonor2.VerifyBitcoinSignature(0)
	bDonor3.VerifyBitcoinSignature(1)
	bDonor4.VerifyBitcoinSignature(0)

	//Ethereum Donors
	eDonor1 := EthDonor1()
	eDonor2 := EthDonor2()
	eDonor3 := EthDonor3()

	//Ethereum Verification
	eDonor1.VerifyEthereumSignature(0)
	eDonor2.VerifyEthereumSignature(0)
	eDonor3.VerifyEthereumSignature(0)

}
