import * as fs from 'fs';
import * as util from 'util';
// import parse from  'csv-parse';

// import {BigNumber} from "bignumber.js"
// import * as bech32 from "bech32"

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

interface genesis{
    app_state: app_state
}

interface app_state{
    accounts: account[]
    staking: staking
}

interface staking{
    pool: pool
}

interface pool{
    loose_tokens:string
    bonded_tokens:string

}
interface account{
    address: string
    sequence_number:string
    account_number:string
    coins:coin[]
    original_vesting: coin[],
    delegated_free: coin[],
    delegated_vesting: null[],
    start_time:string,
    end_time: string
}

interface coin{
    amount:string
    denon:string
}


// function convertBech32(encoded:string, new_hrp:string):string{
//     let data = bech32.decode(encoded)
//     return bech32.encode(new_hrp,data.words)
// }

async function getAddresses() {

    return [await readFile('../../gaia-9002/genesis.json'),await readFile('../genesis-template.json')];
  }



  getAddresses().then(data => {
    

    let accounts_9001:genesis = JSON.parse(data[0].toString());

    let template:genesis = JSON.parse(data[1].toString())

    let accounts = [];

    let new_acc ={
        "address":"cosmos1uclv9m6xuh4m8puxd8ndwhxhf968gxyhk5udyx",
        "coins":[
            {
                "amount":"10000000000",
                "denom":"photinos",
            },
            {
                "amount":"10000",
                "denom":"stake",
            }
        ],
        "sequence_number":"0",
        "account_number":"0",
        "original_vesting": null,
        "delegated_free": null,
        "delegated_vesting": null,
        "start_time": "0",
        "end_time": "0"
    }

    accounts.push(new_acc)

    for (let account of accounts_9001.app_state.accounts){

        let acc = {
            "address":account.address,
            "coins":[
                {
                    "amount":"10000000000",
                    "denom":"photinos"
                },
                {
                    "amount":"10000",
                    "denom":"stake"
                }
            ],
            "sequence_number":"0",
            "account_number":"0",
            "original_vesting": null,
            "delegated_free": null,
            "delegated_vesting": null,
            "start_time": "0",
            "end_time": "0"
        }
        accounts.push(acc);          
    }

    

    template.app_state.accounts= accounts

    fs.writeFileSync("../genesis.json",JSON.stringify(template,undefined, 4))


    })



    // let accounts = genesis.app_state.accounts
    // for (let account of accounts){
    //     account.address = convertBech32(account.address, "cosmos")
    //     account.coins = [{
    //         "denom":"steak",
    //         "amount":"50"
    //     }]
    // }
    // let tokens  = accounts.length * 50

    // console.log(JSON.stringify(accounts, null, 4))

    // console.log(tokens)


  