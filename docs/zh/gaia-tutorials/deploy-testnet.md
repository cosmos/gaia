# 部署你自己的测试网

这篇文章介绍了三种创建`gaiad`节点的测试网的方式，每种针对不同的使用场景：

1. 单节点，本地的，手动的测试网
2. 多节点，本地的，自动的测试网
3. 多节点，远程的，自动的测试网

支持代码可以在[networks 目录](https://github.com/cosmos/cosmos-sdk/tree/develop/networks)中找到，还可以在`local`或`remote`子目录中找到。

> 注意：`remote`网络引导教程可能与最新版本不同步，不可完全依赖。

## 可获取的 Docker 镜像

如果你需要使用或部署 gaia 作为容器，你可以跳过`build`步骤并使用官方镜像，\$TAG 标识你感兴趣的版本：

- `docker run -it -v ~/.gaia:/root/.gaia -v ~/.gaia:/root/.gaia tendermint:$TAG gaiad init`
- `docker run -it -p 26657:26657 -p 26656:26656 -v ~/.gaia:/root/.gaia -v ~/.gaia:/root/.gaia tendermint:$TAG gaiad start`
- ...
- `docker run -it -v ~/.gaia:/root/.gaia -v ~/.gaia:/root/.gaia tendermint:$TAG gaiad version`

相同的镜像也可以用于构建你自己的 docker-compose 栈

## 单节点，本地的，手动的测试网

本教程可帮助你创建一个在本地运行网络的验证人节点，以进行测试和其他相关的用途。

### 需要

- [安装 gaia](./installation.md)
- [安装`jq`](https://stedolan.github.io/jq/download/)(可选的)

### 创建 genesis 文件并启动网络

```bash
# You can run all of these commands from your home directory
cd $HOME

# Initialize the genesis.json file that will help you to bootstrap the network
gaiad init --chain-id=testing testing

# Create a key to hold your validator account
gaiad keys add validator

# Add that key into the genesis.app_state.accounts array in the genesis file
# NOTE: this command lets you set the number of coins. Make sure this account has some coins
# with the genesis.app_state.staking.params.bond_denom denom, the default is staking
gaiad add-genesis-account $(gaiad keys show validator -a) 1000000000stake,1000000000validatortoken

# Generate the transaction that creates your validator
gaiad gentx --name validator

# Add the generated bonding transaction to the genesis file
gaiad collect-gentxs

# Now its safe to start `gaiad`
gaiad start
```

启动将会把`gaiad`相关的所有数据放在`~/.gaia`目录。你可以检查所创建的 genesis 文件——`~/.gaia/config/genesis.json`。同时`gaiad`也已经配置完成并且有了一个拥有 token 的账户(stake 和自定义的代币)。

## 多节点，本地的，自动的测试网

在[networks/local 目录](https://github.com/cosmos/cosmos-sdk/tree/develop/networks/local)中运行如下命令:

### 需要

- [安装 gaia](./installation.md)
- [安装 docker](https://docs.docker.com/install/)
- [安装 docker-compose](https://docs.docker.com/compose/install/)

### 编译

编译`gaiad`二进制文件(linux)和运行`localnet`命令所需的`tendermint/gaianode` docker images。这个二进制文件将被安装到 container 中，并且可以更新重建 image，因此您只需要构建一次 image。

```bash
# Clone the gaia repo
git clone https://github.com/cosmos/gaia.git

# Work from the SDK repo
cd cosmos-sdk

# Build the linux binary in ./build
make build-linux

# Build tendermint/gaiadnode image
make build-docker-gaiadnode
```

### 运行你的测试网

运行一个拥有 4 个节点的测试网络:

```bash
make localnet-start
```

此命令使用 gaiadnode image 创建了一个 4 节点网络。每个节点的端口可以在下表中找到：

| `Node ID`   | `P2P Port` | `RPC Port` |
| ----------- | ---------- | ---------- |
| `gaianode0` | `26656`    | `26657`    |
| `gaianode1` | `26659`    | `26660`    |
| `gaianode2` | `26661`    | `26662`    |
| `gaianode3` | `26663`    | `26664`    |

更新可执行程序，只需要重新编译并重启节点:

```bash
make build-linux localnet-start
```

### 配置

`make localnet-start`命令通过调用`gaiad testnet`命令在`./build`中创建了一个 4 节点测试网络的文件。输出`./build`目录下一些文件:

```bash
$ tree -L 2 build/
build/
├── gaiad
├── gaiad
├── gentxs
│   ├── node0.json
│   ├── node1.json
│   ├── node2.json
│   └── node3.json
├── node0
│   ├── gaiad
│   │   ├── key_seed.json
│   │   └── keys
│   └── gaiad
│       ├── ${LOG:-gaiad.log}
│       ├── config
│       └── data
├── node1
│   ├── gaiad
│   │   └── key_seed.json
│   └── gaiad
│       ├── ${LOG:-gaiad.log}
│       ├── config
│       └── data
├── node2
│   ├── gaiad
│   │   └── key_seed.json
│   └── gaiad
│       ├── ${LOG:-gaiad.log}
│       ├── config
│       └── data
└── node3
    ├── gaiad
    │   └── key_seed.json
    └── gaiad
        ├── ${LOG:-gaiad.log}
        ├── config
        └── data
```

每个`./build/nodeN`目录被挂载到对应 container 的`/gaiad`目录。

### 日志输出

日志被保存在每个`./build/nodeN/gaiad/gaia.log`文件中。你也可以直接通过 Docker 来查看日志：

```bash
docker logs -f gaiadnode0
```

### 密钥&账户

你需要使用指定节点的`gaiad`目录作为你的`home`来同`gaiad`交互，并执行查询或者创建交易:

```bash
gaiad keys list --home ./build/node0/gaiad
```

现在账户已经存在了，你可以创建新的账户并向其发送资金！

::: 提示
注意：每个节点的密钥种子放在`./build/nodeN/gaiad/key_seed.json`中，可以通过`gaiad keys add --restore`命令来回复。
:::

### 特殊的可执行程序

如果你拥有多个不同名称的可执行程序，则可以使用 BINARY 环境变量指定要运行的可执行程序。可执行程序的路径是相对于挂载的卷。例如：

```
# Run with custom binary
BINARY=gaiafoo make localnet-start
```

## 多节点，远程的，自动的测试网

应该从[networks 目录](https://github.com/cosmos/cosmos-sdk/tree/develop/networks)运行下面的命令。

### Terraform & Ansible

使用[Terraform](https://www.terraform.io/)在 AWS 上创建服务器然后用[Ansible](https://www.ansible.com/)创建并管理这些服务器上的测试网来完成自动部署。

### 前提

- 在一台 Linux 机器上安装[Terraform](https://www.terraform.io/downloads.html)和[Ansible](https://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html)
- 创建一个具有 EC2 实例创建能力的[ASW API token](https://docs.aws.amazon.com/general/latest/gr/managing-aws-access-keys.html)
- 创建 SSH 密钥

```
export AWS_ACCESS_KEY_ID="2345234jk2lh4234"
export AWS_SECRET_ACCESS_KEY="234jhkg234h52kh4g5khg34"
export TESTNET_NAME="remotenet"
export CLUSTER_NAME= "remotenetvalidators"
export SSH_PRIVATE_FILE="$HOME/.ssh/id_rsa"
export SSH_PUBLIC_FILE="$HOME/.ssh/id_rsa.pub"
```

`terraform`和`ansible`都会使用到。

### 创建一个远程网络

```
SERVERS=1 REGION_LIMIT=1 make validators-start
```

测试网络的名称将由`--chain-id`定义，集群的名称则是 AWS 中服务器管理标识。该代码将在每个可用区中创建服务器数量的服务器，最多为 REGION_LIMIT，从 us-east-2 开始。（us-east-1 被排除在外）下面的 BaSH 脚本也是如此，但更便于输入。

```
./new-testnet.sh "$TESTNET_NAME" "$CLUSTER_NAME" 1 1
```

### 快速查询状态入口

```
make validators-status
```

### 删除服务器

```
make validators-stop
```

### 日志输出

你可以将日志发送到 Logz.io，一个 Elastic 栈（Elastic 搜索，Logstash 和 Kibana）服务提供商。你可以将节点设置为自动登录。创建一个帐户并从此页面上的说明中获取你的 API 密钥，然后：

```
yum install systemd-devel || echo "This will only work on RHEL-based systems."
apt-get install libsystemd-dev || echo "This will only work on Debian-based systems."

go get github.com/mheese/journalbeat
ansible-playbook -i inventory/digital_ocean.py -l remotenet logzio.yml -e LOGZIO_TOKEN=ABCDEFGHIJKLMNOPQRSTUVWXYZ012345
```

### 监控

你可以安装 DataDog 代理：

```
make datadog-install
```
