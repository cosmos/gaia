FROM fedora:33
ENV GOPATH=/go
ENV PATH=$PATH:/go/bin
RUN dnf install -y git make gcc gcc-c++ which iproute iputils procps-ng vim-minimal tmux net-tools htop tar jq npm openssl-devel perl golang rust cargo
ADD https://gethstore.blob.core.windows.net/builds/geth-linux-amd64-1.10.3-991384a7.tar.gz /geth/
RUN cd /geth && tar -xvf * && mv /geth/**/geth /usr/bin/geth
# this is the Gravity bridge release, update it along with your go.mod entry
ADD https://github.com/althea-net/cosmos-gravity-bridge/releases/download/v0.0.29/test-runner /usr/bin/test-runner
RUN chmod +x /usr/bin/test-runner
ADD https://github.com/althea-net/cosmos-gravity-bridge/releases/download/v0.0.29/Gravity.json /solidity/Gravity.json
ADD https://github.com/althea-net/cosmos-gravity-bridge/releases/download/v0.0.29/TestERC20A.json /solidity/TestERC20A.json
ADD https://github.com/althea-net/cosmos-gravity-bridge/releases/download/v0.0.29/TestERC20B.json /solidity/TestERC20B.json
ADD https://github.com/althea-net/cosmos-gravity-bridge/releases/download/v0.0.29/TestERC20C.json /solidity/TestERC20C.json
ADD https://github.com/althea-net/cosmos-gravity-bridge/releases/download/v0.0.29/contract-deployer /solidity/contract-deployer
RUN chmod +x /solidity/contract-deployer
# the actual source code for this repo, this **only** includes checked in files!
# this is a bit of a pain but it does speed things up a lot
ADD althea.tar.gz /
# build steps for all codebases in this repo, must be below the add statement
RUN pushd /althea/ && PATH=$PATH:/usr/local/go/bin GOPROXY=https://proxy.golang.org PATH=$PATH:/usr/local/go/bin make install
