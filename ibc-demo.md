# IBC instruction

## Dependencies

This branch uses non-canonical branch of cosmos-sdk. Before building, run `go mod vendor` on the root directory to retrieve the dependencies. To build:

```shell
git clone git@github.com:cosmos/gaia
cd gaia
git checkout fedekunze/ibc
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

### Set `gaiad` and `gaiacli` Configuation

Fix the configuration files for both `gaiad` and `gaiacli` to allow both chains/nodes to run on the same machine:

```bash
# Configure the proper database backend for each node and different listening ports
sed -i'.orig' -e 's/"leveldb"/"goleveldb"/g' ibc0/n0/gaiad/config/config.toml
sed -i 's/"leveldb"/"goleveldb"/g' ibc1/n0/gaiad/config/config.toml
sed -i 's#"tcp://0.0.0.0:26656"#"tcp://0.0.0.0:26556"#g' ibc1/n0/gaiad/config/config.toml
sed -i 's#"tcp://0.0.0.0:26657"#"tcp://0.0.0.0:26557"#g' ibc1/n0/gaiad/config/config.toml
sed -i 's#"localhost:6060"#"localhost:6061"#g' ibc1/n0/gaiad/config/config.toml
sed -i 's#"tcp://127.0.0.1:26658"#"tcp://127.0.0.1:26558"#g' ibc1/n0/gaiad/config/config.toml
gaiacli config --home ibc0/n0/gaiacli/ chain-id ibc0
gaiacli config --home ibc1/n0/gaiacli/ chain-id ibc1
gaiacli config --home ibc0/n0/gaiacli/ output json
gaiacli config --home ibc1/n0/gaiacli/ output json
gaiacli config --home ibc0/n0/gaiacli/ node http://localhost:26657
gaiacli config --home ibc1/n0/gaiacli/ node http://localhost:26557
```

Add keys from each chain to the other and make sure that the key at `ibc1/n0/gaiacli/key_seed.json` is named `n1` on each `gaiacli` instance and the same for `n0`. After this is complete the results of `gaiacli keys list` from each chain should be identical. The following commands will do the trick:

```bash
gaiacli --home ibc1/n0/gaiacli keys delete n0
gaiacli keys test --home ibc0/n0/gaiacli n1 "$(jq -r '.secret' ibc1/n0/gaiacli/key_seed.json)" 12345678
gaiacli keys test --home ibc1/n0/gaiacli n0 "$(jq -r '.secret' ibc0/n0/gaiacli/key_seed.json)" 12345678
gaiacli keys test --home ibc1/n0/gaiacli n1 "$(jq -r '.secret' ibc1/n0/gaiacli/key_seed.json)" 12345678
```

After this operation, check to make sure the keys match:

```bash
gaiacli --home ibc0/n0/gaiacli keys list | jq -r '.[].address'
gaiacli --home ibc1/n0/gaiacli keys list | jq -r '.[].address'
```

After configuration is complete, you will be able to start two `gaiad` processes:

```bash
nohup gaiad --home ibc0/n0/gaiad start > ibc0.log &
nohup gaiad --home ibc1/n0/gaiad start > ibc1.log &
```

> NOTE: If you would like to look at the logs from the instances just `tail -f ibc0.log`.

## IBC Command Sequence

### Client Creation

Create IBC clients on each chain using the following commands. Note that we are using the consensus state of `ibc1` to create the client on `ibc0` and visa-versa. These "roots of trust" are used to validate transactions coming from the other chain. They will be updated periodically during handshakes and will require update at least once per unbonding period:

```bash
# client for chain ibc1 on chain ibc0
echo -e "12345678\n" | gaiacli --home ibc0/n0/gaiacli \
  tx ibc client create ibconeclient \
  $(gaiacli --home ibc1/n0/gaiacli q ibc client node-state) \
  --from n0 -y -o text

# client for chain ibc0 on chain ibc1
echo -e "12345678\n" | gaiacli --home ibc1/n0/gaiacli \
  tx ibc client create ibczeroclient \
  $(gaiacli --home ibc0/n0/gaiacli q ibc client node-state) \
  --from n1 -y -o text
```

To query details about the clients use the following commands:

```bash
gaiacli --home ibc0/n0/gaiacli q ibc client consensus-state ibconeclient --indent
gaiacli --home ibc1/n0/gaiacli q ibc client consensus-state ibczeroclient --indent
```

### Connection Creation

In order to send transactions using IBC there are two different handshakes that must be performed. First there is a `connection` created between the two chains. Once the connection is created, an application specific `channel` handshake is performed which allows the transfer of application specific data. Examples of applications are token transfer, cross-chain validation, cross-chain accounts, and in this tutorial `ibc-mock`.

Create a `connection` with the following command:

> NOTE: This command broadcasts a total of 7 transactions between the two chains from 2 different wallets. At the start of the command you will be prompted for passwords for the two different keys (`12345678` for both). The command will then take some time, please wait for it to return!

```shell
gaiacli \
  --home ibc0/n0/gaiacli \
  tx ibc connection handshake \
  connectionzero ibconeclient $(gaiacli --home ibc1/n0/gaiacli q ibc client path) \
  connectionone ibczeroclient $(gaiacli --home ibc0/n0/gaiacli q ibc client path) \
  --chain-id2 ibc1 \
  --from1 n0 --from2 n1 \
  --node1 tcp://localhost:26657 \
  --node2 tcp://localhost:26557
```

After the password input, you should see output like the following:

```
ibc0 <- connection_open_init    [OK] txid(B41C15A8F31524CB34EE061BA4418F48A3A37A7348BF8F818E67F5EE90AED45F) client(ibconeclient) connection(connectionzero)
ibc1 <- update_client           [OK] txid(CD9F9AFD311DDF6B3604BCDD3DF371FAE93F76A01E66815AD6E2DCDDA942976D) client(ibconeclient)
ibc1 <- connection_open_try     [OK] txid(F364081569703146978605D9079B751FA83E8766C749B1DA96054D2728DBD715) client(ibczeroclient) connection(connectionone)
ibc0 <- update_client           [OK] txid(58A8E012E623303AA2B4A73099D00A6A3F34220E10FA5445251B1F7E6435EFF5) client(ibczeroclient)
ibc0 <- connection_open_ack     [OK] txid(9535B4E25E204C129B91CF2FBDBD3E9AC17881AB4D4BFD4F57731E96D348B053) connection(connectionzero)
ibc1 <- update_client           [OK] txid(50D737D7798E1D0A2E0452B7EDDC2D06A06524424F10B54F6400F148D9255DAF) client(ibconeclient)
ibc0 <- connection_open_confirm [OK] txid(CF0F7E54481D90A438A625E16EAB48F5480334AF090CE05736BDFACB69B8F798) connection(connectionone)
```

Once the connection is established you should be able to query it:

```bash
gaiacli --home ibc0/n0/gaiacli q ibc connection end connectionzero --indent --trust-node
gaiacli --home ibc1/n0/gaiacli q ibc connection end connectionone --indent --trust-node
```

### Channel

Now that the `connection` has been created, it's time to establish a `channel` for the `ibc-mock` application protocol. This will allow sending of data between `ibc0` and `ibc1`. To create the `channel`, run the following command:

> NOTE: This command broadcasts a total of 7 transactions between the two chains from 2 different wallets. At the start of the command you will be prompted for passwords for the two different keys (`12345678` for both). The command will then take some time, please wait for it to return!

```bash
gaiacli \
  --home ibc0/n0/gaiacli \
  tx ibc channel handshake \
  ibconeclient bank channelzero connectionzero \
  ibczeroclient bank channelone connectionone \
  --node1 tcp://localhost:26657 \
  --node2 tcp://localhost:26557 \
  --chain-id2 ibc1 \
  --from1 n0 --from2 n1
```

You should see output like the following:

```
ibc0 <- channel_open_init       [OK] txid(792E51E0455A8E0C85705C61A638A4D7C5399B3BA5AF6F29C85BB4E090FCA1B7) portid(bankbankbank) chanid(channelzero)
ibc1 <- update_client           [OK] txid(CEA961B9BE931E7B06E6D5643486D267677E66A253F104BC00E2BCE1F9343C03) client(ibczeroclient)
ibc1 <- channel_open_try        [OK] txid(D6BC3B03646EF61D1DA153C6678FE047DF76CB12981AC1D524C69C22124967D7) portid(bankbankbank) chanid(channelone)
ibc0 <- update_client           [OK] txid(FA22E93601218CEA839FDEB7BD0D8F47D81E1172E18A8B21717675FF5C4BCF40) client(ibconeclient)
ibc0 <- channel_open_ack        [OK] txid(1840343AFB2D5666F52440C199A1356C63A884A959F0A9A53175773CDB83006B) portid(bankbankbank) chanid(channelzero)
ibc1 <- update_client           [OK] txid(BBE212C5041AC366C018BB97F8DF8A495562EAFFFF51BFDA7BBAE9952BE589D0) client(ibczeroclient)
ibc1 <- channel_open_confirm    [OK] txid(69F50CA44AE6AD84BD24866E7DB7FE8ADFD9C484171662CB9E6F0C71BFC222A9) portid(bankbankbank) chanid(channelone)
```

You can query the `channel` after establishment by running the following command:

```bash
gaiacli --home ibc0/n0/gaiacli q ibc channel end bank channelzero --indent --trust-node
gaiacli --home ibc1/n0/gaiacli q ibc channel end bank channelone --indent --trust-node
```

### Send Packet

To send a packet using the `bank` application protocol, you need to know the `channel` you plan to send on, as well as the `port` on the channel. You also need to provide an `address` and `amount`. Use the following command to send the packet:

```bash
gaiacli \
  --home ibc0/n0/gaiacli \
  tx ibc transfer transfer \
  bank channelzero \
  $(gaiacli --home ibc0/n0/gaiacli keys show n1 -a) 1stake \
  --from n0 \
  --source
```

> NOTE: This commands returns the `height` at which it was committed, this should be at the beginning of the JSON output. The enviornment variable `TIMEOUT`.

### Receive Packet

Now, try querying the account on `ibc1` that you sent the `1stake` to, the account will be empty:

```bash
gaiacli --home ibc1/n0/gaiacli q account $(gaiacli --home ibc0/n0/gaiacli keys show n1 -a) --indent --trust-node
```

To complete the transfer once packets are sent, receipt must be confirmed on the destination chain. To `recv-packet` from `ibc0` on `ibc1`, run the following command:

```bash
gaiacli \
  tx ibc transfer recv-packet \
  bank channelzero ibczeroclient \
  --home ibc1/n0/gaiacli \
  --packet-sequence 1 \
  --timeout $TIMEOUT \
  --from n1 \
  --node2 tcp://localhost:26657 \
  --chain-id2 ibc0 \
  --source
```

Once the packets have been recieved you should see the `1stake` in your account on `ibc1`:

```bash
gaiacli --home ibc1/n0/gaiacli q account $(gaiacli --home ibc0/n0/gaiacli keys show n1 -a)
```
