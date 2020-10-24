package app

import "testing"

func TestVerifySignatures(t *testing.T) {
	//Bitcoin Donors
	bDonor1 := btcDonor1()
	bDonor2 := btcDonor2()
	bDonor3 := btcDonor3()
	bDonor4 := btcDonor4()

	//Bitcoin Verification
	bDonor1.verifyBitcoinSignature(0)
	bDonor2.verifyBitcoinSignature(0)
	bDonor3.verifyBitcoinSignature(1)
	bDonor4.verifyBitcoinSignature(0)

	//Ethereum Donors
	eDonor1 := ethDonor1()
	eDonor2 := ethDonor2()
	eDonor3 := ethDonor3()

	//Ethereum Verification
	eDonor1.verifyEthereumSignature(0)
	eDonor2.verifyEthereumSignature(0)
	eDonor3.verifyEthereumSignature(0)

}
