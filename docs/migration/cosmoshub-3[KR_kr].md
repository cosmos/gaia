# 코스모스 허브 3 업그레이드 매뉴얼

이 문서는 풀노드 운영자가 `cosmoshub-3`에서 `cosmoshub-4` 업그레이드를 진행하기 위한 과정을 설명합니다.
텐더민트 팀이 공식적인 업데이트된 제네시스 파일을 업로드할 예정이나, 각 검증인이 해당 제네시스 파일을 각자 검증할 것을 권장합니다.

현재 `cosmoshub-3`에서는 `Cosmos Hub 4 Upgrade Proposal`에 대한 사회적 합의가 도달된 것으로 판단됩니다.
프로포절 #[27](https://www.mintscan.io/cosmos/proposals/27), #[35](https://www.mintscan.io/cosmos/proposals/35) and #[36](https://www.mintscan.io/cosmos/proposals/36)의 내용에 따라 업그레이드 과정은 `2021년 2월 18일, 06:00 UTC`에 진행될 예정입니다.

  - [마이그레이션](#마이그레이션)
  - [사전 정보](#사전-정보)
  - [주요 업데이트](#주요-업데이트)
  - [위험 고지](#위험-고지)
  - [복구](#복구)
  - [업그레이드 절차](#업그레이드-절차)
  - [서비스 제공자에 대한 공지](#서비스-제공자에-대한-공지)

# 마이그레이션

다음 항목들은 애플리케이션 및 모듈을 코스모스 v0.41 스타게이트로 마이그레이션 하는 절차를 정보를 포함하고 있습니다.

만약 코스모스 허브 또는 코스모스 생태계 블록체인의 블록 익스플로러, 지갑, 거래소, 검증인, 등 서비스 (예, 커스터디 제공자)를 운영하시는 경우, 이번 업그레이드에서 상당한 변경사항이 있음으로 꼭 다음 정보를 참고하십시오.

1. [앱 및 모듈 마이그레이션](https://github.com/cosmos/cosmos-sdk/blob/master/docs/migrations/app_and_modules.md)
1. [코스모스 v0.40 체인 업그레이드 가이드](https://github.com/cosmos/cosmos-sdk/blob/master/docs/migrations/chain-upgrade-guide-040.md)
1. [REST 엔드포인트 마이그레이션](https://github.com/cosmos/cosmos-sdk/blob/master/docs/migrations/rest.md)
1. [각 버전 체인지로그의 breaking change 모음](breaking_changes.md)
1. [Inter-Blockchain Communication (IBC)– 체인간 트랜잭션](https://figment.network/resources/cosmos-stargate-upgrade-overview/#ibc)
1. [Protobuf 마이그레이션 – 블록체인 성능 및 개발 과정 개선](https://figment.network/resources/cosmos-stargate-upgrade-overview/#proto)
1. [State Sync – 몇 분 내에 완료되는 노드 동기화](https://figment.network/resources/cosmos-stargate-upgrade-overview/#sync)
1. [강력한 기능을 포함한 라이트 클라이언트](https://figment.network/resources/cosmos-stargate-upgrade-overview/#light)
1. [체인 업그레이드 모듈 – 업그레이드 자동화](https://figment.network/resources/cosmos-stargate-upgrade-overview/#upgrade)

만약 2월 18일 전에 업그레이드 과정을 미리 테스트 진행을 희망하시는 경우 [이 글](https://github.com/cosmos/gaia/issues/569#issuecomment-767910963)을 참고하세요

## 사전 정보

지난 코스모스 허브 업그레이드(`cosmoshub-3`) 이후 코스모스 SDK와 Gaia 애플리케이션에 상당한 양의 변경사항이 적용되었습니다.
변경사항에는 신규 기능, 프로토콜 변경사항, 애플리케이션 구조 변경 등이 포함되었으며, 애플리케이션 개발 과정의 개선이 기대됩니다.

우선, [인터체인 표준](https://github.com/cosmos/ics#ibc-quick-references)를 따른 [IBC](https://docs.cosmos.network/master/ibc/overview.html)이 활성화될 예정입니다. 또한 효율성, 노드 동기화, 추후 블록체인 업데이트 과정이 개선됩니다. 자세한 내용은 [스타게이트 웹사이트](https://stargate.cosmos.network/)를 참고하세요.

__이번 업그레이드에서 풀 노드 운영자 업그레이드를 진행하는 것은 [Gaia](https://github.com/cosmos/gaia) 애플리케이션 v4.0.0입니다. 이번 버전의 Gaia 애플리케이션은 코스모스 SDK v0.41.0 그리고 텐더민트 v0.34.3 기반으로 빌드되었습니다.

## 주요 업데이트

이번 SDK의 릴리즈에서는 다수의 기능 및 변경사항이 적용되어 있습니다. 이에 대한 설명은 [여기](https://github.com/cosmos/stargate)에서 확인하실 수 있습니다.

개발자 또는 클라이언트로서 참고해야할 점은 다음과 같습니다:

- **프로토콜 버퍼(Protocol BufferS)**: 이전 버전의 코스모스 SDK에서는 인코딩 및 디코딩 과정에서 아미노 코덱을 사용했습니다.
이번 코스모스 SDK 버전에서는 프로토콜 버퍼가 내장되어있습니다. 프로토콜 버퍼를 통해 애플리케이션은 속도, 가독성, 편의성, 프로그래밍 언어 간 상호호환성 등의 부분에서 상당한 개선이 있을 것으로 기대됩니다. [더 읽기](https://github.com/cosmos/cosmos-sdk/blob/master/docs/migrations/app_and_modules.md#protocol-buffers)
- **CLI**: 이전 버전의 코스모스 SDK에서는 블록체인의 CLI와 데몬은 별도의 바이너리로 구성되었으며, 실행하는 블록체인 인터랙션에 따라 `gaiad`와 `gaiacli` 바이너리가 구분되었습니다. 이번 버전의 코스모스 SDK에서는 두 바이너리가 하나의 `gaiad` 바이너리로 통합되었으며 해당 바이너리 내에서 기존에 `gaiacli`에서 사용했던 명령어를 지원합니다.
- **노드 구성**: 이전 버전의 코스모스 SDK에서는 블록체인 데이터와 노드 설정이 `~/.gaia/`에 저장되었지만, 이번 버전에서는 해당 정보다 `~/.gaia/` 디렉토리에 보관됩니다. 만약 블록체인 데이터 또는 노드 설정을 관리하는 스크립트를 사용하시는 경우 해당 스크립트에서 패스를 변경해야합니다.


## 위험 고지

검증인이 컨센서스 노드 업그레이드를 진행하는 절차에서 이중서명에 따른 슬래싱의 위험이 존재합니다. 이 과정에서 가장 중요한 것은 검증인을 가동하고 서명을 시작하기 전 소프트웨어 버전을 확인하고 제네시스 파일의 해시를 확인하시기를 바랍니다.

블록체인 검증인이 할 수 있는 가장 위험한 행동은 네트워크 시작 과정에서 존재했던 실수를 인지하고 업그레이드 과정을 처음부터 다시 시작하는 것입니다. 만약 업그레이드 과정에서 실수가 발생했다면 네트워크가 시작되는 것을 기다린 후에 실수를 고치는 것을 권장합니다. 만약 네트워크가 중단되었고 본인의 검증인을 실제 시작된 네트워크가 아닌 다른 제네시스 파일로 가동한 경우, 검증인을 리셋하는 과정에 대해 텐더민트 검증인으로 부터 조언을 구할 것을 권장합니다.



## 복구

각 검증인은 `cosmoshub-3` 상태를 내보내기(export) 전에 내보내는 블록 하이트의 풀 데이터 스냅샷을 진행할 것을 권장합니다. 스냅샷 과정은 각 검증인의 인프라에 따라 다를 수 있지만, 통상 `.gaia` 디렉토리를 백업하는 것으로 진행합니다.

`gaiad` 프로세스를 멈춘 후 `.gaia/data/priv_validator_state.json` 파일을 백업하는 것은 매우 중요합니다. 이 파일은 검증인이 컨센서스 라운드에 참여할 때마다 업데이트됩니다. 만약 업그레이드 과정이 실패하여 이전 체인을 다시 시작해야되는 경우 검증인의 이중서명을 방지하기 위해서 이 파일은 필수입니다.

만약 업그레이드 과정이 실패하는 경우, 검증인과 노드 운영자는 gaia v2.0.15(코스모스 SDK v0.37.15 기반)으로 다운그레이드를 진행하고 가장 최근 진행했던 스냅샷을 복구한 이후에 노드를 시작해야합니다.


## 업그레이드 

__참고__: 이 가이드는 코스모스 SDK의 v0.37.15 기반의 gaia v2.0.15를 운영한다는 가정에 작성된 가이드입니다.

Gaia v2.0.15의 버전/커밋 해시값: `89cf7e6fc166eaabf47ad2755c443d455feda02e`

1. 올바른 _gaiad_ 버전 (v2.0.15)를 운영하고 있는 것을 확인하세요:

   ```bash
    $ gaiad version --long
    name: gaia
    server_name: gaiad
    client_name: gaiacli
    version: 2.0.15
    commit: 89cf7e6fc166eaabf47ad2755c443d455feda02e
    build_tags: netgo,ledger
    go: go version go1.15 darwin/amd64
   ```

1. 체인이 올바른 날짜와 시간에 멈추도록 설정하세요:

    2021년 2월 18일 06:00 UTC의 UNIX seconds 시간 값: `1613628000`

    ```bash
    perl -i -pe 's/^halt-time =.*/halt-time = 1613628000/' ~/.gaia/config/app.toml
    ```

 1. 체인이 멈춘 후 `.gaia` 디렉토리를 백업하세요
 

    ```bash
    mv ~/.gaia ./gaiad_backup
    ```

    **참고**: 업그레이드 과정이 예상 외로 실패하거나 합의된 시간 내에 새로운 체인에 충분한 보팅 파워가 참여하지 않는 경우를 대비해 검증인과 노드 운영자는 export height의 풀 데이터 스냅샷을 진행하는 것을 권장합니다. 이 경우에는 체인 업그레이드 과정은 보류되고 `cosmoshub-3`의 운영이 재개됩니다. 해당 과정에 대해서는 [복구](#복구) 항목을 참고하세요.
    
1. 기존 `cosmoshub-3`의 상태를 내보내기:

   다음 명령어를 사용하여 상태를 내보내기 전 `gaiad` 바이너리가 꼭 멈춰있어야 합니다!
   검증인으로서 가장 최근 생성된 블록은 `~/.gaia/config/data/priv_validator_state.json`에서 확인하실 수 있습니다 (또는 이전 과정에서 백업을 진행한 경우 `gaiad_backup`). 블록 높이는 다음과 같이 확인하세요
   
   ```bash
   cat ~/.gaia/config/data/priv_validator_state.json | jq '.height'
   ```

   ```bash
   $ gaiad export --for-zero-height --height=<height> > cosmoshub_3_genesis_export.json
   ```
   _이 과정은 상당한 시간이 (약 1시간) 소요될 수 있습니다_

1. 내보낸 제네시스 파일의 SHA256 값을 검증하세요:

    본인의 제네시스 파일의 값을 네트워크 내 다른 검증인 / 풀 노드 운영자와 비교하세요.
    이 후 과정에서는 모든 인원이 동일한 제네시스 파일을 생성하는 것이 상당이 중요합니다.

   ```bash
   $ jq -S -c -M '' cosmoshub_3_genesis_export.json | shasum -a 256
   [SHA256_VALUE]  cosmoshub_3_genesis_export.json
   ```

1. 이 단계 까지 오셨다면 올바른 제네시스 상태를 내보내셨습니다! 이후 과정부터는 [Gaia](https://github.com/cosmos/gaia) v4.0.0을 필요로 합니다. 그룹 채팅 방의 다른 검증인들/피어와 새로운 제네시스 파일의 해시를 비교/검증하세요.

   **참고**: Go [1.15+](https://golang.org/dl/) 버전이 설치되어야 합니다!

   ```bash
   $ git clone https://github.com/cosmos/gaia.git && cd gaia && git checkout v4.0.0; make install
   ```

1. _Gaia_의 올바른 버전(v4.0.0)을 운영하고 있는 것을 확인하세요:

   ```bash
    $ gaiad version --long
    name: gaia
    server_name: gaiad
    version: 4.0.0
    commit: 2bb04266266586468271c4ab322367acbf41188f
    build_tags: netgo,ledger
    go: go version go1.15 darwin/amd64
    build_deps:
    ...
   ```
    Gaia v4.0.0 버전/커밋 해시 : `2bb04266266586468271c4ab322367acbf41188f`

1. 내보낸 상태를 기존 v2.0.15 버전에서 v4.0.0 버전으로 마이그레이션 하세요:

   ```bash
   $ gaiad migrate cosmoshub_3_genesis_export.json --chain-id=cosmoshub-4 --initial-height [last_cosmoshub-3_block+1] > genesis.json
   ```

   이 과정은 이전 체인에서 내보낸 상태를 기반으로 `cosmoshub-4`로 시작하기 위한 `genesis.json` 파일을 생성합니다.

1. 최종 제네시스 JSON의 SHA256 해시 값을 검증하세요:

   ```bash
   $ jq -S -c -M '' genesis.json | shasum -a 256
   [SHA256_VALUE]  genesis.json
   ```

    해당 값을 네트워크의 다른 검증인 / 풀 노드 운영자와 비교하세요.
    과정에서 모든 참여자가 같은 genesis.json 파일을 생성하는 것이 중요합니다.

1. 상태 리셋:

   **참고**: 이 과정을 진행하기 전에 꼭 노드의 상태를 백업하세요. 백업 과정은 [복구](#복구) 항복을 참고하세요

   ```bash
   $ gaiad unsafe-reset-all
   ```

1. 새로운 `genesis.json`을 `.gaia/config/` 디렉토리로 옮기세요:

    ```bash
    cp genesis.json ~/.gaia/config/
    ```

1. 블록체인을 가동하세요

    ```bash
    gaiad start
    ```

    Crisis 모듈의 제네시스 상태 자동 검증은 30-120분 소요될 수 있습니다. 해당 기능은 `gaiad start --x-crisis-skip-assert-invariants`로 비활성화할 수 있습니다.

## 서비스 제공자를 위한 정보

# REST 서버

만약 이전까지 `gaiacli rest-server` 명령어로 REST 서버를 구동하신 경우, 해당 명령어는 이번 버전부터 비활성화 됩니다. API 서버는 데몬 내에서 활성화되며, `.gaia/config/app.toml` 설정 내에서 활성화됩니다:

```
[api]
# Enable defines if the API server should be enabled.
enable = false
# Swagger defines if swagger documentation should automatically be registered.
swagger = false
```

`swagger` 설정은 Swagger 문서 API를 활성화/비활성화 하는 여부를 관리합니다 (예, /swagger/ API 엔드포인트)

# gRPC 설정

gRPC 설정은 `.gaia/config/app.toml`에 있습니다.

```yaml
[grpc]
# Enable defines if the gRPC server should be enabled.
enable = true
# Address defines the gRPC server address to bind to.
address = "0.0.0.0:9090"
```

# 스테이트 싱크

스테이트 싱크 설정은 `.gaia/config/app.toml`에 있습니다.

```yaml
# State sync snapshots allow other nodes to rapidly join the network without replaying historical
# blocks, instead downloading and applying a snapshot of the application state at a given height.
[state-sync]
# snapshot-interval specifies the block interval at which local state sync snapshots are
# taken (0 to disable). Must be a multiple of pruning-keep-every.
snapshot-interval = 0
# snapshot-keep-recent specifies the number of recent snapshots to keep and serve (0 to keep all).
snapshot-keep-recent = 2
```
