# License FAQ

Find below an overview of Permissions and Limitations of the
Evmos Non-Commercial License 1.0. For more information, check out the full ENCL-1.0 FAQ here.

### Permissions

- Private Use, including distribution and modification
- Commercial use on designated blockchains
- Commercial use with Evmos permit (to be separately negotiated)

### Prohibited

- Commercial use, other than on designated blockchains, without Evmos permit

Evmos Non-Commercial License 1.0 (ENCL-1.0) was created by Tharsis Labs Ltd. (Evmos) to provide a mutually beneficial
balance between the user benefits of our open software that is free of charge and provides open access to all of the
product code for modification, distribution, auditability, etc., and the sustainability needs of our software developers
to continue delivering product innovation and maintenance.
The ENCL is structured to allow free of charge usage in all non-commercial cases and limited commercial use cases.
ENCL gives users complete access to the source code so users can modify, distribute and enhance it,
within the permitted purposes.
This FAQ is designed to address questions for developers and companies interested in working on
ENCL Software or adopting ENCL Software for commercial use.

**Q: What is Evmos Non-Commercial License 1.0 (ENCL-1.0)?**

**A:** ENCL is an alternative to closed-source or fully open source licensing models. Our licensing model includes
both open source elements, under LGPL3 and source-available elements, under ENCL-1.0. Under ENCL-1.0,
the source code is always publicly available. You have the right to “use, modify, and redistribute the software”
for any non-commercial purpose, or for specific commercial purposes.

By implication, there is only one primary limitation. You MAY NOT make commercial use of the software,
with exceptions that the commercial use is on Designated Blockchains or you have obtained a prior commercial permit.

The right of “use” here includes forking and using ENCL code as
a code dependency as use of the dependency doesn’t violate the ENCL-1.0.

Here is the specific language of the license:

“Commercial Use” is any use that is not a Non-Commercial Use.

“Non-Commercial Use” means academic, scientific, or research and development use,
or evaluating the Software (such as. through auditing),
but does not include the creation of a publicly available blockchain, precompiled smart contracts, or other
distributed-ledger technology systems that facilitate any transaction of economic value.

"Designated Blockchains" refers to the version of the digital blockchain ledger that, at any given time,
is recognized as canonical in accordance with the blockchain consensus.The initial Designated Blockchains
shall be the Evmos blockchains,
identified by chain identifiers 9000 (testing network or testnet) and 9001 (main network or mainnet).

Apart from Evmos repository, Evmos is currently building the Evmos Software Development Kit (Evmos SDK),
to help developers to create their own EVM-compatible blockchain network with custom parameters.
Once it is publicly available, you may use the Evmos SDK instead for your commercial project,
subject to the applicable Evmos SDK license. For more information about Evmos SDK, check out the Evmos Manifesto.

**Q: What is the purpose of ENCL-1.0?**

**A:** To create a license that strikes a balance between being able to maintain sustainable
software development while still supporting the original tenets of open source,
such as empowering all non-Evmos software developers to be part of the innovation cycle
– giving them open access to the code so they can audit,
modify or distribute the software by making the entire source code available from the start.
Note that ENCL 1.0 has not been approved by the OSI, and we do not refer to it as an Open Source license.

**Q:How do I apply for a commercial use permit or obtain different licensing terms or inquire about terms of this license?**

**A:** You may contact the legal department of licensor: legal@thars.is.
They may be able to partner with you to answer your questions and figure out what will work best for you and your needs.
You only need this kind of permit if you cannot meet the limitations of ENCL-1.0.

**Q: What if I am currently using Evmos code commercially in my business/project?**

**A:**You are allowed to continue using the code from older versions (<v13.0.0) of Evmos repository under LGPL 3.0,
however, you must obtain a commercial permit from the licensor for commercial use not allowed under ENCL-1.0,
for version 13 and onward.

**Q: What is a License and Copyright Notice?**

**A:** When you distribute the software, you must include the license and copyright notice.
This obligation extends not only to you as the licensee, but also to anyone who receives a copy of the software from you.
This is necessary so that all recipients of the software understand its license terms.
Removing copyright or licensing notices is a serious problem and a violation of the license.
This means that you must ensure that anyone who gets a copy of the software from you also receives a copy of the
license terms or a link to the license, as well as any plain-text lines beginning with "Required Notice" that the
licensor provided with the software.

For example:

"Required Notice: Copyright Tharsis Labs Ltd. (Evmos)(https://github.com/evmos)"

**Q: What is the difference between LGPL v3 and ENCL-1.0?**

**A:** The main differences are:
ENCL 1.0 is not an Open Source license and Evmos does not claim it to be one.
The legal phrasing of the ENCL 1.0 has been reviewed and edited for consistency and simplicity by Heather Meeker,
author of Open (Source) For Business: A Practical Guide to Open Source Licensing,
Heather has been a pro-bono counsel to the Mozilla, GNOME, and Python foundations, as well as many other for-profit
and non-profit open source projects.

**Q: What will happen to Ethermint? Will it continue to be maintained?**

**A:** Evmos stays the maintainer of the Ethermint repository, which will remain under the LGPL v3 license.
New features developments that have been previously planned to be included
on Ethermint are instead going to be part of Evmos SDK.
For more information on the Evmos SDK, check out the Evmos Manifesto.

**Q: Do I need a commercial permit when testing ENCL software?**

**A:** No. ENCL specifically allows non-commercial use (e.g. use, modification and distribution). That includes testing.

**Q: If I have one copy of the software I am using for commercial purposes
and other copies of the software I am only using for non-commercial purposes,
for which ones do I need to obtain a commercial permit?**

**A:** You only need a commercial permit for any copies of the software that are running in commercial use,
unless they are running on the Designated Blockchains.
If you negotiate a commercial permit with us, it will contain details about the relationship between the two licenses.

**Q: What if Tharsis Labs  decides to change the use limitation in the future?**

**A:** We hope the current limitations will work for the foreseeable future. However, a licensor may change the use
limitation in future releases, and you will always be able to use any previous version of the ENCL software under its
conditions applied at the time of release. In other words,
the licensor’s changes would be forward-looking only and would not apply retroactively.

**Q: What will occur if I combine ENCL-1.0 code with other code?**

**A:** The portion of the code that is ENCL will continue to be subject to ENCL restrictions.
The license does not affect surrounding code, in the way LGPL or GPL might.

**Q: Can I use ENCL-1.0 software in an Open Source project?**

**A:**  No. The ENCL has a usage limitation that is not allowed with open source software.
You can eliminate the usage limitation by obtaining a commercial permit from Evmos for the ENCL code,
but we will not grant rights to release the code under an open source license.

**Q: Can I use ENCL-1.0 code in the code base for my commercial, closed-source product?**

**A:** No, unless you have obtained a commercial permit for the ENCL code
or can ensure compliance with the limitations in ENCL-1.0.

**Q: If I modify the source code of software licensed under the ENCL-1.0,
can I redistribute my modified version under an Open Source license, e.g. MIT or Apache 2.0?**

**A:** No. Your modified version consists of the original software (which is under the ENCL-1.0) and your modifications,
which together constitute a derivative work of the original software.
The license does not grant you the right to redistribute under a permissive license.
However, you can distribute a combined program that contains portions under each license.
Both ENCL-1.0 and permissive licenses allow that.

**Q: Can I use ENCL-1.0 products to develop software that will be licensed under different licenses?**

**A:** Yes, as long as you don’t include any of the ENCL code in the code for the software that’s being developed,
and you do not make commercial use of the ENCL code in violation of the conditions of the license.

**Q: I have written a code patch to an ENCL-1.0 project and would like the ENCL-1.0 vendor
to maintain the code as part of the ENCL-1.0 project. How do I contribute it?**

**A:** First, a big thank you! You can contribute the code to the official Evmos repository
by following the contributing guidelines a
nd the official Code of Conduct (see CONTRIBUTING).
You will need to sign a Contributor License Agreement (CLA) before your contribution
is peer-reviewed and accepted. A bot has been implemented to assist you through the signing process when contributing.

**Q: Can I backport any ENCL-1.0 code to an older, open source, version of the same software?**

**A:** No. In this circumstance, you would either violate the ENCL licensor’s copyright
by re-releasing the code under open source, or you would violate the open source project’s license
by introducing incompatible ENCL code (i.e., code subject to a use limitation not allowed by the open source project’s license).
