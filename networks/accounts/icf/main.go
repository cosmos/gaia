package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"

	pkg "github.com/cosmos/launch/pkg"
)

const (
	earlyJSON    = "early.json" // list [address, float64,]*
	gosJSON      = "gos.json"   // list [address, float64,]*
	multisigJSON = "multisig.json"
	outJSON      = "contributors.json" // map address->float64
)

func main() {
	// load all files into maps
	earlyMap := pkg.ListToMap(earlyJSON)
	gosMap := pkg.ListToMap(gosJSON)
	multisigMap := multisigToMap(multisigJSON)

	// accumulate them in all and sum
	all := make(map[string]float64)
	earlyTotal := pkg.AccumMap(earlyMap, all)
	gosTotal := pkg.AccumMap(gosMap, all)
	multisigTotal := pkg.AccumMap(multisigMap, all)

	fmt.Printf("Early: %d, %f\n", len(earlyMap), earlyTotal)
	fmt.Printf("GoS: %d, %f\n", len(gosMap), gosTotal)
	fmt.Printf("Multisig: %d, %f\n", len(multisigMap), multisigTotal)
	fmt.Println("-------------")
	fmt.Println("Total addrs", len(all))
	total := earlyTotal + gosTotal + multisigTotal
	fmt.Printf("Total: %f\n", total)

	// write the consolidated file
	bz, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(outJSON, bz, 0600)
	if err != nil {
		panic(err)
	}

	// reload the file and check it matches
	contribs := pkg.ObjToMap(outJSON)
	if len(contribs) != len(all) {
		panic(fmt.Sprintf("contribs len (%v) doesnt match computed length (%v)", len(contribs), len(all)))
	}
	var precision float64 = 100
	if math.Round(total*precision) != math.Round(pkg.SumMap(contribs)*precision) {
		panic(fmt.Sprintf("contribs sum (%v) doesnt match computed sum (%v)", pkg.SumMap(contribs), total))
	}
}

type MultisigAccount struct {
	Threshold int      `json:"threshold"`
	Pubs      []string `json:"pubs"`
	Amount    float64  `json:"amount"`
}

func multisigToMap(file string) map[string]float64 {
	// read multisig
	bz, err := ioutil.ReadFile(multisigJSON)
	if err != nil {
		panic(err)
	}

	// expects a map from address to struct with pubkeys
	multisigMap := make(map[string]MultisigAccount)
	err = json.Unmarshal(bz, &multisigMap)
	if err != nil {
		panic(err)
	}

	// build a map with just address->amt
	// check the pubkeys match the address
	// and theres no duplicates
	addrMap := make(map[string]float64)
	for addr, act := range multisigMap {
		pkg.CheckMultisigAddress(act.Threshold, act.Pubs, addr)
		if _, ok := addrMap[addr]; ok {
			panic("duplicate multisig addr")
		}
		addrMap[addr] += act.Amount
	}

	return addrMap
}
