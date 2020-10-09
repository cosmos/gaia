package app

import "testing"

func TestVerifySignatures(t *testing.T) {
	//Bitcoin Donors
	donor1 := BTCDonor1()
	donor2 := BTCDonor2()
	donor3 := BTCDonor3()
	donor4 := BTCDonor4()

	//Bitcoin Verification
	donor1.VerifyBitcoinSignature(0)
	donor2.VerifyBitcoinSignature(0)
	donor3.VerifyBitcoinSignature(1)
	donor4.VerifyBitcoinSignature(0)

	//Ethereum Donors
	eDonor1 := EthDonor1()

	//Ethereum Verification
	eDonor1.VerifyEthereumSignature(0)

}
