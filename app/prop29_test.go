package gaia

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVerifySignatures(t *testing.T) {
	//Bitcoin Donors
	bDonor1 := btcDonor1()
	bDonor2 := btcDonor2()
	bDonor3 := btcDonor3()
	bDonor4 := btcDonor4()

	//Bitcoin Verification
	_, err := bDonor1.verifyBitcoinSignature(0)
	require.NoError(t, err, "bitcoin donor 1")

	_, err = bDonor2.verifyBitcoinSignature(0)
	require.NoError(t, err, "bitcoin donor 2")

	_, err = bDonor3.verifyBitcoinSignature(1)
	require.NoError(t, err, "bitcoin donor 3")

	_, err = bDonor4.verifyBitcoinSignature(0)
	require.NoError(t, err, "bitcoin donor 4")

	//Ethereum Donors
	eDonor1 := ethDonor1()
	eDonor2 := ethDonor2()
	eDonor3 := ethDonor3()

	//Ethereum Verification
	_, err = eDonor1.verifyEthereumSignature(0)
	require.NoError(t, err, "ethereum donor 1")

	_, err = eDonor2.verifyEthereumSignature(0)
	require.NoError(t, err, "ethereum donor 2")

	_, err = eDonor3.verifyEthereumSignature(0)
	require.NoError(t, err, "ethereum donor 3")
}
