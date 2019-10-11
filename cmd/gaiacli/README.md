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

## Environment Setup

Stub out testnet files for 2 nodes, this example does so in your $HOME directory:

```shell
cd ~ && mkdir ibc-testnets && cd ibc-testnets
gaiad testnet -o ibc0 --v 1 --chain-id ibc0 --node-dir-prefix n
gaiad testnet -o ibc1 --v 1 --chain-id ibc1 --node-dir-prefix n
```

### Set `gaiacli` Configuation

Fix the configuration files for both `gaiad` and `gaiacli` to allow both chains/nodes to run on the same machine:

```bash
# Configure the proper database backend for each node and different listening ports
sed -i '' 's/"leveldb"/"goleveldb"/g' ibc0/n0/gaiad/config/config.toml
sed -i '' 's/"leveldb"/"goleveldb"/g' ibc1/n0/gaiad/config/config.toml
sed -i '' 's#"tcp://0.0.0.0:26656"#"tcp://0.0.0.0:26556"#g' ibc1/n0/gaiad/config/config.toml
sed -i '' 's#"tcp://0.0.0.0:26657"#"tcp://0.0.0.0:26557"#g' ibc1/n0/gaiad/config/config.toml
sed -i '' 's#"localhost:6060"#"localhost:6061"#g' ibc1/n0/gaiad/config/config.toml
sed -i '' 's#"tcp://127.0.0.1:26658"#"tcp://127.0.0.1:26558"#g' ibc1/n0/gaiad/config/config.toml
gaiacli config --home ibc0/n0/gaiacli/ chain-id ibc0
gaiacli config --home ibc1/n0/gaiacli/ chain-id ibc1
gaiacli config --home ibc0/n0/gaiacli/ output json
gaiacli config --home ibc1/n0/gaiacli/ output json
gaiacli config --home ibc0/n0/gaiacli/ node http://localhost:26657
gaiacli config --home ibc1/n0/gaiacli/ node http://localhost:26557
```

Add keys from each chain to the other and make such that the key at `ibc1/n0/gaiacli/key_seed.json` is named `n1` on each `gaiacli` instance and the same for `n0`. After this is complete the results of `gaiacli keys list` from each chain should be identical. The following are instructions for how to do this on Mac:

```bash
# These commands copy the seed phrase from each dir into the clipboard on mac
jq -r '.secret' ibc0/n0/gaiacli/key_seed.json | pbcopy
jq -r '.secret' ibc1/n0/gaiacli/key_seed.json | pbcopy

# Remove the key n0 on ibc1
gaiacli --home ibc1/n0/gaiacli keys delete n0

# seed from ibc1/n0/gaiacli/key_seed.json -> ibc0/n1
gaiacli --home ibc0/n0/gaiacli keys add n1 --recover

# seed from ibc0/n0/gaiacli/key_seed.json -> ibc1/n0
gaiacli --home ibc1/n0/gaiacli keys add n0 --recover

# seed from ibc1/n0/gaiacli/key_seed.json -> ibc1/n1
gaiacli --home ibc1/n0/gaiacli keys add n1 --recover

# Ensure keys match
gaiacli --home ibc0/n0/gaiacli keys list | jq '.[].address'
gaiacli --home ibc1/n0/gaiacli keys list | jq '.[].address'
```

After configuration is complete, start your `gaiad` processes:

```bash
nohup gaiad --home ibc0/n0/gaiad start > ibc0.log &
nohup gaiad --home ibc1/n0/gaiad start > ibc1.log &
```

## IBC Command Sequence

### Client Creation

Create IBC clients on each chain using the following commands. Note that we are using the consensus state of `ibc1` to create the client on `ibc0` and visa-versa. These "roots of trust" are used to validate transactions coming from the other chain. They will be updated periodically during handshakes and will require update at least once per unbonding period:

```bash
# client for chain ibc1 on chain ibc0
gaiacli --home ibc0/n0/gaiacli \
  tx ibc client create c0 \
  $(gaiacli --home ibc1/n0/gaiacli q ibc client consensus-state) \
  --from n0 -y -o text

# client for chain ibc0 on chain ibc1
gaiacli --home ibc1/n0/gaiacli \
  tx ibc client create c1 \
  $(gaiacli --home ibc0/n0/gaiacli q ibc client consensus-state) \
  --from n1 -y -o text
```

To query details about the clients use the following commands :

```bash
gaiacli --home ibc0/n0/gaiacli q ibc client client c0 --indent
gaiacli --home ibc1/n0/gaiacli q ibc client client c1 --indent
```

### Connection Creation

In order to send transactions using IBC there are two differnt handshakes that must be preformed. First there is a `connection` created between the two chains. Once the connection is created, an application specific `channel` handshake is preformed which allows the transfer of application specific data. Examples of applications are token transfer, cross-chain validation, cross-chain accounts, and in this tutorial `ibc-mock`.

Create a `connection` with the following command:

> NOTE: This command broadcasts a total of 7 transactions between the two chains from 2 different wallets. At the start of the command you will be prompted for passwords for the two different keys. The command may then take some time. Please wait for the command to return!

```shell
gaiacli \
  --home ibc0/n0/gaiacli \
  tx ibc connection handshake \
  conn0 c0 $(gaiacli --home ibc1/n0/gaiacli q ibc client path) \
  conn1 c1 $(gaiacli --home ibc0/n0/gaiacli q ibc client path) \
  --chain-id2 ibc1 \
  --from1 n0 --from2 n1 \
  --node1 tcp://localhost:26657 \
  --node2 tcp://localhost:26557
```

Once the connection is established you should be able to query it:

```bash
gaiacli --home ibc0/n0/gaiacli q ibc connection connection conn0 --indent --trust-node
gaiacli --home ibc1/n0/gaiacli q ibc connection connection conn1 --indent --trust-node
```

### Channel

Now that the `connection` has been created, its time to establish a `channel` for the `ibc-mock` application protocol. This will allow sending of data between `ibc0` and `ibc1`. To create the `channel`, run the following command:

> NOTE: This command broadcasts a total of 7 transactions between the two chains from 2 different wallets. At the start of the command you will be prompted for passwords for the two different keys. The command may then take some time. Please wait for the command to return!

```bash
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

You can query the `channel` after establishment by running the following command:

```bash
gaiacli --home ibc0/n0/gaiacli query ibc channel channel ibcmocksend chan0 --indent --trust-node
gaiacli --home ibc1/n0/gaiacli query ibc channel channel ibcmockrecv chan1 --indent --trust-node
```

## Send Packet

To send a packet using the `ibc-mock` application protocol, you need to know the channel you plan to send on, as well as the sequence number on the channel. To get the sequence you use the following commands:

```bash
# Returns the last sequence number
gaiacli --home ibc0/n0/gaiacli q ibcmocksend sequence chan0

# Returns the next expected sequence number, for use in scripting
gaiacli --home ibc0/n0/gaiacli q ibcmocksend next chan0
```

Now you are ready to send an `ibc-mock` packet down the channel (`chan0`) from chain `ibc0` to chain `ibc1`! To do so run the following command:

```bash
gaiacli --home ibc0/n0/gaiacli tx ibcmocksend sequence chan0 $(gaiacli --home ibc0/n0/gaiacli q ibcmocksend next chan0) --from n0 -o text
```

### Receive Packet

Once packets are sent, reciept must be confirmed on the destination chain. To receive the packets you just sent, run the following command:

```bash
gaiacli \
  --home ibc0/n0/gaiacli \
  tx ibc channel flush ibcmocksend chan0 \
  --node1 tcp://localhost:26657 \
  --node2 tcp://localhost:26557 \
  --chain-id2 ibc1 \
  --from1 n0 --from2 n1 -y -o text
```

Once the packets have been sent, check the To see the updated sequence run the following command:

```
gaiacli --home ibc1/n0/gaiacli q ibcmockrecv sequence chan1 --trust-node
```