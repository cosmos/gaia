# The Cosmos Hub Roadmap 2.0

This Cosmos Hub Product Roadmap incorporates input from product owners, key Cosmos stakeholders (ATOM holders, Interchain Foundation, and Cosmos Hub operations), the internal Gaia / ATOM call, the virtual Gaia Org Call, the [Cosmos Hub as a Port City](https://blog.cosmos.network/the-cosmos-hub-is-a-port-city-5b7f2d28debf) article, and the [ATOM2021](https://github.com/cosmosdevs/atom2021) presentation.

This roadmap gives a one-year guideline in which stakeholders can anticipate updated features on the Cosmos Hub, with the greatest degree of specificity available for the most immediate upgrades, and decreasing precision available the further out the timeline goes. For example, the roadmap is most precise for the upcoming Liquidity upgrades (the Gravity DEX and Gravity Bridge modules).

The upgrades aim to add features such as liquidity, economic security, usability, and participation. To highlight our focus on DeFi, we have chosen to use the [Greeks from Finance](https://en.wikipedia.org/wiki/Greeks_(finance)) in naming upcoming upgrades.

## Delta Upgrade

Gravity DEX:
- A scalable AMM model for token swaps
- Drives liquidity for tokens on the Cosmos Hub
- Delivers price consistency and order execution

## Vega Upgrade

Gravity Ethereum Bridge:
- Transfer ATOM, ETH, ERC-20, and tokens on the Cosmos Hub between Ethereum- and Cosmos-compatible chains
- Fee and reward model hosted across Comos and Ethereum
- Adds light-weight infrastructure and operational requirements with minimal slashing conditions to all Hub validators

## Theta Upgrade

Fee grant module:
- Allows paying fees on behalf of another account

Authz module:
- Provide governance functions to execute transactions on behalf of another account

Groups module:
- Enables higher-level multisig permissioned accounts, e.g., weight-based voting policies

Interchain accounts:
- A requirement in order to manage accounts across multiple blockchains
- Aims to provide locking/unlocking mechanisms across IBC-enabled blockchains

## Rho Upgrade

NFT module:
- Enable simple management of NFT identifiers, their owners, and associated data, such as URIs, content, and provenance
- An extensible base module for extensions including collectibles, custody, provenance, and marketplaces

Chain name service:
- Allows registration of unique chainids for IBC
- Interoperable with cross-chain validation and interchain staking

## Lambda Upgrade

Staking derivatives:
- Frees secure and low-risk delegations for use in other parts of the Cosmos ecosystem
- Features include enabling transfer of rewards and voting rights

Interchain staking:
- Cosmos solution to shared security using cross chain validation and interchain accounts
- Enables a parent chain, e.g., Cosmos Hub, to be in charge of producing blocks for a baby chain
- Validators of a baby chain will have their ATOM stake on the Cosmos Hub slashed for misbehaviour

Token Issuance:
- Enables creation of tokens directly on the Hub
- Aims to provide ERC20 capabilities

## Future Upgrades

- Cross-chain bridges (non-IBC)
- Atomic Exchange
- Decentralized identifiers (DID)
- Privacy
- Virtual machines
- Smart contract languages
- Zero knowledge and optimistic rollups

The Cosmos Hub Roadmap is maintained by the Interchain Cosmos Hub team as a living document, and is updated in collaboration with key stakeholders from the multi-entity Cosmos community. 