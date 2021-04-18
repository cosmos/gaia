printf "############################################################################################# \n"
printf "#     Script for installing the Althea node by Paul D. Lovette | 04-02-2021                 # \n"
printf "#     Find me at https://skynet.paullovette.com/ or E-Mail: skynet@paullovette.com          # \n"
printf "#     While I have tested this script and it works great on my Ubuntu 20.0.4 install        # \n"
printf "#                                                                                           # \n"
printf "#               >>  No warranty is expressed. User assumes all risk  <<                     # \n"
printf "############################################################################################# \n \n \n"

printf "\n";
printf "$(date) *** Initialize the Althea directory *** \n";
read -p 'Enter a name for your node/validator: ' MONIKER;

cd $HOME
althea init $MONIKER --chain-id althea-testnet2v1

printf "\n";
printf "$(date) *** Let create a wallet without a Ledger device *** \n";

read -p 'Enter a name for your new Key: ' KEY_NAME;

printf "\n";
printf "$(date) Key name entered: $KEY_NAME \n";

althea keys add $KEY_NAME;

printf "\n";
printf "$(date) *** STOP *** Copy the displayed key above! *** \n";

printf "\n";
printf "$(date) *** STOP *** Copy the displayed mnemonic above! *** \n";

printf "\n";
printf "$(date) *** PAUSING 30 SECONDS *** Copy/Write down the displayed ADRESS & MNEMONIC above! *** \n";

printf "\n";
printf "$(date) *** Allow your node to completely sync before proceeding with Enable and start the Gravity Bridge and Geth system services: *** \n";

printf "\n";
printf "$(date) You can view what Althea is doing by typing 'sudo journalctl -fu althea-chain' \n";

printf "\n";
printf "$(date) Type 'althea status' and press ENTER.  In the output you should see "Catching Up = False" once synced. \n";

printf "\n";
printf "$(date) *** Once synced. Proceed to the next step in the instructions *** \n";
