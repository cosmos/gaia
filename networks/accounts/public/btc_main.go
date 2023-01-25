package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/alfg/blockchain"
	"github.com/cosmos/launch/pkg"
)

const (
	exodusAddress = "35ty8iaSbWsj4YVkoHzs9pZMze6dapeoZ8"
	startHeight   = 460654
	endHeight     = 460661

	atomsPerBTC   = 11635
	satoshiPerBTC = 100 * 1000 * 1000

	btcAtomsFile     = "btc_atoms.json"
	btcDonationsFile = "btc_donations.json"
)

type BTCContribution struct {
	BlockHeight int
	TxIndex     int
	Amount      int // satoshis
	Address     string
}

func (contrib BTCContribution) String() string {
	return fmt.Sprintf("%d/%d: %d %s", contrib.BlockHeight, contrib.TxIndex, contrib.BTC, contrib.Address)
}

func (contrib BTCContribution) BTC() float64 {
	return float64(contrib.Amount) / satoshiPerBTC
}

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("requires 'data' or 'atoms' argument")
		os.Exit(1)
	}

	switch args[1] {
	case "data":
		getBTCData()
	case "atoms":
		writeBTCAtoms()
	default:
		fmt.Println("requires 'data' or 'atoms' argument")
	}
}

func getBTCData() {
	client, err := blockchain.New()
	if err != nil {
		panic(err)
	}

	var contribs []BTCContribution
	for height := startHeight; height < endHeight+1; height++ {
		b, err := client.GetBlockHeight(fmt.Sprintf("%d", height))
		if err != nil {
			panic(err)
		}

		if len(b.Blocks) > 1 {
			fmt.Println("DETECTED FORK!")
		}

		block := b.Blocks[0]
		txs := block.Tx
		for txIndex, tx := range txs {
			out := tx.Out
			// valid tx has two outputs
			if len(out) != 2 {
				continue
			}
			out1, out2 := out[0], out[1]

			// first output should be to exodus
			if out1.Addr != exodusAddress {
				if out2.Addr == exodusAddress {
					fmt.Println("found exodus addr in output 2!", tx)
				}
				continue
			}

			// second output should be empty
			if out2.Value != 0 {
				fmt.Println("found value in out 2!", tx)
				continue
			}

			contrib := BTCContribution{
				BlockHeight: height,
				TxIndex:     txIndex,
				Amount:      out1.Value,
				Address:     out2.Script[4:], // shave the op codes
			}
			contribs = append(contribs, contrib)
		}
	}

	bz, err := json.Marshal(contribs)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(btcDonationsFile, bz, 0600)
	if err != nil {
		panic(err)
	}
}

func writeBTCAtoms() {
	bz, err := ioutil.ReadFile(btcDonationsFile)
	if err != nil {
		panic(err)
	}

	var contribs []BTCContribution
	err = json.Unmarshal(bz, &contribs)
	if err != nil {
		panic(err)
	}

	var totalSatoshi int
	var totalBTC float64
	var totalATOM float64
	addrs := make(map[string]float64)
	for _, contrib := range contribs {
		satoshi := contrib.Amount
		btc := contrib.BTC()
		// atoms := pkg.Round2(btc * atomsPerBTC)
		atoms := float64(satoshi*atomsPerBTC) / satoshiPerBTC

		if _, ok := addrs[contrib.Address]; ok {
			fmt.Println("Duplicate addr", contrib)
		}
		addrs[contrib.Address] += atoms

		totalSatoshi += satoshi
		totalBTC += btc
		totalATOM += atoms
	}

	for addr, amt := range addrs {
		addrs[addr] = pkg.Round2(amt)
	}

	fmt.Println("total contributions", len(contribs))
	fmt.Println("total addrs", len(addrs))
	fmt.Println("total satoshi", totalSatoshi)
	fmt.Println("total btc", totalBTC)
	fmt.Println("total atom", totalATOM)

	bz, err = json.MarshalIndent(addrs, "", "  ")
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(btcAtomsFile, bz, 0600)
	if err != nil {
		panic(err)
	}
}
