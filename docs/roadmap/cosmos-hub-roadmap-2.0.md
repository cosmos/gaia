# The Cosmos Hub Roadmap

This Cosmos Hub roadmap serves as a reference for the current planned features of upcoming releases. For past releases, please see the following [document](./README.md).

The Cosmos Hub is decentralized and other teams contribute to it as well. The broad categories are as follows:

## Replicated Security v1.1:  

These are the many changes we want to make but could not get to before launch. These include:

- Random code cleanups and protocol simplifications
- Untrusted consumer chain protocol: Make it so that there is no way a consumer chain can cause inconvenience or harm to the Cosmos Hub, while bringing back automatic slashing as detailed above.
- Soft opt-in: Make it so that the bottom 10-20% of validators on the Hub do not need to validate every consumer chain, making things easier for the smaller validators.

## Interchain Security future:

Work on the next versions of ICS.

- Opt-in security: Make it so that every validator can choose which consumer chains they want to validate. This requires fraud proofs or something similar.
- Mesh security: Participate in the design and adoption of Mesh security. This likely also requires fraud proofs. We’ll be releasing an analysis soon.


## Hub governance improvements: 
There are a lot of improvements to the Cosmos SDK governance system that could make deliberation easier and more constructive, and make community pool funding more accountable.
- Streaming/vesting funding: Allow a funding recipient to vest their funds over time, subject to claw backs if they don’t perform, or set up a streaming payment of future funds, also subject to clawback.
- Stabilized funding: Automatically exchange Atoms from the community pool for a stable asset, or adjust streaming funding amounts over time in response to an Atom price feed.
- Multichoice proposals: Allow voters to choose between several acceptable options- this would make it easier to pass proposals that accurately reflect what voters want.

