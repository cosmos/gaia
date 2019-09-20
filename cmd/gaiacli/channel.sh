CHANID1="chan-$(openssl rand -hex 2)"
CHANID2="chan-$(openssl rand -hex 2)"
echo "establishing channel..."
echo "channel 1: $CHANID1"
echo "channel 2: $CHANID2"
./gaiacli tx ibc channel handshake ibc-mock $CHANID1 $1 ibc-mock $CHANID2 $2 --from1 node0 --from2 node0 --home ../node0/gaiacli --chain-id $CID
