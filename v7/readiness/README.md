# Architecture Decision Records (ADR)

This is a location to record all high-level architecture decisions for new feature and module proposals in the Cosmos Hub.

An Architectural Decision (**AD**) is a software design choice that addresses a functional or non-functional requirement that is architecturally significant.
An Architecturally Significant Requirement (**ASR**) is a requirement that has a measurable effect on a software system’s architecture and quality.
An Architectural Decision Record (**ADR**) captures a single AD, such as often done when writing personal notes or meeting minutes; the collection of ADRs created and maintained in a project constitute its decision log. All these are within the topic of Architectural Knowledge Management (AKM).

You can read more about the ADR concept in this [blog post](https://product.reverb.com/documenting-architecture-decisions-the-reverb-way-a3563bb24bd0#.78xhdix6t).

## Rationale

ADRs are intended to be the primary mechanism for proposing new feature designs and new processes, for collecting community input on an issue, and for documenting the design decisions.
An ADR should provide:

- Context on the relevant goals and the current state
- Proposed changes to achieve the goals
- Summary of pros and cons
- References
- Changelog

Note the distinction between an ADR and a spec. The ADR provides the context, intuition, reasoning, and
justification for a change in architecture, or for the architecture of something
new. The spec is much more compressed and streamlined summary of everything as
it stands today.

If recorded decisions turned out to be lacking, convene a discussion, record the new decisions here, and then modify the code to match.

## Creating new ADR

### Process
1. Copy the `template.md` file. Use the following filename pattern: `adr-next_number-title.md`
2. Link the ADR in the related [feature epic](../../.github/ISSUE_TEMPLATE/module-readiness.md)
2. Create a draft Pull Request if you want to get early feedback.
3. Make sure the context and a solution is clear and well documented.
4. Add an entry to a list in the README file [Table of Contents](#ADR-Table-of-Contents).
5. Create a Pull Request to publish the ADR proposal.

### Life cycle

ADR creation is an **iterative** process. Rather than solving all decisions in a single PR, it's best to first understand the problem and then solicit feedback through Github Issues.

1. Every proposal should start with a new GitHub Issue and be linked to the corresponding Feature Epic. The Issue should contain a brief proposal summary.

2. Once the motivation is validated, a GitHub Pull Request (PR) is created with a new document based on the `template.md`.

3. An ADR doesn't have to arrive to `master` with an `accepted` status in a single PR. If the motivation is clear and the solution is sound, we should be able to merge it and keep a `proposed` status.

4. If a `proposed` ADR is merged, then it should clearly document outstanding issues in the Feature Epic.

5. The PR should always be merged. In the case of a faulty ADR, it's still preferable to merge it with a `rejected` status. The only time the ADR should not be merged is if the author abandons it.

6. Merged ADRs **should not** be pruned.

### Status

Status has two components:

```
{CONSENSUS STATUS} {IMPLEMENTATION STATUS}
```

IMPLEMENTATION STATUS is either `Implemented` or `Not Implemented`.

#### Consensus Status

```
DRAFT -> PROPOSED -> LAST CALL yyyy-mm-dd -> ACCEPTED | REJECTED -> SUPERSEDED by ADR-xxx
                  \        |
                   \       |
                    v      v
                     ABANDONED
```

+ `DRAFT`: [optional] an ADR which is work in progress, not being ready for a general review. This is to present an early work and get an early feedback in a Draft Pull Request form.
+ `PROPOSED`: an ADR covering a full solution architecture and still in the review - project stakeholders haven't reached an agreed yet.
+ `LAST CALL <date for the last call>`: [optional] clear notify that we are close to accept updates. Changing a status to `LAST CALL` means that social consensus (of Cosmos Hub maintainers) has been reached and we still want to give it a time to let the community react or analyze.
+ `ACCEPTED`: ADR which will represent a currently implemented or to be implemented architecture design.
+ `REJECTED`: ADR can go from PROPOSED or ACCEPTED to rejected if the consensus among project stakeholders will decide so.
+ `SUPERSEEDED by ADR-xxx`: ADR which has been superseded by a new ADR.
+ `ABANDONED`: the ADR is no longer pursued by the original authors.

### Language used in ADR

+ The context/background should be written in the present tense.
+ Avoid using a first, personal form.

**Use RFC 2119 Keywords**

When writing ADRs, follow the same best practices for writing RFCs. When writing RFCs, key words are used to signify the requirements in the specification. These words are often capitalized: "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL. They are to be interpreted as described in [RFC 2119](https://datatracker.ietf.org/doc/html/rfc2119).

## ADR Table of Contents

### Accepted

- [ADR 000: <Accepted Module or Feature>]()

### Proposed

- [ADR 001: <Proposed Module or Feature>]()


### Draft

- [ADR 002: <Draft Module or Feature>]()
