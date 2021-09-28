package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/cosmos/launch/pkg"
)

const (
	earlyJSON  = "early.json"
	seedJSON   = "seed.json"
	outputFile = "contributors.json"
)

func main() {

	seedMap := pkg.ListToMap(seedJSON)
	earlyMap := pkg.ListToMap(earlyJSON)

	// accumulate them in all and sum
	all := make(map[string]float64)
	seedTotal := pkg.AccumMap(seedMap, all)
	earlyTotal := pkg.AccumMap(earlyMap, all)

	fmt.Printf("Seed: %f\n", seedTotal)
	fmt.Printf("Strategic/Early: %f\n", earlyTotal)
	fmt.Println("----------")
	fmt.Println("Total Addrs", len(all))
	total := seedTotal + earlyTotal
	fmt.Printf("Total Atoms %f\n", total)

	// marshal to json and write to file
	bz, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(outputFile, bz, 0600)
	if err != nil {
		panic(err)
	}
}
