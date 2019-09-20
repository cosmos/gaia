# IBC instruction

// temporal document

## Dependencies

This branch uses non-canonical branch of cosmos-sdk. Run `go mod vendor` on the root directory to retrive the dependencies.

Move to `gaiad` directory, run `./gaiad testnet --v 1` and `mv ./mytestnet/node0 ../`

Currently the testing is done on a single chain with a pseudo-loopback connection, meaning that 
- there is only one client living on single chain, pointing itself
- there are two connections living on single chain, bonded to the client and pointing each other
- there are two channels living on single chain, bonded to each connections and pointing each other

but all lightclient verification and merkle proving are actually processed.

Check the chain-id of the testnet under `node0/gaiad/config/genesis.json` and store it as the environment variable `CID`.

```bash
> export CID=chain-attv4e
```

This will be used by the shell scripts.

Run gaia daemon by

```bash
> ./gaiad/gaiad --home ./node0/gaiad start
```

## Client

Client can be instantiated with `client.sh` command. It will print

```bash
creating client client-09b6
```

export that identifier as an env variable.

```bash
> export CLIENTID=client-b438
```

You can query the client after creation by 

```bash
> ./gaiacli query ibc client client $CLIENTID --home ../node0/gaiacli --trust-node
{
  "type": "ibc/client/tendermint/ConsensusState",
  "value": {
    "ChainID": "chain-attv4e",
    "Height": "1006",
    "Root": {
      "type": "ibc/commitment/merkle/Root",
      "value": {
        "hash": "RDYMrUY6z9UBtPk9+qKl2Vujm8dOyePj/9dUlh6VvWM="
      }
    },
    "NextValidatorSet": {
      "validators": [
        {
          "address": "9A4B3DF37C5F60517397410AE705B68652275ECF",
          "pub_key": {
            "type": "tendermint/PubKeyEd25519",
            "value": "g61k/wo7hKejV8qDPVKKYKF9RbK9NH4G+5ioRlDkha4="
          },
          "voting_power": "100",
          "proposer_priority": "0"
        }
      ],
      "proposer": {
        "address": "9A4B3DF37C5F60517397410AE705B68652275ECF",
        "pub_key": {
          "type": "tendermint/PubKeyEd25519",
          "value": "g61k/wo7hKejV8qDPVKKYKF9RbK9NH4G+5ioRlDkha4="
        },
        "voting_power": "100",
        "proposer_priority": "0"
      }
    }
  }
}
```

See [script log](./client.txt)

## Connection

Connections can be established with `connection.sh $CLIENTID` command. It will print

```bash
connection 1: conn-c91b
connection 2: conn-b49a
```

export that identifier as an env variable.

```bash
> export CONNID1=conn-c91b
> export CONNID2=conn-b49a
```

You can query the connection after establishment by

```bash
> ./gaiacli query ibc connection connection $CONNID1 --home ../node0/gaiacli --trust-node
{
  "connection": {
    "client": "client-09b6",
    "counterparty": "conn-e358",
    "path": {
      "type": "ibc/commitment/merkle/Path",
      "value": {
        "key_path": [
          "aWJj"
        ],
        "key_prefix": "djEv"
      }
    }
  },
  "available": true,
  "kind": "handshake"
}
```

See [script log](./conn.txt)

## Channel

Channels can be established with `channel.sh $CONNID1 $CONNID2` command.

You can query the channel after establishment by

```bash
> ./gaiacli query ibc channel channel ibc-mock $CHANID1 --home ../node0/gaiacli --trust-node
{
  "channel": {
    "Counterparty": "chan-f7b8",
    "CounterpartyPort": "ibc-mock",
    "ConnectionHops": [
      "conn-c91b"
    ]
  },
  "available": true,
  "sequence_send": "0",
  "sequence_receive": "0"
}
```

See [script log](./chan.txt)
