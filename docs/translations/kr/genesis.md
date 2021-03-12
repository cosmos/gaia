# 제네시스 파일

이 문서는 코스모스 허브 메인넷의 제네시스 파일의 구조를 설명합니다. 또한, 자체 gaia 테스트넷을 운영하기 위해 자체적으로 제네시스 파일을 작성하는 방법을 설명합니다.

참고로 기본 설정이 적용된 제네시스 파일을 사용해 테스트넷을 운영하기 위해서는 다음 명령어를 입력할 수 있습니다:

```bash
gaiad init <명칭(moniker)> --chain-id <체인_아이디(chain-id)>
```

제네시스 파일은 `~/.gaia/config/genesis.toml`에 저장됩니다.

## 제네시스 파일은 무엇인가?

제네시스 파일은 블록체인 초기 상태(state)를 정의하는 JSON 파일입니다. 이는 실질적으로 블록 높이 `0`을 뜻하며, 첫 블록이 생성되는 `1` 블록은 제네시스 파일을 패런트(parent)로 참조합니다.

제네시스 파일에 정의된 상태는 토큰 분배, 제네시스 시간, 기본 파라미터 값 등의 모든 필수 정보를 포함하고 있습니다. 각 정보를 항목별로 정리합니다.

## 제네시스 시간과 체인 아이디

`genesis_time`은 제네시스 파일 상단에 정의됩니다. 제네시스 타임은 블록체인이 시작되는 `UTC` 기준 시간을 정의합니다. 해당 시간에 도달하면 제네시스 검증인은 온라인어 컨센서스 과정에 참여를 시작합니다. 블록체인은 제네시스 검증인의 2/3 이상이 (보팅 파워 기준으로) 온라인될 경우에 시작됩니다.

```json
"genesis_time": "2019-03-13T17:00:00.000000000Z",
```

`chain_id`는 블록체인의 고유 식별 정보입니다. 동일한 소프트웨어를 운영하는 다양한 체인을 구별하기 위해 사용됩니다.

```json
"chain_id": "cosmoshub-2",
```

## 컨센서스 파라미터

이후 제네시스 파일은 컨센서스 파라미터 값을 정의합니다. 컨센서스 파라미터는 모든 합의 계층(`gaia`의 경우 `Tendermint` 계층) 관련 값을 리그룹(regroup)합니다. 파라미터 값에 대해 알아보겠습니다:

- `block`
    + `max_bytes`: 블록 최대 바이트 크기
    + `max_gas`: 블록 가스 한도(gas limit). 블록에 포함되는 모든 트랜잭션은 가스를 소모합니다. 블록에 포함되어있는 트랜잭션들의 총 가스 사용량은 이 한도를 초과할 수 없습니다.
- `evidence`
    + `max_age`: 증거(evidence)는 검증인이 동일한 블록 높이와 합의 라운드(round)에서 두개의 블록을 동시했다는 증거입니다. 위와 같은 행동은 명백한 악의적 행동으로 간주되며 스테이트 머신 계층에서 페널티를 부과합니다. `max_age` 값은 증거 유효성이 유지되는 최대 _블록_ 개수를 의미합니다.
- `validator`
    + `pub_key_types`: 검증인이 사용할 수 있는 pubkey 형태입니다(`ed25519`, `secp256k1`, ...). 현재 `ed25519`만 지원됩니다.

```json
"consensus_params": {
    "block_size": {
      "max_bytes": "150000",
      "max_gas": "1500000"
    },
    "evidence": {
      "max_age": "1000000"
    },
    "validator": {
      "pub_key_types": [
        "ed25519"
      ]
    }
  },
```

## 애플리케이션 상태

애플리케이션 상태(application state)는 스테이트 머신(상태 기계, state machine)의 초기 상태를 정의합니다.

### 제네시스 계정

이 항목에서는 초기 토큰 분배가 정의됩니다. 수동으로 제네시스 파일에 계정을 추가할 수 있지만, 다음 명령어를 통해 계정을 추가할 수도 있습니다:

```bash
// 예시: gaiad add-genesis-account cosmos1qs8tnw2t8l6amtzvdemnnsq9dzk0ag0z37gh3h 10000000uatom

gaiad add-genesis-account <계정_주소(account_address)> <수량(amount)> <단위(denom)>
```

위 명령어는 `app_state` 항목 내 `accounts` 리스트에 아이템을 추가합니다.

```json
"accounts": [
      {
        "address": "cosmos1qs8tnw2t8l6amtzvdemnnsq9dzk0ag0z37gh3h",
        "coins": [
          {
            "denom": "uatom",
            "amount": "10000000"
          }
        ],
        "sequence_number": "0",
        "account_number": "0",
        "original_vesting": [
          {
            "denom": "uatom",
            "amount": "26306000000"
          }
        ],
        "delegated_free": null,
        "delegated_vesting": null,
        "start_time": "0",
        "end_time": "10000"
      }
]
```

각 파라미터 값을 항목별로 설명하겠습니다:

- `sequence_number`: 이 숫자는 계정이 전송한 트랜잭션 수를 추적하는데 사용됩니다. 트랜잭션이 블록에 포함될 때마다 숫자가 증가하며 리플레이 공격(replay attack)을 방지하기 위해 사용됩니다. 기본 값은 `0` 입니다.
- `account_number`: 계정 고유 식별번호입니다. 해당 계정의 첫 트랜잭션이 블록에 포함될때 생성됩니다.
- `original_vesting`: `gaia`는 베스팅(락업) 기능을 지원합니다. 락업 계정이 소유한 토큰 수량을 지정하고, 토큰 전송이 가능할때까지의 시간을 정할 수 있습니다. 락업된 토큰의 스테이킹은 지원됩니다. 기본 값은 `null`입니다.
- `delegated_free`: 락업이 풀린 후 전송될 수 있는 위임된 수량을 뜻합니다. 대부분 제네시스 상황에서는 `null`이 표준 값입니다.
- `delegated_vesting`: 아직 락업이 진행중인 위임된 수량을 뜻합니다. 대분분 제네시스 상황에서는 `null`이 표준 값입니다.
- `start_time`: 락업이 풀리는 시점의 블록 높이입니다. 대부분 제네시스 상황에서는 `0`이 표준 값입니다.
- `end_time`: 락업 기간이 마감되는 시점의 블록 높이입니다. 락업이 없는 계정의 표준 값은 `0`입니다.

### 뱅크(Bank)

`bank` 모듈은 토큰을 관리합니다. 이 항목에서 정의될 파라미터는 제네시스 시작시 전송 가능여부를 정의하는 것입니다.

```json
"bank": {
      "send_enabled": false
    }
```

### 스테이킹(Staking)

`staking` 모듈은 스테이트 머신의 지분증명(proof-of-stake) 로직의 대다수를 처리합니다. 이 항목은 다음과 같습니다:

```json
"staking": {
      "pool": {
        "not_bonded_tokens": "10000000",
        "bonded_tokens": "0"
      },
      "params": {
        "unbonding_time": "1814400000000000",
        "max_validators": 100,
        "max_entries": 7,
        "bond_denom": "uatom"
      },
      "last_total_power": "0",
      "last_validator_powers": null,
      "validators": null,
      "bonds": null,
      "unbonding_delegations": null,
      "redelegations": null,
      "exported": false
    }
```

각 파라미터 값에 대해 알아보겠습니다:

- `pool`
    + `not_bonded_tokens`: 제네시스 시점에서 위임되지 않은 토큰의 수량을 정의합니다. 대부분의 상황에서 이 값은 스테이킹 토큰의 총 발행량을 뜻합니다 (이 예시에서는 `uatom` 단위로 정의됩니다).
    + `bonded_tokens`: 제네시스 시점에서 위임된 토큰의 수량입니다. 대부분 상황에서 이 값은 `0`입니다.
- `params`
    + `unbonding_time`: 토큰의 언본딩이 완료될 때까지의 기간을 _나노초(nanosecond)_ 단위로 정의합니다.
    + `max_validators`: 최대 검증인 수입니다.
    + `max_entries`: 특정 검증인/위임자 쌍에서 동시에 진행될 수 있는 최대 언본딩/재위임 회수.
    + `bond_denom`: 스테이킹 토큰 단위.
- `last_total_power`: 보팅 파워 수치. 통상 제네시스 시점에서 `0`입니다 (다만, 과거 블록체인 상태를 기반으로 제네시스가 생성되었을 경우 다를 수 있습니다).
- `last_validator_powers`: 각 검증인의 가장 최근 보팅 파워 수치입니다. 통상 제네시스 시점에서 `null`입니다. (다만, 과거 블록체인 상태를 기반으로 제네시스가 생성되었을 경우 다를 수 있습니다).
- `validators`: 가장 최근 검증인 목록. 통상 제네시스 시점에서 `null`입니다. (다만, 과거 블록체인 상태를 기반으로 제네시스가 생성되었을 경우 다를 수 있습니다).
- `bonds`: 가장 최근 위임 목록입니다. 통상 제네시스 시점에서 `null`입니다. (다만, 과거 블록체인 상태를 기반으로 제네시스가 생성되었을 경우 다를 수 있습니다).
- `unbonding_delegations`: 가장 최근 위임 취소 목록입니다. 통상 제네시스 시점에서 `null`입니다. (다만, 과거 블록체인 상태를 기반으로 제네시스가 생성되었을 경우 다를 수 있습니다).
- `redelegations`: 가장 최근 재위임 목록입니다. 통상 제네시스 시점에서 `null`입니다. (다만, 과거 블록체인 상태를 기반으로 제네시스가 생성되었을 경우 다를 수 있습니다).
- `exported`: 제네시스 파일이 과거 상태를 기반을 내보내어 작성된 여부.

### 민트(mint)

`mint` 모듈은 토큰 발행량의 인플레이션 관련 로직을 처리합니다. 제네시스 파일의 `mint` 항목은 다음과 같습니다:

```json
"mint": {
      "minter": {
        "inflation": "0.070000000000000000",
        "annual_provisions": "0.000000000000000000"
      },
      "params": {
        "mint_denom": "uatom",
        "inflation_rate_change": "0.130000000000000000",
        "inflation_max": "0.200000000000000000",
        "inflation_min": "0.070000000000000000",
        "goal_bonded": "0.670000000000000000",
        "blocks_per_year": "6311520"
      }
    }
```

각 파라미터 값에 대해 알아보겠습니다:

- `minter`
    + `inflation`: 토큰 총 발행량의 기본 연간 인플레이션 % (주 단위 복리 기준). `0.070000000000000000` 값은 주 단위 복리 기준으로 연간 `7%` 인플레이션을 뜻합니다.
    + `annual_provisions`: 매 블록마다 계산됨. 기본 값은 `0.000000000000000000`으로 시작합니다.
- `params`
    + `mint_denom`: 인플레이션 대상 스테이킹 토큰의 단위.
    + `inflation_rate_change`: 연간 인플레이션 변화율.
    + `inflation_max`: 인플레이션 최대 수치.
    + `inflation_min`: 인플레이션 최소 수치.
    + `goal_bonded`: 총 발행량 중 위임 목표 % 수치. 만약 현재 위임 비율이 해당 이 값보다 낮은 경우, 인플레이션은 `inflation_rate_change` 값을 따라 `inflation_max`까지 증가합니다. 반대로 현재 위임 비율이 이 수치보다 높을 경우 `inflation_rate_change` 값을 따라 `inflation_min`까지 감소합니다.
    + `blocks_per_year`: 연간 생성되는 블록 예상 수치. 스테이킹 토큰 인플레이션으로 발생하는 토큰을 각 블록당 지급(블록 프로비젼, block provisions)하는데 계산하는 용도로 사용됩니다.

### 분배(distribution)

`distribution` 모듈은 블록당 위임 보상(block provision)을 검증인과 위임자에게 분배하는 로직을 처리합니다. 제네시스 파일의 `distribution` 항목은 다음과 같습니다:

```json
    "distribution": {
      "fee_pool": {
        "community_pool": null
      },
      "community_tax": "0.020000000000000000",
      "base_proposer_reward": "0.010000000000000000",
      "bonus_proposer_reward": "0.040000000000000000",
      "withdraw_addr_enabled": false,
      "delegator_withdraw_infos": null,
      "previous_proposer": "",
      "outstanding_rewards": null,
      "validator_accumulated_commissions": null,
      "validator_historical_rewards": null,
      "validator_current_rewards": null,
      "delegator_starting_infos": null,
      "validator_slash_events": null
    }
```

각 파라미터 값에 대해 알아보겠습니다:

- `fee_pool`
    + `community_pool`: 커뮤니티 풀은 임무 수행(개발, 커뮤니티 빌딩, 등)의 보상으로 제공될 수 있는 토큰 자금입니다. 토큰 풀의 사용은 거버넌스 프로포절을 통해 결정됩니다. 통상 제네시스 시점에서 `null`입니다.
- `community_tax`: 블록 보상과 수수료 중 커뮤니티 풀로 예치될 '세금' %.
- `base_proposer_reward`: 유효한 블록의 트랜잭션 수수료 중 블록 프로포저에게 지급될 리워드. 값이 `0.010000000000000000`인 경우, 수수료의 1%가 블록 프로포저에게 지급됩니다.
- `bonus_proposer_reward`: 유효한 블록의 트랜잭션 수수료 중 블록 프로포저에게 지급될 리워드의 _최대 한도_. 보너스 수량은 블록 프로포저가 포함한 `precommit` 수량에 비례합니다. 만약 프로포저가 보팅 파워 기준으로`precommit`의 2/3을 포함한 경우 (2/3는 유효한 블록을 생성하기 위한 최소 한도입니다), `base_proposer_reward` 만큼의 보너스를 지급 받습니다. 보너스는 블록 프로포저가 `precommit`의 100%를 포함하는 경우 최대 `bonus_proposer_reward`까지 선의적(linearly)으로 증가합니다.
- `withdraw_addr_enabled`: 파라미터 값이 `true`인 경우, 위임자는 위임 보상을 받을 별도의 주소를 지정할 수 있습니다. 제네시스에서 토큰 전송 기능을 비활성화하기 원하시는 경우, 토큰 전송 잠금 기능을 우회할 수 있으니 `false`로 설정하세요.
- `delegator_withdraw_infos`: 위임자들의 보상 출금 주소 목록입니다. 과거 상태에서 제네시스가 생성되지 않은 경우 `null` 값이 기본 값입니다.
- `previous_proposer`: 지난 블록의 프로포저입니다. 과거 상태에서 제네시스가 생성되지 않은 경우 `""` 값으로 입력하세요
- `outstanding_rewards`: 보상 출금이 진행되지 않은 리워드입니다. 과거 상태에서 제네시스가 생성되지 않은 경우 `null` 값이 기본 설정입니다.
- `validator_accumulated_commission`: 검증인 커미션 중 출금되지 않은 커미션입니다. 과거 상태에서 제네시스가 생성되지 않은 경우 `null` 값이 기본 설정입니다.
- `validator_historical_rewards`: `distribution` 모듈 연산 용도로 사용되는 검증인 과거 리워드 정보입니다. 과거 상태에서 제네시스가 생성되지 않은 경우 `null` 값이 기본 설정입니다.
- `validators_current_rewards`: `distribution` 모듈 연산 용도로 사용되는 검증인 현재 리워드 정보입니다. 과거 상태에서 제네시스가 생성되지 않은 경우 `null` 값이 기본 설정입니다.
- `delegator_starting_infos`: 검증인 검증 기간, 위임자 스테이킹 토큰 수량, creation height (슬래싱이 발생한 경우 확인용) 정보입니다. `distribution` 모듈 연산 용도로 사용되는 검증인 과거 리워드 정보입니다. 과거 상태에서 제네시스가 생성되지 않은 경우 `null` 값이 기본 설정입니다.
- `validator_slash_events`: 과거 검증인의 슬래싱 정보입니다. `distribution` 모듈 연산 용도로 사용되는 검증인 과거 리워드 정보입니다. 과거 상태에서 제네시스가 생성되지 않은 경우 `null` 값이 기본 설정입니다.

### 거버넌스(governance)

`gov` 모듈은 모든 거버넌스 관련 트랜잭션을 처리합니다. 제네시스 파일의 기본 `gov` 항목은 다음과 같습니다:

```json
"gov": {
      "starting_proposal_id": "1",
      "deposits": null,
      "votes": null,
      "proposals": null,
      "deposit_params": {
        "min_deposit": [
          {
            "denom": "uatom",
            "amount": "512000000"
          }
        ],
        "max_deposit_period": "1209600000000000"
      },
      "voting_params": {
        "voting_period": "1209600000000000"
      },
      "tally_params": {
        "quorum": "0.4",
        "threshold": "0.5",
        "veto": "0.334",
        "governance_penalty": "0.0"
      }
    }
```

각 파라미터 값에 대해 알아보겠습니다:

- `starting_proposal_id`: 이 파라미터는 첫 프로포절의 ID를 정의합니다. 각 프로포절은 고유한 ID를 보유합니다.
- `deposits`: 각 프로포절 ID에 대한 보증금 목록입니다. 과거 상태에서 제네시스가 생성되지 않은 경우 `null` 값이 기본 설정입니다.
- `votes`: 각 프로포절 ID에 대한 투표 목록입니다. 과거 상태에서 제네시스가 생성되지 않은 경우 `null` 값이 기본 설정입니다.
- `votes`: 각 프로포절 ID에 대한 투표 목록입니다. 과거 상태에서 제네시스가 생성되지 않은 경우 `null` 값이 기본 설정입니다.
- `proposals`: 각 프로포절 ID에 대한 프로포절 목록입니다. 
- `deposit_params`
    + `min_deposit`: 프로포절의 `voting period` 단계를 시작하기 위해 필요한 최소 보증금 수량입니다. 만약 다수 단위를 적용할 경우 `OR` 연산자를 사용하세요.
    + `max_deposit_period`: 프로포절 보증금 추가가 가능한 기간 (**나노초(nanosecond)** 단위로 입력). 이 기간이 지난 이후에는 프로포절 보증금 추가가 불가능합니다.
- `voting_params`
    + `voting_period`: 프로포절의 투표 기간(**나노초(nanosecond)** 단위로 입력).
- `tally_params`
    + `quorum`: 프로포절 투표 결과가 유효하기 위한 위임된 스테이킹 토큰의 투표율.
    + `threshold`: 프로포절 투표가 통과하기 위해 필요한 최소 `YES` 투표 %.
    + `veto`: 프로포절 투표 결과가 유효하기 위한 `NO_WITH_VETO` 투표 %의 최대 한도.
    + `governance_penalty`: 프로포절 투표에 참여하지 않은 검증인들에 부과하는 페널티.

### 슬래싱(slashing) 

`slashing` 모듈은 검증인의 악의적인 행동으로 발생하는 위임자 슬래싱 페널티 로직을 처리합니다.

```json
"slashing": {
      "params": {
        "max_evidence_age": "1814400000000000",
        "signed_blocks_window": "10000",
        "min_signed_per_window": "0.050000000000000000",
        "downtime_jail_duration": "600000000000",
        "slash_fraction_double_sign": "0.050000000000000000",
        "slash_fraction_downtime": "0.000100000000000000"
      },
      "signing_infos": {},
      "missed_blocks": {}
    }
```

각 파라미터 값에 대해 알아보겠습니다:

- `params`
    + `max_evidence_age`: 증거 최대 유효기간 (**나노초(nanosecond)** 단위) 
    + `signed_blocks_window`: 오프라인 검증인 판단을 위해 검토되는 최근 블록 개수.
    + `min_signed_per_window`: 검증인이 온라인으로 간주되기 위해`singed_blocks_window` 내에 포함되어야하는 최소 `precommit` %.
    + `downtime_jail_duration`: 다운 타임 슬래싱으로 발생하는 제일(jail) 기간(**나노초(nanosecond)** 단위.
    + `slash_fraction_double_sign`: 검증인이 더블 사이닝을 할 경우 슬래싱되는 위임자 위임량의 % 단위.
    + `slash_fraction_downtime`: 검증인이 오프라인인 경우 슬래싱되는 위임자 워임량의 % 단위.
- `signing_infos`: `slashing` 모듈이 사용하는 각 검증인의 정보. 과거 상태에서 제네시스가 생성되지 않은 경우 `{}` 값이 기본 설정입니다.
- `missed_blocks`: `slashing` 모듈이 사용하는 missed blocks 정보. 과거 상태에서 제네시스가 생성되지 않은 경우 `{}` 값이 기본 설정입니다.

### 제네시스 트랜잭션(genesis transactions)

기본적인 상태에서 제네시스 파일은 `gentx`를 포함하지 않습니다. `gentx`는 제네시스 파일 내 `accounts` 항목에 있는 스테이킹 토큰을 검증인에게 위임하는 트랜잭션입니다. 실질적으로 제네시스에서 검증인을 생성하는데 사용됩니다. 유효한 `gentx`를 보유한 검증인(보팅 파워 기준)의 2/3가 `genesis_time` 이후 온라인되면 블록체인이 시작됩니다.

`gentx`는 수동으로 제네시스 파일에 추가되거나 다음 명령어를 사용해 추가할 수 있습니다:

```bash
gaiad collect-gentxs
```

위 명령어는 `~/.gaia/config/gentx`에 있는 모든 `gentxs`를 제네시스 파일에 추가합니다. 제네시스 트랜잭션을 생성하기 위해서는 [여기](./validators/validator-setup.md#participate-in-genesis-as-a-validator)를 확인하세요.