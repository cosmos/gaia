# 최신 퍼블릭 테스트넷에 참가하기

::: tip 최신 테스트넷
최신 테스트넷에 대한 정보는 다음의 [테스트넷 리포](https://github.com/cosmos/testnets)를 참고하세요. 어떤 코스모스 SDK 버전과 제네시스 파일에 대한 정보가 포합되어있습니다.
:::

::: warning
**여기에서 더 진행하시기 전에 [gaia](./installation.md)가 꼭 설치되어있어야 합니다.**
:::

## 새로운 노드 세팅하기


다음 절차는 새로운 풀노드를 처음부터 세팅하는 절차입니다.

우선 노드를 실행하고 필요한 config 파일을 생성합니다:


```bash
gaiad init <your_custom_moniker>
```

::: warning 참고
`--moniker`는 ASCII 캐릭터만을 지원합니다. Unicode 캐릭터를 이용하는 경우 노드 접근이 불가능할 수 있으니 참고하세요.
:::

`moniker`는 `~/.gaia/config/config.toml` 파일을 통해 추후에 변경이 가능합니다:

```toml
# A custom human readable name for this node
moniker = "<your_custom_moniker>"
```

최소 가스 가격보다 낮은 트랜잭션을 거절하는 스팸 방지 메커니즘을 활성화 하시려면 `~/.gaia/config/gaiad.toml` 파일을 변경하시면 됩니다:

```
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

##### main base config options #####

# The minimum gas prices a validator is willing to accept for processing a
# transaction. A transaction's fees must meet the minimum of any denomination
# specified in this config (e.g. 10uatom).

minimum-gas-prices = ""
```

이제 풀노드가 활성화 되었습니다!

## 제네시스와 시드

### 제네시스 파일 복사하기

테스트넷의 `genesis.json`파일을 `gaiad`의 config 디렉토리로 가져옵니다.

```bash
mkdir -p $HOME/.gaia/config
curl https://raw.githubusercontent.com/cosmos/launch/master/genesis.json > $HOME/.gaia/config/genesis.json
```

위 예시에서는 최신 테스트넷에 대한 정보가 포함되어있는 [launch repo](https://github.com/cosmos/launch)의 `latest` 디렉토리를 이용하는 것을 참고하세요. 

::: tip
만약 다른 퍼블릭 테스트넷에 연결하신다면 [여기](./join-testnet.md)에 있는 정보를 확인하세요.
:::

설정이 올바르게 작동하는지 확인하기 위해서는 다음을 실행하세요:

```bash
gaiad start
```
### 시드 노드 추가하기

이제 노드가 다른 피어들을 찾는 방법을 알아야합니다. `$HOME/.gaia/config/config.toml`에 안정적인 시드 노드들을 추가할 차례입니다. [`launch`](https://github.com/cosmos/launch) repo에 몇 개 시드 노드 링크가 포함되어있습니다.

만약 해당 시드가 작동하지 않는다면, 추가적인 시드와 피어들을 코스모스 허브 익스플로러에서 확인하세요(목록은 [launch](https://cosmos.network/launch) 페이지에 있습니다.)

이 외에도 [밸리데이터 라이엇 채팅방](https://riot.im/app/#/room/#cosmos-validators:matrix.org)을 통해서 피어 요청을 할 수 있습니다.

시드와 피어에 대한 더 많은 정보를 원하시면 [여기](https://github.com/tendermint/tendermint/blob/develop/docs/tendermint-core/using-tendermint.md#peers)를 확인하세요.

### 가스와 수수료에 대해서

::: warning
코스모스 메인넷에서는 `uatom` 단위가 표준 단위로 사용됩니다. `1atom = 1,000,000uatom`으로 환산됩니다.
:::

코스모스 허브 네트워크는 트랜잭션 처리를 위해 트랜잭션 수수료를 부과합니다. 해당 수수료는 트랜잭션을 실행하기 위한 가스로 사용됩니다. 공식은 다음과 같습니다:


```
수수료(Fee) = 가스(Gas) * 가스 값(GasPrices)
```

위 공식에서 `gas`는 전송하는 트랜잭션에 따라 다릅니다. 다른 형태의 트랜잭션은 각자 다른 `gas`량을 필요로 합니다. `gas` 수량은 트랜잭션이 실행될때 계산됨으로 사전에 정확한 값을 확인할 수 있는 방법은 없습니다. 다만, `gas` 플래그의 값을 `auto`로 설정함으로 예상 값을 추출할 수는 있습니다. 예상 값을 수정하기 위해서는 `--gas-adjustment` (기본 값 `1.0`) 플래그 값을 변경하셔서 트랜잭션이 충분한 가스를 확보할 수 있도록 하십시오.

`gasPrice`는 각 `gas` 유닛의 가격입니다. 각 검증인은 직접 최소 가스 가격인 `min-gas-price`를 설정하며, 트랜잭션의 `gasPrice`가 설정한 `min-gas-price`보다 높을때 트랜잭션을 처리합니다.

트랜잭션 피(`fees`)는 `gas` 수량과 `gasPrice`를 곱한 값입니다. 유저는 3개의 값 중 2개의 값을 입력하게 됩니다. `gasPrice`가 높을수록 트랜잭션이 블록에 포함될 확률이 높아집니다.

::: tip
메인넷 권장 `gas-prices`는 `0.0025uatom` 입니다.
:::

## 최소 가스 가격(`minimum-gas-prices`) 설정하기

풀노드는 컨펌되지 않은 트랜잭션을 멤풀에 보관합니다. 스팸 트랜잭션으로부터 풀노드를 보호하기 위해서 노드 멤풀에 보관되기 위한 트랜잭션의 최소 가스 가격(`minimum-gas-prices`)을 설정할 것을 권장합니다. 해당 파라미터는 `~/.gaia/config/gaiad.toml`에서 설정하실 수 있씁니다.

기본 권장 `minimum-gas-prices`는 `0.0025uatom`이지만, 추후 바꾸실 수 있습니다. 

## 풀노드 운영하기

다음 커맨드로 풀노드를 시작하세요:

```bash
gaiad start
```

모든 것이 잘 작동하고 있는지 확인하기 위해서는:

```bash
gaiad status
```

네트워크 상태를 [코스모스 익스플로러](https://cosmos.network/launch)에서 확인하세요.

## 상태 내보내기(Export State)

Gaia는 현재 애플리케이션의 상태를 JSON파일 형태로 내보낼 수 있습니다. 이런 데이터는 수동 분석과 새로운 네트워크의 제네시스 파일로 이용될 때 유용할 수 있습니다.

현재 상태를 내보내기 위해서는:

```bash
gaiad export > [filename].json
```

특정 블록 높이의 상태를 내보낼 수 있습니다(해당 블록 처리 후 상태):

```bash
gaiad export --height [height] > [filename].json
```

만약 해당 상태를 기반으로 새로운 네트워크를 시작하시려 한다면, `--for-zero-height` 플래그를 이용하셔서 내보내기를 실행해주세요:

```bash
gaiad export --height [height] --for-zero-height > [filename].json
```

## 메인넷 검증하기

각 폴노드에서 invariant를 실행하여 검증 중 위험한 상황이 발생하는 것을 방지하세요. Invariant를 사용하여 메인넷의 상태(state)가 올바른 상태인 것을 확인합니다. 중요한 invariant 검증 중 하나는 프로토콜 예상 범위 밖에서 새로운 아톰이 생성되거나 사라지는 행위를 미리 감지하고 예빵합니다. 이 외에도 다양한 invariant check가 모듈 내 내장되어있습니다.

Invariant check는 블록체인 연산력을 상당하게 소모하기 때문에, 기본적으로 비활성화 되어있습니다. Invariant check를 실행한 상태로 노드를 시작하기 원하시는 경우 `assert-invariants-blockly` 플래그를 추가하세요:

```bash
gaiad start --assert-invariants-blockly
```

만약 노드 내 invariant가 문제를 감지하는 경우, 노드는 패닉하여 메인넷을 중지하는 트랜잭션을 전송합니다. 예시 메시지는 다음과 같습니다:

```bash
invariant broken:
    loose token invariance:
        pool.NotBondedTokens: 100
        sum of account tokens: 101
    CRITICAL please submit the following transaction:
        gaiad tx crisis invariant-broken staking supply

```

Invariant-broken 트랜잭션을 전송하는 경우 블록체인이 중지되기 떄문에 수수료가 없습니다.

## 검증인 노드로 업그레이드 하기

이제 풀노드 설정을 완료하셨습니다. 원하신다면 풀노드를 코스모스 검증인으로 업그레이드 하실 수 있습니다. 보팅 파워 상위 100위 검증인은 코스모스 허브의 새로운 블록 프로포즈 과정에 참여할 수 있습니다. [검증인 세팅하기](./validators/validator-setup.md)를 확인하세요.
