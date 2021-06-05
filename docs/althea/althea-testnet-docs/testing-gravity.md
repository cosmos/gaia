# Testing Gravity

Now that we've made it this far it's time to actually play around with the bridge

This first command will send some ERC20 tokens to an address of your choice on the Althea
chain. Notice that the Ethereum key is pre-filled. This address has both some test ETH and
a large balance of ERC20 tokens from the contracts listed here.

```
0xD7600ae27C99988A6CD360234062b540F88ECA43 - Bitcoin MAX (MAX)
0x7580bFE88Dd3d07947908FAE12d95872a260F2D8 - 2 Ethereum (E2H)
0xD50c0953a99325d01cca655E57070F1be4983b6b - Byecoin (BYE)
```

Note that the 'amount' field for this command is now in whole coins rather than wei like the previous testnets

```
gbt -a althea client eth-to-cosmos \
        --ethereum-key "0xb1bab011e03a9862664706fc3bbaa1b16651528e5f0e7fbfcbfdd8be302a13e7" \
        --gravity-contract-address "0xFA2f45c5C8AcddFfbA0E5228bDf7E8B8f4fD2E84" \
        --token-contract-address "any of the three values above" \
        --amount=1 \
        --destination "any Cosmos address, I suggest your delegate Cosmos address"
```

You should see a message like this on your Orchestrator. The details of course will be different but it means that your Orchestrator has observed the event on Ethereum and sent the details into the Cosmos chain!

```
[2021-02-13T12:35:54Z INFO  orchestrator::ethereum_event_watcher] Oracle observed deposit with sender 0xBf660843528035a5A4921534E156a27e64B231fE, destination cosmos1xpfu40gseet70wfeazds773v05pjx3dwe7e03f, amount
999999984306749440, and event nonce 3
```

Once the event has been observed we can check our balance on the Cosmos side. We will see some peggy<ERC20 address> tokens in our balance. We have a good bit of code in flight right now so the module renaming from 'Peggy' to 'Gravity' has been put on hold until we're feature complete.

```
althea query bank balances <any cosmos address>
```

Now that we have some tokens on the Althea chain we can try sending them back to Ethereum. Remember to use the Cosmos phrase for the address you actually sent the tokens to. Alternately you can send Cosmos native tokens with this command.

The denom of a bridged token will be

```
gravity0xD7600ae27C99988A6CD360234062b540F88ECA43
```

```
gbt -a althea client cosmos-to-eth \
        --cosmos-phrase "the phrase containing the Gravity bridged tokens (delegate keys mnemonic)" \
        --fees 100footoken \
        --amount 100000000000gravity0xD7600ae27C99988A6CD360234062b540F88ECA43 \
        --eth-destination "any eth address, try your delegate eth address"
```

It will take a moment or two for Etherescan to catch up, but once it has you'll see the new ERC20 token balance reflected at https://goerli.etherscan.io/

# Really testing Gravity

Now that we have the basics out of the way we can get into the fun testing, including hundreds of transactions across the bridge, upgrades, and slashing. Depending on how the average participant is doing we may or may not get to this during our chain start call.

- Send a 100 transaction batch - [x]
- Send 100 deposits to the Althea chain from Ethereum - [x]
- IBC bridge some tokens to another chain - [x]
- Exchange those bridged tokens on the Gravity DEX - []
- Have a governance vote to reduce the slashing period to 1 hr downtime, then have a volunteer get slashed - []
- Stretch goal, upgrade the testnet with Gravity V2 features. This may end up not being practical depending on the amount of changes made. -[x]
