# Security

This security document is the process that the Gaia team follows regarding security issues 
that can impact partners and users of Gaia and the Cosmos Hub network. 

## Prerequisites

If a vulnerability and its exploit are both publicly known, the security process may not apply. 
However, in such cases, resolutions and mitigation strategies may still be eligible for rewards through a bounty program.

## Reporting a Bug

As part of our [Coordinated Vulnerability Disclosure
Policy](https://tendermint.com/security), we operate a [bug
bounty](https://hackerone.com/tendermint).
See the policy for more details on submissions and rewards, and see "Example Vulnerabilities" (below) for examples of the kinds of bugs we're most interested in.

### Guidelines

We require that all researchers:

* Use the bug bounty to disclose all vulnerabilities, and avoid posting vulnerability information in public places, including Github, Discord, Telegram, and Twitter
* Make every effort to avoid privacy violations, degradation of user experience, disruption to production systems (including but not limited to the Cosmos Hub), and destruction of data
* Keep any information about vulnerabilities that you’ve discovered confidential between yourself and the Gaia engineering team until the issue has been resolved and disclosed
* Avoid posting personally identifiable information, privately or publicly

If you follow these guidelines when reporting an issue to us, we commit to:

* Not pursue or support any legal action related to your research on this vulnerability
* Work with you to understand, resolve and ultimately disclose the issue in a timely fashion

## Disclosure Process

Gaia uses the following disclosure process:

1. Once a security report is received, the Gaia team works to verify the issue and confirm its severity level using CVSS.
2. The team collaborates with the Cosmos SDK and Tendermint teams to determine the vulnerability’s potential impact on the Cosmos Hub.
3. Patches are prepared for eligible releases of Gaia in private repositories. See “Supported Releases” below for more information on which releases are considered eligible.
4. If it is determined that a CVE-ID is required, we request a CVE through a CVE Numbering Authority.
5. We provide the community with a 24 hour notice that a security release is impending, giving partners time to prepare their systems for the update. Notifications can include Discord, Telegram, Twitter, forum posts, and emails, including emails sent to the [Tendermint Security Mailing List](https://berlin.us4.list-manage.com/subscribe?u=431b35421ff7edcc77df5df10&id=3fe93307bc).
6. 24 hours following this notification, the fixes are applied publicly and new releases are issued.
7. Once new security releases are available, we notify the community, through the same channels as above. <!-- We also publish a Security Advisory on Github and publish the CVE, as long as neither the Security Advisory nor the CVE include any information on how to exploit these vulnerabilities beyond what information is already available in the patch itself. -->
8. Once the community is notified, we will pay out any relevant bug bounties to submitters.
9. Approximately one week after the releases go out, we will publish a post with further details on the vulnerability as well as our response to it. This timeline may be adjusted depending on the severity of the issue and the need to inform and update partner Cosmos zones.

This process can take some time. Every effort will be made to handle the bug in as timely a manner as possible, however it's important that we follow the process described above to ensure that disclosures are handled consistently and to keep Gaia and its partner projects--including but not limited to the Cosmos Hub--as secure as possible.

### Disclosure Communications

Communications to Cosmos Hub Validators will include the following details:
1. Affected version(s)
1. New release version
1. Impact on user funds
1. For timed releases, a date and time that the new release will be made available
1. Impact on the hub if upgrades are not completed in a timely manner
1. Potential actions to take if an adverse condition arises during the security release process

An example notice looks like:
```
Dear Cosmos Hub Validators,

A critical security vulnerability has been identified in Gaia v4.0.x. 
User funds are NOT at risk; however, the vulnerability can result in a chain halt.

This notice is to inform you that on [[**March 1 at 1pm EST/6pm UTC**]], we will be releasing Gaia v4.1.x, which patches the security issue. 
We ask all validators to upgrade their nodes ASAP.

If the chain halts, validators with sufficient voting power need to upgrade and come online in order for the chain to resume.
```

### Example Timeline

The following is an example timeline for the triage and response. The required roles and team members are described in parentheses after each task; however, multiple people can play each role and each person may play multiple roles.

#### 24+ Hours Before Release Time

1. Request CVE number (ADMIN)
2. Gather emails and other contact info for validators (COMMS LEAD)
3. Test fixes on a testnet  (GAIA ENG, TENDERMINT ENG, COSMOS SDK ENG)
4. Write “Security Advisory” for forum (GAIA LEAD)

#### 24 Hours Before Release Time

1. Post “Security Advisory” pre-notification on forum (GAIA LEAD)
2. Post Tweet linking to forum post (COMMS LEAD)
3. Announce security advisory/link to post in various other social channels (Telegram, Discord) (COMMS LEAD)
4. Send emails to validators or other users (PARTNERSHIPS LEAD)

#### Release Time

1. Cut Tendermint releases for eligible versions (TENDERMINT ENG)
2. Cut Cosmos SDK release for eligible versions (COSMOS SDK ENG)
3. Cut Gaia release for eligible versions (GAIA ENG)
4. Post “Security releases” on forum (GAIA LEAD)
5. Post new Tweet linking to forum post (COMMS LEAD)
6. Remind everyone via social channels (Telegram, Discord)  that the release is out (COMMS LEAD)
7. Send emails to validators or other users (COMMS LEAD)
8. Publish Security Advisory and CVE, if CVE has no sensitive information (ADMIN)

#### After Release Time

1. Write forum post with exploit details (GAIA LEAD)
2. Approve pay-out on HackerOne for submitter (ADMIN)

#### 7 Days After Release Time

1. Publish CVE if it has not yet been published (ADMIN)
2. Publish forum post with exploit details (GAIA ENG, GAIA LEAD)

## Supported Releases

The Gaia team commits to releasing security patch releases for both the latest minor release as well for the major/minor release that the Cosmos Hub is running.

If you are running older versions of Gaia, we encourage you to upgrade at your earliest opportunity so that you can receive security patches directly from the Gaia repo. While you are welcome to backport security patches to older versions for your own use, we will not publish or promote these backports.

## Scope

The full scope of our bug bounty program is outlined on our [Hacker One program page](https://hackerone.com/tendermint). Please also note that, in the interest of the safety of our users and staff, a few things are explicitly excluded from scope:

* Any third-party services
* Findings from physical testing, such as office access
* Findings derived from social engineering (e.g., phishing)

## Example Vulnerabilities

The following is a list of examples of the kinds of vulnerabilities that we’re most interested in. It is not exhaustive: there are other kinds of issues we may also be interested in!

* Injection exploits
* Privilege escalation
* IBC
* Inter-module interactions
* Web exploits
* Network channel attacks
* Replay attacks

Future vulnerabilities could include:

* Zero-knowledge libraries
* Dependencies
