#!/bin/bash

set -ue

# check if buf is installed and install otherwise
if ! command -v buf > /dev/null
then
    echo "buf will be installed"
    # Substitute PREFIX for your preferred buf install prefix.
    # Substitute VERSION for buf's current released version.
    PREFIX="/usr/local" && \
    VERSION="1.6.0" && \
    curl -sSL \
        "https://github.com/bufbuild/buf/releases/download/v${VERSION}/buf-$(uname -s)-$(uname -m).tar.gz" | \
        tar -xvzf - -C "${PREFIX}" --strip-components 1
fi

# ensure that you authenticate with the BSR by generating an API token
# also make sure that you are added to the Evmos organization on buf
# run `buf registry login` and input your details

# below are the module addresses (directories containing `buf.yaml` files)
# paths are relative to the scripts/ directory
THIRDPARTYPROTO="../third_party/proto/"
EVMOSPROTO="../proto/"

# first, push the third party module and documentation as dependencies (order matters)
buf push $THIRDPARTYPROTO

# update the dependencies
buf mod update $EVMOSPROTO

# then, push the evmos proto module and documentation 
buf push $EVMOSPROTO

# two commit addresses should be printed to the command line
# the evmos documentation will have links to the third party documentation
# third party documentation can be viewed by looking at the previous commit history in the buf repository, as well