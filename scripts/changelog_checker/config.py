"""
This file contains the configurations for the changelog checker.
You can adjust the variables in this file to change the results of the checker.

Things that can be adjusted are:

    - the allowed change types with a release
    - the allowed description categories (i.e. the `(...)` portion at the beginning of an entry)
    - PRs that are allowed to occur twice in the changelog (e.g. backports of bug fixes)
    - a set of known patterns in PR descriptions and their preferred way of spelling
    - known exceptions that do not need to follow the formatting rules
    - the legacy version at which to stop the checking

"""

import os
import re
from typing import List


def get_allowed_categories() -> List[str]:
    """
    Returns a list of allowed categories for an individual changelog entry.

    It is using a set of predefined categories, that are extended by the entries in
     - the `x/...` modules
     - the `precompiles/...` subdirectories
     - the `precompiles/outposts/...` subdirectories

    :return: a list of allowed categories for an individual changelog entry
    """

    allowed_categories = [
        "all",
        "ante",
        "api",
        "app",
        "build",
        "ci",
        "cli",
        "crisis",
        "db",
        "deps",
        "docs",
        "docker",
        "eip712",
        "fees",
        "go",
        "make",
        "metrics",
        "outposts",
        "post",
        "precompiles",
        "proto",
        "release",
        "rpc",
        "swagger",
        "testnet",
        "tests",
        "types",
        "utils",
        "upgrade",
        # third party modules
        "bank",
        "distribution",
        "gov",
        "ics20",
        "staking",
        # outdated modules (we have to keep them since they're in the changelog)
        "claims",
        "consensus",
        "recovery",
        "incentives",
    ]

    base_path = os.path.dirname(os.path.dirname(os.path.dirname(__file__)))

    module_path = os.path.join(base_path, "x")
    for module in os.listdir(module_path):
        if os.path.isdir(os.path.join(module_path, module)):
            allowed_categories.append(module)

    precompile_path = os.path.join(base_path, "precompiles")
    for precompile in os.listdir(precompile_path):
        if os.path.isdir(os.path.join(precompile_path, precompile)):
            allowed_categories.append(precompile + "-precompile")

    outpost_path = os.path.join(base_path, "precompiles", "outposts")
    for outpost in os.listdir(outpost_path):
        if os.path.isdir(os.path.join(outpost_path, outpost)):
            allowed_categories.append(outpost + "-outpost")

    return allowed_categories


# List of allowed categories for an individual changelog entry.
ALLOWED_CATEGORIES = get_allowed_categories()

# A dictionary of allowed spellings for some common patterns in changelog entries.
ALLOWED_SPELLINGS = {
    "ABI": re.compile("abi", re.IGNORECASE),
    "API": re.compile("api", re.IGNORECASE),
    "CI": re.compile("ci", re.IGNORECASE),
    "Cosmos-SDK": re.compile(r"cosmos[\s-]*sdk", re.IGNORECASE),
    "CLI": re.compile("cli", re.IGNORECASE),
    "EIP-712": re.compile(r"eip[\s-]*712", re.IGNORECASE),
    "ERC-20": re.compile(r"erc[\s-]*20", re.IGNORECASE),
    "EVM": re.compile("evm", re.IGNORECASE),
    "IBC": re.compile("ibc", re.IGNORECASE),
    "ICS": re.compile("ics", re.IGNORECASE),
    "ICS-20": re.compile(r"ics[\s-]*20", re.IGNORECASE),
    "outpost": re.compile("outpost", re.IGNORECASE),
    "Osmosis": re.compile("osmosis", re.IGNORECASE),
    "PR": re.compile(r"(pr)(\s|$)", re.IGNORECASE),
    "precompile": re.compile("precompile", re.IGNORECASE),
    "SDK": re.compile("sdk", re.IGNORECASE),
    "Stride": re.compile("stride", re.IGNORECASE),
    "WERC-20": re.compile(r"werc[\s-]*20", re.IGNORECASE),
}

# Collection of allowed change types and the matching patterns.
ALLOWED_CHANGE_TYPES = {
    "API Breaking": re.compile(r"api\s*breaking", re.IGNORECASE),
    "Bug Fixes": re.compile(r"bug\s*fixes", re.IGNORECASE),
    "Features": re.compile("features", re.IGNORECASE),
    "Improvements": re.compile("improvements", re.IGNORECASE),
    "State Machine Breaking": re.compile(r"state\s*machine\s*breaking", re.IGNORECASE),
}

# A list of pull requests that are allowed to be mentioned multiple times in the changelog.
# Usually, this only applies to bug fixes that were patched on two versions (e.g. v12.1.6 and v13.0.0).
ALLOWED_DUPLICATES = [
    1370,
    1635,
]

# A list of known exceptions to the formattiing. This usually applies to PRs that e.g. merged contents from
# a security advisory.
KNOWN_EXCEPTIONS = [
    "- (vesting) Refactor vesting flow.",
    "- (vesting) Fix vesting bug.",
    "- (vesting) [GHSA-2q3r-p2m3-898g](https://github.com/evmos/evmos/commit/39b750cdaf1d69158ab93da85bd43ae4a7da1456"
    + ") Apply ClawbackVestingAccount Barberry patch & Bump SDK to v0.46.13",
]

# The legacy major version at which to stop the checking.
LEGACY_VERSION: int = 2
