# Gaia 제네시스 상태

Gaia genesis state, 또는 `GenesisState`는 다수의 계정, 모듈 상태 그리고 제네시스 트랜잭션 같은 메타데이터가 포함되어 있습니다. 각 모듈은 각자의 `GenesisState`를 지정할 수 있습니다. 또한, 각 모듈은 제네시스 상태 검증, 불러오기(import), 내보내기(export) 기능을 지정할 수 있습니다.

Gaia 제네시스 상태는 다음과 같이 정의됩니다:

```go
type GenesisState struct {
  AuthData     auth.GenesisState     `json:"auth"`
  BankData     bank.GenesisState     `json:"bank"`
  StakingData  staking.GenesisState  `json:"staking"`
  MintData     mint.GenesisState     `json:"mint"`
  DistrData    distr.GenesisState    `json:"distribution"`
  GovData      gov.GenesisState      `json:"gov"`
  SlashingData slashing.GenesisState `json:"slashing"`
  GenTxs       []json.RawMessage     `json:"gentxs"`
}
```

ABCI에서는 `initFromGenesisState`의 `initChainer`의 정의가 호출되며 내부적으로 각 모듈의 `InitGenesis`를 호출하여 각자의 `GenesisState`를 파라미터 값으로 지정합니다.

## 계정(Accounts)

제네시스 계정은 `x/auth` 모듈의 `GenesisState`에 정의되며 `accounts` 키 내에 존재합니다. 제네시스 계정의 표준 정의는 없으나 모든 제네시스 계정은 `x/auth`가 정의한 `GenesisAccount`를 사용합니다.

각 계정은 유효하고 고유한 계정 번호(account number), 시퀀스 번호(sequence number / nonce)와 주소가 있어야 합니다.

또한, 계정은 베스팅(락업) 형태로 존재할 수 있으며, 이 경우 필수 락업 정보를 정의해야 합니다. 베스팅 계정은 최소 `OriginalVesting`과 `EndTime` 값을 제공해야 합니다. 만약 `StartTime`이 정의되는 경우, 해당 계정은 지속적 베스팅 계정으로 취급되며 지정된 스케줄에 따라 지속적으로 토큰의 락업을 해제합니다. 정의되는 `StartTime`은 `EndTime`값보다 적어야 하지만, 미래에 시작될 수도 있습니다. 즉, 락업 해제는 제네시스 시간과 동일하지 않아도 무관합니다. 만약 Export된 상태(state)가 아닌 신규 상태에서 시작되는 체인일 경우, `OriginalVesting` 값은 `Coin` 값과 동일하거나 적어야 합니다.

<!-- TODO: Remaining modules and components in GenesisState -->
