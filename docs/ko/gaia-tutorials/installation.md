## Gaia 설치하기

이 가이드는 `gaiad`와 `gaiad`를 엔트리포인트를 시스템에 설치하는 방법을 설명합니다. `gaiad`와 `gaiad`가 설치된 서버를 통해 [풀노드](./gaia-tutorials/join-testnet.md#run-a-full-node) 또는 [밸리데이터로](./validators/validator-setup.md)써 최신 테스트넷에 참가하실 수 있습니다.

### Go 설치하기

공식 [Go 문서](https://golang.org/doc/install)를 따라서 `go`를 설치하십시오. 그리고 `$PATH`의 환경을 꼭 세팅하세요. 예시:

```bash
mkdir -p $HOME/go/bin
echo "export PATH=$PATH:$(go env GOPATH)/bin" >> ~/.bash_profile
source ~/.bash_profile
```

::: tip
코스모스 SDK를 운영하기 위해서는 **Go 1.12+** 이상 버전이 필요합니다.
:::

### 바이너리 설치하기

다음은 최신 Gaia 버전을 설치하는 것입니다. 필요에 따라 `git checkout`을 통해 올바른 [릴리즈 버전](https://github.com/cosmos/gaia/releases)이 설치되어있는지 확인하세요.

```bash
git clone -b <latest-release-tag> https://github.com/cosmos/gaia
cd gaia && make install
```

만약 다음과 같은 에러 메시지로 명령어가 실패하는 경우, 이미 `LDFLAGS`를 설정하셨을 수 있습니다.

```
# github.com/cosmos/gaia/cmd/gaiad
flag provided but not defined: -L
usage: link [options] main.o
...
make: *** [install] Error 2
```

해당 환경변수를 언세팅 하신 후 다시 시도해보세요.

```
LDFLAGS="" make install
```

> _참고_: 여기에서 문제가 발생한다면, Go의 최신 스테이블 버전이 설치되어있는지 확인하십시오.

위 절차를 따라하시면 `gaiad`와 `gaiad` 바이너리가 설치될 것입니다. 설치가 잘 되어있는지 확인하십시오:

```bash
$ gaiad version --long
$ gaiad version --long
```

`gaiad` 명령어는 다음과 비슷한 아웃풋을 내보냅니다:

```bash
name: gaia
server_name: gaiad
client_name: gaiad
version: 1.0.0
commit: 89e6316a27343304d332aadfe2869847bf52331c
build_tags: netgo,ledger
go: go version go1.12.5 darwin/amd64
```

### 빌드 태그

빌드 태그는 해당 바이너리에서 활성화된 특별 기능을 표기합니다.

| 빌드 태그 | 설명                                          |
| --------- | --------------------------------------------- |
| netgo     | Name resolution이 오직 Go 코드만을 사용합니다 |
| ledger    | 렛저 기기(하드웨어 지갑)이 지원됩니다         |

### snap을 사용해 바이너리 설치하기 (리눅스에만 해당)

**재현가능한 바이너리 시스템이 완벽하게 구현되기 전까지 snap 바이너리를 실제 노드 운용에 사용하지 않으시는 것을 추천드립니다.**

## 개발자 워크플로우

코스모스 SDK 또는 텐더민트의 변경 사항을 테스팅하기 위해서는 `replace` 항목이 `go.mod`에 추가하여 올바른 import path를 제공해야합니다.

- 변경 사항 적용
- `go.mod`에 `replace github.com/cosmos/cosmos-sdk => /path/to/clone/cosmos-sdk` 추가
- `make clean install` 또는 `make clean build` 실행
- 변경 사항 테스트

### 다음 절차

축하합니다! 이제 [메인넷](./gaia-tutorials/join-mainnet.md)에 참가하거나[퍼블릭 테스트넷](./join-testnet.md)에 참가하시거나 [자체 테스트넷](./gaia-tutorials/deploy-testnet.md)을 운영하실 수 있습니다.
