import json
import os
import shutil
import subprocess
import time

working_directory = os.getcwd() + "/"

template_input = working_directory + 'templates/replacement_defaults.txt'

template_file_config = working_directory + 'templates/config.toml'
template_file_app = working_directory + 'templates/app.toml'

target_dir = working_directory + 'mytestnet/node0/gaiad/config/'

target_files = []
target_configs = []
target_apps = []

# port sequence parameters
port_sequence = 0
port_increment = 10

# take a template
# replace template parameters
# increment ports
# set peers
# apply genesis if one is specified
# overwrite configs for each target

# collect testnet validator pub keys
# copy genesis file to targets

# get genesis validators info
# choose which validators to replace
# create pubkey_replacement file
# use pubkey_replacement option on start

# read the template as a single string
with open(template_file_config, 'r') as file:
    # template_config = file.read().replace('\n', '')
    template_config = file.read()

with open(template_file_app, 'r') as file:
    template_app = file.read()
    # .replace('\n', '')

# populate template replacements from input file
template_replacements = {}
with open(template_input, 'r') as file:
    template_lines = file.readlines()
    for lin in template_lines:
        lin = lin.strip()
        if len(lin) == 0:
            continue
        line_separations = lin.split("=")
        template_replacements[line_separations[0]] = line_separations[1]


def make_replacements(template, port_sequence):
    intermediate_template = template
    # make template replacements
    for k in template_replacements:
        # allow special care for ports and their increments
        if k.endswith("_PORT"):
            intermediate_template = intermediate_template.replace("<" + k + ">", str(int(template_replacements[k]) + port_sequence))
        else:
            intermediate_template = intermediate_template.replace("<" + k + ">", template_replacements[k])
    return intermediate_template


local_sequence = 0


def get_validator_pubkey(target_dir):
    val_pub_key = subprocess.check_output(['gaiad', 'tendermint', 'show-validator', '--home', target_dir.rstrip('/config')])
    val_pub_key = val_pub_key.decode("utf-8").rstrip('\n')
    return val_pub_key


def get_validator_id(target_dir):
    global local_sequence
    # subprocess.call(['gaiad', 'init', '--home', target_dir])
    nodeid = subprocess.check_output(['gaiad', 'tendermint', 'show-node-id', '--home', str(target_dir).rstrip("config/")])
    peer_id = str(nodeid.decode("utf-8").rstrip('\n') + '@' + template_replacements['P2P_PEERID_IP'] + ':' + str(int(template_replacements['P2P_LADDR_PORT']) + int(local_sequence)))
    local_sequence += 10
    return peer_id


common_genesis = working_directory + template_replacements['replacement_genesis']
# support compressed genesis files
if common_genesis.endswith(".tar.gz"):
    unzip_cmd = "tar zxvf " + common_genesis + " --cd " + working_directory + "templates"
    print("unzip_cmd:" + unzip_cmd)
    subprocess.call(unzip_cmd, shell=True)
    common_genesis = common_genesis.rstrip(".tar.gz")

if len(template_replacements['replacement_genesis']) > 0:
    # cat genesis.cosmoshub-4.json| jq -s '.[].validators[] | { address: .address, power: .power, name: .name }'
    # genesis_validator_set = subprocess.check_output(['cat ' + str(common_genesis) + " | jq -s '.[].validators[] | { address: .address, power: .power, name: .name }'"], shell=True)
    genesis_validator_set = subprocess.check_output(['cat ' + str(common_genesis) + " | jq -s '.[].app_state.staking.validators[] | { address: .operator_address, power: .tokens, name: .description.moniker }'"], shell=True)
    print("genesis validator set:" + str(genesis_validator_set))
    genesis_valset_python = json.loads("[" + genesis_validator_set.decode("utf-8").replace("}", "},") + "{}]")
    # sort validator records by decreasing power
    # genesis_valset_sorted2 = sorted(genesis_valset_python[:-1], key=lambda k: print(k))
    genesis_valset_sorted2 = sorted(genesis_valset_python[:-1], key=lambda k: -int(k["power"]))
    print("sorted valset:" + str(genesis_valset_sorted2))

    total_power = 0
    for r in genesis_valset_sorted2:
        total_power += int(r["power"])

    safe_percentage = 0.66
    safe_absolute = total_power * safe_percentage
    safe_index = 0
    safe_index_scan_stop = False

    rolling_percentage = 0
    print("rolling percentage:")
    for i, r in enumerate(genesis_valset_sorted2):
        rolling_percentage += int(r["power"]) / total_power
        print(str(i) + ":" + str(rolling_percentage))
        if rolling_percentage > safe_percentage and not safe_index_scan_stop:
            safe_index_scan_stop = True
            safe_index = i
    print("liveness index:" + str(safe_index))

    if len(template_replacements['replacement_genesis_make_safe']) > 0:
        # gaiad testnet --keyring-backend test --v 4
        print("Creating testnet subdirectories")
        subprocess.call(['rm', '-rf', working_directory + 'mytestnet'])
        subprocess.call(['gaiad', 'testnet', '--keyring-backend', 'test', '--v', str(safe_index)])

    # specify the output
    for node_num in range(safe_index):
        target_file = target_dir.replace("node0", "node" + str(node_num))
        target_files.append(target_file)
        # subprocess.call("gaiad init node" + str(node_num) + " -o --home "+target_file.rstrip('/config'), shell=True)
        subprocess.call('gaiad unsafe-reset-all --home ' + target_file.rstrip('/config'), shell=True)

    print("target_files:"+str(target_files))

    # collect validator pubkeys for replacement
    testnet_validator_pubkeys = [get_validator_pubkey(t) for t in target_files]
    print('testnet validator pubkeys:' + str(testnet_validator_pubkeys))

    output_els = []
    for v_index in range(safe_index):
        output_els.append({
            "validator_name": genesis_valset_sorted2[v_index]["name"],
            "validator_address": genesis_valset_sorted2[v_index]["address"],
            "stargate_consensus_public_key": testnet_validator_pubkeys[v_index]
        })
    print("replacement keys:" + str(json.dumps(output_els)))

    with open(working_directory + 'templates/validator_replacement_output.json', 'w') as f:
        f.write(str(json.dumps(output_els)))

    # gaiad migrate cosmoshub_3_genesis_export.json --chain-id=cosmoshub-4 --initial-height [last_cosmoshub-3_block+1] > genesis.json
    print("migration genesis:" + str(common_genesis))
    cmd_string = 'gaiad migrate ' + common_genesis + ' --chain-id cosmoshub-4 --initial-height 0  --replacement-cons-keys ' + working_directory + 'templates/validator_replacement_output.json > ' + working_directory + 'templates/genesis_replaced.json'
    print("cmd_string:" + cmd_string)
    subprocess.call([cmd_string], shell=True)

    # compress genesis
    subprocess.call('tar zcvf ' + working_directory + 'templates/genesis_replaced.json.tar.gz --cd ' + working_directory + 'templates genesis_replaced.json', shell=True)

    common_genesis = working_directory + 'templates/genesis_replaced.json'

    # create each target's config files
    for target_file in target_files:
        # copy genesis if a file path to a genesis file is set
        print("common_genesis:" + common_genesis)
        shutil.copy2(common_genesis, target_file + 'genesis.json')
else:
    # gaiad testnet --keyring-backend test --v 4
    print("Creating testnet subdirectories")
    subprocess.call(['rm', '-rf', working_directory + 'mytestnet'])

    num_of_nodes_to_apply = int(template_replacements["num_of_nodes_to_apply"])

    # specify the output
    # target_files = []
    for node_num in range(num_of_nodes_to_apply):
        target_files.append(target_dir.replace("node0", "node" + str(node_num)))

    subprocess.call(['gaiad', 'testnet', '--keyring-backend', 'test', '--v', str(num_of_nodes_to_apply)])

peer_ids = [get_validator_id(t) for t in target_files]
peers = ",".join(peer_ids)
print("testnet peer ids:" + peers)
main_template_config = template_config.replace("<SEEDS>", peers)

# collect validator pubkeys for replacement
testnet_validator_pubkeys = [get_validator_pubkey(t) for t in target_files]
print('testnet validator pubkeys:' + str(testnet_validator_pubkeys))


# give the node some time to start if this is a genesis file with a lot of state, Cosmos Hub 4 mainnet requires at least 10 minutes
# time.sleep(60 * 10)
# tendermint_validator_set = subprocess.check_output(['gaiad', 'query', 'tendermint-validator-set']).decode("utf-8").rstrip('\n')
# print("tendermint_validator_set:" + tendermint_validator_set)

for target_file in target_files:
    # make replacements to app and config toml files
    current_template_config = make_replacements(main_template_config, port_sequence)
    target_configs.append(current_template_config)
    current_template_app = make_replacements(template_app, port_sequence)
    target_apps.append(current_template_app)
    port_sequence += port_increment

    # backup current file, but we choose to overwrite instead
    # shutil.copy2(file, file+"-"+str(time.time_ns())+".bak")

    # print(current_template_config)
    # print(current_template_app)

    # make sure target path exists
    os.makedirs(os.path.dirname(target_file), exist_ok=True)

    # save the config.toml production
    with open(target_file + 'config.toml', 'w') as f:
        f.write(current_template_config)

    # save the app.toml production
    with open(target_file + 'app.toml', 'w') as f:
        f.write(current_template_app)

    proc = subprocess.Popen(['gaiad', 'start', '--home', target_file.rstrip('/config'), '--x-crisis-skip-assert-invariants'])

    # automatically terminate program (and thus all gaiad instances) after some time

time.sleep(300)
