# 레저(Ledger) 나노 하드웨어 지갑 지원

암호화폐 자산을 하드웨어 지갑을 사용하여 보관하는 것은 보안을 상당히 향상합니다. 렛저 기기는 시드와 프라이빗 키를 보관하는 '영역' 역할을 하며, 트래잭션을 기기 내에서 서명합니다. 민감한 정보는 절대로 렛저 기기 밖으로 노출되지 않습니다. 이 문서는 코스모스 렛저 앱을 Gaia CLI 환경에서 사용하거나 [Lunie.io](https://lunie.io/#/) 웹 지갑에서 사용하는 방법을 설명합니다.

모든 렛저 기기의 핵심에는 프라이빗 키를 생성하는데 사용되는 네모닉 시드 단어가 있습니다. 이 시드 단어는 처음 렛저 기기를 활성화할때 생성됩니다(또는 직접 입력됩니다). 이 네모닉 시드는 코스모스와 호환되며 이를 기반으로 코스모스 계정을 생성하실 수 있쓰빈다.

::: danger
12 단어 시드키를 분실하거나 그 누구와도 공유하지 마세요. 자금 탈취와 손실을 예방하기 위해서는 다수의 시드키 사본을 만드시고 금고 같이 본인만이 알 수 있는 안전한 곳에 보관하는 것을 추천합니다. 누군가 시드키를 가지게 된 경우, 관련 프라이빗 키와 모든 계정의 소유권을 가지게 됩니다.
:::

## Gaia CLI + Ledger Nano

코스모스 허브 네트워크에서 새로운 계정을 생성하고 트랜잭션을 전송하는데 사용되는 도구는 `gaiad`입니다. 다음은 `gaiad`를 사용하는데 필요한 정보입니다. 만약 CLI 기반 도구가 익숙하지 않으신 경우, 하단에 있는 Lunie.io 지갑 사용법을 참고하세요.

### 시작하기 전에

- [렛저 기기에 코스모스 앱 설치하기](https://github.com/cosmos/ledger-cosmos/blob/master/README.md#installing)
- [Golang 설치하기](https://golang.org/doc/install)
- [Gaia 설치하기](https://cosmos.network/docs/cosmos-hub/installation.html)

다음 명령어를 입력하여 Gaiacli가 올바르게 설치된 것을 확인하세요:

```bash
gaiad version --long

➜ cosmos-sdk: 0.34.3
git commit: 67ab0b1e1d1e5b898c8cbdede35ad5196dba01b2
vendor hash: 0341b356ad7168074391ca7507f40b050e667722
build tags: netgo ledger
go version go1.11.5 darwin/amd64

```

### 렛저 키 추가하기

- 렛저 기기를 연결하시고 잠금해제 하세요
- 렛저 기기에서 코스모스 앱을 실행하세요
- 렛저키를 사용해 Gaiacli에서 새로운 계정을 생성하세요

::: tip
_키 명칭(keyName)_ 파라미터에 의미있는 값을 입력하세요. `ledger` 플래그는 `gaiad`가 렛저 기기의 시드를 사용해 계정을 생성할 것을 알립니다.
:::

```bash
gaiad keys add <keyName> --ledger

➜ NAME: TYPE: ADDRESS:     PUBKEY:
<키_명칭(keyName)> ledger cosmos1... cosmospub1...
```

코스모스는 [HD Wallets](./hd-wallets.md) 표준을 사용합니다. HD Wallet은 하나의 렛저 시드로부터 다수의 계정을 생성할 수 있게 합니다. 같은 렛저 기기에서 추가 계정을 생성하기 위해서는 다음 명령어를 실행하세요:

```bash
gaiad keys add <새로운_키_명칭(secondKeyName)> --ledger
```

### 주소 확인하기

다음 명령어를 실행하여 렛저 기기에서 주소를 확인하세요. 렛저 키 명칭을 `키_명칭` 값에 입력하여 해당 키의 주소를 확인하세요. `-d` 플래그는 렛저 `1.5.0` 버전 이상 기기에서만 지원됩니다.

```bash
gaiad keys show <keyName> -d
```

키를 새로 생성했을때 표기된 주소와 기기에서 표기된 주소가 일치하는지 확인하세요.

### 풀노드에 연결하기

이제 gaiacli를 코스모스 풀노드의 주소와 `chain-id`값을 설정해야 합니다. 이 예시에서는 코러스원 검증인이 운영하는 공개 노드를 사용해 `cosmoshub-2`에 연결하는 방법을 알아보겠습니다. 단, `gaiad`는 다른 풀노드에 연결하실 수 있다는 점을 참고하세요. Gaiacli에서 설정하는 `chain-id`와 풀노드의 `chain-id`은 동일해야합니다.

```bash
gaiad config node https://cosmos.chorus.one:26657
gaiad config chain_id cosmoshub-2
```

다음과 같은 명령어를 입력하여 연결 상태를 조회하세요:

``` bash
gaiad query staking validators
```

::: tip
자체 풀노드를 로컬 환경에서 운영하기 위해서는 다음 [글](https://cosmos.network/docs/cosmos-hub/join-mainnet.html#setting-up-a-new-node)을 참고하세요.
:::

### 트랜잭션 서명하기

이제 트랜잭션을 서명하고 전송할 수 있습니다. Gaiacli를 사용해 트랜잭션을 전송하기 위해서는 `tx send` 명령어를 사용하세요.

``` bash
gaiad tx send --help # to see all available options.
```

::: tip
다음 명령어를 실행하기 전 렛저 기기에 비밀번호를 입력하시고 코스모스 앱을 실행하세요
:::

렛저의 `키_명칭(keyName)`을 지정하여 Gaia와 코스모스 렛저 앱을 연결하고 트랜잭션을 서명하세요.

```bash
gaiad tx send <키_명칭(keyName)> <수신자_주소(destinationAddress)> <수량(amount)><단위(denomination)>
```

만약 `confirm transaction before signing`이 표기되는 경우, `Y`를 입력하여 진행하세요.

이후, 렛저 기기에서 트랜잭션 내용을 확인하고 승인합니다. 화면에 표기되는 트랜잭션 내용 JSON을 확인하세요. 각 값을 하나씩 확인하실 수 있습니다. 하단 내용을 확인하여 표준 트랜잭션 내용을 확인하세요.

이제 [네트워크에서 트랜잭션을 전송](./delegator-guide-cli.md#sending-transactions)할 준비가 되었습니다.

### 자산 받기

렛저 기기의 코스모스 계정으로 자산을 받기 위해서는 다음 명령어를 입력하여 주소를 확인하세요 (`TYPE ledger`로 표기되는 주소):

```bash
gaiad keys list

➜ NAME: TYPE: ADDRESS:     PUBKEY:
<키_명칭(keyName)> ledger cosmos1... cosmospub1...
```

### 추가 참고 문서

`gaiad`를 어떻게 사용해야되실지 모르시겠나요? 명령어 값을 비운 상태로 입력하여 각 명령어의 정보를 확인하실 수 있습니다.

::: tip
`gaiad` 명령어는 중첩된 형태로 존재합니다. `$ gaiad` 명령어는 최상위 명령어만을 표기합니다(status, config, query, tx). 하위 명령어에 대한 정보는 해당 명령어에 help 명령어를 추가하여 확인할 수 있습니다.

예를 들어 `query` 명령어에 대한 정보를 확인하기 위해서는:

```bash
gaiad query --help
```

또는 `tx`(트랜잭션) 명령어를 확인하기 위해서는:

```bash
gaiad tx --help
```

를 입력하세요.

# Lunie.io

Lunie 웹 지갑은 렛저 나노 S 기기를 사용해 서명하는 것을 지원합니다. 다음은 (Lunie.io)[https://lunie.io] 지갑을 Ledger 기기로 사용하는 방법을 정리합니다.

### 기기 연결하기

- 렛저 기기를 컴퓨터에 연결하시고, 비밀번호를 입력하여 잠금해제한 후 코스모스 앱을 실행하세요
- 웹 브라우저에서 [https://lunie.io](https://lunie.io)를 가세요
- "Sign In"을 클릭하세요
- "Sign in with Ledger Nano S"를 선택하세요

### 주소 확인하기

다음 명령어를 실행하여 렛저 기기에서 주소를 확인하세요. 렛저 키 명칭을 `키_명칭` 값에 입력하여 해당 키의 주소를 확인하세요. `-d` 플래그는 렛저 `1.5.0` 버전 이상 기기에서만 지원됩니다.

```bash
gaiad keys show <keyName> -d
```

렛저에 표기되는 주소와 Lunie.io에 표기되는 주소가 동일한지 먼저 확인하시고 다음 단계를 진행하세요. 확인이 된 경우, Lunie에서 렛저 키를 사용해 트랜잭션을 서명할 수 있습니다.

Lunie에 대해 더 알고싶으시면 이 [투토리얼](https://medium.com/easy2stake/how-to-delegate-re-delegate-un-delegate-cosmos-atoms-with-the-lunie-web-wallet-eb72369e52db)을 참고하셔서 아톰 위임과 Lunie 웹 지갑에 대해서 더 알아보세요.

# 코스모스 표준 트랜잭션

코스모스의 트랜잭션은 코스모스 SDK의[표준 트랜잭션 타입(Standard Transaction Type)](https://godoc.org/github.com/cosmos/cosmos-sdk/x/auth#StdTx)를 응용합니다. 렛저 기기는 이 오브젝트의 시리얼화된 JSON을 기기에서 표기하며, 트랜잭션 서명 전 검토하실 수 있습니다. 각 필드에 대한 설명은 다음과 같습니다:

- `chain-id`: 트랜잭션을 전송할 체인, (예, `gaia-13003` 테스트넷 또는 `cosmoshub-2` 메인넷)
- `account_number`: 계정에 최초로 자금을 입금할때 생성되는 계정의 고유 글로벌 ID
- `sequence`: 계정의 논스 값. 이후 발생하는 트랜잭션 마다 이 값은 증가합니다.
- `fee`: 트랜잭션 수수료, 가스 수량, 그리고 수수료의 단위를 설명하는 JSON 오브젝트
- `memo`: (선택 값) 트랜잭션 태깅 등의 용도로 사용되는 텍스트 값
- `msgs_<index>/<field>`: 트랜잭션에 포함된 메시지 어레이. 더블클릭하여 하위 JSON 값을 확인하실 수 있습니다.

# 지원

추가적인 지원이 필요하신 경우, 코스모스 포럼의 [과거 글](https://forum.cosmos.network/search?q=ledger)을 참고하세요.

[텔레그램 채널](https://t.me/cosmosproject)을 통해 문의하시거나 다음 커뮤니티 가이드를 참고하세요:

- [Ztake](https://medium.com/@miranugumanova) - [How to Redelegate Cosmos Atoms with the Lunie Web Wallet](https://medium.com/@miranugumanova/how-to-re-delegate-cosmos-atoms-with-lunie-web-wallet-8303752832c5)
- [Cryptium Labs](https://medium.com/cryptium-cosmos) - [How to store your ATOMS on your Ledger and delegate with the command line](https://medium.com/cryptium-cosmos/how-to-store-your-cosmos-atoms-on-your-ledger-and-delegate-with-the-command-line-929eb29705f)
