package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	slashing "github.com/cosmos/cosmos-sdk/x/slashing/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pkg/errors"
	tmtypes "github.com/tendermint/tendermint/types"
)

type replacementConfigs []replacementConfig

func (r *replacementConfigs) isReplacedValidator(validatorAddress string) (int, replacementConfig) {

	for i, replacement := range *r {
		if replacement.ValidatorAddress == validatorAddress {
			return i, replacement
		}
	}

	return -1, replacementConfig{}
}

type replacementConfig struct {
	Name             string `json:"validator_name"`
	ValidatorAddress string `json:"validator_address"`
	ConsensusPubkey  string `json:"stargate_consensus_public_key"`
}

func loadKeydataFromFile(clientCtx client.Context, replacementrJSON string, genDoc *tmtypes.GenesisDoc) *tmtypes.GenesisDoc {
	jsonReplacementBlob, err := ioutil.ReadFile(replacementrJSON)
	if err != nil {
		log.Fatal(errors.Wrapf(err, "failed to read replacement keys from file %s", replacementrJSON))
	}

	var replacementKeys replacementConfigs

	err = json.Unmarshal(jsonReplacementBlob, &replacementKeys)

	if err != nil {
		log.Fatal("Could not unmarshal replacement keys ")
	}

	var state types.AppMap
	if err := json.Unmarshal(genDoc.AppState, &state); err != nil {
		log.Fatal(errors.Wrap(err, "failed to JSON unmarshal initial genesis state"))
	}

	var stakingGenesis staking.GenesisState
	var slashingGenesis slashing.GenesisState

	clientCtx.JSONMarshaler.MustUnmarshalJSON(state[staking.ModuleName], &stakingGenesis)
	clientCtx.JSONMarshaler.MustUnmarshalJSON(state[slashing.ModuleName], &slashingGenesis)

	for i, val := range stakingGenesis.Validators {
		idx, replacement := replacementKeys.isReplacedValidator(val.OperatorAddress)

		if idx != -1 {
			toReplaceVal, err := val.ToTmValidator()

			if err != nil {
				log.Fatal(fmt.Errorf("failed construnct a tendermint validator to replace from export:%s %w", val, err))
			}

			consPubKey, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, replacement.ConsensusPubkey)

			if err != nil {
				log.Fatal(fmt.Errorf("failed to decode key:%s %w", consPubKey, err))
			}

			val.ConsensusPubkey, err = codectypes.PackAny(consPubKey)
			if err != nil {
				log.Fatal(fmt.Errorf("failed to decode key:%s %w", consPubKey, err))
			}

			replaceVal, err := val.ToTmValidator()

			if err != nil {
				log.Fatal(fmt.Errorf("failed construnct a tendermint validator to replace with from list:%s %w", val, err))
			}

			toReplaceValConsAddress, _ := sdk.ConsAddressFromHex(toReplaceVal.Address.String())
			replaceValConsAddress, _ := sdk.ConsAddressFromHex(replaceVal.Address.String())

			for i, signingInfo := range slashingGenesis.SigningInfos {
				if signingInfo.Address == toReplaceValConsAddress.String() {
					slashingGenesis.SigningInfos[i].Address = replaceValConsAddress.String()
					slashingGenesis.SigningInfos[i].ValidatorSigningInfo.Address = replaceValConsAddress.String()
				}
			}

			for i, missedInfo := range slashingGenesis.MissedBlocks {
				if missedInfo.Address == toReplaceValConsAddress.String() {
					slashingGenesis.MissedBlocks[i].Address = replaceValConsAddress.String()
				}
			}

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
	state[slashing.ModuleName] = clientCtx.JSONMarshaler.MustMarshalJSON(&slashingGenesis)

	genDoc.AppState, err = json.Marshal(state)

	if err != nil {
		log.Fatal("Could not marshal App State")
	}
	return genDoc

}
