# IBC instruction

// temporal document

## Dependencies

This branch uses non-canonical branch of cosmos-sdk. Before building, run `go mod vendor` on the root directory to retrive the dependencies. To build:

```shell
git clone git@github.com:cosmos/gaia
cd gaia
git checkout joon/ibc-gaia-interface
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
# Configure the proper database backend for each node and different listening ports
sed -i '' 's/"leveldb"/"goleveldb"/g' ibc0/n0/gaiad/config/config.toml
sed -i '' 's/"leveldb"/"goleveldb"/g' ibc1/n0/gaiad/config/config.toml
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

# Add the key from ibc1 to the ibc0 cli
jq -r '.secret' ibc1/n0/gaiacli/key_seed.json | pbcopy

# Paste the mnemonic from the above command after setting password (12345678)
gaiacli --home ibc0/n0/gaiacli keys add n1 --recover
```

After configuration is complete, start each node in a seperate terminal window:

```bash
gaiad --home ibc0/n0/gaiad start
gaiad --home ibc1/n0/gaiad start
```

## Client

Create a client on ibc1:

```bash
gaiacli --home ibc0/n0/gaiacli q ibc client path > path0.json
gaiacli --home ibc0/n0/gaiacli q ibc client consensus-state > state0.json
gaiacli --home ibc1/n0/gaiacli tx ibc client create c1 ./state0.json --from n0 -y
gaiacli --home ibc1/n0/gaiacli q ibc client client c1
```

Create a client on ibc0:

```bash
gaiacli --home ibc1/n0/gaiacli q ibc client path > path1.json
gaiacli --home ibc1/n0/gaiacli q ibc client consensus-state > state1.json
gaiacli --home ibc0/n0/gaiacli tx ibc client create c0 ./state1.json --from n0 -y
gaiacli --home ibc0/n0/gaiacli q ibc client client c0
```

## Connection

Create a connection with the following command:

```shellls

gaiacli \
  --home ibc0/n0/gaiacli \
  tx ibc connection handshake \
  conn0 c0 path1.json \
  conn1 c1 path0.json \
  --chain-id2 ibc1 \
  --from1 n0 --from2 n1 \
  --node1 tcp://localhost:26657 \
  --node2 tcp://localhost:26557
```

Once the connection is established you should be able to query it:

```bash
gaiacli --home ibc0/n0/gaiacli q ibc connection connection conn0 --trust-node
gaiacli --home ibc1/n0/gaiacli q ibc connection connection conn1 --trust-node
```

## Channel

To establish a channel using the `ibc-mock` application protocol run the following command:

```
gaiacli \
  --home ibc0/n0/gaiacli \
  tx ibc channel handshake \
  ibcmocksend chan0 conn0 \
  ibcmockrecv chan1 conn1 \
  --node1 tcp://localhost:26657 \
  --node2 tcp://localhost:26557 \
  --chain-id2 ibc1 \
  --from1 n0 --from2 n1
```

You can query the channel after establishment by running the following command

```bash
gaiacli --home ibc0/n0/gaiacli query ibc channel channel ibcmocksend chan0 --trust-node
gaiacli --home ibc1/n0/gaiacli query ibc channel channel ibcmockrecv chan1 --trust-node
```

## Send Packet

To send a packet using the `ibc-mock` application protocol run the following command:

```
gaiacli --home ibc0/n0/gaiacli q ibcmocksend sequence chan0
```

The command will return the latest sent sequence, `0` if not exists. Run command with next sequence (n+1).

```
gaiacli --home ibc0/n0/gaiacli tx ibcmocksend sequence chan0 1 --from n0
```

## Receive Packet

To receive packets using the `ibc-mock` application protocol run the following command:

```
gaiacli \
  --home ibc0/n0/gaiacli \
  tx ibc channel flush ibcmocksend chan0 \
  --node1 tcp://localhost:26657 \
  --node2 tcp://localhost:26557 \
  --chain-id2 ibc1 \
  --from1 n0 --from2 n1
```

To see the updated sequence run the following command:

```
gaiacli --home ibc1/n0/gaiacli q ibcmockrecv sequence chan1 --trust-node
```
