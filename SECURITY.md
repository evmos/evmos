# Security

As part of our vulnerability disclosure policy. This document serves as a complementary guideline
for reporting vulnerabilities and how the disclosure process is managed.

## Guidelines

We require that all whitehat hackers and researchers:

- Use the Evmos security email ([security@evmos.org](mailto:security@evmos.org)) to disclose all vulnerabilities,
and avoid posting vulnerability information in public places, including GitHub, Discord, Telegram, X (Twitter) or
other non-private channels.
- Make every effort to avoid privacy violations, degradation of user experience, disruption to production systems,
and destruction of data.
- Keep any information about vulnerabilities that you’ve discovered confidential between yourself and the engineering
team until the issue has been resolved and disclosed
- Avoid posting personally identifiable information, privately or publicly

If you follow these guidelines when reporting an issue to us, we commit to:

- Not pursue or support any legal action related to your research on this vulnerability
- Work with you to understand, resolve and ultimately disclose the issue in a timely fashion

## Disclosure Process

Evmos uses the following disclosure process:

1. Once a security report is received via the security email, the team works to verify the issue and confirm its
severity level using [CVSS](https://nvd.nist.gov/vuln-metrics/cvss) in its latest version (v4 at the time of writing).
    1. Two people from the affected project will review, replicate and acknowledge the report
       within 48-96 hours of the alert according to the table below:

        | Security Level       | Hours to First Response (ACK) from Escalation |
        | -------------------- | --------------------------------------------- |
        | Critical             | 48                                            |
        | High                 | 96                                            |
        | Medium               | 96                                            |
        | Low or Informational | 96                                            |
        | None                 | 96                                            |

    2. If the report is not applicable or the vulnerability is not able to be reproduced,
       the Security Lead will revert to the reporter to request more info or close the report.
    3. The report is confirmed by the Security Lead to the reporter.

2. The team determines the vulnerability’s potential impact on Evmos.

    1. Vulnerabilities with `Informational` and `Low` categorization will result in creating a public issue.
    2. Vulnerabilities with `Medium` categorization will result
       in the creation of an internal ticket and patch of the code.
    3. Vulnerabilities with `High` or `Critical` will result in the [creation of a new Security Advisory](https://docs.github.com/en/code-security/repository-security-advisories/creating-a-repository-security-advisory)

Once the vulnerability severity is defined, the following steps apply:

- For `High` and `Critical`:
    1. Patches are prepared for supported releases of Evmos in a
       [temporary private fork](https://docs.github.com/en/code-security/repository-security-advisories/collaborating-in-a-temporary-private-fork-to-resolve-a-repository-security-vulnerability)
       of the repository.
    2. Only relevant parties will be notified about an upcoming upgrade.
       These being validators, the core developer team, and users directly affected by the vulnerability.
    3. 24 hours following this notification, relevant releases with the patch will be made public.
    4. The nodes and validators update their Evmos and Ethermint dependencies to use these releases.
    5. A week (or less) after the security vulnerability has been patched on Evmos,
       we will disclose that the mentioned release contained a security fix.
    6. After an additional 2 weeks, we will publish a public announcement of the vulnerability.
       We also publish a security Advisory on GitHub and publish a
       [CVE](https://en.wikipedia.org/wiki/Common_Vulnerabilities_and_Exposures)

- For `Informational` , `Low` and `Medium` severities:
    1. `Medium` and `Low` severity bug reports are included in a public issue
       and will be incorporated in the current sprint and patched in the next release.
       `Informational` reports are additionally categorized as with low or medium priority
       and might not be included in the next release.
    2. One week after the releases go out, we will publish a post
       with further details on the vulnerability as well as our response to it.

This process can take some time.
Every effort will be made to handle the bug in as timely a manner as possible,
however, it's important that we follow the process described above
to ensure that disclosures are handled consistently
and to keep Ethermint and its downstream dependent projects,
including but not limited to Evmos,
as secure as possible.

### Payment Process

The payment process will be executed according to Evmos SAFU for `Critical` and `High` severity vulnerabilities.
Payouts can only be executed in accordance and under supervision of the Evmos Operations team and only once the
following requirements have been completed:

- The whitehat hacker or organization successfully completes the KYC/KYB process (i.e KYC/KYB accepted).
- The vulnerability is patched in production (eg. mainnet).

#### KYC/KYB Process

The Operations team will get in contact with the whitehat hacker to coordinate the submission of KYC/KYC with
the Service Provider [Provenance](http://provenancecompliance.com).

The KYC/KYB process is performed independently by the Service Provider, which submits a report with the
KYC/KYB result
(Accepted or Rejected) to the Evmos Core Team. The Evmos Core team does not have access to any of the information
provided to the Service Provider.

The following information is to be submitted to the independent Service Provider:

- **Email**
- **Physical Address**
- **Proof of Address**: Utility bill (with exception of mobile phone invoice) or bank statement with no
more than 3 months old from the current date.
- **Passport** (National Identification) + Selfie photo.
- **Receiving Address**: The on-chain address account that will receive the Payouts.

#### Supported Releases

The team commits to releasing security patch releases for the latest release that Evmos is running.

If evmOS licensees are running older versions, we encourage them to upgrade at the earliest opportunity
so that you can receive
security patches directly from the repo, according to the terms set in the License Agreement. While project
are welcomed to backport security patches to older versions for their own use, the Evmos team reserves
the right to prioritize patches for
latest versions being used by projects.

#### Scope of Vulnerabilities

We’re interested in a full range of bugs with demonstrable security risk: from those that can be proven
with a simple unit test,
to those that require a full cluster and a complex sequence of transactions.

Please note that, in the interest of the safety of our users and staff, a few things are explicitly
excluded from scope:

- Any third-party services.
- Findings derived from social engineering (e.g., phishing).

Examples of vulnerabilities that are of interest to us include memory allocation bugs, race conditions,
timing attacks,information leaks, authentication bypasses, denial of service
(specifically at the application- or protocol-layer),
lost-write bugs, unauthorized account or capability access, stolen or loss of funds, token inflation bugs,
payloads/transactions that cause panics, non deterministic logic, etc.

##### JSON-RPC

- Write-access to anything besides sending transactions
- Bypassing transactions authentication
- Denial-of-Service
- Leakage of secrets

##### Denial-of-Service

Attacks may come through the P2P network or the RPC layer:

- Amplification attacks
- Resource abuse
- Deadlocks and race conditions

##### Precompiles

- Override of state due to misuse of `DELEGATECALL`, `STATICCALL`, `CALLCODE`
- Unauthorized transactions via precompiles (eg. ERC-20 token approvals)

##### EVM Module

- Memory allocation bugs
- Payloads that cause panics
- Authorization of invalid transactions

##### Fee Market Module (EIP-1559)

- Memory allocation bugs
- Improper / unpenalized manipulation of the BaseFee value

### Contact

The Evmos Security Team is constantly being monitored.
If you need to reach out to the team directly,
please reach out via email: [security@evmos.org](mailto:security@evmos.org)
