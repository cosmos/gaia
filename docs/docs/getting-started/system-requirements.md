---
title: System requirements
sidebar_position: 4
---


# System requirements

<!-- 
## Validator Nodes

## Archive Nodes 
-->

## Gaia Upgrades

The Gaia application typically needs at least 32GB RAM, for smooth operation for upgrade, as there may be lengthy migrations to perform.

If you have less than 32GB RAM, you might try creating a swapfile to swap an idle program onto the hard disk to free up memory. This can allow your machine to run the binary than it could run in RAM alone.

```shell
# Linux instructions
sudo fallocate -l 16G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
```
