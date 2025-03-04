package operator

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	abi "github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/cosmos/cosmos-sdk/codec"

	tmclient "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"

	"github.com/cosmos/solidity-ibc-eureka/abigen/sp1ics07tendermint"

	"github.com/srdtrk/solidity-ibc-eureka/e2e/v8/testvalues"
)

type GenesisFixture struct {
	TrustedClientState    string `json:"trustedClientState"`
	TrustedConsensusState string `json:"trustedConsensusState"`
	UpdateClientVkey      string `json:"updateClientVkey"`
	MembershipVkey        string `json:"membershipVkey"`
	UcAndMembershipVkey   string `json:"ucAndMembershipVkey"`
	MisbehaviourVKey      string `json:"misbehaviourVkey"`
}

// membershipFixture is a struct that contains the membership proof and proof height
type membershipFixture struct {
	GenesisFixture
	// hex encoded height
	ProofHeight string `json:"proofHeight"`
	// hex encoded proof
	MembershipProof string `json:"membershipProof"`
}

type misbehaviourFixture struct {
	GenesisFixture
	SubmitMsg string `json:"submitMsg"`
}

// binaryPath is a function that returns the path to the operator binary
func binaryPath() string {
	return "operator"
}

// RunGenesis is a function that runs the genesis script to generate genesis.json
func RunGenesis(args ...string) error {
	args = append([]string{"genesis"}, args...)
	cmd := exec.Command(binaryPath(), args...)
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

// StartOperator is a function that runs the operator
func StartOperator(args ...string) error {
	args = append([]string{"start"}, args...)
	cmd := exec.Command(binaryPath(), args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// MembershipProof is a function that generates a membership proof and returns the proof height and proof
func MembershipProof(trusted_height uint64, paths string, writeFixtureName string, args ...string) (*sp1ics07tendermint.IICS02ClientMsgsHeight, []byte, error) {
	args = append([]string{"fixtures", "membership", "--trusted-block", strconv.FormatUint(trusted_height, 10), "--key-paths", paths}, args...)

	cmd := exec.Command(binaryPath(), args...)
	output, err := execOperatorCommand(cmd)
	if err != nil {
		return nil, nil, err
	}

	// eliminate non-json characters
	jsonStartIdx := strings.Index(string(output), "{")
	if jsonStartIdx == -1 {
		panic("no json found in output")
	}
	output = output[jsonStartIdx:]

	if writeFixtureName != "" {
		fixtureFileName := fmt.Sprintf("%s/%s_fixture.json", testvalues.SP1ICS07FixturesDir, writeFixtureName)
		if err := os.WriteFile(fixtureFileName, output, 0o600); err != nil {
			return nil, nil, err
		}
	}

	var membership membershipFixture
	err = json.Unmarshal(output, &membership)
	if err != nil {
		return nil, nil, err
	}

	heightBz, err := hex.DecodeString(membership.ProofHeight)
	if err != nil {
		return nil, nil, err
	}

	heightType, err := abi.NewType("tuple", "IICS02ClientMsgsHeight", []abi.ArgumentMarshaling{
		{Name: "revisionNumber", Type: "uint32"},
		{Name: "revisionHeight", Type: "uint32"},
	})
	if err != nil {
		return nil, nil, err
	}

	heightArgs := abi.Arguments{
		{Type: heightType, Name: "param_one"},
	}

	// abi encoding
	heightI, err := heightArgs.Unpack(heightBz)
	if err != nil {
		return nil, nil, err
	}

	height := abi.ConvertType(heightI[0], new(sp1ics07tendermint.IICS02ClientMsgsHeight)).(*sp1ics07tendermint.IICS02ClientMsgsHeight)

	if height.RevisionHeight != uint32(trusted_height) {
		return nil, nil, errors.New("heights do not match")
	}

	proofBz, err := hex.DecodeString(membership.MembershipProof)
	if err != nil {
		return nil, nil, err
	}

	return height, proofBz, nil
}

// UpdateClientAndMembershipProof is a function that generates an update client and membership proof
func UpdateClientAndMembershipProof(trusted_height, target_height uint64, paths string, args ...string) (*sp1ics07tendermint.IICS02ClientMsgsHeight, []byte, error) {
	args = append([]string{"fixtures", "update-client-and-membership", "--trusted-block", strconv.FormatUint(trusted_height, 10), "--target-block", strconv.FormatUint(target_height, 10), "--key-paths", paths}, args...)

	output, err := execOperatorCommand(exec.Command(binaryPath(), args...))
	if err != nil {
		return nil, nil, err
	}

	// eliminate non-json characters
	jsonStartIdx := strings.Index(string(output), "{")
	if jsonStartIdx == -1 {
		panic("no json found in output")
	}
	output = output[jsonStartIdx:]

	var membership membershipFixture
	err = json.Unmarshal(output, &membership)
	if err != nil {
		return nil, nil, err
	}

	heightBz, err := hex.DecodeString(membership.ProofHeight)
	if err != nil {
		return nil, nil, err
	}

	heightType, err := abi.NewType("tuple", "IICS02ClientMsgsHeight", []abi.ArgumentMarshaling{
		{Name: "revisionNumber", Type: "uint32"},
		{Name: "revisionHeight", Type: "uint32"},
	})
	if err != nil {
		return nil, nil, err
	}

	heightArgs := abi.Arguments{
		{Type: heightType, Name: "param_one"},
	}

	// abi encoding
	heightI, err := heightArgs.Unpack(heightBz)
	if err != nil {
		return nil, nil, err
	}

	height := abi.ConvertType(heightI[0], new(sp1ics07tendermint.IICS02ClientMsgsHeight)).(*sp1ics07tendermint.IICS02ClientMsgsHeight)

	if height.RevisionHeight != uint32(target_height) {
		return nil, nil, errors.New("heights do not match")
	}

	proofBz, err := hex.DecodeString(membership.MembershipProof)
	if err != nil {
		return nil, nil, err
	}

	return height, proofBz, nil
}

// MisbehaviourProof is a function that generates a misbehaviour proof and returns the submit message
func MisbehaviourProof(cdc codec.Codec, misbehaviour tmclient.Misbehaviour, writeFixtureName string, args ...string) ([]byte, error) {
	misbehaviourBz, err := marshalMisbehaviour(cdc, misbehaviour)
	if err != nil {
		return nil, err
	}

	// write misbehaviour to file for the operator to use
	misbehaviourFileName := "misbehaviour.json"
	if err := os.WriteFile(misbehaviourFileName, misbehaviourBz, 0o600); err != nil {
		return nil, err
	}
	defer os.Remove(misbehaviourFileName)

	args = append([]string{"fixtures", "misbehaviour", "--misbehaviour-path", misbehaviourFileName}, args...)
	output, err := execOperatorCommand(exec.Command(binaryPath(), args...))
	if err != nil {
		return nil, err
	}

	// eliminate non-json characters
	jsonStartIdx := strings.Index(string(output), "{")
	if jsonStartIdx == -1 {
		panic("no json found in output")
	}
	output = output[jsonStartIdx:]

	var misbehaviourFixture misbehaviourFixture
	err = json.Unmarshal(output, &misbehaviourFixture)
	if err != nil {
		return nil, err
	}

	if writeFixtureName != "" {
		fixtureFileName := fmt.Sprintf("%s/misbehaviour_%s_fixture.json", testvalues.SP1ICS07FixturesDir, writeFixtureName)
		if err := os.WriteFile(fixtureFileName, output, 0o600); err != nil {
			return nil, err
		}
	}

	submitMsgBz, err := hex.DecodeString(misbehaviourFixture.SubmitMsg)
	if err != nil {
		return nil, err
	}

	return submitMsgBz, nil
}

// ToBase64KeyPaths is a function that takes a list of key paths and returns a base64 encoded string
// that the operator can use to generate a membership proof
func ToBase64KeyPaths(paths ...[][]byte) string {
	var keyPaths []string
	for _, path := range paths {
		if len(path) != 2 {
			panic("path must have 2 elements")
		}
		keyPaths = append(keyPaths, base64.StdEncoding.EncodeToString(path[0])+"\\"+base64.StdEncoding.EncodeToString(path[1]))
	}
	return strings.Join(keyPaths, ",")
}

// TODO: This is a mighty ugly piece of code. Hopefully there is a better way to do this.
// marshalMisbehaviour takes a MisbehaviourProof struct and marshals it into a JSON byte slice that can be unmarshalled by the operator.
// It first marshals to JSON directly, and then modifies all the incompatible types (mostly base64 encoded bytes) to be hex encoded.
// Ideally, we can update the types in the operator to be more compatible with the type we have here.
// It might be enough to get out a new version of the rust crate "ibc-proto" and update the operator to use it.
func marshalMisbehaviour(cdc codec.Codec, misbehaviour tmclient.Misbehaviour) ([]byte, error) {
	misbehaviour.ClientId = "07-tendermint-0" // We just have to set it to something to make the unmarshalling to work :P
	bzIntermediary, err := cdc.MarshalJSON(&misbehaviour)
	if err != nil {
		return nil, err
	}
	var jsonIntermediary map[string]interface{}
	if err := json.Unmarshal(bzIntermediary, &jsonIntermediary); err != nil {
		return nil, err
	}
	headerHexPaths := []string{
		"validator_set.proposer.address",
		"trusted_validators.proposer.address",
		"signed_header.header.last_block_id.hash",
		"signed_header.header.last_block_id.part_set_header.hash",
		"signed_header.header.app_hash",
		"signed_header.header.consensus_hash",
		"signed_header.header.data_hash",
		"signed_header.header.evidence_hash",
		"signed_header.header.last_commit_hash",
		"signed_header.header.last_results_hash",
		"signed_header.header.next_validators_hash",
		"signed_header.header.proposer_address",
		"signed_header.header.validators_hash",
		"signed_header.commit.block_id.hash",
		"signed_header.commit.block_id.part_set_header.hash",
	}

	var hexPaths []string
	for _, path := range headerHexPaths {
		hexPaths = append(hexPaths, "header_1."+path)
		hexPaths = append(hexPaths, "header_2."+path)
	}

	for _, path := range hexPaths {
		pathParts := strings.Split(path, ".")
		tmpIntermediary := jsonIntermediary
		for i := 0; i < len(pathParts)-1; i++ {
			var ok bool
			tmpIntermediary, ok = tmpIntermediary[pathParts[i]].(map[string]interface{})
			if !ok {
				fmt.Printf("path not found: %s\n", path)
				continue
			}
		}
		base64str, ok := tmpIntermediary[pathParts[len(pathParts)-1]].(string)
		if !ok {
			return nil, fmt.Errorf("path not found: %s", path)
		}
		bz, err := base64.StdEncoding.DecodeString(base64str)
		if err != nil {
			return nil, err
		}
		tmpIntermediary[pathParts[len(pathParts)-1]] = hex.EncodeToString(bz)
	}

	validators1 := jsonIntermediary["header_1"].(map[string]interface{})["validator_set"].(map[string]interface{})["validators"].([]interface{})
	validators2 := jsonIntermediary["header_2"].(map[string]interface{})["validator_set"].(map[string]interface{})["validators"].([]interface{})
	trustedValidators1 := jsonIntermediary["header_1"].(map[string]interface{})["trusted_validators"].(map[string]interface{})["validators"].([]interface{})
	trustedValidators2 := jsonIntermediary["header_2"].(map[string]interface{})["trusted_validators"].(map[string]interface{})["validators"].([]interface{})
	validators := validators1
	validators = append(validators, validators2...)
	validators = append(validators, trustedValidators1...)
	validators = append(validators, trustedValidators2...)
	for _, val := range validators {
		val := val.(map[string]interface{})
		valAddressBase64Str, ok := val["address"].(string)
		if !ok {
			return nil, fmt.Errorf("address not found in path: %s", val)
		}
		valAddressBz, err := base64.StdEncoding.DecodeString(valAddressBase64Str)
		if err != nil {
			return nil, err
		}
		val["address"] = hex.EncodeToString(valAddressBz)

		pubKey, ok := val["pub_key"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("pub_key not found in path: %s", val)
		}
		ed25519PubKey := pubKey["ed25519"].(string)
		pubKey["type"] = "tendermint/PubKeyEd25519"
		pubKey["value"] = ed25519PubKey
	}

	var pubKeys []map[string]interface{}
	pubKeys = append(pubKeys, jsonIntermediary["header_1"].(map[string]interface{})["validator_set"].(map[string]interface{})["proposer"].(map[string]interface{})["pub_key"].(map[string]interface{}))
	pubKeys = append(pubKeys, jsonIntermediary["header_1"].(map[string]interface{})["trusted_validators"].(map[string]interface{})["proposer"].(map[string]interface{})["pub_key"].(map[string]interface{}))
	pubKeys = append(pubKeys, jsonIntermediary["header_2"].(map[string]interface{})["validator_set"].(map[string]interface{})["proposer"].(map[string]interface{})["pub_key"].(map[string]interface{}))
	pubKeys = append(pubKeys, jsonIntermediary["header_2"].(map[string]interface{})["trusted_validators"].(map[string]interface{})["proposer"].(map[string]interface{})["pub_key"].(map[string]interface{}))

	for _, proposerPubKey := range pubKeys {
		ed25519PubKey := proposerPubKey["ed25519"].(string)
		proposerPubKey["type"] = "tendermint/PubKeyEd25519"
		proposerPubKey["value"] = ed25519PubKey
	}

	header1Sigs := jsonIntermediary["header_1"].(map[string]interface{})["signed_header"].(map[string]interface{})["commit"].(map[string]interface{})["signatures"].([]interface{})
	header2Sigs := jsonIntermediary["header_2"].(map[string]interface{})["signed_header"].(map[string]interface{})["commit"].(map[string]interface{})["signatures"].([]interface{})
	sigs := header1Sigs
	sigs = append(sigs, header2Sigs...)
	for _, sig := range sigs {
		sig := sig.(map[string]interface{})
		if sig["block_id_flag"] == "BLOCK_ID_FLAG_COMMIT" {
			sig["block_id_flag"] = 2
		} else {
			return nil, fmt.Errorf("unexpected block_id_flag: %s", sig["block_id_flag"])
		}

		valAddressBase64Str, ok := sig["validator_address"].(string)
		if !ok {
			return nil, fmt.Errorf("validator_address not found")
		}
		valAddressBz, err := base64.StdEncoding.DecodeString(valAddressBase64Str)
		if err != nil {
			return nil, err
		}
		sig["validator_address"] = hex.EncodeToString(valAddressBz)
	}

	return json.Marshal(jsonIntermediary)
}

func execOperatorCommand(c *exec.Cmd) ([]byte, error) {
	var outBuf bytes.Buffer

	// Create a MultiWriter to write to both os.Stdout and the buffer
	multiWriter := io.MultiWriter(os.Stdout, &outBuf)

	// Set the command's stdout and stderror to the MultiWriter
	c.Stdout = multiWriter
	c.Stderr = multiWriter

	// Run the command
	if err := c.Run(); err != nil {
		return nil, fmt.Errorf("operator command '%s' failed: %s", strings.Join(c.Args, " "), outBuf.String())
	}

	return outBuf.Bytes(), nil
}
