BINARY=gaiad
MNEMONIC_1="alley afraid soup fall idea toss can goose become valve initial strong forward bright dish figure check leopard decide warfare hub unusual join cart"

MNEMONIC_a="captain six loyal advice caution cost orient large mimic spare radar excess quote orchard error biology choice shop dish master quantum dumb accident between"
CHAINID1=test-1
HOME1=/Users/yaruwang/code/interchain-accounts-gaia/test-1
P2PPORT_1=16656
RPCPORT_1=16657
GRPCPORT_1=9095
GRPCWEBPORT_1=9081
RESTPORT_1=1316
ROSETTA_1=8080

# Stop if it is already running
if pgrep -x  "$BINARY" >/dev/null; then
    echo "Terminating $BINARY..."
    killall gaiad
fi

echo "Removing previous data..."
rm -rf $HOME1 &> /dev/null

# Add directories for both chains, exit if an error occurs
if ! mkdir -p $HOME1 2>/dev/null; then
    echo "Failed to create gaiad folder. Aborting..."
    exit 1
fi

echo "Initializing $CHAINID1..."
gaiad init test --chain-id=$CHAINID1 --home=$HOME1

echo "Adding genesis accounts..."
$BINARY keys add val1 --keyring-backend=test --home=$HOME1
echo $MNEMONIC_a | $BINARY keys add a  --recover --keyring-backend=test --home=$HOME1
echo $MNEMONIC_1 | $BINARY keys add rly --recover --keyring-backend=test --home=$HOME1
$BINARY add-genesis-account $($BINARY keys show val1 --keyring-backend test -a --home=$HOME1) 100000000000stake --home=$HOME1  --keyring-backend test
$BINARY add-genesis-account $($BINARY keys show a --keyring-backend test -a --home=$HOME1) 100000000000stake --home=$HOME1 --keyring-backend test
$BINARY add-genesis-account $($BINARY keys show rly --keyring-backend test -a --home=$HOME1) 100000000000stake --home=$HOME1 --keyring-backend test

echo "Creating and collecting gentx..."
$BINARY gentx val1 7000000000stake --chain-id $CHAINID1 --keyring-backend test --home=$HOME1
$BINARY collect-gentxs --home=$HOME1

echo "Changing defaults and ports in app.toml and config.toml files..."
sed -i -e 's#"tcp://0.0.0.0:26656"#"tcp://0.0.0.0:'"$P2PPORT_1"'"#g' $HOME1/config/config.toml
sed -i -e 's#"tcp://127.0.0.1:26657"#"tcp://0.0.0.0:'"$RPCPORT_1"'"#g' $HOME1/config/config.toml
sed -i -e 's/"0.0.0.0:9090"/"0.0.0.0:'"$GRPCPORT_1"'"/g' $HOME1/config/app.toml
sed -i -e 's/"0.0.0.0:9091"/"0.0.0.0:'"$GRPCWEBPORT_1"'"/g' $HOME1/config/app.toml
sed -i -e 's/timeout_commit = "5s"/timeout_commit = "1s"/g' $HOME1/config/config.toml
sed -i -e 's/timeout_propose = "3s"/timeout_propose = "1s"/g' $HOME1/config/config.toml
sed -i -e 's/index_all_keys = false/index_all_keys = true/g' $HOME1/config/config.toml
sed -i -e 's/enable = false/enable = true/g' $HOME1/config/app.toml
sed -i -e 's/swagger = false/swagger = true/g' $HOME1/config/app.toml
sed -i -e 's#"tcp://0.0.0.0:1317"#"tcp://0.0.0.0:'"$RESTPORT_1"'"#g' $HOME1/config/app.toml
sed -i -e 's#":8080"#":'"$ROSETTA_1"'"#g' $HOME1/config/app.toml

# Update host chain genesis to allow all msg types
sed -i '' 's%\"allow_messages\":.*%\"allow_messages\": [\"\/cosmos.crypto.ed25519.PubKey\", \"/cosmos.crypto.secp256k1.PubKey\", \"/cosmos.crypto.multisig.LegacyAminoPubKey\", \"/cosmos.crypto.secp256r1.PubKey\", \"/cosmos.vesting.v1beta1.PeriodicVestingAccount\", \"/cosmos.vesting.v1beta1.PermanentLockedAccount\", \"/cosmos.auth.v1beta1.BaseAccount\", \"/cosmos.auth.v1beta1.ModuleAccount\", \"/cosmos.vesting.v1beta1.BaseVestingAccount\", \"/cosmos.vesting.v1beta1.DelayedVestingAccount\", \"/cosmos.vesting.v1beta1.ContinuousVestingAccount\", \"/cosmos.bank.v1beta1.SendAuthorization\", \"/cosmos.authz.v1beta1.GenericAuthorization\", \"/cosmos.staking.v1beta1.StakeAuthorization\", \"/ibc.lightclients.tendermint.v1.ClientState\", \"/ibc.lightclients.localhost.v1.ClientState\", \"/ibc.lightclients.solomachine.v2.ClientState\", \"/ibc.core.connection.v1.Version\", \"/ibc.core.channel.v1.Counterparty\", \"/ibc.core.channel.v1.Packet\", \"/ibc.core.commitment.v1.MerklePrefix\", \"/cosmos.feegrant.v1beta1.PeriodicAllowance\", \"/cosmos.feegrant.v1beta1.AllowedMsgAllowance\", \"/cosmos.feegrant.v1beta1.BasicAllowance\", \"/ibc.lightclients.solomachine.v2.ConsensusState\", \"/ibc.lightclients.tendermint.v1.ConsensusState\", \"/ibc.core.client.v1.Height\", \"/ibc.core.connection.v1.ConnectionEnd\", \"/ibc.core.commitment.v1.MerklePath\", \"/ibc.core.commitment.v1.MerkleProof\", \"/ibc.lightclients.solomachine.v2.Misbehaviour\", \"/ibc.lightclients.tendermint.v1.Misbehaviour\", \"/ibc.core.commitment.v1.MerkleRoot\", \"/ibc.core.connection.v1.MsgConnectionOpenAck\", \"/ibc.core.channel.v1.MsgTimeoutOnClose\", \"/cosmos.feegrant.v1beta1.MsgRevokeAllowance\", \"/ibc.applications.transfer.v1.MsgTransfer\", \"/cosmos.authz.v1beta1.MsgGrant\", \"/ibc.core.channel.v1.MsgChannelCloseInit\", \"/cosmos.crisis.v1beta1.MsgVerifyInvariant\", \"/cosmos.evidence.v1beta1.MsgSubmitEvidence\", \"/cosmos.distribution.v1beta1.MsgSetWithdrawAddress\", \"/ibc.core.connection.v1.MsgConnectionOpenConfirm\", \"/ibc.core.channel.v1.MsgChannelOpenAck\", \"/ibc.core.channel.v1.MsgChannelCloseConfirm\", \"/ibc.core.channel.v1.MsgRecvPacket\", \"/cosmos.feegrant.v1beta1.MsgGrantAllowance\", \"/cosmos.gov.v1beta1.MsgSubmitProposal\", \"/cosmos.authz.v1beta1.MsgRevoke\", \"/cosmos.authz.v1beta1.MsgExec\", \"/ibc.core.connection.v1.MsgConnectionOpenTry\", \"/cosmos.staking.v1beta1.MsgEditValidator\", \"/tendermint.liquidity.v1beta1.MsgSwapWithinBatch\", \"/cosmos.bank.v1beta1.MsgSend\", \"/cosmos.vesting.v1beta1.MsgCreateVestingAccount\", \"/ibc.core.client.v1.MsgUpgradeClient\", \"/ibc.core.channel.v1.MsgChannelOpenInit\", \"/ibc.core.channel.v1.MsgAcknowledgement\", \"/cosmos.staking.v1beta1.MsgCreateValidator\", \"/tendermint.liquidity.v1beta1.MsgCreatePool\", \"/tendermint.liquidity.v1beta1.MsgDepositWithinBatch\", \"/cosmos.distribution.v1beta1.MsgFundCommunityPool\", \"/ibc.core.client.v1.MsgUpdateClient\", \"/ibc.core.connection.v1.MsgConnectionOpenInit\", \"/ibc.core.channel.v1.MsgChannelOpenTry\", \"/cosmos.staking.v1beta1.MsgBeginRedelegate\", \"/cosmos.bank.v1beta1.MsgMultiSend\", \"/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission\", \"/cosmos.slashing.v1beta1.MsgUnjail\", \"/ibc.core.channel.v1.MsgTimeout\", \"/cosmos.gov.v1beta1.MsgVoteWeighted\", \"/cosmos.gov.v1beta1.MsgDeposit\", \"/ibc.core.client.v1.MsgCreateClient\", \"/ibc.core.client.v1.MsgSubmitMisbehaviour\", \"/ibc.core.channel.v1.MsgChannelOpenConfirm\", \"/cosmos.staking.v1beta1.MsgDelegate\", \"/cosmos.staking.v1beta1.MsgUndelegate\", \"/tendermint.liquidity.v1beta1.MsgWithdrawWithinBatch\", \"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward\", \"/cosmos.gov.v1beta1.MsgVote\", \"/cosmos.vesting.v1beta1.BaseVestingAccount\", \"/cosmos.vesting.v1beta1.DelayedVestingAccount\", \"/cosmos.vesting.v1beta1.ContinuousVestingAccount\", \"/cosmos.vesting.v1beta1.PeriodicVestingAccount\", \"/cosmos.vesting.v1beta1.PermanentLockedAccount\", \"/cosmos.auth.v1beta1.BaseAccount\", \"/cosmos.auth.v1beta1.ModuleAccount\", \"/cosmos.vesting.v1beta1.ContinuousVestingAccount\", \"/cosmos.vesting.v1beta1.DelayedVestingAccount\", \"/cosmos.vesting.v1beta1.PeriodicVestingAccount\", \"/cosmos.vesting.v1beta1.PermanentLockedAccount\", \"/ibc.lightclients.solomachine.v2.Header\", \"/ibc.lightclients.tendermint.v1.Header\", \"/ibc.core.connection.v1.Counterparty\", \"/ibc.core.channel.v1.Channel\", \"/cosmos.tx.v1beta1.Tx\", \"/cosmos.evidence.v1beta1.Equivocation\", \"/cosmos.bank.v1beta1.Supply\", \"/cosmos.params.v1beta1.ParameterChangeProposal\", \"/ibc.core.client.v1.ClientUpdateProposal\", \"/ibc.core.client.v1.UpgradeProposal\", \"/cosmos.upgrade.v1beta1.SoftwareUpgradeProposal\", \"/cosmos.upgrade.v1beta1.CancelSoftwareUpgradeProposal\", \"/cosmos.distribution.v1beta1.CommunityPoolSpendProposal\", \"/cosmos.gov.v1beta1.TextProposal\"]%g' $HOME1/config/genesis.json
 echo "Starting $CHAINID1 in ~/.gaia ..."
 echo "Creating log file at gaia.log"
 $BINARY start --home=$HOME1 --log_level=trace --log_format=json --pruning=nothing > gaia.log 2>&1 &
