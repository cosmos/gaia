package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
)

const (
	atomsPerETH = 452.3
	million     = 1000 * 1000
	weiPerETH   = million * million * million
	weiPerATOM  = 2210921954455007

	startBlock   = 3482440
	endBlock     = 3571380
	contractAddr = "0xCF965Cfe7C30323E9C9E41D4E398e2167506f764"
	topic        = "0x14432f6e1dc0e8c1f4c0d81c69cecc80c0bea817a74482492b0211392478ab9b"

	etherscanFmt = "https://api.etherscan.io/api?module=logs&action=getLogs&fromBlock=%d&toBlock=%d&address=%s&topic0=%s"

	ethAtomsFile     = "eth_atoms.json"
	ethDonationsFile = "eth_donations.json"
)

type ETHContribution struct {
	BlockHeight int    `json:"height"`
	TxIndex     int    `json:"tx_index"`
	Wei         string `json:"wei"`
	Address     string `json:"address"`
}

func (contrib ETHContribution) String() string {
	return fmt.Sprintf("%d/%d: %s Wei %s", contrib.BlockHeight, contrib.TxIndex, contrib.Wei, contrib.Address)
}

type ETHAPIResponse struct {
	Result []ETHRaw `json:"result"`
}

type ETHRaw struct {
	Address  string   `json:"address"`
	Data     string   `json:"data"`
	BlockNum string   `json:"blockNumber"`
	TxIndex  string   `json:"transactionIndex"`
	Topics   []string `json:"topics"`
}

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("requires 'data' or 'atoms' argument")
		os.Exit(1)
	}

	switch args[1] {
	case "data":
		getETHData()
	case "atoms":
		writeETHAtoms()
	default:
		fmt.Println("requires 'data' or 'atoms' argument")
	}
}

func getETHData() {

	resp, err := http.Get(fmt.Sprintf(etherscanFmt, startBlock, endBlock, contractAddr, topic))
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	bz, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var apiRes ETHAPIResponse
	err = json.Unmarshal(bz, &apiRes)
	if err != nil {
		panic(err)
	}

	var contribs []ETHContribution
	for i, raw := range apiRes.Result {

		if len(raw.Topics) > 2 {
			panic(fmt.Sprintf("too many topics (%d) in result %d: %v", len(raw.Topics), i, raw))
		}
		if raw.Topics[0] != topic {
			panic(fmt.Sprintf("bad first topic, got %s", raw.Topics[0]))
		}

		blockNum := hexToInt(raw.BlockNum)
		txIndex := hexToInt(raw.TxIndex)

		// address and value are encoded in the event.
		// addr is the second topic value and the amount of wei
		// is the second value in the data.
		addr := stripHex(raw.Topics[1])[24:]
		wei := dataToWei(raw.Data)

		contrib := ETHContribution{
			BlockHeight: blockNum,
			TxIndex:     txIndex,
			Wei:         wei,
			Address:     addr,
		}
		contribs = append(contribs, contrib)
	}

	bz, err = json.Marshal(contribs)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(ethDonationsFile, bz, 0600)
	if err != nil {
		panic(err)
	}
}

func writeETHAtoms() {
	bz, err := ioutil.ReadFile(ethDonationsFile)
	if err != nil {
		panic(err)
	}

	var contribs []ETHContribution
	err = json.Unmarshal(bz, &contribs)
	if err != nil {
		panic(err)
	}

	var totalETH float64
	var totalATOM float64
	addrs := make(map[string]float64)
	for _, contrib := range contribs {
		wei := contrib.Wei
		eth, atoms := weiStringToAmts(wei)

		if _, ok := addrs[contrib.Address]; ok {
			fmt.Println("Duplicate addr", contrib)
		}
		addrs[contrib.Address] += atoms

		totalETH += eth
		totalATOM += atoms
	}

	// refund whale
	whaleRefundETH := float64(17890)
	whaleAtoms := float64(10000000)
	whaleAddr := "aff9f5a716cdd701304eae6fc7f42c80fdeea584"
	addrs[whaleAddr] = whaleAtoms
	totalETH -= whaleRefundETH

	fmt.Println("total contributions", len(contribs)-1)
	fmt.Println("total addrs", len(addrs))
	fmt.Println("total eth", totalETH)
	fmt.Println("total atom", totalATOM)

	bz, err = json.MarshalIndent(addrs, "", "  ")
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(ethAtomsFile, bz, 0600)
	if err != nil {
		panic(err)
	}
}

//------------

func hexToInt(h string) int {
	h = stripHex(h)
	bz, err := hex.DecodeString(h)
	if err != nil {
		panic(err)
	}
	z := new(big.Int)
	z.SetBytes(bz)
	return int(z.Int64())
}

// return wei as decimal string
func hexWeiToString(h string) string {
	h = stripHex(h)
	bz, err := hex.DecodeString(h)
	if err != nil {
		panic(err)
	}
	z := new(big.Int)
	z.SetBytes(bz)
	return z.String()
}

// return eth and atoms from wei
func weiStringToAmts(weiString string) (float64, float64) {

	bigWei, suc := new(big.Float).SetString(weiString)
	if !suc {
		panic("failed to set weiString")
	}

	bigEth := new(big.Float)
	bigEth.Quo(bigWei, big.NewFloat(float64(weiPerETH)))
	eth, _ := bigEth.Float64()

	bigAtom := new(big.Float)
	bigAtom.Quo(bigWei, big.NewFloat(float64(weiPerATOM)))
	atoms, _ := bigAtom.Float64()
	atoms = float64(int64(atoms))
	return eth, atoms
}

func stripHex(h string) string {
	if h == "0x" {
		return "00"
	}
	if len(h) > 2 && h[:2] == "0x" {
		h = h[2:]
	}
	if len(h)%2 != 0 {
		h = "0" + h
	}
	return h
}

func dataToWei(h string) string {
	h = stripHex(h)
	if len(h) != 64*3 {
		panic("expected 64*3")
	}
	msgValueHex := h[64 : 64*2]
	return hexWeiToString(msgValueHex)
}
