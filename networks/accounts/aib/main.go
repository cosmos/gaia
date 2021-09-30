package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/cosmos/launch/pkg"
)

const (
	employeeJSON = "employees.json"
	multisigJSON = "multisig.json"
)

func main() {

	// get AiB allocations
	employees := getAccounts(employeeJSON)
	multisigAcc := getMultisig(multisigJSON)

	// check allocations, duplicates, multisig
	checkAddrs(employees, multisigAcc)

	// sum employees
	sum := sumAccounts(employees)

	fmt.Println("Num Employee Addresses", len(employees))
	fmt.Printf("Employee Atoms %f\n", sum)
	fmt.Printf("Multisig Atoms %f\n", multisigAcc.Amount)
	fmt.Println("------------")
	sum += multisigAcc.Amount
	fmt.Printf("Total %f\n", sum)
}

type Account struct {
	Address string  `json:"addr"`
	Amount  float64 `json:"amount"`
	Lock    string  `json:"lock"`
}

type MultisigAccount struct {
	Address   string   `json:"addr"`
	Threshold int      `json:"threshold"`
	Pubs      []string `json:"pubs"`
	Amount    float64  `json:"amount"`
}

func getMultisig(file string) MultisigAccount {
	bz, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	var multisigAcc MultisigAccount
	err = json.Unmarshal(bz, &multisigAcc)
	if err != nil {
		panic(err)
	}
	return multisigAcc
}

func getAccounts(file string) []Account {
	bz, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	var accs []Account
	err = json.Unmarshal(bz, &accs)
	if err != nil {
		panic(err)
	}
	return accs
}

// check for duplicates, non positive values, and multisig addr
func checkAddrs(accs []Account, multi MultisigAccount) {
	accsMap := make(map[string]struct{})

	for _, acc := range accs {
		// duplicate check
		if _, ok := accsMap[acc.Address]; ok {
			panic(fmt.Sprintf("duplicate account: %v", acc.Address))
		}
		accsMap[acc.Address] = struct{}{}

		// value check
		if acc.Amount <= 0 {
			panic(fmt.Sprintf("employee with 0 atoms: %v", acc))
		}
	}

	// check multisig for dupl and non positive value
	if _, ok := accsMap[multi.Address]; ok {
		panic(fmt.Sprintf("duplicate account: %v", multi.Address))
	}
	if multi.Amount <= 0 {
		panic(fmt.Sprintf("multisig with 0 atoms: %v", multi))
	}

	// check multisig address
	pkg.CheckMultisigAddress(multi.Threshold, multi.Pubs, multi.Address)
}

func sumAccounts(accs []Account) float64 {
	var sum float64
	for _, acc := range accs {
		sum += acc.Amount
	}
	return sum
}
