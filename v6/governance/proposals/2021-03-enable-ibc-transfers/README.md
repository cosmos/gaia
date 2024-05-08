# Enable IBC Transfers

The Proposal enables transferring and receiving assets using the ICS20 standard on the Cosmos Hub. If this proposal passes, there will be IBC assets available in the Bank module of the Hub and ATOM will be available on Zones connected over IBC.
Iqlusion believes that the IBC software is sufficiently stable for small amounts of value transfer. We expect there to be issues with stuck funds and UX confusion but overcoming these issues will only happen once IBC is live.

## Security Model

Tendermint full nodes produce agreement under the assumption that at most ⅓ of the voting power held by validators is Byzantine.

## IBC

IBC is a protocol for authenticated message passing between heterogeneous sovereign blockchains. IBC requires trusting that chains on both sides of the connections operate within their security model.

## Incentive Security Extensions

IBC has a facility to support freezing connections once a violation of the security model has occurred. The set of criteria for detecting such attacks continues to evolve and is a constant focus of research.
