# 자체 테스트넷 구축하기

해당 문서는 `gaiad` 노드 네트워크를 구축하는 세가지 방법을 설명합니다. 각 테스트넷 모델은 다른 테스트 목적에 최적화 되어있습니다.

1. 싱글-노드, 로컬, 수동 테스트넷
2. 멀티-노드, 로컬, 자동 테스트넷
3. 멀티-노드, 리모트, 자동 테스트넷

관련 코드는 [네트워크 디렉토리](https://github.com/cosmos/cosmos-sdk/tree/develop/networks)와 하단의 `local`과 `remote` 서브 디렉토리에서 찾으실 수 있습니다.

> 참고: 현재 `remote` 관련 정보는 최신 릴리즈와 호환성이 맞지 않을 수 있으므로 참고하시기 바랍니다.

## Docker 이미지

컨테이너 형태로 Gaia 디플로이를 원하시는 경우, `build` 단계를 건너뛰시고 공식 이미지 파일을 설치하실 수 있습니다. \$TAG은 설치하시려는 버전을 의미합니다.

- `docker run -it -v ~/.gaia:/root/.gaia -v ~/.gaia:/root/.gaia tendermint:$TAG gaiad init`
- `docker run -it -p 26657:26657 -p 26656:26656 -v ~/.gaia:/root/.gaia -v ~/.gaia:/root/.gaia tendermint:$TAG gaiad start`
- `docker run -it -v ~/.gaia:/root/.gaia -v ~/.gaia:/root/.gaia tendermint:$TAG gaiad version`

각 이미지는 자체적인 docker-compose 스택을 빌드하는데 사용하실 수 있습니다.

## 싱글-노드, 로컬, 수동 테스트넷

이 가이드는 한 개의 검증인 노드로 구성된 테스트넷을 로컬 환경에서 운영하는 방식을 알려드립니다. 테스트 용도 또는 개발 용도로 이용될 수 있습니다.

### 필수 사항

- [gaia 설치](./installation.md)
- [`jq` 설치](https://stedolan.github.io/jq/download/) (선택 사항)

### 제네시스 파일 만들기, 네트워크 시작하기

```bash
# 모든 명령어는 홈 디렉토리에서 실행하실 수 있습니다
cd $HOME

# 네트워크를 시작할 genesis.json 파일을 초기화하기
gaiad init --chain-id=testing testing

# 밸리데이터 키 생성하기
gaiad keys add validator

# 해당 키를 제네시스 파일에 있는 genesis.app_state.accounts 어레이(array)에 추가하세요
# 참고: 이 명령어는 코인 수량을 설정할 수 있게 합니다. 위 계정에 코인 잔고를 포함하세요
# genesis.app_state.staking.params.bond_denom의 기본 설정은 is staking gaiad add-genesis-account $(gaiad keys show validator -a) 1000stake,1000validatortoken 입니다.

# 밸리데이터 생성 트랜잭션 실행
gaiad gentx --name validator

# 제네시스 파일에 초기 본딩 트랜잭션 추가하기
gaiad collect-gentxs

# 이제 `gaiad`를 실행하실 수 있습니다.
gaiad start
```

이 셋업은 모든 `gaiad` 정보를 `~/.gaia`에 저장힙니다. 생성하신 제네시스 파일을 확인하고 싶으시다면 `~/.gaia/config/genesis.json`에서 확인이 가능합니다. 위의 세팅으로 `gaiad`가 이용이 가능하며, 토큰(스테이킹/커스텀)이 있는 계정 또한 함께 생성됩니다.

## 멀티 노드, 로컬, 자동 테스트넷

관련 코드 [networks/local 디렉토리](https://github.com/cosmos/cosmos-sdk/tree/develop/networks/local):

### 필수 사항

- [gaia 설치](./installation.md)
- [docker 설치](https://docs.docker.com/engine/installation/)
- [docker-compose 설치](https://docs.docker.com/compose/install/)

### 빌드

`localnet` 커맨드를 운영하기 위한 `gaiad` 바이너리(리눅스)와 `tendermint/gaiadnode` docker 이미지를 생성합니다. 해당 바이너리는 컨테이너에 마운팅 되며 업데이트를 통해 이미지를 리빌드 하실 수 있습니다.

```bash
# Clone the gaia repo
git clone https://github.com/cosmos/gaia.git

# Work from the SDK repo
cd gaia

# Build the linux binary in ./build
make build-linux

# Build tendermint/gaiadnode image
make build-docker-gaiadnode
```

### 테스트넷 실행하기

4개 노드 테스트넷을 실행하기 위해서는:

```
make localnet-start
```

이 커맨드는 4개 노드로 구성되어있는 네트워크를 gaiadnode 이미지를 기반으로 생성합니다. 각 노드의 포트는 하단 테이블에서 확인하실 수 있습니다:

| 노드 ID     | P2P 포트 | RPC 포트 |
| ----------- | -------- | -------- |
| `gaianode0` | `26656`  | `26657`  |
| `gaianode1` | `26659`  | `26660`  |
| `gaianode2` | `26661`  | `26662`  |
| `gaianode3` | `26663`  | `26664`  |

바이너리를 업데이트 하기 위해서는 리빌드를 하신 후 노드를 재시작 하시면 됩니다:

```
make build-linux localnet-start
```

### 설정

`make localnet-start`는 `gaiad testnet` 명령을 호출하여 4개 노드로 구성된 테스트넷에 필요한 파일을 `./build`에 저장합니다. 이 명령은 `./build` 디렉토리에 다수의 파일을 내보냅니다.

```bash
$ tree -L 2 build/
build/
├── gaiad
├── gaiad
├── gentxs
│   ├── node0.json
│   ├── node1.json
│   ├── node2.json
│   └── node3.json
├── node0
│   ├── gaiad
│   │   ├── key_seed.json
│   │   └── keys
│   └── gaiad
│       ├── ${LOG:-gaiad.log}
│       ├── config
│       └── data
├── node1
│   ├── gaiad
│   │   └── key_seed.json
│   └── gaiad
│       ├── ${LOG:-gaiad.log}
│       ├── config
│       └── data
├── node2
│   ├── gaiad
│   │   └── key_seed.json
│   └── gaiad
│       ├── ${LOG:-gaiad.log}
│       ├── config
│       └── data
└── node3
    ├── gaiad
    │   └── key_seed.json
    └── gaiad
        ├── ${LOG:-gaiad.log}
        ├── config
        └── data
```

각 `./build/nodeN` 디렉토리는 각자 컨테이너 안에 있는 `/gaiad`에 마운팅 됩니다.

### 로깅

로그는 각 `./build/nodeN/gaiad/gaia.log`에 저장됩니다. 로그는 docker를 통해서 바로 확인하실 수도 있습니다:

```
docker logs -f gaiadnode0
```

### 키와 계정

`gaiad`를 이용해 tx를 생성하거나 상태를 쿼리 하시려면, 특정 노드의 `gaiad` 디렉토리를 `home`처럼 이용하시면 됩니다. 예를들어:

```bash
gaiad keys list --home ./build/node0/gaiad
```

이제 계정이 존재하니 추가로 새로운 계정을 만들고 계정들에게 토큰을 전송할 수 있습니다.

::: tip
**참고**: 각 노드의 시드는 `./build/nodeN/gaiad/key_seed.json`에서 확인이 가능하며 `gaiad keys add --restore` 명령을 통해 CLI로 복원될 수 있습니다.
:::

### 특수 바이너리

다수의 이름을 가진 다수의 바이너리를 소유하신 경우, 어떤 바이너리의 환경 변수(environment variable)를 기준으로 실행할지 선택할 수 있습니다. 바이너리의 패스(path)는 관련 볼륨(volume)에 따라 달라집니다. 예시:

```
# Run with custom binary
BINARY=gaiafoo make localnet-start
```

## 멀티 노드, 리모트, 자동 테스트넷

다음 환경은 [네트워크 디렉터리](https://github.com/cosmos/cosmos-sdk/tree/develop/networks)에서 실행하셔야 합니다.

### Terraform과 Ansible

자동 디플로이멘트(deployment)는 [Terraform](https://www.terraform.io/)를 이용해 AWS 서버를 만든 후 [Ansible](http://www.ansible.com/)을 이용해 해당 서버에서 테스트넷을 생성하고 관리하여 운영됩니다.

### 필수 사항

- [Terraform](https://www.terraform.io/downloads.html) 과 [Ansible](http://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html)를 리눅스 머신에 설치.
- EC2 create 권한이 있는 [AWS API 토큰](https://docs.aws.amazon.com/general/latest/gr/managing-aws-access-keys.html) 생성
- SSH 키 생성.

```
export AWS_ACCESS_KEY_ID="2345234jk2lh4234"
export AWS_SECRET_ACCESS_KEY="234jhkg234h52kh4g5khg34"
export TESTNET_NAME="remotenet"
export CLUSTER_NAME= "remotenetvalidators"
export SSH_PRIVATE_FILE="$HOME/.ssh/id_rsa"
export SSH_PUBLIC_FILE="$HOME/.ssh/id_rsa.pub"
```

해당 명령은 `terraform` 과 `ansible`에서 이용됩니다..

### 리모트 네트워크 생성하기

```
SERVERS=1 REGION_LIMIT=1 make validators-start
```

테스트넷 이름은 --chain-id에서 이용될 값이며, 클러스터 이름은 AWS 서버 관리 태그에서 이용될 값입니다. cluster name은 서버의 AWS 관리용 태그입니다. 해당 코드는 SERVERS 개수에 비례하는 서버를 REGION_LIMIT 값까지 us-east-2 리전부터 생성합니다 (us-east-1는 제외됩니다). 다음 BaSH 스크립트는 동일한 명령을 실행하나, 개인 선호도에 따라 입력하기 편하실 수 있습니다.

```
./new-testnet.sh "$TESTNET_NAME" "$CLUSTER_NAME" 1 1
```

### /status 엔드포인트 빠르게 확인하기

```
make validators-status
```

### 서버 삭제하기

```
make validators-stop
```

### 로깅

로그는 Elastic stack (Elastic search, Logstash, Kibana) 서비스를 제공하는 Logz.io로 내보내실 수 있습니다. 또한, 노드가 자동으로 해당 서비스에 로깅을 진행하실 수 있습니다. 계정을 생성하시고 [이 페이지](https://app.logz.io/#/dashboard/data-sources/Filebeat)의 노트에서 API키를 받으신 후:

```
yum install systemd-devel || echo "This will only work on RHEL-based systems."
apt-get install libsystemd-dev || echo "This will only work on Debian-based systems."

go get github.com/mheese/journalbeat
ansible-playbook -i inventory/digital_ocean.py -l remotenet logzio.yml -e LOGZIO_TOKEN=ABCDEFGHIJKLMNOPQRSTUVWXYZ012345
```

### 모니터링

다음과 같이 DataDog 에이전트를 설치하실 수 있습니다:

```
make datadog-install
```
