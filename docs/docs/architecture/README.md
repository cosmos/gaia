---
title: Architecture Decision Records (ADR)
order: 1
---

This is a location to record all high-level architecture decisions for new feature and module proposals in the Cosmos Hub.

An Architectural Decision (**AD**) is a software design choice that addresses a functional or non-functional requirement that is architecturally significant.
An Architecturally Significant Requirement (**ASR**) is a requirement that has a measurable effect on a software system’s architecture and quality.
An Architectural Decision Record (**ADR**) captures a single AD, such as often done when writing personal notes or meeting minutes; the collection of ADRs created and maintained in a project constitute its decision log. All these are within the topic of Architectural Knowledge Management (AKM).

You can read more about the ADR concept [here](https://adr.github.io/).

## Rationale

ADRs are intended to be the primary mechanism for proposing new feature designs and new processes, for collecting community input on an issue, and for documenting the design decisions.
An ADR should provide:

- Context on the relevant goals and the current state
- Proposed changes to achieve the goals
- Summary of pros and cons
- Discarded solution spaces and why they were discarded
- References
- Changelog

Note the distinction between an ADR and a spec. The ADR provides the context, intuition, reasoning, and
justification for a change in architecture, or for the architecture of something
new. The spec is much more compressed and streamlined summary of everything as
it stands today.

If recorded decisions turn out to be lacking, convene a discussion, record the new decisions here, and then modify the code to match.

## Creating new ADR

Read about the [PROCESS](./PROCESS.md).

### Use RFC 2119 Keywords

When writing ADRs, follow the same best practices for writing RFCs.
When writing RFCs, key words are used to signify the requirements in the specification.
These words are often capitalized: "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL.
They are to be interpreted as described in [RFC 2119](https://datatracker.ietf.org/doc/html/rfc2119).

## ADR Table of Contents

### Accepted

- n/a

### Proposed

- n/a

### Draft

- n/a

### Rejected

- [ADR 001: Interchain Accounts](./adr/adr-001-interchain-accounts.md)

### Deprecated

- [ADR 002: Globalfee Module](./adr/adr-002-globalfee.md)