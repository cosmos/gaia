echo "writing path..."
./gaiacli query ibc client path --chain-id $CID --home ../node0/gaiacli > path.json
echo "writing state..."
./gaiacli query ibc client consensus-state --chain-id $CID --home ../node0/gaiacli > state.json

CLIENTID="client-$(openssl rand -hex 2)"
echo "creating client $CLIENTID"
./gaiacli tx ibc client create $CLIENTID ./state.json --from node0 --home ../node0/gaiacli --chain-id $CID
