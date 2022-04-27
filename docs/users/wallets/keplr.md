<!--
order: 4
-->

# Keplr

Connect your Keplr wallet with Evmos {synopsis}

## Pre-requisite Readings

- [Install Keplr](https://www.keplr.app/) {prereq}

The Keplr browser extension is a wallet for accessing applications in the Cosmos ecosystem and managing user identities. It can be used to connect to {{ $themeConfig.project.name }} through the official testnet and request Funds from the Faucet.

## Install Keplr

Add the Keplr browser extension following the instructions on the [Keplr website](https://www.keplr.app/). The Keplr extension is officially supported only on Chromium-based explorers.

## Create/Import Account

Open the Keplr extension on your browser. If you are setting up Keplr for the first time, you can either create a new account or import an existing account. Refer to the [Keplr documentation](https://keplr.crunch.help/getting-started) for further information.

## Connect Keplr to Emvos Mainnet

Once you are signed in to the Keplr extension, you can connect the wallet with the Evmos network.

To connect Keplr to mainnet, visit the Evmos [Dashboard](https://app.evmos.org/). On the dashboard, click `CONNECT WALLET` and select `Keplr`. A popup will prompt you `Requesting Connection` to add the Evmos mainnet chain (`evmos_{{ $themeConfig.project.chain_id }}-{{ $themeConfig.project.version_number }}`) to Keplr and approve the connection.
