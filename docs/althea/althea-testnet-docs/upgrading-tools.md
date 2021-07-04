# Update the Althea chain tools

The gravity bridge tools package contains a number of programs for interacting with the bridge. These are updated much more frequently.

In order to download ARM binaries change the name in the wget link from ‘gbt' to gbt-arm'

```

mkdir althea-bin
cd althea-bin

# Tools for the gravity bridge from the gravity repo

wget https://github.com/althea-net/althea-chain/releases/download/v0.2.3/gbt
chmod +x *
sudo mv * /usr/bin/

```

At specific points during the testnet you may be told to 'update your orchestrator'. In order to do that you can simply repeat the above instructions and then restart the affected software.

to check what version of the tools you have run `gbt --version` the current latest version is `gbt 0.5.6`
