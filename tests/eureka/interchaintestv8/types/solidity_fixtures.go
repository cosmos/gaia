package types

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/cosmos/solidity-ibc-eureka/abigen/ics26router"

	"github.com/srdtrk/solidity-ibc-eureka/e2e/v8/testvalues"
)

// GenericSolidityFixture is the fixture to be unmarshalled into a test case in Solidity tests
type GenericSolidityFixture struct {
	// Hex encoded bytes for sp1 genesis fixture
	Sp1GenesisFixture string `json:"sp1_genesis_fixture"`
	// Hex encoded bytes to be fed into the router contract
	Msg string `json:"msg"`
	// Hex encoded bytes for the IICS26RouterMsgsPacket in the context of this fixture
	Packet string `json:"packet"`
	// The contract address of the ERC20 token
	Erc20Address string `json:"erc20_address"`
	// The timestamp in seconds around the time of submitting the Msg to the router contract
	Timestamp int64 `json:"timestamp"`
}

// GenerateAndSaveSolidityFixture generates a fixture and saves it to a file
func GenerateAndSaveSolidityFixture(fileName, erc20Address string, msgBz []byte, packet ics26router.IICS26RouterMsgsPacket) error {
	fixture, err := generateFixture(erc20Address, msgBz, packet)
	if err != nil {
		return err
	}

	fixtureBz, err := json.Marshal(fixture)
	if err != nil {
		return err
	}

	filePath := testvalues.SolidityFixturesDir + "/" + fileName
	// nolint:gosec
	return os.WriteFile(filePath, fixtureBz, 0o644)
}

func generateFixture(erc20Address string, msgBz []byte, packet ics26router.IICS26RouterMsgsPacket) (GenericSolidityFixture, error) {
	genesisBz, err := getGenesisFixture()
	if err != nil {
		return GenericSolidityFixture{}, err
	}

	packetBz, err := abiEncodePacket(packet)
	if err != nil {
		return GenericSolidityFixture{}, err
	}

	// Generate the fixture
	fixture := GenericSolidityFixture{
		Sp1GenesisFixture: hex.EncodeToString(genesisBz),
		Msg:               hex.EncodeToString(msgBz),
		Erc20Address:      erc20Address,
		Timestamp:         time.Now().Unix(),
		Packet:            hex.EncodeToString(packetBz),
	}
	return fixture, nil
}

func getGenesisFixture() ([]byte, error) {
	genesisBz, err := os.ReadFile(testvalues.Sp1GenesisFilePath)
	if err != nil {
		return nil, err
	}

	// Because the genesis json has line breaks and spaces, we need to unmarshal and marshal it again to get the compact version
	var jsonData interface{}
	if err := json.Unmarshal(genesisBz, &jsonData); err != nil {
		return nil, err
	}
	compactGenesisBz, err := json.Marshal(jsonData)
	if err != nil {
		return nil, err
	}

	return compactGenesisBz, nil
}

func abiEncodePacket(packet ics26router.IICS26RouterMsgsPacket) ([]byte, error) {
	structType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "sequence", Type: "uint32"},
		{Name: "sourceClient", Type: "string"},
		{Name: "destClient", Type: "string"},
		{Name: "timeoutTimestamp", Type: "uint64"},
		{Name: "payloads", Type: "tuple[]", Components: []abi.ArgumentMarshaling{
			{Name: "sourcePort", Type: "string"},
			{Name: "destPort", Type: "string"},
			{Name: "version", Type: "string"},
			{Name: "encoding", Type: "string"},
			{Name: "value", Type: "bytes"},
		}},
	})
	if err != nil {
		return nil, err
	}

	args := abi.Arguments{
		{Type: structType},
	}

	return args.Pack(packet)
}
