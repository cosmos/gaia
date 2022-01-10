BINARY=gaiad
MNEMONIC_2="record gift you once hip style during joke field prize dust unique length more pencil transfer quit train device arrive energy sort steak upset"
MNEMONIC_b="uphold train large action document mixed exact cherry input evil sponsor digital used child engine fire attract sing little jeans decrease despair unfair what"
CHAINID2=test-2
HOME2=/Users/yaruwang/code/interchain-accounts-gaia/test-2
P2PPORT_2=26656
RPCPORT_2=26657
GRPCPORT_2=9096
GRPCWEBPORT_2=9082
RESTPORT_2=1317
ROSETTA_2=8081


# Stop if it is already running
if pgrep -x  "$BINARY" >/dev/null; then
    echo "Terminating $BINARY..."
    killall gaiad
fi

echo "Removing previous data..."
rm -rf $HOME2 &> /dev/null

# Add directories for both chains, exit if an error occurs
if ! mkdir -p $HOME2 2>/dev/null; then
    echo "Failed to create gaiad folder. Aborting..."
    exit 1
fi

echo "Initializing $CHAINID2..."
gaiad init test --chain-id=$CHAINID2 --home=$HOME2

echo "Adding genesis accounts..."
$BINARY keys add val2 --keyring-backend=test --home=$HOME2
echo $MNEMONIC_b | $BINARY keys add b  --recover --keyring-backend=test --home=$HOME2
echo $MNEMONIC_2 | $BINARY keys add rly --recover --keyring-backend=test --home=$HOME2

$BINARY add-genesis-account $($BINARY keys show val2 --keyring-backend test -a --home=$HOME2) 100000000000stake --home=$HOME2  --keyring-backend test
$BINARY add-genesis-account $($BINARY keys show b --keyring-backend test -a --home=$HOME2) 100000000000stake --home=$HOME2 --keyring-backend test
$BINARY add-genesis-account $($BINARY keys show rly --keyring-backend test -a --home=$HOME2) 100000000000stake --home=$HOME2 --keyring-backend test

echo "Creating and collecting gentx..."
$BINARY gentx val2 7000000000stake --chain-id $CHAINID2 --keyring-backend test --home=$HOME2
$BINARY collect-gentxs --home=$HOME2

echo "Changing defaults and ports in app.toml and config.toml files..."
sed -i -e 's#"tcp://0.0.0.0:26656"#"tcp://0.0.0.0:'"$P2PPORT_2"'"#g' $HOME2/config/config.toml
sed -i -e 's#"tcp://127.0.0.1:26657"#"tcp://0.0.0.0:'"$RPCPORT_2"'"#g' $HOME2/config/config.toml
sed -i -e 's/"0.0.0.0:9090"/"0.0.0.0:'"$GRPCPORT_2"'"/g' $HOME2/config/app.toml
sed -i -e 's/"0.0.0.0:9091"/"0.0.0.0:'"$GRPCWEBPORT_2"'"/g' $HOME2/config/app.toml
sed -i -e 's/timeout_commit = "5s"/timeout_commit = "1s"/g' $HOME2/config/config.toml
sed -i -e 's/timeout_propose = "3s"/timeout_propose = "1s"/g' $HOME2/config/config.toml
sed -i -e 's/index_all_keys = false/index_all_keys = true/g' $HOME2/config/config.toml
sed -i -e 's/enable = false/enable = true/g' $HOME2/config/app.toml
sed -i -e 's/swagger = false/swagger = true/g' $HOME2/config/app.toml
sed -i -e 's#"tcp://0.0.0.0:1317"#"tcp://0.0.0.0:'"$RESTPORT_2"'"#g' $HOME2/config/app.toml
sed -i -e 's#":8080"#":'"$ROSETTA_2"'"#g' $HOME2/config/app.toml

# Update host chain genesis to allow all msg types
sed -i '' 's%\"allow_messages\":.*%\"allow_messages\": [\"\/cosmos.crypto.ed25519.PubKey\", \"/cosmos.crypto.secp256k1.PubKey\", \"/cosmos.crypto.multisig.LegacyAminoPubKey\", \"/cosmos.crypto.secp256r1.PubKey\", \"/cosmos.vesting.v1beta1.PeriodicVestingAccount\", \"/cosmos.vesting.v1beta1.PermanentLockedAccount\", \"/cosmos.auth.v1beta1.BaseAccount\", \"/cosmos.auth.v1beta1.ModuleAccount\", \"/cosmos.vesting.v1beta1.BaseVestingAccount\", \"/cosmos.vesting.v1beta1.DelayedVestingAccount\", \"/cosmos.vesting.v1beta1.ContinuousVestingAccount\", \"/cosmos.bank.v1beta1.SendAuthorization\", \"/cosmos.authz.v1beta1.GenericAuthorization\", \"/cosmos.staking.v1beta1.StakeAuthorization\", \"/ibc.lightclients.tendermint.v1.ClientState\", \"/ibc.lightclients.localhost.v1.ClientState\", \"/ibc.lightclients.solomachine.v2.ClientState\", \"/ibc.core.connection.v1.Version\", \"/ibc.core.channel.v1.Counterparty\", \"/ibc.core.channel.v1.Packet\", \"/ibc.core.commitment.v1.MerklePrefix\", \"/cosmos.feegrant.v1beta1.PeriodicAllowance\", \"/cosmos.feegrant.v1beta1.AllowedMsgAllowance\", \"/cosmos.feegrant.v1beta1.BasicAllowance\", \"/ibc.lightclients.solomachine.v2.ConsensusState\", \"/ibc.lightclients.tendermint.v1.ConsensusState\", \"/ibc.core.client.v1.Height\", \"/ibc.core.connection.v1.ConnectionEnd\", \"/ibc.core.commitment.v1.MerklePath\", \"/ibc.core.commitment.v1.MerkleProof\", \"/ibc.lightclients.solomachine.v2.Misbehaviour\", \"/ibc.lightclients.tendermint.v1.Misbehaviour\", \"/ibc.core.commitment.v1.MerkleRoot\", \"/ibc.core.connection.v1.MsgConnectionOpenAck\", \"/ibc.core.channel.v1.MsgTimeoutOnClose\", \"/cosmos.feegrant.v1beta1.MsgRevokeAllowance\", \"/ibc.applications.transfer.v1.MsgTransfer\", \"/cosmos.authz.v1beta1.MsgGrant\", \"/ibc.core.channel.v1.MsgChannelCloseInit\", \"/cosmos.crisis.v1beta1.MsgVerifyInvariant\", \"/cosmos.evidence.v1beta1.MsgSubmitEvidence\", \"/cosmos.distribution.v1beta1.MsgSetWithdrawAddress\", \"/ibc.core.connection.v1.MsgConnectionOpenConfirm\", \"/ibc.core.channel.v1.MsgChannelOpenAck\", \"/ibc.core.channel.v1.MsgChannelCloseConfirm\", \"/ibc.core.channel.v1.MsgRecvPacket\", \"/cosmos.feegrant.v1beta1.MsgGrantAllowance\", \"/cosmos.gov.v1beta1.MsgSubmitProposal\", \"/cosmos.authz.v1beta1.MsgRevoke\", \"/cosmos.authz.v1beta1.MsgExec\", \"/ibc.core.connection.v1.MsgConnectionOpenTry\", \"/cosmos.staking.v1beta1.MsgEditValidator\", \"/tendermint.liquidity.v1beta1.MsgSwapWithinBatch\", \"/cosmos.bank.v1beta1.MsgSend\", \"/cosmos.vesting.v1beta1.MsgCreateVestingAccount\", \"/ibc.core.client.v1.MsgUpgradeClient\", \"/ibc.core.channel.v1.MsgChannelOpenInit\", \"/ibc.core.channel.v1.MsgAcknowledgement\", \"/cosmos.staking.v1beta1.MsgCreateValidator\", \"/tendermint.liquidity.v1beta1.MsgCreatePool\", \"/tendermint.liquidity.v1beta1.MsgDepositWithinBatch\", \"/cosmos.distribution.v1beta1.MsgFundCommunityPool\", \"/ibc.core.client.v1.MsgUpdateClient\", \"/ibc.core.connection.v1.MsgConnectionOpenInit\", \"/ibc.core.channel.v1.MsgChannelOpenTry\", \"/cosmos.staking.v1beta1.MsgBeginRedelegate\", \"/cosmos.bank.v1beta1.MsgMultiSend\", \"/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission\", \"/cosmos.slashing.v1beta1.MsgUnjail\", \"/ibc.core.channel.v1.MsgTimeout\", \"/cosmos.gov.v1beta1.MsgVoteWeighted\", \"/cosmos.gov.v1beta1.MsgDeposit\", \"/ibc.core.client.v1.MsgCreateClient\", \"/ibc.core.client.v1.MsgSubmitMisbehaviour\", \"/ibc.core.channel.v1.MsgChannelOpenConfirm\", \"/cosmos.staking.v1beta1.MsgDelegate\", \"/cosmos.staking.v1beta1.MsgUndelegate\", \"/tendermint.liquidity.v1beta1.MsgWithdrawWithinBatch\", \"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward\", \"/cosmos.gov.v1beta1.MsgVote\", \"/cosmos.vesting.v1beta1.BaseVestingAccount\", \"/cosmos.vesting.v1beta1.DelayedVestingAccount\", \"/cosmos.vesting.v1beta1.ContinuousVestingAccount\", \"/cosmos.vesting.v1beta1.PeriodicVestingAccount\", \"/cosmos.vesting.v1beta1.PermanentLockedAccount\", \"/cosmos.auth.v1beta1.BaseAccount\", \"/cosmos.auth.v1beta1.ModuleAccount\", \"/cosmos.vesting.v1beta1.ContinuousVestingAccount\", \"/cosmos.vesting.v1beta1.DelayedVestingAccount\", \"/cosmos.vesting.v1beta1.PeriodicVestingAccount\", \"/cosmos.vesting.v1beta1.PermanentLockedAccount\", \"/ibc.lightclients.solomachine.v2.Header\", \"/ibc.lightclients.tendermint.v1.Header\", \"/ibc.core.connection.v1.Counterparty\", \"/ibc.core.channel.v1.Channel\", \"/cosmos.tx.v1beta1.Tx\", \"/cosmos.evidence.v1beta1.Equivocation\", \"/cosmos.bank.v1beta1.Supply\", \"/cosmos.params.v1beta1.ParameterChangeProposal\", \"/ibc.core.client.v1.ClientUpdateProposal\", \"/ibc.core.client.v1.UpgradeProposal\", \"/cosmos.upgrade.v1beta1.SoftwareUpgradeProposal\", \"/cosmos.upgrade.v1beta1.CancelSoftwareUpgradeProposal\", \"/cosmos.distribution.v1beta1.CommunityPoolSpendProposal\", \"/cosmos.gov.v1beta1.TextProposal\"]%g' $HOME2/config/genesis.json


 echo "Starting $CHAINID2 in ~/.gaia ..."
 echo "Creating log file at gaia.log"
 $BINARY start --home=$HOME2 --log_level=trace --log_format=json --pruning=nothing > gaia.log 2>&1 &
