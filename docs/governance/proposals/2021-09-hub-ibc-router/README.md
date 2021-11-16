# ⚛️ Make the Cosmos Hub the IBC Router ⚛️

The following is a selection from the [Cosmos Whitepaper](https://v1.cosmos.network/resources/whitepaper):

```
The Cosmos Hub connects to many other blockchains (or zones) via a novel inter-blockchain communication protocol. The Cosmos Hub tracks numerous token types and keeps record of the total number of tokens in each connected zone. Tokens can be transferred from one zone to another securely and quickly without the need for a liquid exchange between zones, because all inter-zone coin transfers go through the Cosmos Hub.

...

Any of the zones can themselves be hubs to form an acyclic graph, but for the sake of clarity we will only describe the simple configuration where there is only one hub, and many non-hub zones.
```

The Hub has long been envisioned as a central point in the IBC architecture. In the battle to build and ship IBC this central vision has remained unchanged, but with so much focus on the need to build out other zones with real economies to support this network (the CosmosSDK is the result of this effort), the idea of the hub as an Interchain Router hasn't been discussed in a serious context for quite a while.

This is understandable: Cosmos needed so many other pieces to come together before the Hub had a chance to even start performing this function. Those other zones have been created, they each have products and economies. The bootstrapping era of IBC is well underway. 

These new zones joining are noticing a problem: they need to maintain a large amount of infrastructure (archive nodes and relayers for each counterparty chain) to connect with all the chains in the ecosystem, a number that is continuing to increase quickly.

Luckly this problem has been anticipated and IBC architected to accomodate multi-hop transactions. However, a packet forwarding/routing feature was not in the initial IBC release. This proposal aims to fix this for the Hub.

This is a proposal to include a new feature to IBC on the Hub that allows for multi-hop packet routing for ICS20 transfers. By appending an intermediate address, and the port/channel identifiers for the final destination, clients will be able to outline more than one transfer at a time. The following example shows routing from Terra to Osmosis through the Hub:

```json
// Packet sent from Terra to the hub, note the format of the forwaring info
// {intermediate_refund_address}|{foward_port}/{forward_channel}:{final_destination_address}
{
    "denom": "uluna",
    "amount": "100000000",
    "sender": "terra15gwkyepfc6xgca5t5zefzwy42uts8l2m4g40k6",
    "receiver": "cosmos1vzxkv3lxccnttr9rs0002s93sgw72h7ghukuhs|transfer/channel-141:osmo1vzxkv3lxccnttr9rs0002s93sgw72h7gl89vpz",
}

// When OnRecvPacket on the hub is called, this packet will be modified for fowarding to transfer/channel-141.
// Notice that all fields execept amount are modified as follows:
{
    "denom": "ibc/FEE3FB19682DAAAB02A0328A2B84A80E7DDFE5BA48F7D2C8C30AAC649B8DD519",
    "amount": "100000000",
    "sender": "cosmos1vzxkv3lxccnttr9rs0002s93sgw72h7ghukuhs",
    "receiver": "osmo1vzxkv3lxccnttr9rs0002s93sgw72h7gl89vpz",
}
```

Strangelove Ventures has delivered an [IBC Middleware module](https://github.com/cosmos/ibc-go/pull/373) that will allow the hub to play the role of IBC Router that was always envisioned for it. Passing of this propsal will begin the era of the Hub offering interchain services to other chains and profiting from those relationships.

To pay the hub validators and stakers, this proposal implements a governance configurable fee (which we propose should be initially set to 0.0 to encourage adoption) that will be taken out of each packet and given to the community pool. The community pool will then periodically trade these fees for ATOM and distribute them to staked holders. The exact distribution method of these fees is left TBD in this proposal as it is not initially required and can be implemented in a future governance proposal. One way to do this would be using the [Groups module](https://docs.cosmos.network/master/architecture/adr-042-group-module.html), Community spend proposals and the Gravity DEX.

A vote YES on this proposal indicates that this feature should be included in the next hub upgrade. We (as the Hub) believe that time is critical right now and we cannot wait to begin providing this service to other chains. A NO vote indicates that this shouldn't be included in the next upgrade.