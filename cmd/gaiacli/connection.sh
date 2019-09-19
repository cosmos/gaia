CONNID1="conn-$(openssl rand -hex 2)"
CONNID2="conn-$(openssl rand -hex 2)"
echo "establishing connection..."
echo "connection 1: $CONNID1"
echo "connection 2: $CONNID2"
./gaiacli tx ibc connection handshake $CONNID1 $1 ./path.json $CONNID2 $1 ./path.json --from1 node0 --from2 node0 --home ../node0/gaiacli --chain-id $CID
