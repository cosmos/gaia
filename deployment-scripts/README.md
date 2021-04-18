# Gravity Bridge Validator Setup Instructions:

## Prerequisites

### Setup base install of Linux.  I use Ubuntu 20.04.  Follow this guide:

[Linux setup and hardening page](https://github.com/lightiv/SkyNet/wiki/Ubuntu-Linux-Install-Guide)


### Fork this repository and clone it to your local PC

### Setup Ansible Control Workstation

Install Ansible:
```
sudo apt install ansible
```

Edit the inventory file to the IP address for the server(s) that your will me managing with Ansible

Edit the ansible.cfg to point to your SSH private key: 
```
private_key_file = 
```

## Step 1 - Configure your remote Gravity-Bridge node

From the Gravity Bridge git directory run:

```
ansible-playbook --ask-become-pass gravity-full-v005.yml
```

You will be prompted for various username and/or passwords.  Use the NON-ROOT user from the above "Setup base install of Linux."

## Step 2 - Log into your Gravity Bridge node with the NON-ROOT user from above

Run the following script and answer the prompts appropiately
```
~/althea-bin/setup1.sh
```

After the script finishes, move the the next step.


### Let your node sync before moving forward

Run the following command and look for "catching_up":false. If this says 'true' you are still syncing.

```
althea status
```

## Request some funds be sent to your address

Find your address.  Copy it from above when you created it or type the following:

```
althea keys list
```

Enter your keyring password and press ENTER

```
Copy your address from the 'address:' field and paste it into the command below
```

Looks like this: cosmos1xxxxxxxxxxxx.....xxxx

```
curl -vv -XPOST http://testnet1.althea.net/get_altg/cosmos1xxxxxxxxxxxx.....xxxx
```

This will provide you 10 ALTG from the faucet storage.


## Step 3 - Configure Validator

Return to your Gravity Bridge node and continue the setup.  Adjust these values and run the following command.
```
althea tx staking create-validator \
  --amount=9000000ualtg \
  --pubkey=$(althea tendermint show-validator) \
  --moniker=$MONIKER \
  --chain-id=althea-testnet1v5 \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1" \
  --fees 5000ualtg
  --gas="auto" \
  --gas-adjustment=1.5 \
  --gas-prices="0.025ualtg" \
  --from=<KEY_NAME_FROM_STEP_2>
```

## Confirm that you are validating

Lets get your validator address.  If it does not return an address you are not a validator:

```
althea keys show skynet --bech val --address
```

You will be prompted for your keyring passphrase.  Enter it and your should get your validator address:

```
cosmosvaloper1xxxxxxxxxxxx.....xxxx
```

If you did not get your validator address above you have not created a validator.  If you did get an address go to the next step to see if you are signing blocks:

Now that we have our validator address we can see if we are validating.

```
althea query staking validator cosmosvaloper1xxxxxxxxxxxx.....xxxx
```

In the output from the above command and look for the line ```status:``` It should say ```BOND_STATUS_BONDED```  If it does you are a validator and signing blocks.  If it says ```BOND_STATUS_UNBONDED``` your are jailed and not signing blocks.

If you do not get Orchestrator going in a timely manner your will also get ```jailed```

## Step 4 - Create our COSMOS and ETHEREUM delegate addresses, fund them, deploy our ETHEREUM contract address and finally start Orchestrator

From your Gravity Bridge node.  Finish the Gravity Bridge setup:
```
~/althea-bin/setup1.sh
```

** Answer the prompts when you get to messages below you have successfull set up a Gravity Bridge Validator node!


If you made it to the following messages.  Congratulation.  You are Gravity Bridge Validator!
```
*** It is finished!  You should now have a fully running Gravity Bridge Validator configuration! ***
The view the status of Gravity-Bridge, Orchestrator or Geth, type: sudo journal -fu <service_name> (gravity-bridge, orchestratord, or geth)
```
