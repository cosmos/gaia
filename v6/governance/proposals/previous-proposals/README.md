# Previous proposals

This is a record of past proposals, including ones that weren't drafted in the governance repository. These proposals are present in the cosmoshub-4 genesis file.

## 1 Adjustment of blocks_per_year to come aligned with actual block time

* **Submitted:** 2019-03-20 02:41:27
* **Status:** PROPOSAL_STATUS_PASSED
* **Type:** /cosmos.gov.v1beta1.TextProposal

This governance proposal is for adjustment of blocks_per_year parameter to normalize the inflation rate and reward rate.\n ipfs link: https://ipfs.io/ipfs/QmXqEBr56xeUzFpgjsmDKMSit3iqnKaDEL4tabxPXoz9xc

## 2 ATOM Transfer Enablement

* **Submitted:** 2019-03-25T21:42:19.240550245Z
* **Status:** PROPOSAL_STATUS_REJECTED
* **Type:** /cosmos.gov.v1beta1.TextProposal

A plan is proposed to set up a testnet using the Cosmos SDK v0.34.0 release, along with mainnet conditions, plus transfer enablement and increased block size, as a testing ground. Furthermore, a path for upgrading the cosmoshub-1 chain to use the Cosmos SDK release v0.34.0, along with the necessary updates to the genesis file, at block 425000, is outlined. IPFS: https://ipfs.io/ipfs/QmaUaMjXPE6i4gJR1NakQc15TZpSqjSrXNmrS1vA5veF9W

## 3 ATOM Transfer Enablement v2

* **Submitted:** 2019-04-03T10:15:22.291176064Z
* **Status:** PROPOSAL_STATUS_PASSED
* **Type:** /cosmos.gov.v1beta1.TextProposal

A plan for enabling ATOM transfers is being proposed, which involves the release and test of Cosmos SDK v0.34.0 and a strategy for the network to accept the release and upgrade the mainnet once testing has been deemed to be successful. Read the full proposal at https://ipfs.io/ipfs/Qmam1PU39qmLBzKv3eYA3kMmSJdgR6nursGwWVjnmovpSy or formatted at https://ipfs.ink/e/Qmam1PU39qmLBzKv3eYA3kMmSJdgR6nursGwWVjnmovpSy

## 4 Proposal for issuance of fungible tokens directly on the Cosmos Hub

* **Submitted:** 2019-04-15T08:45:39.072577509Z
* **Status:** PROPOSAL_STATUS_REJECTED
* **Type:** /cosmos.gov.v1beta1.TextProposal

This proposal is a first step towards enabling fungible token issuance on the Cosmos Hub, with listing of new tokens requiring governance approval. Read the full proposal at https://github.com/validator-network/cosmoshub-proposals/blob/0d306f1fcc841a0ac6ed1171af96e6869d6754b6/issuance-proposal.md

## 5 Expedited Cosmos Upgrade Proposal

* **Submitted:** 2019-04-19T00:49:55.251313656Z
* **Status:** PROPOSAL_STATUS_PASSED
* **Type:** /cosmos.gov.v1beta1.TextProposal

Proposal to upgrade the Cosmos Hub at block 500,000 on April 22nd 5pm GMT. Details:https://ipfs.io/ipfs/QmS13GPNs1cRKSojete5y9RgW7wyf1sZ1BGqX3zjTGs7sX

## 6 Don't Burn Deposits for Rejected Governance Proposals Unless Vetoed

* **Submitted:** 2019-05-03T18:14:33.209053883Z
* **Status:** PROPOSAL_STATUS_PASSED
* **Type:** /cosmos.gov.v1beta1.TextProposal

Read here, or on https://ipfs.ink/e/QmRtR7qkeaZCpCzHDwHgJeJAZdTrbmHLxFDYXhw7RoF1ppnnThe Cosmos Hub's state machine handles spam prevention of governance proposals by means of a deposit system. A governance proposal is only considered eligible for voting by the whole validator set if a certain amount of staking token is deposited on the proposal. The intention is that the deposit will be burned if a proposal is spam or has caused a negative externality to the Cosmos community (such as wasting stakeholders’ time having to review the proposal).nnIn the current implementation of the governance module used in the Cosmos Hub, the deposit is burned if a proposal does not pass, regardless of how close the final tally result may have been. For example, if 49% of stake votes in favor of a proposal while 51% votes against it, the deposit will still be burned. This seems to be an undesirable behavior as it disincentivizes anyone from creating or depositing on a proposal that might be slightly contentious but not spam, due to fear of losing the deposit minimum (currently 512 atoms). This will especially be the case as TextProposals will be used for signaling purposes, to gauge the sentiment of staked Atom holders. Disincentivizing proposals for which the outcome is uncertain would undermine that effort.nnWe instead propose that the deposit be returned on failed votes, and that the deposit only be burned on vetoed votes. If a proposal seems to be spam or is deemed to have caused a negative externality to Cosmos communninty, voters should vote NoWithVeto on the proposal. If >33% of the stake chooses to Veto a proposal, the deposits will then be burned. However, if a proposal gets rejected without being vetoed, the deposits should be returned to the depositors. This proposal does not make any change to the current behavior for proposals that fail to meet quorum; if a proposal fails to meet quorum its deposit will be burned.

## 7 Activate the Community Pool

* **Submitted:** 2019-05-03T21:08:25.443199036Z
* **Status:** PROPOSAL_STATUS_PASSED
* **Type:** /cosmos.gov.v1beta1.TextProposal

Enable governance to spend funds from the community pool. Full proposal: https://ipfs.io/ipfs/QmNsVCsyRmEiep8rTQLxVNdMHm2uiZkmaSHCR6S72Y1sL1

## 8 Notification for Security Critical Hard Fork at Block 482100

* **Submitted:** 2019-05-30T19:43:02.870666885Z
* **Status:** PROPOSAL_STATUS_PASSED
* **Type:** /cosmos.gov.v1beta1.TextProposal

As described by user @Jessysaurusrex on Cosmos Forum in https://forum.cosmos.network/t/critical-cosmossdk-security-advisory/2211, All in Bits has learned of a critical security vulnerability in the codebase for the Cosmos Hub. We deem the issue to be of high severity, as if exploited it can potentially degrade the security model of the chain's Proof of Stake system. This vulnerability CANNOT lead to the theft of Atoms or creation of Atoms out of thin air. nn All in Bits has released a source code patch, Gaia v0.34.6, that closes the exploitable code path starting at block 482100. nn The proposed upgrade code Git hash is: 80234baf91a15dd9a7df8dca38677b66b8d148c1 nn As a proof of stake, we are putting some collateral behind this legitimacy of this bug and patch and encourage others familiar with the report to do so as well. If the disclosed bug turns out to be fabricated or malicious in some way, we urge the Cosmos Hub governance to slash these Atoms by voting NoWithVeto on this proposal. nn We encourage validators and all users to upgrade their nodes to Gaia v0.34.6 before block 482100. In the absence of another public bulletin board, we request validators to please vote Yes on this proposal AFTER they have upgraded their nodes to v0.34.6, as a method of signalling the readiness of the network for the upgrade.

## 10 Increase Max Validator Set Size to 125

* **Submitted:** 2019-07-01T14:09:25.508939113Z
* **Status:** PROPOSAL_STATUS_PASSED
* **Type:** /cosmos.gov.v1beta1.TextProposal

Read here, or on https://ipfs.ink/e/QmRhQycV19QiTQGLuPzPHfJwCioj1wDeHHtZvxiHegTFDd nnThis proposal supercedes proposal number 9, which contains conflicting numbers in the title and body. nnIn the Cosmos Hub, the total number of active validators is currently capped at 100, ordered by total delegated Atoms. This number was originally proposed in the Cosmos whitepaper section titled [Limitations on the Number of Validators](https://github.com/cosmos/cosmos/blob/master/WHITEPAPER.md#limitations-on-the-number-of-validators 4). This number was chosen as a relatively conservative estimate, as at the time of writing, it was unclear how many widely distributed nodes Tendermint consensus could scale to over the public internet. nnHowever, since then, we have seen empirically through the running of the Game of Stakes incentivized testnet that Tendermint Core with Gaia state machine can operate with over 180 validators at reasonable average block times of <7 seconds. The Game of Stakes results empirically show that adding validators should not delay consensus at small block sizes. At large block sizes, the time it takes for the block to gossip to all validators may increase depending on the newfound network topology. However we view this as unlikely, and if it did become a problem, it could later be solved by known improvements at the p2p layer. The other tradeoff to increasing the number of validators is that the size of commits becomes ~25% larger due to more precommits being included, increasing the network and storage costs for nodes. This can also be resolved in the future with the integration of aggregate signatures. At the time of submission of this proposal, the minimum delegation to become a top 100 validator is 30,600 Atoms, a fairly high barrier to entry for new validators looking to enter the active validator set. nnIn the Cosmos whitepaper, it states that the number of validators on the Hub will increase at a rate of 13% a year until it hits a cap of 300 validators. We propose scrapping this mechanism and instead increasing the max validators to 125 validators in the next chain upgrade with no further planned increases. Future increases to the validator set size will be originated through governance.

## 12 Are Validators Charging 0% Commission Harmful to the Success of the Cosmos Hub?

* **Submitted:** 2019-07-23T00:28:15.881319915Z
* **Status:** PROPOSAL_STATUS_PASSED
* **Type:** /cosmos.gov.v1beta1.TextProposal

This governance proposal is intended to act purely as a signalling proposal. Throughout this history of the Cosmos Hub, there has been much debate about the impact that validators charging 0% commission has on the Cosmos Hub, particularly with respect to the decentralization of the Cosmos Hub and the sustainability for validator operations. nn Discussion around this topic has taken place in many places including numerous threads on the Cosmos Forum, public Telegram channels, and in-person meetups. Because this has been one of the primary discussion points in off-chain Cosmos governance discussions, we believe it is important to get a signal on the matter from the on-chain governance process of the Cosmos Hub. nn There have been past discussions on the Cosmos Forum about placing an in-protocol restriction on validators from charging 0% commission. https://forum.cosmos.network/t/governance-limit-validators-from-0-commission-fee/2182 nn This proposal is NOT proposing a protocol-enforced minimum. It is merely a signalling proposal to query the viewpoint of the bonded Atom holders as a whole. nn We encourage people to discuss the question behind this governance proposal in the associated Cosmos Hub forum post here: https://forum.cosmos.network/t/proposal-are-validators-charging-0-commission-harmful-to-the-success-of-the-cosmos-hub/2505 nn Also, for voters who believe that 0% commission rates are harmful to the network, we encourage optionally sharing your belief on what a healthy minimum commission rate for the network using the memo field of their vote transaction on this governance proposal or linking to a longer written explanation such as a Forum or blog post. nn The question on this proposal is “Are validators charging 0% commission harmful to the success of the Cosmos Hub?”. A Yes vote is stating that they ARE harmful to the network's success, and a No vote is a statement that they are NOT harmful.

## 13 Cosmos Hub 3 Upgrade Proposal A

* **Submitted:** 2019-07-26T18:04:10.416760069Z
* **Status:** PROPOSAL_STATUS_PASSED
* **Type:** /cosmos.gov.v1beta1.TextProposal

This is a proposal to approve these high-level changes for a final vote for what will become Cosmos Hub 3. Please read them carefully: nhttps://github.com/cosmos/cosmos-sdk/blob/rc1/v0.36.0/CHANGELOG.mdnn-=-=-nnIf approved, and assuming that testing is successful, there will be a second proposal called Cosmos Hub 3 Upgrade Proposal B. Cosmos Hub 3 Upgrade Proposal B should specify 1) the software hash; 2) the block height state export from cosmoshub-2; 3) the genesis time; 4) instructions for generating the new genesis file.nn-=-=-nnFull proposal: nhttps://ipfs.io/ipfs/QmbXnLfx9iSDH1rVSkW5zYC8ErRZHUK4qUPfaGs4ZdHdc7n

## 14 Cosmos Hub 3 Upgrade Proposal B

* **Submitted:** 2019-08-23T16:16:19.814900321Z
* **Status:** PROPOSAL_STATUS_REJECTED
* **Type:** /cosmos.gov.v1beta1.TextProposal

This proposal is intended to signal acceptance/rejection of the precise software release that will contain the changes to be included in the Cosmos Hub 3 upgrade. A high level overview of these changes was successfully approved by the voters signalling via Cosmos Hub 3 Upgrade Proposal A: https://hubble.figment.network/cosmos/chains/cosmoshub-2/governance/proposals/13nnWe are proposing to use this code https://github.com/cosmos/gaia/releases/tag/v2.0.0 to upgrade the Cosmos Hub. We are proposing to export the ledger's state at Block Height 1823000, which we expect to occur on Sunday, September 15, 2019 at or around 2:00 pm UTC. We are proposing to launch Cosmos Hub 3 at 3:57 pm UTC on Sunday, September 15, 2019. nnInstructions for migration: https://github.com/cosmos/gaia/wiki/Cosmos-Hub-2-UpgradennFull proposal: https://ipfs.io/ipfs/Qmf54mwb8cSRf316jS4by96dL91fPCabvB9V5i2Sa1hxdznn

## 16 Cosmos Hub 3 Upgrade Proposal D

* **Submitted:** 2019-09-05T21:32:32.253341577Z
* **Status:** PROPOSAL_STATUS_PASSED
* **Type:** /cosmos.gov.v1beta1.TextProposal

Figment Networks (https://figment.network)nn-=-=-nnThis proposal is intended to supersede flawed Cosmos Hub 3 Upgrade Proposal B (https://hubble.figment.network/cosmos/chains/cosmoshub-2/governance/proposals/14) and Cosmos Hub 3 Upgrade Proposal C (https://hubble.figment.network/cosmos/chains/cosmoshub-2/governance/proposals/15), regardless of their outcomes. This proposal will make both Proposal 14 and 15 void.nnThis proposal is intended to signal acceptance/rejection of the precise software release that will contain the changes to be included in the Cosmos Hub 3 upgrade. A high overview of these changes was successfully approved by the voters signalling via Cosmos Hub 3 Upgrade Proposal A:nhttps://hubble.figment.network/cosmos/chains/cosmoshub-2/governance/proposals/13nn-=-=-nnWe are proposing to use this code https://github.com/cosmos/gaia/releases/tag/v2.0.0 to upgrade the Cosmos Hub. We are proposing to export the ledger’s state at Block Height 1,933,000, which we expect to occur on September 24, 2019 at or around 1:53 pm UTC. Please note that there will likely be a variance from this target time, due to changes in block time (https://forum.cosmos.network/t/cosmos-hub-3-upgrade-proposal-d/2675/18?u=gavin). We are proposing to launch Cosmos Hub 3 at 60 minutes after Block Height 1,933,000.nn-=-=-nnInstructions for migration: https://github.com/cosmos/gaia/wiki/Cosmos-Hub-2-UpgradenPlease note the recovery scenario in the case that the chain fails to start.nn-=-=-nnFull proposal:nhttps://ipfs.io/ipfs/QmPbSLvAgY8m7zAgSLHzKHfDtV4wx5XaGt1S1cDzqvXqJg

## 19 Cosmos Hub 3 Upgrade Proposal E

* **Submitted:** 2019-11-14T17:13:31.985706216Z
* **Status:** PROPOSAL_STATUS_PASSED
* **Type:** /cosmos.gov.v1beta1.TextProposal

Figment Networks (https://figment.network)nn-=-=-nnFull proposal:nhttps://ipfs.io/ipfs/QmfJyd64srJSX824WoNnF6BbvF4wvPGqVBynZeN98C7ygqnn-=-=-nn_Decision_nnWe are signalling that:nn1. The Gaia 2.0.3 implementation is aligned with the list of high-level changes approved in Cosmos Hub 3 Upgrade Proposal A.nn2. We are prepared to upgrade the Cosmos Hub to cosmoshub-3 based uponnta. Commit hash: 2f6783e298f25ff4e12cb84549777053ab88749a;ntb. The state export from cosmoshub-2 at Block Height 2902000;ntc. Genesis time: 60 minutes after the timestamp at Block Height 2902000.nn3. We are prepared to relaunch cosmoshub-2nta. In the event of:ntti. A non-trivial error in the migration procedure and/ornttii. A need for ad-hoc genesis file changesnttiii. The failure of cosmoshub-3 to produce two (2) blocks by 180 minutes after the timestamp of Block Height 2902000;ntb. Using:ntti. The starting block height: 2902000nttii. Software version: Cosmos SDK v0.34.6+ https://github.com/cosmos/cosmos-sdk/releases/tag/v0.34.10nttiii. The full data snapshot at export Block Height 2902000;ntc. And will consider the relaunch complete after cosmoshub-2 has reached consensus on Block 2902001.nn4. The upgrade will be considered complete after cosmoshub-3 has reached consensus on Block Height 2 within 120 minutes of genesis time.nn5. This proposal is void if the voting period has not concluded by Block Height 2852202.nn-=-=-nn_Context_nThis proposal follows Cosmos Hub 3 Upgrade Proposal D (https://hubble.figment.network/cosmos/chains/cosmoshub-2/governance/proposals/16) aka Prop 16, which passed in vote, but failed in execution (https://forum.cosmos.network/t/cosmos-hub-3-upgrade-post-mortem/2772). This proposal is intended to succeed where Prop 16 failed.nnThis proposal is intended to signal acceptance/rejection of the precise software release that will contain the changes to be included in the Cosmos Hub 3 upgrade. A high level overview of these changes was successfully approved by the voters signalling via Cosmos Hub 3 Upgrade Proposal A:nhttps://hubble.figment.network/cosmos/chains/cosmoshub-2/governance/proposals/13nnWe are proposing to use this code https://github.com/cosmos/gaia/releases/tag/v2.0.3 to upgrade the Cosmos Hub.nWe are proposing to export the ledger’s state at Block Height 2,902,000, which we expect to occur on December 11, 2019 at or around 14:27 UTC assuming an average of 6.94 seconds per block. Please note that there will likely be a variance from this target time, due to deviations in block time.nnWe are proposing that the Cosmos Hub 3 genesis time be set to 60 minutes after Block Height 2,902,000.nn-=-=-nnCo-ordination in case of failure will happen in this channel: https://riot.im/app/#/room/#cosmos_validators_technical_updates:matrix.org

## 23 Cosmos Governance Working Group - Q1 2020

* **Submitted:** 2020-01-15T06:51:48.001168602Z
* **Status:** PROPOSAL_STATUS_PASSED
* **Type:** /cosmos.distribution.v1beta1.CommunityPoolSpendProposal

Cosmos Governance Working Group - Q1 2020 fundingnnCommunity-spend proposal submitted by Gavin Birch (https://twitter.com/Ether_Gavin) of Figment Networks (https://figment.network)nn-=-=-nnFull proposal: https://ipfs.io/ipfs/QmSMGEoY2dfxADPfgoAsJxjjC6hwpSNx1dXAqePiCEMCbYnn-=-=-nnAmount to spend from the community pool: 5250 ATOMsnnTimeline: Q1 2020nnDeliverables:n1. A governance working group community & chartern2. A template for community spend proposalsn3. A best-practices document for community spend proposalsn4. An educational wiki for the Cosmos Hub parametersn5. A best-practices document for parameter changesn6. Monthly governance working group community calls (three)n7. Monthly GWG articles (three)n8. One Q2 2020 GWG recommendations articlennMilestones:nBy end of Month 1, the Cosmos Governance Working Group (GWG) should have been initiated and led by Gavin Birch of Figment Networks.nBy end of Month 2, Gavin Birch is to have initiated and led GWG’s education, best practices, and Q2 recommendations.nBy end of Month 3, Gavin Birch is to have led and published initial governance education, best practices, and Q2 recommendations.nnDetailed milestones and funding:nhttps://docs.google.com/spreadsheets/d/1mFEvMSLbiHoVAYqBq8lo3qQw3KtPMEqDFz47ESf6HEg/edit?usp=sharingnnBeyond the milestones, Gavin will lead the GWG to engage in and answer governance-related questions on the Cosmos Discourse forum, Twitter, the private Cosmos VIP Telegram channel, and the Cosmos subreddit. The GWG will engage with stake-holders to lower the barriers to governance participation with the aim of empowering the Cosmos Hub’s stakeholders. The GWG will use this engagement to guide recommendations for future GWG planning.nnRead more about the our efforts to launch the Cosmos GWG here: https://figment.network/resources/introducing-the-cosmos-governance-working-group/nn-=-=-nn_Problem_nPerhaps the most difficult barrier to effective governance is that it demands one of our most valuable and scarce resources: our attention. Stakeholders may be disadvantaged by informational or resource-based asymmetries, while other entities may exploit these same asymmetries to capture value controlled by the Cosmos Hub’s governance mechanisms.nnWe’re concerned that without establishing community standards, processes, and driving decentralized delegator-based participation, the Cosmos Hub governance mechanism could be co-opted by a centralized power. As governance functionality develops, potential participants will need to understand how to assess proposals by knowing what to pay attention to.nn_Solution_nWe’re forming a focused, diverse group that’s capable of assessing and synthesizing the key parts of a proposal so that the voting community can get a fair summary of what they need to know before voting.nnOur solution is to initiate a Cosmos governance working group that develops decentralized community governance efforts alongside the Hub’s development. We will develop and document governance features and practices, and then communicate these to the broader Cosmos community.nn_Future_nAt the end of Q1, we’ll publish recommendations for the future of the Cosmos GWG, and ideally we’ll be prepared to submit a proposal based upon those recommendations for Q2 2020. We plan to continue our work in blockchain governance, regardless of whether the Hub passes our proposals.nn-=-=-nnCosmos forum: https://forum.cosmos.network/c/governancenCosmos GWG Telegram channel: https://t.me/hubgovnTwitter: https://twitter.com/CosmosGov

## 25 CosmWasm Integration 1 - Permissions and Upgrades

* **Submitted:** 2020-05-12T17:10:00.465282299Z
* **Status:** PROPOSAL_STATUS_PASSED
* **Type:** /cosmos.distribution.v1beta1.CommunityPoolSpendProposal

CosmWasm Integration 1 - Permissions and UpgradesnnCommunity-spend proposal submitted by Ethan Frey (https://github.com/ethanfrey) of Confio UO (http://confio.tech/) and CosmWasm (https://www.cosmwasm.com)nn-=-=-nnFull proposal: https://ipfs.io/ipfs/QmbD3bMajQCFmtDmkuRVWhmMWVdN2sK8QP2FoFCz9cjPiCnForum Post: https://forum.cosmos.network/t/proposal-cosmwasm-on-cosmos-hub/3629nn-=-=-nnAmount to spend from the community pool: 25000 ATOMsnnTimeline: 2-4 months from approvalnnDeliverables:n1. Adding governance control to all aspects of the CosmWasm contract lifecycle to make it compatible with the hub. Allowing governance to control code upload, contract instantiation, upgrades, and destruction (if needed).n2. Adding ability to upgrade contracts along with migrations (also allowing orderly shutdowns). This controlled by a governance vote.n3. Launch a testnet with working version of this code (Cosmos SDK 0.38 or 0.39) to enable all interested parties to trial the process and provide feedback.n4. Provide sample contracts to demo on the testnet, along with some migration scenariosnnWithin 2 months, the working code and binaries should be delivered and open for public review. Within 4 months, these binaries will be used on a testnet, with sufficient staking tokens given to all active voters on the Cosmos Hub, and we will go through a few governance voting cycles to trial contract deployment and migrations (with a shorter voting cycles, eg. 3 days)nnDetailed milestones in the full proposal:nhttps://ipfs.io/ipfs/QmbD3bMajQCFmtDmkuRVWhmMWVdN2sK8QP2FoFCz9cjPiCnnBeyond the milestones, CosmWasm will enhance documentation of the platform and offer technical support on our Telegram channel.nn-=-=-nn_Problem_nWith the upcoming launch of IBC, the hub will need to adapt more rapidly to the needs of the ecosystem, while also limiting chain restarts, which may be detrimental to IBC connections. In particular support for relaying Dynamic IBC Protocols and Rented Security, using ATOMs as collateral for smaller zones, would greatly benefit from CosmWasm's flexibility.nn_Solution_nWe’re adding some key features to CosmWasm to convert it from a permissionless, immutable smart contract platform to a permissioned platform with governance control for upgrading or shutting down contracts. This is a key requirement to be able to integrate CosmWasm to the Cosmos Hub with minimal disruption.nn_Future_nWe will continue development of CosmWasm, especially adding IBC integration as well as working towards a stable 1.0 release that can be audited and safely deployed (Q3/Q4 2020).nn-=-=-nnTwitter: https://twitter.com/CosmWasmnMedium: https://medium.com/confionTelegram: https://t.me/joinchat/AkZriEhk9qcRw5A5U2MapAnWebsite: https://www.cosmwasm.comnGithub: https://github.com/CosmWasm

## 26 Takeoff Proposal from Cyber to Cosmos

* **Submitted:** 2020-05-21T18:00:11.292428073Z
* **Status:** PROPOSAL_STATUS_REJECTED
* **Type:** /cosmos.distribution.v1beta1.CommunityPoolSpendProposal

cyber Congress (https://cybercongress.ai) developed Cyber (https://github.com/cybercongress/go-cyber): a software for replacing existing internet behemoth monopolies, such as Google, which exploited outdated internet protocols using the common patterns of our semantic interaction. These corps lock the information, produced by the users, from search, social and commercial knowledge graphs in private databases, and then sell this knowledge back as advertisement. They stand as an insurmountable wall between content creators and consumers extracting an overwhelming majority of the created value.nnWe propose ATOM holders to invest 10,000 ATOM from the community pool into the Takeoff of Cyber. In exchange, at the end of its donation round (https://cyber.page/gol/takeoff), and when an IBC connection will become possible, cyber Congress will transfer CYB tokens back to the community pool. Passing this proposal will transfer 10,000 ATOMs from the community pool to cyber Congress multisig (https://www.mintscan.io/account/cosmos1latzme6xf6s8tsrymuu6laf2ks2humqv2tkd9a).nnFull Proposal-Manifest text: https://ipfs.io/ipfs/QmUYDQt9tqLQJwxnUck7dQY3XmZA3tDtpFh3Hchkg7oH46nnor at https://cyber.page/gol/takeoffnnThe software we offer resembles a decentralized google (https://github.com/cybercongress):n- A protocol spec and the rationale behind itn- go-cyber: our implementation using cosmos-sdkn- cyber.page: PoC reference web interfacen- launch-kit: useful tools for launching cosmos-sdk based chainsn- cyberindex: GraphQL middleware for cybern- euler Foundation: mainnet predecessor of cyber Foundation: the DAO, which will handle all the donated ETHn- documentation and various side toolsnnCyber solves the problem of opening up the centralised semantics core of the Internet. It does so by opening up access to evergrowing semantics core taught to it by the users.nnEconomics of the protocol are built around the idea that feedback loops between the number of links and the value of the knowledge graph exist. The more usage => the bigger the knowledge graph => the more value => the better the quality of the knowledge => the more usage. Transaction fees for basic operations are replaced by lifetime bandwidth, which means usability for both, end-users and developers. You can think of Cyber as a shared ASIC for search.nnYou already see that the idea of Cyber evolves around content identifiers and its ranks. From here, welcome to Decentralized Marketing, or DeMa. You've certainly heard of DeFi. DeFi is built around a simple idea that you can use a collateral for something that will be settled based on a provided price feed. Here comes the systematic problem of DeFi: price oracles. DeMa is based on the same idea of using collateral, but the input for settlement can be information regarding the content identifier itself.nnWith the help of DeMa and IBC chains will be able to prove relevance using content identifiers and their ranks one to another. This will help to grow the IBC ecosystem, where each chain has multiple possibilities to exchange data, which is provably valued.nnCosmos was created to become the internet of blockchains. A protocol that propagates the spirit of decentralization and governed by the community. For such technology to succeed, a lot is required. One thing is a solid foundation it can build on. One virtue of such foundation is monetary flow of income that has to feed this machine for as long as it exists.nnA good question that arises is how to turn the community pool into a pool that isn’t (a) a pot of money which goes solely to network security, (b) a pool that isn’t solely a build-up of inflationary rewards and (с) has long term prosperity value (its value rises).nnThe solution to the above problem is to establish a fund, that is managed and processed collectively and consists of a diversified number of assets that can bring long term value to its stakeholders.nnThis means using the funds to support exceptional projects that are building with Tendermint and Cosmos-SDK. After all, is we want to glorify the ecosystem, we need for it to grow. How will it grow? It will have projects with a clear utility, amazing a product and provable distribution. This will attract users, developers and large stakeholders to the ecosystem. Together we already did one very successful investment decision. We all participated in cosmos fundraizer. So let us move the idea forward.nnIf this proposal is successful and stands for more demand from the public, we will open another proposal using the community pool. However, anyone can participate in Game of Links (https://cyber.page/gol/) or Takeoff https://cyber.page/gol/takeoff independently. If you have question you can ask them either on Cyber topic on Cosmos forum (https://forum.cosmos.network/t/cyber-a-decentralized-google-for-provable-and-relevant-answers) or Cyber forum (https://ai.cybercongress.ai).nnProposal results: https://www.mintscan.io/proposals/26

## 27 Stargate Upgrade Proposal 1

* **Submitted:** 2020-07-12T06:23:02.440964897Z
* **Status:** PROPOSAL_STATUS_PASSED
* **Type:** /cosmos.gov.v1beta1.TextProposal

Stargate is our name for the process of ensuring that the widely integrated public network known as the Cosmos Hub is able to execute the cosmoshub-3 -> cosmoshub-4 upgrade with the minimum disruption to its existing ecosystem. This upgrade will also realize the Internet of Blockchains vision from the Cosmos whitepaper.nIntegrations from ecosystem partners are at risk of breaking changes due to the Stargate changes. These changes drive the need for substantial resource and time requirements to ensure successful migration. Stargate represents a unique set of circumstances and is not intended to set precedent for future upgrades which are expected to be less dramatic.nThere is a widespread consensus from many Cosmos stakeholders that these changes to core software components will enhance the performance and composability of the software and the value of the Cosmos Hub in a world of many blockchains.nA Yes result on this proposal provides a clear signal that the Cosmos Hub accepts and understands the Stargate process and is prepared to approve an upgrade with proposed changes if the plan below is executed successfully.nA No result would force a reconsideration of the tradeoffs in the Alternatives section and the forming a new plan to deliver IBC.nSee the full proposal here: https://ipfs.io/ipfs/Qmbo3fF54tX3JdoHZNVLcSBrdkXLie56Vh2u29wLfs4PnW

## 29 Genesis fund recovery proposal on behalf of fundraiser participants unable to access their ATOMs

* **Submitted:** 2020-09-09T06:47:46.521375251Z
* **Status:** PROPOSAL_STATUS_PASSED
* **Type:** /cosmos.gov.v1beta1.TextProposal

The purpose of this proposal is to restore access to geneis ATOMs for a subset of donors who have been active participants in our community through the last year.n The view of iqlusion is that this is an important moment for the Cosmos Hub. Stargate brings the fundraiser period to the end with delivery of IBC. This proposal resolves the open business of active members of our community who cannot access their ATOM. This is an opportunity is opporunity to bring this business to a close and setup the agenda for IBC powered innovation comming in 2021.We strongly encourage the Cosmos Community to verify the cryptographic evidence and bring these community members to full ATOM holder status.nnnFull Proposal:https://ipfs.io/ipfs/QmV6pBgDppN7X3BdVW197EUe7dpcmcdLMivPa6xxtPj3aW nThe original authors of the proposal will be available to answer questions on the Cosmos forum.nhttps://forum.cosmos.network/t/updated-genesis-atoms-recovery-request-proposal/3905


## 31 Governance Split Votes

* **Submitted:** 2020-11-23T00:53:38.508414880Z
* **Status:** PROPOSAL_STATUS_PASSED
* **Type:** /cosmos.gov.v1beta1.TextProposal

In the Cosmos Hub governance system, each address can only cast a vote for one option (Yes/No/Abstain/NoWithVeto) which uses their full voting power behind that choice.nnThis proposal proposes an upgrade to the Cosmos Hub governance module that would allow a staker to optionally split their votes into several voting options. For example, a single address could use 70% of its voting power to vote Yes and 30% of its voting power to vote No. Clients may opt into supporting this feature, as the existing UX of voting for a single option is preserved.nnThis is beneficial because oftentimes the entity owning that address might not be a single individual. For example, a company or organization that owns an address might have different stakeholders who want to vote differently, and so it makes sense to allow them to split their voting power.nnAnother example use case is exchanges and custodians. Many custodians and exchanges custody multiple customers’ ATOMs in the same address and use this address to stake on behalf of them. However, because of this, it makes it infeasible to do 'passthrough voting' and give their customers voting rights over their tokens, if different customers have different voting preferences. With this new proposal, custodians can use split votes to accurately reflect the preferences of their customers in on-chain governance.nnThe technical architecture for this feature can be seen in ADR 037 to the Cosmos SDK: https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-037-gov-split-vote.md nnAcceptance of this governance proposal is signalling approval to adopt this feature in a future upgrade of the Cosmos Hub.

## 32 Funding for Development of Governance Split Votes

* **Submitted:** 2020-11-24T17:22:36.584208993Z
* **Status:** PROPOSAL_STATUS_PASSED
* **Type:** /cosmos.distribution.v1beta1.CommunityPoolSpendProposal

Sikka is requesting 1776 ATOMs from the community pool to architect and implement the Governance Split Votes feature proposed in Cosmos Hub Proposal #31. This community fund proposal is dependent on the passing of Proposal #31 and thus should only be approved if Proposal #31 is approved. We request 1776 ATOMs, valuing each atom at $5.1 nnSikka has already begun the design of this feature and submitted it as ADR 037 to the Cosmos Hub: https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-037-gov-split-vote.md nn As past contributors to the codebase that runs the Cosmos Hub, we are familiar with the security and code quality requirements to be included in the Cosmos Hub. Sikka will implement & test this feature and will work with the maintainers of the github.com/cosmos/cosmos-sdk repo to get it merged into the x/gov module.

## 34 Luna Mission - Funding $ATOM

* **Submitted:** 2021-01-05T23:09:26.477112871Z
* **Status:** PROPOSAL_STATUS_PASSED
* **Type:** /cosmos.distribution.v1beta1.CommunityPoolSpendProposal

The Cosmos Hub (ATOM) community is requesting a community pool spend amount of 129,208 ATOM in order to implement a comprehensive ATOM marketing plan that will be executed in collaboration with AiB (Tendermint). The marketing efforts will be initiated immediately upon passing of proposal #34.nn The distribution of funds will be administered by 5 community members, that have been carefully selected by the community via the Cosmos governance working group to administer the marketing plan and release funds to either AiB that will act as a liaison between Cosmos Hub community and third parties or directly to parties that will be in charge of executing the marketing plan based on a majority multisignature approval. At least 3 members will have to approve each milestone-spend for it to be released to AiB based on the expected proposal scope &completion. nn More details can be found in the long form proposal here: https://cloudflare-ipfs.com/ipfs/QmWAxtxf7fUprPVWx1jWyxSKjBNqkcbA3FG6hRps7QTu3k and https://github.com/cosmos/governance/pull/10 and https://forum.cosmos.network/t/draft-governance-proposal-for-a-community-pool-spend-proposal-33-luna-mission-funding-atom/4244/15 nn The multisig administration includes: n @johnniecosmos, @JoeDirtay, @jackzampolin (Jack Zampolin, Pylon Validator), @immasssi (SG-1 Validator), @zakimanian (Zaki Manian, Iqlusion Validator).

## 35 Cosmos Stargate Hub Upgrade Proposal 2: Time to Upgrade.

* **Submitted:** 2021-01-12T01:37:07.471992293Z
* **Status:** PROPOSAL_STATUS_PASSED
* **Type:** /cosmos.gov.v1beta1.TextProposal

Proposal to complete the Stargate upgrade, halt `cosmoshub-3` at 06:00 UTC on Jan 28th, export the state and start `cosmoshub-4` based on gaia 3.0.nn Gaia Commit hash: n d974b27a8caf8cad3b06fbe4678871e4b0b69a51 Proposal details can be found on n github: https://github.com/cosmos/governance/pull/5 n ipfs: https://cloudflare-ipfs.com/ipfs/QmPww2PSmkmuLLu12GGwRdu5ur1Etf9u3Nt3Z6NqB7BQP1 n sia: https://siasky.net/EAALGMzFCafvbKkQjnAieo2cA1mpxk-JLpKsiC4XxuM6eQ

## 36 Delay of Hub Stargate Upgrade for approximately 2 weeks

* **Submitted:** 2021-01-24T15:51:52.051468824Z
* **Status:** PROPOSAL_STATUS_PASSED
* **Type:** /cosmos.gov.v1beta1.TextProposal

The Stargate team is recommending that the Cosmos Hub reschedule the next upgrade to a new commit hash. The new commit hash is expected to be available on Tuesday Jan 26th with a new upgrade proposal immediately after.nnThis governance proposal will signal that [proposal 35](https://www.mintscan.io/cosmos/proposals/35) will not be executed. The Hub governance will vote on the forthcoming proposal aiming for a final upgrade. The earliest target date would be February 11th. Given that Lunar New Year is on Feb 12th. The next best date is Feb 18th 06:00UTC.nnWe are recommending the delay for the following reasons.nn* Bugs have been identified in the Proposal 29 implementation. They are resolved in this pull request[Additional review of prop 29 and migration testing by zmanian · Pull Request #559 · cosmos/gaia · GitHub](https://github.com/cosmos/gaia/pull/559)n* A balance validation regression was identified during Prop 29 code review. [x/bank: balance and metadata validation by fedekunze · Pull Request #8417 · cosmos/cosmos-sdk · GitHub](https://github.com/cosmos/cosmos-sdk/pull/8417)n* The IBC Go To Market Working Group has [identified Ledger hardware wallet](https://github.com/cosmos/cosmos-sdk/issues/8266) support as a necessary feature for the initial launch of IBC on the Hub. We have an opportunity to provide this support in this upgrade. The SDK believes this can be quickly remediated in the time available with merged PRs on Monday.n* The number of Stargate related support requests from integrators has increased significantly since the governance proposal went live but some teams have already announced a period of reduced $ATOM support while they upgrade like <https://twitter.com/Ledger_Support/status/1352247403605356551?s=20>. The additional time should minimize the disruption for $ATOM holders. Thank so much to the $IRIS team whom is fielding a similar request volume among our non-English community.

## 37 Stargate Upgrade- Second time is a charm!

* **Submitted:** 2021-01-28T21:07:30.044676129Z
* **Status:** PROPOSAL_STATUS_PASSED
* **Type:** /cosmos.gov.v1beta1.TextProposal

Proposal to complete the Stargate upgrade, halt `cosmoshub-3` at 06:00 UTC on Feb 18th, export the state and start `cosmoshub-4` based on gaia 4.0.nn Gaia Commit hash: n a279d091c6f66f8a91c87943139ebaecdd84f689 Proposal details can be found on n github: https://github.com/cosmos/governance/pull/13 n Rendered: https://ipfs.io/ipfs/QmTkzDwWqPbnAh5YiV5VwcTLnGdwSNsNTn2aDxdXBFca7D/example#/ipfs/QmYn2SxCMYk5SWs5GMcXdbXR8wMCCXRmCyW19SFyzSeZp1 n ipfs: https://cloudflare-ipfs.com/ipfs/QmYn2SxCMYk5SWs5GMcXdbXR8wMCCXRmCyW19SFyzSeZp1 n sia: https://siasky.net/EACAsPcUjpTEpQlG9_nRI1OR07gNeRiudfEWAvKnf0tj_Q n 

