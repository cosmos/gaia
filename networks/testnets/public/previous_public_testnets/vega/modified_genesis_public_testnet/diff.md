The difference between original exported genesis and the modified genesis for public testnet:
``` diff
690c690
<           "address": "cosmos1qq9ydrjeqalqa3zyqqtdczvuugsjlcc3c7x4d4",
---
>           "address": "cosmos10aak94tfdl3pgt8qe6ga75qh3zkf3anpq8aqg0",
693c693
<             "key": "AjEkAHzQakRnyUppiM5/hnA6h2D7NkdxExxgiCG+NiDh"
---
>             "key": "A81DhG/5sB6RA8dl/6jtmX0svTc0xJL5NjPPI/q4jJWP"
3395038c3395038
<           "address": "cosmos1qq9ydrjeqalqa3zyqqtdczvuugsjlcc3c7x4d4",
---
>           "address": "cosmos10aak94tfdl3pgt8qe6ga75qh3zkf3anpq8aqg0",
3395041c3395041
<               "amount": "160896",
---
>               "amount": "300000000160896",
4039796c4039796
<               "amount": "194616098248861",
---
>               "amount": "6194616098248861",
5464404c5464404
<           "amount": "277834757180509",
---
>           "amount": "6577834757180509",
6282421c6282421
<           "delegator_address": "cosmos1qq9ydrjeqalqa3zyqqtdczvuugsjlcc3c7x4d4",
---
>           "delegator_address": "cosmos10aak94tfdl3pgt8qe6ga75qh3zkf3anpq8aqg0",
6282425c6282425
<             "stake": "11316631.000000000000000000"
---
>             "stake": "6000000011316631.000000000000000000"
8638136c8638136
<         "voting_period": "1209600s"
---
>         "voting_period": "86400s"
12473131c12473131
<           "address": "cosmosvalcons1s0686a68krmr8f46ph6fklw0v8us4gdsm7nhz3",
---
>           "address": "cosmosvalcons10j0slt0nqmldje3lsytpj9qlny28ueezg92w6g",
12976124c12976124
<           "address": "cosmosvalcons1kq9xxgmn0uepav9c6kwxl4yh599kpyu28e7ee6",
---
>           "address": "cosmosvalcons1zarj77frdpvj97qktselfcppwcvf4ujn35a2ps",
13933841c13933841
<           "address": "cosmosvalcons1lrqpcp5p2792wqxhxmt8tjveypjlvh378gkddu",
---
>           "address": "cosmosvalcons1zv4650adjtwm9g3mk87gz38j6p83dhxgytprlj",
14011124c14011124
<           "address": "cosmosvalcons1s0686a68krmr8f46ph6fklw0v8us4gdsm7nhz3",
---
>           "address": "cosmosvalcons10j0slt0nqmldje3lsytpj9qlny28ueezg92w6g",
14011575c14011575
<           "address": "cosmosvalcons1kq9xxgmn0uepav9c6kwxl4yh599kpyu28e7ee6",
---
>           "address": "cosmosvalcons1zarj77frdpvj97qktselfcppwcvf4ujn35a2ps",
14012554c14012554
<           "address": "cosmosvalcons1lrqpcp5p2792wqxhxmt8tjveypjlvh378gkddu",
---
>           "address": "cosmosvalcons1zv4650adjtwm9g3mk87gz38j6p83dhxgytprlj",
14012556c14012556
<             "address": "cosmosvalcons1lrqpcp5p2792wqxhxmt8tjveypjlvh378gkddu",
---
>             "address": "cosmosvalcons1zv4650adjtwm9g3mk87gz38j6p83dhxgytprlj",
14012857,14012858c14012857,14012858
<           "delegator_address": "cosmos1qq9ydrjeqalqa3zyqqtdczvuugsjlcc3c7x4d4",
<           "shares": "11316631.000000000000000000",
---
>           "delegator_address": "cosmos10aak94tfdl3pgt8qe6ga75qh3zkf3anpq8aqg0",
>           "shares": "6000000011316631.000000000000000000",
14733643c14733643
<       "last_total_power": "194616038",
---
>       "last_total_power": "6194616038",
14733955c14733955
<           "power": "13944328"
---
>           "power": "6013944328"
14830337c14830337
<             "key": "cOQZvh/h9ZioSeUMZB/1Vy1Xo5x2sjrVjlE/qHnYifM="
---
>             "key": "p6ihCq31IZUeY6z00G9ROoHTphnhi1J7wFrZ+5F2epU="
14835615c14835615
<             "key": "W459Kbdx+LJQ7dLVASW6sAfdqWqNRSXnvc53r9aOx/o="
---
>             "key": "bf5gFMl/dQxJFVE4jReOYxbVeux8UcFJ9lj1+qDZDGs="
14835617c14835617
<           "delegator_shares": "13944328343563.000000000000000000",
---
>           "delegator_shares": "6013944328343563.000000000000000000",
14835629c14835629
<           "tokens": "13944328343563",
---
>           "tokens": "6013944328343563",
14838167c14838167
<             "key": "NK3/1mb/ToXmxlcyCK8HYyudDn4sttz1sXyyD+42x7I="
---
>             "key": "un9hBl/53UOx5oFOu7+eOY1C0wOsdoVDfUW5VCH8TyA="
14838929c14838929
<   "chain_id": "cosmoshub-4",
---
>   "chain_id": "vega-testnet",
14838952c14838952
<       "address": "B00A6323737F321EB0B8D59C6FD497A14B60938A",
---
>       "address": "17472F7923685922F8165C33F4E02176189AF253",
14838957c14838957
<         "value": "cOQZvh/h9ZioSeUMZB/1Vy1Xo5x2sjrVjlE/qHnYifM="
---
>         "value": "p6ihCq31IZUeY6z00G9ROoHTphnhi1J7wFrZ+5F2epU="
14839645c14839645
<       "address": "83F47D7747B0F633A6BA0DF49B7DCF61F90AA1B0",
---
>       "address": "7C9F0FADF306FED9663F811619141F99147E6722",
14839647c14839647
<       "power": "13944328",
---
>       "power": "6013944328",
14839650c14839650
<         "value": "W459Kbdx+LJQ7dLVASW6sAfdqWqNRSXnvc53r9aOx/o="
---
>         "value": "bf5gFMl/dQxJFVE4jReOYxbVeux8UcFJ9lj1+qDZDGs="
14840023c14840023
<       "address": "F8C01C0681578AA700D736D675C9992065F65E3E",
---
>       "address": "132BAA3FAD92DDB2A23BB1FC8144F2D04F16DCC8",
14840028c14840028
<         "value": "NK3/1mb/ToXmxlcyCK8HYyudDn4sttz1sXyyD+42x7I="
---
>         "value": "un9hBl/53UOx5oFOu7+eOY1C0wOsdoVDfUW5VCH8TyA="
```
