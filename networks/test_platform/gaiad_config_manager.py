import os
import subprocess
import time

template_input = '/Users/shahank/git_interchain/gaia2/networks/test_platform/templates/replacement_defaults.txt'

template_file_config = '/Users/shahank/git_interchain/gaia2/networks/test_platform/templates/config.toml'
template_file_app = '/Users/shahank/git_interchain/gaia2/networks/test_platform/templates/app.toml'

target_dir = '/Users/shahank/git_interchain/gaia2/mytestnet/node0/gaiad/config/'

# take a template
# replace template parameters
# increment ports
# set peers
# take a target directory
# overwrite configs for each target

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

num_of_nodes_to_apply = int(template_replacements["num_of_nodes_to_apply"])

# gaiad testnet --keyring-backend test --v 4
print("Creating testnet subdirectories")
subprocess.call(['gaiad', 'testnet', '--keyring-backend', 'test', '--v', str(num_of_nodes_to_apply)])

# specify the output
target_files = []
for node_num in range(num_of_nodes_to_apply):
    target_files.append(target_dir.replace("node0", "node"+str(node_num)))

target_configs = []
target_apps = []

# port sequence parameters
port_sequence = 0
port_increment = 10


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


def get_validator_id(target_dir):
    global peers
    global local_sequence
    # subprocess.call(['gaiad', 'init', '--home', target_dir])
    nodeid = subprocess.check_output(['gaiad', 'tendermint', 'show-node-id', '--home', str(target_dir).rstrip("config/")])
    peer_id = str(nodeid.decode("utf-8").rstrip('\n') + '@' + template_replacements['P2P_PEERID_IP'] + ':' + str(int(template_replacements['P2P_LADDR_PORT']) + int(local_sequence)))
    local_sequence += 10
    return peer_id
    # peers = peers + "," + peer_id
    # subprocess.call(['gaiad', 'start', '--home', target_dir])


peer_ids = [get_validator_id(t) for t in target_files]
peers = ",".join(peer_ids)
print("peer_ids:" + peers)
main_template_config = template_config.replace("<SEEDS>", peers)

# create each target's config files
for target_file in target_files:
    # make replacements to app and config toml files
    current_template_config = make_replacements(main_template_config, port_sequence)
    target_configs.append(current_template_config)
    current_template_app = make_replacements(template_app, port_sequence)
    target_apps.append(current_template_app)
    port_sequence += port_increment

    # backup current file, but we choose to overwrite instead
    # shutil.copy2(file, file+"-"+str(time.time_ns())+".bak")

    print(current_template_config)
    print(current_template_app)

    # make sure target path exists
    os.makedirs(os.path.dirname(target_file), exist_ok=True)

    # save the config.toml production
    with open(target_file + 'config.toml', 'w') as f:
        f.write(current_template_config)

    # save the app.toml production
    with open(target_file + 'app.toml', 'w') as f:
        f.write(current_template_app)

    proc = subprocess.Popen(['gaiad', 'start', '--home', target_file.rstrip('/config')])


# automatically terminate program (and thus all gaiad instances) after some time
time.sleep(300)
