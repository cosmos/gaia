# IBC instruction

// temporal document

## Dependencies

This branch uses non-canonical branch of cosmos-sdk. Before building, run `go mod vendor` on the root directory to retrive the dependencies. To build:

```shell
git clone git@github.com:cosmos/gaia
cd gaia
go mod vendor
make install
gaiad version
gaiacli version
```

Stub out testnet files for 2 nodes, this example does so in your $HOME directory:

```shell
cd ~ && mkdir ibc-testnets && cd ibc-testnet
gaiad testnet -o ibc0 --v 1 --chain-id ibc0 --node-dir-prefix n
gaiad testnet -o ibc1 --v 1 --chain-id ibc1 --node-dir-prefix n
```

Fix the configuration files to allow both chains/nodes to run on the same machine

```shell
# Configure the proper database backend for each node
sed -i '' 's/"leveldb"/"goleveldb"/g' ibc0/n0/gaiad/config/config.toml
sed -i '' 's/"leveldb"/"goleveldb"/g' ibc1/n0/gaiad/config/config.toml

# Configure chain ibc1 to have different listening ports
sed -i '' 's#"tcp://0.0.0.0:26656"#"tcp://0.0.0.0:26556"#g' ibc1/n0/gaiad/config/config.toml
sed -i '' 's#"tcp://0.0.0.0:26657"#"tcp://0.0.0.0:26557"#g' ibc1/n0/gaiad/config/config.toml
sed -i '' 's#"localhost:6060"#"localhost:6061"#g' ibc1/n0/gaiad/config/config.toml
sed -i '' 's#"tcp://127.0.0.1:26658"#"tcp://127.0.0.1:26558"#g' ibc1/n0/gaiad/config/config.toml
```

Then configure your `gaiacli` instances for each chain:

```bash
gaiacli config --home ibc0/n0/gaiacli/ chain-id ibc0
gaiacli config --home ibc1/n0/gaiacli/ chain-id ibc1
gaiacli config --home ibc0/n0/gaiacli/ node http://localhost:26657
gaiacli config --home ibc1/n0/gaiacli/ node http://localhost:26557
```

After configuration is complete, start each node in a seperate terminal window:

```bash
gaiad --home ibc0/n0/gaiad start
gaiad --home ibc1/n0/gaiad start
```

## Client

Create a client on ibc0:

```bash
gaiacli --home ibc0/n0/gaiacli q ibc client path > path.json
gaiacli --home ibc0/n0/gaiacli q ibc client consensus-state > state0.json
gaiacli --home ibc0/n0/gaiacli tx ibc client create c0 ./state0.json --from n0
gaiacli --home ibc0/n0/gaiacli q ibc client client c0
```

Create a client on ibc1:

```bash
gaiacli --home ibc1/n0/gaiacli q ibc client path > path.json
gaiacli --home ibc1/n0/gaiacli q ibc client consensus-state > state1.json
gaiacli --home ibc1/n0/gaiacli tx ibc client create c1 ./state1.json --from n0
gaiacli --home ibc1/n0/gaiacli q ibc client client c1
```

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
