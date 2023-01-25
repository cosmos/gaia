import * as fs from 'fs';
import * as util from 'util';
// import {BigNumber} from "bignumber.js"
import * as bech32 from "bech32"

// Convert fs.readFile into Promise version of same    
const readFile = util.promisify(fs.readFile);

// function parseFraction(fraction: string):string{
//     let parts = fraction.split("/")

//     if( parts.length == 2){
//         return new BigNumber(parts[0]).div(new BigNumber(parts[1])).toFixed(10)
//     }
//     return new BigNumber(parts[0]).toFixed(10)

// }

// function fractionToDecimal(fraction: string):string{
//     return parseFraction(fraction)
// }


function convertBech32(encoded:string, new_hrp:string):string{
    let data = bech32.decode(encoded)
    return bech32.encode(new_hrp,data.words)
}

async function getStuff() {
    return await readFile('../gaia8001.json');
  }



  getStuff().then(data => {
    let genesis = JSON.parse(data.toString()); 
    // let validators = genesis.app_state.stake.validators;
    // for (let val of validators){
    //     val.tokens = fractionToDecimal(val.tokens)
    //     val.delegator_shares = fractionToDecimal(val.delegator_shares)
    //     val.bond_height = "0";
    //     val.bond_intra_tx_counter = 0;
    //     val.jailed = val.revoked;
    //     delete val.revoked;
    // }
    // let bonds = genesis.app_state.stake.bonds;
    // for (let bond of bonds){
    //     bond.delegator_addr = convertBech32(bond.delegator_addr, "cosmos")
    //     bond.validator_addr = convertBech32(bond.validator_addr, "cosmosvaloper")
    //     bond.shares = fractionToDecimal(bond.shares);
    //     bond.height = 0;
    // }

    let accounts = genesis.app_state.accounts
    for (let account of accounts){
        account.address = convertBech32(account.address, "cosmos")
        account.coins = [{
            "denom":"steak",
            "amount":"50"
        }]
    }
    let tokens  = accounts.length * 50

    // console.log(JSON.stringify(accounts, null, 4))

    console.log(tokens)


  }
  )