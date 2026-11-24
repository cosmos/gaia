package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cometbft/cometbft/crypto/mldsa65"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"

	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v10/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"
	ibctesting "github.com/cosmos/ibc-go/v10/testing"

	"github.com/cosmos/gaia/v29/ante"
	gaiaApp "github.com/cosmos/gaia/v29/app"
)

const (
	// highMlDsaPower is larger than the combined power of the default
	// ibctesting validators (4 x 1). An ML-DSA validator with this power sorts
	// first in the validator set, its signature alone exceeds 2/3 of the voting
	// power, and the ed25519 validators alone cannot reach 1/3. The light client
	// therefore always verifies the ML-DSA signature.
	highMlDsaPower = 10

	// lowMlDsaPower lets the ed25519 validators (4 x 1) reach 2/3 of the voting
	// power without the ML-DSA validator.
	lowMlDsaPower = 1

	errWrongSignature = "wrong signature"
	errInvalidPubKey  = "invalid size for PubKeyMlDsa65"
	errCantTrust      = "cant trust new val set"
	errTooMuchChange  = "too much change from trusted validator set"
	errClientFrozen   = "with status Frozen"
)

// mutation changes a header before it is submitted.
type mutation func(header *ibctm.Header, vals *cmttypes.ValidatorSet)

// MLDSALightClientTestSuite checks a Gaia 07-tendermint client of a
// counterparty chain whose validator set includes ML-DSA-65 keys (CometBFT
// v0.38.26): client updates and misbehaviour handling.
type MLDSALightClientTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator
	chainA      *ibctesting.TestChain
	chainB      *ibctesting.TestChain
	path        *ibctesting.Path

	useFeeMarketDecorator bool
}

func TestMLDSALightClientTestSuite(t *testing.T) {
	suite.Run(t, new(MLDSALightClientTestSuite))
}

func (s *MLDSALightClientTestSuite) SetupSuite() {
	s.useFeeMarketDecorator = ante.UseFeeMarketDecorator
}

func (s *MLDSALightClientTestSuite) TearDownSuite() {
	ante.UseFeeMarketDecorator = s.useFeeMarketDecorator
}

func (s *MLDSALightClientTestSuite) SetupTest() {
	// The feemarket ante decorator rejects the zero-fee txs sent by ibctesting.
	ante.UseFeeMarketDecorator = false

	ibctesting.DefaultTestingAppInit = GaiaAppIniter
	s.coordinator = ibctesting.NewCoordinator(s.T(), 2)
	s.chainA = s.coordinator.GetChain(ibctesting.GetChainID(1))
	s.chainB = s.coordinator.GetChain(ibctesting.GetChainID(2))

	s.path = ibctesting.NewPath(s.chainA, s.chainB)
	s.path.SetupClients()
}

// SetupSubTest gives every s.Run subtest a fresh pair of chains.
func (s *MLDSALightClientTestSuite) SetupSubTest() {
	s.SetupTest()
}

func (s *MLDSALightClientTestSuite) TestUpdateClientMixedEd25519AndMlDsa() {
	nextVals, nextSigners := s.counterpartyValSet(true, 1, highMlDsaPower)

	height := s.rotateAndUpdate(nextVals, nextSigners, nil, "")
	s.requireClientAt(height, nextVals)
}

func (s *MLDSALightClientTestSuite) TestUpdateClientMlDsaOnly() {
	nextVals, nextSigners := s.counterpartyValSet(false, 4, 1)

	height := s.rotateAndUpdate(nextVals, nextSigners, nil, "")
	s.requireClientAt(height, nextVals)
}

func (s *MLDSALightClientTestSuite) TestUpdateClientRejectsCorruptedMlDsaSignature() {
	nextVals, nextSigners := s.counterpartyValSet(true, 1, highMlDsaPower)

	height := s.rotateAndUpdate(nextVals, nextSigners, corruptMlDsaSignatures, errWrongSignature)
	s.requireClientAt(height, nextVals)
}

// A counterparty often starts with a single low-power ML-DSA validator. Its
// public key must be decoded even when it did not sign the header.
func (s *MLDSALightClientTestSuite) TestUpdateClientAbsentLowPowerMlDsaValidator() {
	nextVals, nextSigners := s.counterpartyValSet(true, 1, lowMlDsaPower)

	height := s.rotateAndUpdate(nextVals, nextSigners, omitMlDsaSignatures, "")
	s.requireClientAt(height, nextVals)
}

func (s *MLDSALightClientTestSuite) TestUpdateClientNonAdjacentMixedEd25519AndMlDsa() {
	nextVals, nextSigners := s.counterpartyValSet(true, 1, highMlDsaPower)

	height := s.updateNonAdjacent(s.chainB.NextVals, nextVals, nextSigners, nil, "")
	s.requireClientAt(height, nextVals)
}

func (s *MLDSALightClientTestSuite) TestUpdateClientNonAdjacentRejectsCorruptedMlDsaSignature() {
	trusted := s.clientState().LatestHeight
	nextVals, nextSigners := s.counterpartyValSet(true, 1, highMlDsaPower)

	height := s.updateNonAdjacent(s.chainB.NextVals, nextVals, nextSigners, corruptMlDsaSignatures, errWrongSignature)
	s.Require().Equal(trusted, height)
	s.requireClientAt(height, s.chainB.NextVals)
}

func (s *MLDSALightClientTestSuite) TestUpdateClientNonAdjacentAbsentLowPowerMlDsaValidator() {
	nextVals, nextSigners := s.counterpartyValSet(true, 1, lowMlDsaPower)

	height := s.updateNonAdjacent(s.chainB.NextVals, nextVals, nextSigners, omitMlDsaSignatures, "")
	s.requireClientAt(height, nextVals)
}

// After the client trusts a validator set with ML-DSA keys, a skipping update
// needs the ML-DSA signature to reach 1/3 of the trusted voting power.
//
// The trusted set gives the ML-DSA validator 10 of 14 voting power. The new set
// has the same keys with the power moved to the ed25519 validators (4 x 10 and
// ML-DSA 1), so the new set reaches 2/3 without the ML-DSA signature. A missing
// or corrupted ML-DSA signature is then rejected only by the 1/3 check against
// the trusted set.
func (s *MLDSALightClientTestSuite) TestUpdateClientNonAdjacentFromMlDsaTrustedSet() {
	trustedVals, signers := s.counterpartyValSet(true, 1, highMlDsaPower)
	s.rotateAndUpdate(trustedVals, signers, nil, "")
	newVals := shiftPowerToEd25519(trustedVals)

	height := s.updateNonAdjacent(trustedVals, newVals, signers, nil, "")
	s.requireClientAt(height, newVals)
}

func (s *MLDSALightClientTestSuite) TestUpdateClientNonAdjacentFromMlDsaTrustedSetRequiresMlDsaSignature() {
	trustedVals, signers := s.counterpartyValSet(true, 1, highMlDsaPower)
	trusted := s.rotateAndUpdate(trustedVals, signers, nil, "")
	newVals := shiftPowerToEd25519(trustedVals)

	height := s.updateNonAdjacent(trustedVals, newVals, signers, omitMlDsaSignatures, errCantTrust)
	s.Require().Equal(trusted, height)
	s.requireClientAt(height, trustedVals)
}

func (s *MLDSALightClientTestSuite) TestUpdateClientNonAdjacentFromMlDsaTrustedSetRejectsCorruptedMlDsaSignature() {
	trustedVals, signers := s.counterpartyValSet(true, 1, highMlDsaPower)
	trusted := s.rotateAndUpdate(trustedVals, signers, nil, "")
	newVals := shiftPowerToEd25519(trustedVals)

	height := s.updateNonAdjacent(trustedVals, newVals, signers, corruptMlDsaSignatures, errWrongSignature)
	s.Require().Equal(trusted, height)
	s.requireClientAt(height, trustedVals)
}

// Rejected headers must not change or freeze the client. A valid header for the
// same height is accepted afterwards.
func (s *MLDSALightClientTestSuite) TestUpdateClientRecoversAfterRejectedHeaders() {
	nextVals, nextSigners := s.counterpartyValSet(true, 1, highMlDsaPower)
	h1 := s.announceValSet(nextVals)
	h2 := h1.Increment().(clienttypes.Height)

	for _, bad := range []struct {
		mutate  mutation
		wantErr string
	}{
		{corruptMlDsaSignatures, errWrongSignature},
		{truncateMlDsaPubKeys, errInvalidPubKey},
	} {
		height := s.sendUpdate(s.adjacentHeader(h1, nextVals, nextSigners), nextVals, bad.mutate, bad.wantErr, h1, h2)
		s.Require().Equal(h1, height)

		_, found := s.chainA.GetConsensusState(s.path.EndpointA.ClientID, h2)
		s.Require().False(found, "a rejected header must not store a consensus state")
		s.requireClientAt(h1, nextVals)
	}

	height := s.sendUpdate(s.adjacentHeader(h1, nextVals, nextSigners), nextVals, nil, "", h1, h2)
	s.requireClientAt(height, nextVals)
}

// A valid fork with ML-DSA-65 signatures freezes the client, and the frozen
// client rejects later updates.
func (s *MLDSALightClientTestSuite) TestMisbehaviourFreezesClient() {
	m, trustedVals, newVals, signers := s.mlDsaMisbehaviour()

	s.Require().NoError(s.submitClientMessage(m))
	s.Require().Equal(exported.Frozen, s.clientStatus())

	trusted := s.clientState().LatestHeight
	s.coordinator.IncrementTime()
	header := s.chainB.CreateTMClientHeader(
		s.chainB.ChainID, int64(trusted.RevisionHeight)+10, trusted, s.coordinator.CurrentTime, //nolint:gosec
		newVals, newVals, trustedVals, signers,
	)
	s.Require().ErrorContains(s.submitClientMessage(header), errClientFrozen)
	s.Require().Equal(exported.Frozen, s.clientStatus())
}

// Forged evidence is rejected by the 1/3 check against the trusted validator
// set and leaves the client active. Each header of the misbehaviour is checked.
func (s *MLDSALightClientTestSuite) TestMisbehaviourRejectsForgedMlDsaSignatures() {
	for _, tc := range []struct {
		name    string
		mutate  mutation
		wantErr string
	}{
		{"corrupted", corruptMlDsaSignatures, errWrongSignature},
		{"missing", omitMlDsaSignatures, errTooMuchChange},
	} {
		for _, header := range []int{1, 2} {
			s.Run(fmt.Sprintf("%s ML-DSA signature in header %d", tc.name, header), func() {
				m, _, newVals, _ := s.mlDsaMisbehaviour()
				target := m.Header1
				if header == 2 {
					target = m.Header2
				}
				tc.mutate(target, newVals)

				s.Require().ErrorContains(s.submitClientMessage(m), tc.wantErr)
				s.Require().Equal(exported.Active, s.clientStatus())
			})
		}
	}
}

// A second valid header for a height the client already stores is detected as
// misbehaviour by UpdateClient and freezes the client.
func (s *MLDSALightClientTestSuite) TestUpdateClientConflictingMlDsaHeaderFreezesClient() {
	vals, signers := s.counterpartyValSet(true, 1, highMlDsaPower)
	h1 := s.announceValSet(vals)
	h2 := s.sendUpdate(s.adjacentHeader(h1, vals, signers), vals, nil, "", h1, h1.Increment().(clienttypes.Height))
	s.requireClientAt(h2, vals)

	// adjacentHeader uses a later timestamp, so the header hash differs.
	s.Require().NoError(s.submitClientMessage(s.adjacentHeader(h1, vals, signers)))
	s.Require().Equal(exported.Frozen, s.clientStatus())
}

// mlDsaMisbehaviour rotates chainB to a trusted validator set in which the
// ML-DSA validator holds 10 of 14 voting power, then returns a fork: two
// different headers for the same height, signed by a new set with the same
// keys and the power moved to the ed25519 validators (4 x 10, ML-DSA 1).
//
// Misbehaviour.ValidateBasic checks that each header's own validator set
// signed more than 2/3, which the new set reaches without the ML-DSA
// signature. A missing or corrupted ML-DSA signature is then rejected only by
// the 1/3 check against the trusted set.
func (s *MLDSALightClientTestSuite) mlDsaMisbehaviour() (*ibctm.Misbehaviour, *cmttypes.ValidatorSet, *cmttypes.ValidatorSet, map[string]cmttypes.PrivValidator) {
	trustedVals, signers := s.counterpartyValSet(true, 1, highMlDsaPower)
	trusted := s.rotateAndUpdate(trustedVals, signers, nil, "")
	newVals := shiftPowerToEd25519(trustedVals)
	height := int64(trusted.RevisionHeight) + 5 //nolint:gosec

	s.coordinator.IncrementTime()
	header1 := s.chainB.CreateTMClientHeader(
		s.chainB.ChainID, height, trusted, s.coordinator.CurrentTime,
		newVals, newVals, trustedVals, signers,
	)
	header2 := s.chainB.CreateTMClientHeader(
		s.chainB.ChainID, height, trusted, s.coordinator.CurrentTime.Add(-time.Second),
		newVals, newVals, trustedVals, signers,
	)

	return ibctm.NewMisbehaviour(s.path.EndpointA.ClientID, header1, header2), trustedVals, newVals, signers
}

// submitClientMessage submits clientMsg to chainA in a MsgUpdateClient and
// returns the tx error. It does not check the client status.
func (s *MLDSALightClientTestSuite) submitClientMessage(clientMsg exported.ClientMessage) error {
	_, err := s.chainA.SendMsgs(s.updateClientMsg(clientMsg))
	if err != nil {
		s.syncSenderSequence()
	}
	return err
}

func (s *MLDSALightClientTestSuite) clientStatus() exported.Status {
	return s.chainA.App.GetIBCKeeper().ClientKeeper.GetClientStatus(s.chainA.GetContext(), s.path.EndpointA.ClientID)
}

// counterpartyValSet returns chainB's validator set after a key rotation. If
// keepEd25519 is set, the current ed25519 validators of chainB stay in the set.
// numMlDsa ML-DSA-65 validators with mlDsaPower each are added.
func (s *MLDSALightClientTestSuite) counterpartyValSet(keepEd25519 bool, numMlDsa int, mlDsaPower int64) (*cmttypes.ValidatorSet, map[string]cmttypes.PrivValidator) {
	var vals []*cmttypes.Validator
	signers := make(map[string]cmttypes.PrivValidator)

	if keepEd25519 {
		for _, v := range s.chainB.Vals.Validators {
			vals = append(vals, cmttypes.NewValidator(v.PubKey, v.VotingPower))
			signers[v.Address.String()] = s.chainB.Signers[v.Address.String()]
		}
	}

	for range numMlDsa {
		privKey, err := mldsa65.GenPrivKey()
		s.Require().NoError(err)
		pv := cmttypes.NewMockPVWithParams(privKey, false, false)
		pubKey, err := pv.GetPubKey()
		s.Require().NoError(err)
		vals = append(vals, cmttypes.NewValidator(pubKey, mlDsaPower))
		signers[pubKey.Address().String()] = pv
	}

	return cmttypes.NewValidatorSet(vals), signers
}

// shiftPowerToEd25519 returns a validator set with the same keys as vals, in
// which every ed25519 validator has highMlDsaPower and every ML-DSA validator
// lowMlDsaPower.
func shiftPowerToEd25519(vals *cmttypes.ValidatorSet) *cmttypes.ValidatorSet {
	shifted := make([]*cmttypes.Validator, 0, len(vals.Validators))
	for _, v := range vals.Validators {
		power := int64(highMlDsaPower)
		if v.PubKey.Type() == mldsa65.KeyType {
			power = lowMlDsaPower
		}
		shifted = append(shifted, cmttypes.NewValidator(v.PubKey, power))
	}
	return cmttypes.NewValidatorSet(shifted)
}

// rotateAndUpdate updates chainA's client of chainB in two steps, as a relayer
// would after chainB rotates its validator set:
//  1. h1, signed by chainB's current ed25519 validators, announces nextVals.
//  2. h2, signed by nextVals, is verified against the trusted nextVals of h1.
//
// mutate, if set, changes h2 before it is submitted. If wantErr is set, the
// update of h2 must fail with it. It returns the latest height the client is
// expected to have.
func (s *MLDSALightClientTestSuite) rotateAndUpdate(nextVals *cmttypes.ValidatorSet, nextSigners map[string]cmttypes.PrivValidator, mutate mutation, wantErr string) clienttypes.Height {
	h1 := s.announceValSet(nextVals)
	h2 := h1.Increment().(clienttypes.Height)

	return s.sendUpdate(s.adjacentHeader(h1, nextVals, nextSigners), nextVals, mutate, wantErr, h1, h2)
}

// announceValSet updates the client with a header signed by chainB's current
// ed25519 validators whose NextValidatorsHash commits to nextVals. It returns
// the new client height.
func (s *MLDSALightClientTestSuite) announceValSet(nextVals *cmttypes.ValidatorSet) clienttypes.Height {
	trusted := s.clientState().LatestHeight
	h1 := trusted.Increment().(clienttypes.Height)

	s.coordinator.IncrementTime()
	header := s.chainB.CreateTMClientHeader(
		s.chainB.ChainID, int64(h1.RevisionHeight), trusted, s.coordinator.CurrentTime, //nolint:gosec
		s.chainB.Vals, nextVals, s.chainB.NextVals, s.chainB.Signers,
	)
	_, err := s.chainA.SendMsgs(s.updateClientMsg(header))
	s.Require().NoError(err, "update announcing the ML-DSA validator set must succeed")

	return h1
}

// adjacentHeader returns a header at trusted+1 signed by vals, where vals is
// also the trusted next validator set.
func (s *MLDSALightClientTestSuite) adjacentHeader(trusted clienttypes.Height, vals *cmttypes.ValidatorSet, signers map[string]cmttypes.PrivValidator) *ibctm.Header {
	height := trusted.Increment().(clienttypes.Height)

	s.coordinator.IncrementTime()
	return s.chainB.CreateTMClientHeader(
		s.chainB.ChainID, int64(height.RevisionHeight), trusted, s.coordinator.CurrentTime, //nolint:gosec
		vals, vals, vals, signers,
	)
}

// updateNonAdjacent updates chainA's client of chainB with a header several
// heights above the trusted height, signed by nextVals. The light client checks
// that trustedVals signed more than 1/3 of the trusted voting power and that
// nextVals signed more than 2/3. This is the skipping verification relayers
// commonly use. trustedVals must be the next validator set of the trusted
// height. It returns the latest height the client is expected to have.
func (s *MLDSALightClientTestSuite) updateNonAdjacent(trustedVals, nextVals *cmttypes.ValidatorSet, nextSigners map[string]cmttypes.PrivValidator, mutate mutation, wantErr string) clienttypes.Height {
	trusted := s.clientState().LatestHeight
	height := clienttypes.NewHeight(trusted.RevisionNumber, trusted.RevisionHeight+5)

	s.coordinator.IncrementTime()
	header := s.chainB.CreateTMClientHeader(
		s.chainB.ChainID, int64(height.RevisionHeight), trusted, s.coordinator.CurrentTime, //nolint:gosec
		nextVals, nextVals, trustedVals, nextSigners,
	)
	return s.sendUpdate(header, nextVals, mutate, wantErr, trusted, height)
}

// sendUpdate applies mutate to header, if set, and submits it to chainA. If
// wantErr is set, the update must fail with it and the client stays at
// prevHeight. Otherwise the update must succeed and the client moves to height.
func (s *MLDSALightClientTestSuite) sendUpdate(header *ibctm.Header, vals *cmttypes.ValidatorSet, mutate mutation, wantErr string, prevHeight, height clienttypes.Height) clienttypes.Height {
	if mutate != nil {
		mutate(header, vals)
	}

	_, err := s.chainA.SendMsgs(s.updateClientMsg(header))

	s.Require().Equal(exported.Active, s.clientStatus())

	if wantErr != "" {
		s.Require().ErrorContains(err, wantErr)
		s.syncSenderSequence()
		return prevHeight
	}
	s.Require().NoError(err, "update signed by the ML-DSA validator set must succeed")
	return height
}

// syncSenderSequence sets the sender sequence to the on-chain value.
// ibctesting increments it after every tx, but a tx rejected before the ante
// handler, such as one with a malformed public key, does not use a sequence.
func (s *MLDSALightClientTestSuite) syncSenderSequence() {
	app, ok := s.chainA.App.(*gaiaApp.GaiaApp)
	s.Require().True(ok)
	acc := app.AccountKeeper.GetAccount(s.chainA.GetContext(), s.chainA.SenderAccount.GetAddress())
	s.Require().NoError(s.chainA.SenderAccount.SetSequence(acc.GetSequence()))
}

// corruptMlDsaSignatures flips a byte of every ML-DSA signature in the commit.
func corruptMlDsaSignatures(header *ibctm.Header, vals *cmttypes.ValidatorSet) {
	for i, v := range vals.Validators {
		if v.PubKey.Type() == mldsa65.KeyType {
			header.Commit.Signatures[i].Signature[0] ^= 0xff
		}
	}
}

// omitMlDsaSignatures marks every ML-DSA validator as absent from the commit.
func omitMlDsaSignatures(header *ibctm.Header, vals *cmttypes.ValidatorSet) {
	for i, v := range vals.Validators {
		if v.PubKey.Type() == mldsa65.KeyType {
			header.Commit.Signatures[i] = cmtproto.CommitSig{BlockIdFlag: cmtproto.BlockIDFlagAbsent}
		}
	}
}

// truncateMlDsaPubKeys drops the last byte of every ML-DSA public key in the
// header's validator set.
func truncateMlDsaPubKeys(header *ibctm.Header, _ *cmttypes.ValidatorSet) {
	for _, v := range header.ValidatorSet.Validators {
		if pk, ok := v.PubKey.Sum.(*cmtprotocrypto.PublicKey_Mldsa65); ok {
			pk.Mldsa65 = pk.Mldsa65[:len(pk.Mldsa65)-1]
		}
	}
}

// updateClientMsg wraps clientMsg (a header or misbehaviour) in a
// MsgUpdateClient and round-trips it through Gaia's tx encoder and decoder, as
// a relayer-submitted tx would be. CometBFT v0.38.25 fails this decode because
// of the unknown mldsa65 public key field.
func (s *MLDSALightClientTestSuite) updateClientMsg(clientMsg exported.ClientMessage) *clienttypes.MsgUpdateClient {
	msg, err := clienttypes.NewMsgUpdateClient(
		s.path.EndpointA.ClientID, clientMsg, s.chainA.SenderAccount.GetAddress().String(),
	)
	s.Require().NoError(err)

	txBuilder := s.chainA.TxConfig.NewTxBuilder()
	s.Require().NoError(txBuilder.SetMsgs(msg))
	txBytes, err := s.chainA.TxConfig.TxEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)
	tx, err := s.chainA.TxConfig.TxDecoder()(txBytes)
	s.Require().NoError(err)

	decoded, ok := tx.GetMsgs()[0].(*clienttypes.MsgUpdateClient)
	s.Require().True(ok)
	decodedClientMsg, err := clienttypes.UnpackClientMessage(decoded.ClientMessage)
	s.Require().NoError(err)

	want, err := s.chainA.Codec.MarshalInterface(clientMsg)
	s.Require().NoError(err)
	got, err := s.chainA.Codec.MarshalInterface(decodedClientMsg)
	s.Require().NoError(err)
	s.Require().Equal(want, got, "the client message must survive the tx round trip")

	return decoded
}

func (s *MLDSALightClientTestSuite) clientState() *ibctm.ClientState {
	clientState, ok := s.chainA.GetClientState(s.path.EndpointA.ClientID).(*ibctm.ClientState)
	s.Require().True(ok)
	return clientState
}

// requireClientAt checks the client's latest height and that the consensus
// state at that height commits to nextVals.
func (s *MLDSALightClientTestSuite) requireClientAt(height clienttypes.Height, nextVals *cmttypes.ValidatorSet) {
	s.Require().Equal(height, s.clientState().LatestHeight)

	consState, found := s.chainA.GetConsensusState(s.path.EndpointA.ClientID, height)
	s.Require().True(found)
	tmConsState, ok := consState.(*ibctm.ConsensusState)
	s.Require().True(ok)
	s.Require().Equal(nextVals.Hash(), tmConsState.NextValidatorsHash.Bytes())
}
