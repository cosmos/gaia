package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"

	"github.com/cosmos/launch/pkg"
)

const (
	ethATOMFile = "eth_atoms.json"
	btcATOMFile = "btc_atoms.json"

	outFile = "contributors.json"
)

func main() {
	ethAtoms := pkg.ObjToMap(ethATOMFile)
	btcAtoms := pkg.ObjToMap(btcATOMFile)

	all := make(map[string]float64)

	var totalAtom float64
	for addr, amt := range ethAtoms {
		if _, ok := all[addr]; ok {
			fmt.Println("Duplicate eth/btc", addr)
		}
		amt = float64(math.Round(amt)) // SIGH
		all[addr] += amt
		totalAtom += amt
	}

	for addr, amt := range btcAtoms {
		if _, ok := all[addr]; ok {
			fmt.Println("Duplicate eth/btc", addr)
		}
		all[addr] += amt
		totalAtom += amt
	}

	for addr, amt := range all {
		all[addr] = pkg.Round2(amt)
	}

	fmt.Println("total addr", len(all))
	fmt.Println("total atom", totalAtom)

	bz, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(outFile, bz, 0600)
	if err != nil {
		panic(err)
	}
}
