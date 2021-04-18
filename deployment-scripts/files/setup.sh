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
althea init $MONIKER --chain-id althea-testnet1v5

printf "\n";
printf "$(date) Edit app.toml to prevent spam and activate the api server needed by Gravity Bridge. \n";
sed -E -i 's/minimum-gas-prices = \".*\"/minimum-gas-prices = \"0.025ualtg\"/' ~/.althea/config/app.toml
sed -E -i 's/enable = false/enable = true/' ~/.althea/config/app.toml
sed -E -i 's/persistent_peers = \".*\"/persistent_peers = \"05ded2f258ab158c5526eb53aa14d122367115a7@testnet1.althea.net:26656\"/' ~/.althea/config/config.toml
sed -E -i 's/pex =.*/pex = false/' ~/.althea/config/config.toml
sed -E -i 's/max_open_connections = 3.*/max_open_connections = 20/' ~/.althea/config/config.toml

# Copy the genesis.json to the .althea config directory
printf "\n";
printf "$(date) Copy the genesis.json to the .althea config directory. \n";

sudo cp /home/$USER/althea-bin/genesis.json /home/$USER/.althea/config/genesis.json

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
printf "$(date) Setting up the Geth Ethereum Light Client \n";

cd $HOME
tar -xvf geth-linux-amd64-1.10.1-c2d2f4ed.tar.gz

chmod +x /home/$USER/geth-linux-amd64-1.10.1-c2d2f4ed/geth
sudo cp /home/$USER/geth-linux-amd64-1.10.1-c2d2f4ed/geth /usr/bin/geth

printf "\n";
printf "$(date) Finalizing Gravity Bridge Services \n";

sed -i "s/^User=.*/"User=$USER"/" /home/$USER/althea-bin/gravity-bridge.service
sed -i "s/^User=.*/"User=$USER"/" /home/$USER/althea-bin/geth.service

sudo cp /home/$USER/althea-bin/gravity-bridge.service /lib/systemd/system/gravity-bridge.service
sudo cp /home/$USER/althea-bin/geth.service /lib/systemd/system/geth.service

printf "\n";
printf "$(date) Open inbound firewall port 26656 for P2P network stability: \n";

sudo ufw allow from any to any port 26656  proto tcp

printf "\n";
printf "$(date) Enable and start the Gravity Bridge and Geth system services: \n";

sudo systemctl enable gravity-bridge.service
sudo systemctl start gravity-bridge.service

sudo systemctl enable geth.service
sudo systemctl start geth.service

printf "\n";
printf "$(date) *** Allow your node to completely sync before proceeding with Enable and start the Gravity Bridge and Geth system services: *** \n";

printf "\n";
printf "$(date) You can view what Althea is doing by typing 'sudo journalctl -fu gravity-bridge' \n";

printf "\n";
printf "$(date) Type 'althea status' and press ENTER.  In the output you should see "Catching Up = False" once synced. \n";

printf "\n";
printf "$(date) *** Once synced. type the folloiwng *** \n";
