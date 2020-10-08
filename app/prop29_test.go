package app

import "testing"

func TestVerifyBitcoinSignature(t *testing.T) {
	donor := BTCDonor1()

	donor.VerifyBitcoinSignature(0)
}
