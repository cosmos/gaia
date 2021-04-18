# How to become a validator on the Althea testnet!

## What do I need?

A Linux server with any modern Linux distribution, 4cores, 8gb of ram and at least 20gb of SSD storage.

In theory Althea chain can be run on Windows and Mac. Binaries are provided on the releases page. But instructions are not provided.

I also suggest an open notepad or other document to keep track of the keys you will be generating.

## Bootstrapping steps and commands

Start by logging into your Linux server using ssh. The following commands are intended to be run on that machine

### Install git and ansible

For Ubuntu / Debian

```
sudo apt get update
sudo apt-get install -y ansible git
```

For Centos

```
sudo yum install -y epel-release
sudo yum install -y ansible git
```

### Download the Althea chain repo

```
git clone https://github.com/althea-net/althea-chain
cd althea-chain/deployment-scripts
```

### Run the first time bootstrapping playbook and script

This script will print a lot of instructions, follow them carefully and be sure to copy
down the keys you generate. You will need them later.

```
ansible-playbook validator-prep.yml
bash shell/chain-setup.sh
```

### Get yourself some ALTG

You should have copied down your address in the last step, but if you didn't you can find it again with

```

althea keys list

```

Copy your address from the 'address' field and paste it into the command below remember to remove the `<>`

```

curl -vv -XPOST http://testnet2.althea.net/get_altg/<your address here without the brackets>

```

This will provide you 100 ALTG from the faucet storage.

### Run the delegate key and orchestrator setup script

```
bash shell/gravity-setup.sh
```

### Now it's finally time to set everything up and start it

```
ansible-playbook deploy-validator.yml
```

### Check the status of everything

You should be good to go! You can check the status of the three
major components of Althea chain by running

```
sudo journal -fu <service_name> (althea-chain, orchestrator, or geth)
```
