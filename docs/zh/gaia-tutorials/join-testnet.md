# 加入公共测试网

::: 提示 当前测试网
请查看[testnet repo](https://github.com/cosmos/testnets)获取最新的公共测试网信息，包含了所使用的 Cosmos-SDK 的正确版本和 genesis 文件。
:::

::: 警告
你需要先完成[安装`gaia`](./installation.md)
:::

## 创建一个新节点

> 注意：如果你在之前的测试网中运行过一个全节点，请跳至[升级之前的 Testnet](#upgrading-from-previous-testnet)。

要创建一个新节点，主网的指令同样适用：

- [加入 mainnet](./join-mainnet.md)
- [部署验证人节点](./validators/validator-setup.md)

只有 SDK 的版本和 genesis 文件不同。查看[testnet repo](https://github.com/cosmos/testnets)
获取测试网的信息，包括 Cosmos-SDK 正确的版本和 genesis 文件。

## 升级之前的 Testnet

这些指令用以把运行过以前测试网络的全节点升级至最新的测试网络。

### 重置数据

首先，移除过期的文件并重置数据：

```bash
rm $HOME/.gaia/config/addrbook.json $HOME/.gaia/config/genesis.json
gaiad unsafe-reset-all
```

你的节点现在处于原始状态并保留了最初的`priv_validator.json`文件和`config.toml`文件。如果之前你还有其他的哨兵节点或者全节点，你的节点仍然会连接他们，但是会失败，因为他们还没有升级。

::: 警告
确保每个节点有一个独一无二的`priv_validator.json`文件。不要从一个旧节点拷贝`priv_validator.json`到多个新的节点。运行两个有着相同`priv_validator.json`文件的节点会导致双签。
:::

### 升级软件

现在升级软件：

```bash
git clone https://github.com/cosmos/cosmos-sdk.git
cd cosmos-sdk
git fetch --all && git checkout master
make install
```

::: 提示
_注意_：如果在这一步出现问题，请检查是否安装了最新稳定版本的 Go。
:::

注意这里我们使用的是包含最新稳定发布版本的`master`分支。请查看[testnet repo](https://github.com/cosmos/testnets)查看哪个版本的测试网需要哪一个 Cosmos-SDK 版本，在[SDK 发布版](https://github.com/cosmos/cosmos-sdk/releases)中对应的详细信息。

你的全节点已经升级成功！
