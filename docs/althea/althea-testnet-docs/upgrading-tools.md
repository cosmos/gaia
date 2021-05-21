# Download Althea chain and the Gravity tools

The gravity bridge tools package contains a number of programs for interacting with the bridge. These are updated much more frequently.

```

mkdir althea-bin
cd althea-bin

# Tools for the gravity bridge from the gravity repo

wget https://github.com/althea-net/althea-chain/releases/download/v0.2.1/client
wget https://github.com/althea-net/althea-chain/releases/download/v0.2.1/orchestrator
wget https://github.com/althea-net/althea-chain/releases/download/v0.2.1/register-delegate-keys
wget https://github.com/althea-net/althea-chain/releases/download/v0.2.1/relayer
chmod +x *
sudo mv * /usr/bin/

```

At specific points during the testnet you may be told to 'update your orchestrator' or 'update your althea binary'. In order to do that you can simply repeat the above instructions and then restart the affected software.

to check what version of the tools you have run `orchestrator --version` the current latest version is `0.4.3`
