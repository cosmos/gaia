---
title: Off-Chain Proposal Process
order: 3
---

Once a proposal is on-chain, it cannot be changed to reflect feedback or new information. It's very important to give a proposal time off-chain to receive feedback, input, and edits before going on-chain and asking for votes. 

The process of passing a proposal starts long before it goes on-chain!

There are currently several types of proposals supported by the Cosmos Hub: 
- **[Text](./proposal-types/text-prop.md)** - Proposal to agree to a certain strategy, plan, commitment, future upgrade or other statement. Text proposals do not directly cause any changes, but they can be used to take a record of the community's opinion or commitment to a future idea.
- [**Community Pool Spend**](./proposal-types/community-pool-spend.md) - Proposal to spend funds from the community pool on a project.
- [**Parameter Change**](./proposal-types/param-change.md) - Proposal to change a core on-chain parameter.
- **Software Upgrade** - Proposal to upgrade the chain version.
- **IBC Client Update** - Proposal to update an IBC client.

You'll first want to determine which kind of proposal you are making. Be sure to review all details of your specific proposal type. 

## Engage directly with the voting community and seek feedback

Engagement is likely to be critical to the success of a proposal. The degree to which you engage with the Cosmos Hub community should be relative to the potential impact that your proposal may have on the stakeholders. This guide does not cover all ways of engaging but here are some suggestions: 
 - Post your idea to the [Cosmos Hub Forum](https://forum.cosmos.network/)
<!-- markdown-link-check-disable-next-line -->
 - Mention the idea in a community call (often hosted on [Twitter](https://twitter.com/CosmosHub))
 - Host an AMA on [Reddit](https://www.reddit.com/r/cosmosnetwork) 
 
 We encourage you to experiment and use your strengths to introduce proposal ideas and gather feedback.

There are many different ways to engage. One strategy involves a few stages of engagement before and after submitting a proposal on chain. 

**Why do it in stages?** It's a more conservative approach to save resources. The idea is to check in with key stakeholders at each stage before investing more resources into developing your proposal.

In the first stage of this strategy, you should engage people (ideally experts) informally about your idea. You'll want to start with the minimal, critical components (name, value to Cosmos Hub, timeline, any funding needs) and check:
- Does it make sense? 
- Are there critical flaws? 
- How will this affect other projects or properties of the Hub? 

You should be engaging with key stakeholders (e.g., a large validator operator) with a few short sentences to measure their support. Here's an example:

"We are considering a proposal for funding to work on `project`. We think it will help the Hub to `outcome`. Timeline is `x`, and we're asking for `y` amount. Do you think that this is a proposal that `large validator` may support?"

**Why a large validator?** They tend to be the de facto decision-makers on the Cosmos Hub, since their delegators also delegate their voting power. If you can establish a base layer of off-chain support, you can be more confident that it's worth proceeding to the next stage.

**Note:** Many validators will likely hesitate to commit support, and that's okay. It will be important to reassure these stakeholders that this isn't a binding commitment. You're just canvasing the community to get a feel for whether it's worthwhile to proceed. It's also an opportunity to connect with new people and to answer their questions about what it is you're working on. It will be important for them to clearly understand why you think what you're proposing will be valuable to the Cosmos Hub, and if possible, why it will be valuable to them as long-term stakeholders.

If you're already confident about your idea, [skip to Stage 2](#stage-2-your-draft-proposal).

## Stage 1: Your Idea

### Not yet confident about your idea?

Great! Governance proposals potentially impact many stakeholders. Introduce your idea with known members of the community before investing resources into drafting a proposal. Don't let negative feedback dissuade you from exploring your idea if you think that it's still important. 

If you know people who are very involved with the Cosmos Hub, send them a private message with a concise overview of what you think will result from your idea or proposed changes. Wait for them to ask questions before providing details. Do the same in semi-private channels where people tend to be respectful (and hopefully supportive). 


### Confident with your idea?

Great! However, remember that governance proposals potentially impact many stakeholders, which can happen in unexpected ways. Introduce your idea with members of the community before investing resources into drafting a proposal. At this point you should seek out and carefully consider critical feedback in order to protect yourself from [confirmation bias](https://en.wikipedia.org/wiki/Confirmation_bias). This is the ideal time to see a critical flaw, because submitting a flawed proposal on-chain will waste resources and have reputational costs.

Posting your idea to the [Cosmos Hub Forum](https://forum.cosmos.network/) is a great way to get broad feedback and perspective even if you don't have personal connections to any stakeholders or involved parties.

### Are you ready to draft a governance proposal?

There will likely be differences of opinion about the value of what you're proposing to do and the strategy by which you're planning to do it. If you've considered feedback from broad perspectives and think that what you're doing is valuable and that your strategy should work, and you believe that others feel this way as well, it's likely worth drafting a proposal. However, remember that the largest ATOM stakers have the biggest vote, so a vocal minority isn't necessarily representative or predictive of the outcome of an on-chain vote. 

You could choose to take a conservative approach and wait until you have some confidence that you roughly have initial support from a majority of the voting power before proceeding to drafting the details of your proposal. Or you could propose the idea, or define the problem statement and let the community participate freely in drafting competing solutions to solve the issue.


## Stage 2: Your Draft Proposal

The next major section outlines and describes some potential elements of drafting a proposal. Ensure that you have considered your proposal and anticipated questions that the community will likely ask. **Once your proposal is on-chain, you will not be able to change it.**

### Proposal Elements 

It will be important to balance two things: being detailed and being concise. You'll want to be concise so that people can assess your proposal quickly. You'll want to be detailed so that voters will have a clear, meaningful understanding of what the changes are and how they are likely to be impacted.

Each major proposal type has a rough template available on the forum: [Text](https://forum.cosmos.network/t/about-the-signaling-text-category/5947), [community pool spend](https://forum.cosmos.network/t/about-the-community-spend-category/5949), [parameter change](https://forum.cosmos.network/t/about-the-parameter-change-category/5950), [software upgrade](https://forum.cosmos.network/t/about-the-software-upgrade-category/5951).

Each proposal should contain a summary with key details about what the proposal hopes to change. If you were viewing only the summary with no other context, it should be a good start to being able to make a decision.

Assume that many people will stop reading at this point. However it is important to provide in-depth information. The on-chain proposal text should also include a link to an un-editable version of the text, such as an IPFS pin, and a link to where discussion about the idea is happening.

A few more pointers for Parameter-change and Community Spend proposals are below.

#### Parameter-Change
An example of a successful parameter change proposal is [Proposal #66](https://forum.cosmos.network/t/proposal-66-accepted-increase-active-validator-spots-to-175/6118/53). Note that this proposal went on-chain without the recommended IPFS pin.

1. Problem/Value - The problem or value that's motivating the parameter change(s).
1. Solution - How changing the parameter(s) will address the problem or improve the network.
1. Risks & Benefits - How making this/these change(s) may expose stakeholders to new benefits and/or risks.
   - The beneficiaries of the change(s) (ie. who will these changes impact and how?)
   - Voters should understand the importance of the change(s) in a simple way
1. Supplementary materials - Optional materials eg. models, graphs, tables, research, signed petition, etc


#### Community-Spend Proposal
An example of a successful community spend proposal is [Proposal #63](https://forum.cosmos.network/t/proposal-63-accepted-activate-governance-discussions-on-the-discourse-forum-using-community-pool-funds/5833).

1. Applicant(s) - The profile of the person(s)/entity making the proposal.
   - Who you are and your involvement in Cosmos and/or other blockchain networks.
   - An overview of team members involved and their relevant experience.
1. Problem - What you're solving and/or opportunity you're addressing.
   - Past, present (and possibly a prediction of the future without this work being done).
1. Solution - How you're proposing to deliver the solution.
   - Your plan to fix the problem or deliver value.
   - The beneficiaries of this plan (ie. who will your plan impact and how?).
   - Your reasons for selecting this plan.
   - Your motivation for delivering this solution/value.
1. Funding - amount and denomination proposed eg. 5000 ATOM.
   - The entity controlling the account receiving the funding.
   - Consider an itemized breakdown of funding per major deliverable.
   - Note that the 'budget' of a spend proposal is generally the easiest thing to criticize. If your budget is vague, consider explaining the reasons you're unable to give a detailed breakdown and be clear about what happens if you do not meet your budget.
1. Deliverables and timeline - the specifics of what you're delivering and how, and what to expect.
   - What are the specific deliverables? (be detailed).
   - When will each of these be delivered?
   - How will each of these be delivered?
   - What will happen if you do not deliver on time?
   - Do you have a plan to return the funds if you're under-budget or the project fails?
   - How will you be accountable to the Cosmos Hub stakeholders?
     - How will you communicate updates and how often?
     - How can the community observe your progress?
     - How can the community provide feedback?
   - How should the quality of deliverables be assessed? eg. metrics.
1. Relationships and disclosures.
   - Have you received or applied for grants or funding? for similar work? eg. from the Interchain Foundation.
   - How will you and/or your organization benefit?
   - Do you see this work continuing in the future and is there a plan?
   - What are the risks involved with this work?
   - Do you have conflicts of interest to declare?

### Begin with a well-considered draft proposal
Ideally, a proposal is first sent to the forum in Markdown format so that it can be further edited and available for comments. A changelog is a great tool so that people can see how the idea has developed over time and in response to feedback.

This Markdown-formatted post can eventually become the description text in a proposal sent on-chain.

### Engage the community with your draft proposal

1. Post a draft of your proposal as a topic in the appropriate category of the forum. [Hub Proposals](https://forum.cosmos.network/c/hub-proposals) is a catch-all if you are not sure where to post, but there are categories for all types of proposals.

2. Directly engage key members of the community for feedback. These could be large contributors, those likely to be most impacted by the proposal, and entities with high stake-backing (eg. high-ranked validators; large stakers).

<!-- markdown-link-check-disable-next-line -->
3. Alert the entire community to the draft proposal on other platforms such as Twitter, tagging accounts such as the [Cosmos Hub account](https://twitter.com/cosmoshub), the [Cosmos Governance account](https://twitter.com/CosmosGov), and other governance-focused groups.

### Submit your proposal to the testnet

Before going on mainnet, you can test your proposal on the [testnet](submitting.md#submitting-your-proposal-to-the-testnet). 

This is a great way to make sure your proposal looks the way you want and refine it before heading to mainnet.

## Stage 3: Your On-Chain Proposal

A majority of the voting community should probably be aware of the proposal and have considered it before the proposal goes live on-chain. If you're taking a conservative approach, you should have reasonable confidence that your proposal will pass before risking deposit contributions. Make revisions to your draft proposal after each stage of engagement.

See the [submitting guide](./submitting.md) for more on submitting proposals.

### The Deposit Period

The deposit period currently lasts 14 days. If you submitted your transaction with the minimum deposit (250 ATOM), your proposal will immediately enter the voting period. If you didn't submit the minimum deposit amount (currently 250 ATOM), then this may be an opportunity for others to show their support by contributing (and risking) their ATOMs as a bond for your proposal. You can request contributions openly and also contact stakeholders directly (particularly stakeholders who are enthusiastic about your proposal). Remember that each contributor is risking their funds, and you can [read more about the conditions for burning deposits here](./process.md#burned-deposits).

This is a stage where proposals may begin to get broader attention. Some block explorers display proposals in the deposit period, while others don't show them until they hit voting period.

<!-- markdown-link-check-disable-next-line -->
A large cross-section of the blockchain/cryptocurrency community exists on Twitter. Having your proposal in the deposit period is a good time to engage the so-called 'crypto Twitter' Cosmos community to prepare validators to vote (eg. tag [@cosmosvalidator](https://twitter.com/cosmosvalidator)) and ATOM-holders that are staking (eg. tag [@cosmoshub](https://twitter.com/cosmoshub), [@CosmosGov](https://twitter.com/CosmosGov)). 

### The Voting Period

At this point you'll want to track which validator has voted and which has not. You'll want to re-engage directly with top stake-holders, ie. the highest-ranking validator operators, to ensure that:
1. they are aware of your proposal;
2. they can ask you any questions about your proposal; and
3. they are prepared to vote.

Remember that any voter may change their vote at any time before the voting period ends. That historically doesn't happen often, but there may be an opportunity to convince a voter to change their vote. The biggest risk is that stakeholders won't vote at all (for a number of reasons). Validator operators tend to need multiple reminders to vote. How you choose to contact validator operators, how often, and what you say is up to you--remember that no validator is obligated to vote, and that operators are likely occupied by competing demands for their attention. Take care not to stress any potential relationship with validator operators.

   [forum]: https://forum.cosmos.network
