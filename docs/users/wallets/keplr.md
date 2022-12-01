<!--
order: 4
-->

# Keplr

Connect your Keplr wallet with Evmos. {synopsis}

- [Install Keplr](https://www.keplr.app/) {prereq}

:::tip
**Note**: The Keplr extension is officially supported only on Chromium-based explorers.
:::

The [Keplr](https://www.keplr.app/) browser extension is a wallet for accessing applications in the Cosmos ecosystem and managing user identities. It can be used to connect to {{ $themeConfig.project.name }} mainnet and claim rewards, send and stake tokens, interact with governance, and more.

## Set Up Keplr

:::tip
**Note**: Further information about the process of setting up Keplr can be found in the official [Keplr Documentation](https://keplr.crunch.help/getting-started) or in this [Medium article](https://medium.com/chainapsis/how-to-use-keplr-wallet-40afc80907f6).
:::

Open the Keplr extension on your browser. If you are setting up Keplr for the first time, you can either [create a new account](#create-a-new-account) or [import an existing account](#import-an-existing-account).

### Create a New Account

There are several ways to create a new account:

- via a [mnemonic/seed phrase](#create-an-account-with-a-seed-phrase)
- via [one-click login](#create-an-account-with-one-click-login)

#### Create an Account with a Seed Phrase

1. In the initial pop-up window, choose **Create New Account**
    - If you have used Keplr before, click on the silhouette in the upper-right corner, then the blue box labeled **Add Account**, and select **Create New Account**
2. Choose to have a seed/mnemonic phrase of 24 words, and save the phrase
    - You can change the derivation path by clicking on **Advanced**, but this is optional (learn more in the [Keplr FAQ](https://faq.keplr.app/))
3. Enter a name for your account (can change later)
4. Once you have transcribed your 24 word seed/mnemonic phrase, click on **Next**
5. To confirm the creation of the new account, click on the words on the right order in which they appear in your seed/mnemonic phrase, and press **Register**
6. If you have not used Keplr before, set a password for the Keplr extension, and click **Confirm**

#### Create an Account with One-Click Login

:::tip
**Note**: It is suggested to create an account via mnemonic phrase or delegate via Ledger, not to use One-Click Login.
:::

1. Choose the option **Sign in with Google**
2. Now enter the email/phone number associated with your Google account, the password, and click **Next**
3. If you have not used Keplr before, set a password for the Keplr extension, and click **Confirm**

### Import an Existing Account

There are several ways to import an existing account:

- via a [mnemonic/seed phrase/private key](#import-an-account-with-a-seed-phrase)
- via [ledger](#import-an-account-with-a-ledger)

#### Import an Account with a Seed Phrase

1. In the initial pop-up window, choose **Import Existing Account**
    - If you have used Keplr before, click on the silhouette in the upper-right corner, then the blue box labeled **Add Account**, and select **Import Existing Account**
2. Enter your mnemonic/seed phrase/private key in the appropriate slot, seperating the words with spaces and taking care to check they are spelled correctly
3. Make sure you have imported the account with the correct derivation path, viewable by clicking on **Advanced**
    - Normally, the derivation path should be `m/44'/…’/0/0/0`, but if you see that importing the account via mnemonic on Keplr, the Cosmos Mainnet address displayed is different than yours, it is possible the derivation path ends with 1 (or another number) instead of 0
    - If this is the case, you just have to start the process over, and replace the last 0 with 1
    - Learn more in the [Keplr FAQ](https://faq.keplr.app/)
4. If you have not used Keplr before, set a password for the Keplr extension, and click **Confirm**

#### Import an Account with a Ledger

1. In the initial pop-up window, choose **Import Ledger**
   - If you have used Keplr before, click on the silhouette in the upper-right corner, then the blue box labeled **Add Account**, and select **Import Ledger**
   - Be sure you have both the Cosmos and Ethereum Ledger apps downloaded on your Ledger device
2. To complete the connection with your Ledger Nano Hard Wallet, follow the steps described in the pop-up that appears (a detailed tutorial can be found [here](https://medium.com/chainapsis/how-to-use-ledger-nano-hardware-wallet-with-keplr-9ea7f07826c2))
3. If you have not used Keplr before, set a password for the Keplr extension, and click **Confirm**
4. Switch to the Ethereum app on the Ledger, then select “Evmos” from the Keplr chain registry to connect the public key
   - All signing from Keplr will use the Ledger Ethereum app, with either [EIP-712 transactions](https://eips.ethereum.org/EIPS/eip-712) or standard [Ethereum transactions](https://ethereum.org/en/developers/docs/transactions/).

## Connect Keplr to Mainnet

Once you are signed in to the Keplr extension, you can connect the wallet with the Evmos network. The Evmos mainnet network is already built into Keplr; look for the `Evmos` network.
