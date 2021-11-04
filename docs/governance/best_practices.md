# Best Practices for Drafting a Proposal 

There are currently three types of proposals supported by the Cosmos Hub: 
- [Community Pool Spend](/community-pool-spend) - Proposal to spend funds from the community pool on
  an important project.
- [Parameter Change](/params-change) - Proposal to change a core on-chain parameter.
- [Text](/text) - Proposal to agree to a certain strategy, plan, commitment, future
  upgrade or other statement. Text proposals are exclusively a signalling mechanism and focal point for future coordination - 
  they do not directly cause any changes.

You'll first want to determine which kind of proposal you are making. Be sure to
review all details of your specific proposal type. What follows below are
general best practices, regardless of proposal type.

## Engage directly with the voting community and seek feedback

Engagement is likely to be critical to the success of a proposal. 

The degree to which you engage with the Cosmos Hub community should be relative to the potential impact that your proposal may have on the stakeholders. 

There are many different ways to engage. One strategy involves a few stages of engagement before and after submitting a proposal on chain. **Why do it in stages?** It's a more conservative approach to save resources. The idea is to check in with key stakeholders at each stage before investing more resources into developing your proposal.

In the first stage of this strategy, you should engage people (ideally experts) informally about your idea.
- Does it make sense? 
- Are there critical flaws? 
- Does it need to be reconsidered? 

If you're already confident about your idea, [skip to Stage 2](#stage-2-your-draft-proposal).

**Note**: this guide likely fails to capture all ways of engaging. Perhaps you could bring your idea to a podcast or a hackathon. You could host an AMA on [Reddit](https://www.reddit.com/r/cosmosnetwork) or host a Q&A (questions & answers) video call. Try to go above and beyond what's recommended here--experiment, and use your strengths and connections.

## Stage 1: Your Idea

### Not yet confident about your idea?
Great! Governance proposals potentially impact many stakeholders. Introduce your idea with known members of the community before investing resources into drafting a proposal. Don't let negative feedback dissuade you from exploring your idea if you think that it's still important. 

If you know people who are very involved with the Cosmos Hub, send them a private message with a concise overview of what you think will result from your idea or proposed changes. Wait for them to ask questions before providing details. Do the same in semi-private channels where people tend to be respectful (and hopefully supportive). I recommend [this Cosmos Discord community][discord] and the private Cosmos Network VIP Telegram channel (ask for an invite [on the forum][forum] if you are or would like to be a Cosmos contributor).

### Confident with your idea?
Great! However, remember that governance proposals potentially impact many stakeholders, which can happen in unexpected ways. Introduce your idea with members of the community before investing resources into drafting a proposal. At this point you should seek out and carefully consider critical feedback in order to protect yourself from [confirmation bias](https://en.wikipedia.org/wiki/Confirmation_bias). This is the ideal time to see a critical flaw, because submitting a flawed proposal will waste resources.

### Are you ready to draft a governance proposal?
There will likely be differences of opinion about the value of what you're proposing to do and the strategy by which you're planning to do it. If you've considered feedback from broad perspectives and think that what you're doing is valuable and that your strategy should work, and you believe that others feel this way as well, it's likely worth drafting a proposal. However, remember that the largest ATOM stakers have the biggest vote, so a vocal minority isn't necessarily representative or predictive of the outcome of an on-chain vote. 

A conservative approach is to have some confidence that you roughly have initial support from a majority of the voting power before proceeding to drafting your proposal. However, there are likely other approaches, and if your idea is important enough, you may want to pursue it regardless of whether or not you are confident that the voting power will support it.

## Stage 2: Your Draft Proposal

### Begin with a well-considered draft proposal
The next major section outlines and describes some potential elements of drafting a proposal. Ensure that you have considered your proposal and anticipated questions that the community will likely ask. Once your proposal is on-chain, you will not be able to change it.

The ideal format for a proposal is as a Markdown file (ie. `.md`) in a github repo. Markdown
is a simple and accessible format for writing plain text files that is easy to
learn. See the [Github Markdown
Guide](https://guides.github.com/features/mastering-markdown/) for details on
writing markdown files.

If you don't have a [Github](http://github.com/) account already, register one. Then fork this
repository, draft your proposal in the `proposals` directory, and make a
pull-request back to this repository. For more details on using Github, see the
[Github Forking Guide](https://guides.github.com/activities/forking/). If you
need help using Github, don't be afraid to ask someone!

If you really don't want to deal with Github, you can always draft a proposal in
Word or Google Docs, or directly in the forums, or otherwise. However Markdown
on Github is the ultimate standard for distributed collaboration on text files.

### Engage the community with your draft proposal
1. Post a draft of your proposal as a topic in the 'governance' category of the [Cosmos forum][forum]. Ideally this should contain a link to this repository, either directly to your proposal if it has been merged, or else to a pull-request containing your proposal if it has not been merged yet.
2. Directly engage key members of the community for feedback. These could be large contributors, those likely to be most impacted by the proposal, and entities with high stake-backing (eg. high-ranked validators; large stakers).
3. Engage with the Cosmos Governance Working Group (GWG). These are people focused on Cosmos governance--they won't write your proposal, but will provide feedback and recommend resources to support your work. Members can be contacted on the [forum][forum] (they use the tag 'GWG' in posts), in [Telegram](https://t.me/hubgov), and on [Discord][discord].
4. Target members of the community in a semi-public way before bringing the draft to a full public audience. The burden of public scrutiny in a semi-anonymized environment (eg. Twitter) can be stressful and overwhelming without establishing support. Solicit opinions in places with people who have established reputations first. For example, there is a private Telegram group called Cosmos Network VIP (ask for an invite [on the forum][forum] if you are or would like to be a Cosmos contributor). Let people in the [Discord community][discord] know about your draft proposal.
5. Alert the entire community to the draft proposal via
   - Twitter, tagging accounts such as the All in Bits [Cosmos account](https://twitter.com/cosmos), the [Cosmos GWG](https://twitter.com/CosmosGov), and Today in Cosmos [@adriana_kalpa](https://twitter.com/adriana_kalpa)
   - [Telegram](https://t.me/cosmosproject), [Adriana](https://t.me/adriana_KalpaTech) (All in Bits)
   - [Discord][discord]

### Submit your proposal to the testnet

I intend to expand this [guide to include testnet instructions](/submitting.md#submitting-your-proposal-to-the-testnet). 

You may want to submit your proposal to the testnet chain before the mainnet for a number of reasons, such as wanting to see what the proposal description will look like, to share what the proposal will look like in advance with stakeholders, and to signal that your proposal is about to go live on the mainnet.

Perhaps most importantly, for parameter change proposals, you can test the parameter changes in advance (if you have enough support from the voting power on the testnet).

Submitting your proposal to the testnet increases the likelihood of engagement and the possibility that you will be alerted to a flaw before deploying your proposal to mainnet.

## Stage 3: Your On-Chain Proposal

A majority of the voting community should probably be aware of the proposal and have considered it before the proposal goes live on-chain. If you're taking a conservative approach, you should have reasonable confidence that your proposal will pass before risking deposit contributions. Make revisions to your draft proposal after each stage of engagement.

See the [submitting guide](/submitting.md) for more on submitting proposals.

### The Deposit Period
The deposit period currently lasts 14 days. If you submitted your transaction with the minimum deposit (512 ATOM), your proposal will immediately enter the voting period. If you didn't submit the minimum deposit amount (currently 512 ATOM), then this may be an opportunity for others to show their support by contributing (and risking) their ATOMs as a bond for your proposal. You can request contributions openly and also contact stakeholders directly (particularly stakeholders who are enthusiastic about your proposal). Remember that each contributor is risking their funds, and you can [read more about the conditions for burning deposits here](/overview.md#burned-deposits).

This is a stage where proposals may begin to get broader attention. Most popular explorers currently display proposals that are in the deposit period, but due to proposal spamming, this may change. [Hubble](https://hubble.figment.network/cosmos/chains/cosmoshub-3/governance), for example, only displays proposals that have 10% or more of the minimum deposit, so 51.2 ATOM or more.

A large cross-section of the blockchain/cryptocurrency community exists on Twitter. Having your proposal in the deposit period is a good time to engage the so-called 'crypto Twitter' Cosmos community to prepare validators to vote (eg. tag [@cosmosvalidator](https://twitter.com/cosmosvalidator)) and ATOM-holders that are staking (eg. tag [@cosmos](https://twitter.com/cosmos), [@adriana_kalpa](https://twitter.com/adriana_kalpa)). 

### The Voting Period
At this point you'll want to track which validator has voted and which has not. You'll want to re-engage directly with top stake-holders, ie. the highest-ranking validator operators, to ensure that:
1. they are aware of your proposal;
2. they can ask you any questions about your proposal; and
3. they are prepared to vote.

Remember that any voter may change their vote at any time before the voting period ends. That historically doesn't happen often, but there may be an opportunity to convince a voter to change their vote. The biggest risk is that stakeholders won't vote at all (for a number of reasons). Validator operators tend to need multiple reminders to vote. How you choose to contact validator operators, how often, and what you say is up to you--remember that no validator is obligated to vote, and that operators are likely occupied by competing demands for their attention. Take care not to stress any potential relationship with validator operators.

   [discord]: https://discord.gg/W8trcGV
   [forum]: https://forum.cosmos.network/c/governance
