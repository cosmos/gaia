printf "############################################################################################# \n"
printf "#     Script for installing the Althea node by Paul D. Lovette | 04-02-2021                 # \n"
printf "#     Find me at https://skynet.paullovette.com/ or E-Mail: skynet@paullovette.com          # \n"
printf "#     While I have tested this script and it works great on my Ubuntu 20.0.4 install        # \n"
printf "#                                                                                           # \n"
printf "#               >>  No warranty is expressed. User assumes all risk  <<                     # \n"
printf "############################################################################################# \n \n \n"


printf "\n";
printf "$(date) Now that you have confirmed you are validating. Let's create/register our deleagate keys \n"

read -p 'Enter MNEMONIC from step 2: ' MNEMONIC

RUST_LOG=INFO register-delegate-keys --validator-phrase="\"$MNEMONIC\"" --cosmos-rpc="http://localhost:1317" --fees=footoken

printf "\n";
printf "$(date) Pausing 30 second so you can write down/save the above delegate keys and additional mnemonic for them. \n"
sleep 30;

printf "\n";
printf "$(date) Now lets fund our delegate keys. \n"

read -p 'Copy and paste the DELEGATE COSMOS address just displayed here (cosmos1xxxxxxxxxxxx.....xxxx): ' COSMOS_DELEGATE_ADDRESS

read -p 'Enter the name of your key created earlier with your validator address created in step 2: ' KEY_NAME

althea tx bank send $KEY_NAME $COSMOS_DELEGATE_ADDRESS 5000000footoken --chain-id=althea-testnet1v5

printf "\n";
printf "$(date) We have fund your COSMOS delegate address. \n"

printf "\n";
printf "$(date) Now lets fund our ETHEREUM delegate address. \n"

read -p 'Copy and paste the DELEGATE ETHEREUM address just displayed here (0xxx.....xxxx): ' ETHEREUM_DELEGATE_ADDRESS

curl -vv -XPOST http://testnet1.althea.net/get_eth/$ETHEREUM_DELEGATE_ADDRESS

printf "\n";
printf "$(date) Both delegate addresses are now funded.  Lets view your balances. \n"

read -p 'Enter a name your validator address from step 2 (cosmos1xxxxxxxxxxxx.....xxxx): ' VALIDATOR_ADDRESS

althea query bank balances $VALIDATOR_ADDRESS

printf "\n";
printf "$(date) Pausing 10secs \n"
sleep 10;

printf "\n";
printf "$(date) This is the Gravity Ethereum contract: 0xB48095a68501bC157654d338ce86fdaEF4071B24.  Keep it for reference. \n"

ETHEREUM_CONTRACT_ADDRESS=0xB48095a68501bC157654d338ce86fdaEF4071B24


printf "\n";
printf "$(date) *** Time to create the Orchestrator service and start it! *** \n"

sudo cp /home/$USER/althea-bin/orchestratord.service /lib/systemd/system/orchestratord.service

sed -i "s/^    --cosmos-phrase=.*/    --cosmos-phrase=\"$MNEMONIC\" "\\"\\"" /" /home/$USER/althea-bin/orchestratord.service
sed -i "s/^    --ethereum-key=.*/    --ethereum-key=\"$ETHEREUM_DELEGATE_ADDRESS\" "\\"\\"" /" /home/$USER/althea-bin/orchestratord.service
sed -i "s/^    --contract-address=.*/    --contract-address=\"$ETHEREUM_CONTRACT_ADDRESS\" "\\"\\"" /" /home/$USER/althea-bin/orchestratord.service

sudo systemctl enable orchestratord.service
sudo systemctl start orchestratord.service

printf "\n";
printf "$(date) *** It is finished!  You should now have a fully running Gravity Bridge Validator configuration! *** \n"

printf "\n";
printf "$(date) To view the status of Gravity-Bridge, Orchestrator or Geth, type: sudo journal -fu <service_name> (gravity-bridge, orchestratord, or geth) \n"
