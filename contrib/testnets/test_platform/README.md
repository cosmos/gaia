# Gaiad Testnet Tool

This python tool starts multiple gaiad instances on the same machine without virtualization, i.e., non-conflicting ports are used.

This tool aims to simplify testing of key Cosmos Hub operations, such as module deployments and upgrades.

## Features

1. All ports automatically incremented by 10
1. Gaiad nodes peer with all other nodes
1. Gaiad nodes all started on one machine without conflict
1. All nodes generate, propose, and vote on blocks
1. Stopping app stops all instances
1. Support specifying a pre-existing genesis file
1. Supports taking a pre-existing genesis file and creating a network with a sufficient number of validators. The network
   creates as many validators as needed to attain majority voting power on the new network (and produce new blocks with pre-existing genesis file).
   The validators that are replaced is the set that provides at least 66% of the total voting power given in the genesis file.
   
   **This feature allows testing upgrades and module migrations of existing networks, using their pre-existing genesis** :star:


## Usage

1. Configure `template/replacement_defaults.txt`:
   1. To create a network from scratch:
      1. Set `replacement_genesis` value to blank, e.g., `replacement_genesis=`
      1. Set `num_of_nodes_to_apply` to the _number of nodes to run_, e.g., `num_of_nodes_to_apply=4`
   1. To create a network based on an existing genesis file:
      1. Set `replacement_genesis` to the source genesis file; `.tar.gz` files are also supported
      1. Set `replacement_genesis_make_safe` to `True` in order to create as many nodes as needed to run a majority of validators. 
      1. Otherwise, set `replacement_genesis_make_safe` value to blank to create `num_of_nodes_to_apply` nodes, e.g., `replacement_genesis_make_safe=`. 
         Important: if the `replacement_genesis_make_safe` is not set, then the validator keys in the genesis file aren't replaced and so the network may not produce new blocks.
   1. Optionally, set `LOG_LEVEL` to one of _(trace | debug | info | warn | error | fatal | panic)_; default _info_
1. Start  `gaiad_config_manager.py`

Notes for `template/replacement_defaults.txt`: 
- only the last occurrence of a key and it's value are used, i.e., earlier occurrences are overwritten.
- keys ending in `_PORT` are automatically incremented for each node

## Upcoming features

1. custom network architectures
1. custom failure testing
1. ibc integration testing
1. module integration testing