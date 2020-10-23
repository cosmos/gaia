package app

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pkg/errors"
	tmtypes "github.com/tendermint/tendermint/types"
)

type replacementConfigs []replacementConfig

func (r *replacementConfigs) isReplacedValidator(validatorAddress string) (int, replacementConfig) {

	for i, replacement := range *r {
		if replacement.validatorAddress == validatorAddress {
			return i, replacement
		}
	}

	return -1, replacementConfig{}
}

type replacementConfig struct {
	name             string
	validatorAddress string
	consensusPubkey  string
}

func loadKeydataFromFile(clientCtx client.Context, replacementrJSON string, genDoc *tmtypes.GenesisDoc) *tmtypes.GenesisDoc {
	jsonReplacementBlob, err := ioutil.ReadFile(replacementrJSON)
	if err != nil {
		log.Fatal(errors.Wrapf(err, "failed to read replacement keys from file %s", replacementrJSON))
	}

	var replacementKeys replacementConfigs

	err = json.Unmarshal(jsonReplacementBlob, &replacementKeys)

	var state types.AppMap
	if err := json.Unmarshal(genDoc.AppState, &state); err != nil {
		log.Fatal(errors.Wrap(err, "failed to JSON unmarshal initial genesis state"))
	}

	var stakingGenesis staking.GenesisState

	clientCtx.JSONMarshaler.MustUnmarshalJSON(state[staking.ModuleName], &stakingGenesis)

	for i, val := range stakingGenesis.Validators {
		idx, replacement := replacementKeys.isReplacedValidator(val.OperatorAddress)

		if idx != -1 {
			toReplaceVal := val.ToTmValidator()

			val.ConsensusPubkey = replacement.consensusPubkey

			replaceVal := val.ToTmValidator()

			for tmIdx, tmval := range genDoc.Validators {
				if tmval.Address.String() == toReplaceVal.Address.String() {
					genDoc.Validators[tmIdx].Address = replaceVal.Address
					genDoc.Validators[tmIdx].PubKey = replaceVal.PubKey

				}
			}
			stakingGenesis.Validators[i] = val

		}

	}
	state[staking.ModuleName] = clientCtx.JSONMarshaler.MustMarshalJSON(&stakingGenesis)

	genDoc.AppState, err = json.Marshal(state)

	if err != nil {
		log.Fatal("Could not App State")
	}
	return genDoc

}
