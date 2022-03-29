<!--
order: 1
-->

# Keyring

Create, import, export and delete keys using the CLI keyring {synopsis}

The keyring holds the private/public keypairs used to interact with the node. For instance, a validator key needs to be set up before running the node, so that blocks can be correctly signed. The private key can be stored in different locations, called ["backends"](#keyring-backends), such as a file or the operating system's own key storage.

## Add keys

You can use the following commands for help with the `keys` command and for more information about a particular subcommand, respectively:

```bash
evmosd keys
```

```bash
evmosd keys [command] --help
```

To create a new key in the keyring, run the `add` subcommand with a `<key_name>` argument. You will have to provide a password for the newly generated key. This key will be used in the next section.

```bash
evmosd keys add mykey

# Put the generated address in a variable for later use.
MY_VALIDATOR_ADDRESS=$(evmosd keys show mykey -a)
```

This command generates a new 24-word mnemonic phrase, persists it to the relevant backend, and outputs information about the keypair. If this keypair will be used to hold value-bearing tokens, be sure to write down the mnemonic phrase somewhere safe!

By default, the keyring generates a `eth_secp256k1` key. The keyring also supports `ed25519` keys, which may be created by passing the `--algo` flag. A keyring can of course hold both types of keys simultaneously.

::: warning
**NOTE**: Cosmos `secp256k1` keys are not supported on Evmos due to compatibility issues with Ethereum transactions.
:::

## Keyring Backends

### OS

::: tip
**`os`** is the default option since operating system's default credentials managers are
designed to meet users' most common needs and provide them with a comfortable
experience without compromising on security.
:::

The `os` backend relies on operating system-specific defaults to handle key storage
securely. Typically, an operating system's credential sub-system handles password prompts,
private keys storage, and user sessions according to the user's password policies. Here
is a list of the most popular operating systems and their respective passwords manager:

- macOS (since Mac OS 8.6): [Keychain](https://support.apple.com/en-gb/guide/keychain-access/welcome/mac)
- Windows: [Credentials Management API](https://docs.microsoft.com/en-us/windows/win32/secauthn/credentials-management)
- GNU/Linux:
    - [libsecret](https://gitlab.gnome.org/GNOME/libsecret)
    - [kwallet](https://api.kde.org/frameworks/kwallet/html/index.html)

GNU/Linux distributions that use GNOME as default desktop environment typically come with
[Seahorse](https://wiki.gnome.org/Apps/Seahorse). Users of KDE based distributions are
commonly provided with [KDE Wallet Manager](https://userbase.kde.org/KDE_Wallet_Manager).
Whilst the former is in fact a `libsecret` convenient frontend, the latter is a `kwallet`
client.

The recommended backends for headless environments are `file` and `pass`.

### File

The `file` stores the keyring encrypted within the app's configuration directory. This
keyring will request a password each time it is accessed, which may occur multiple
times in a single command resulting in repeated password prompts. If using bash scripts
to execute commands using the `file` option you may want to utilize the following format
for multiple prompts:

```bash
# assuming that KEYPASSWD is set in the environment
yes $KEYPASSWD | evmosd keys add me
yes $KEYPASSWD | evmosd keys show me
# start evmosd with keyring-backend flag
evmosd --keyring-backend=file start
```

::: tip
The first time you add a key to an empty keyring, you will be prompted to type the password twice.
:::

### Password Store

The `pass` backend uses the [pass](https://www.passwordstore.org/) utility to manage on-disk
encryption of keys' sensitive data and metadata. Keys are stored inside `gpg` encrypted files
within app-specific directories. `pass` is available for the most popular UNIX
operating systems as well as GNU/Linux distributions. Please refer to its manual page for
information on how to download and install it.

::: tip
**`pass`** uses [GnuPG](https://gnupg.org/) for encryption. `gpg` automatically invokes the `gpg-agent`
daemon upon execution, which handles the caching of GnuPG credentials. Please refer to `gpg-agent`
man page for more information on how to configure cache parameters such as credentials TTL and
passphrase expiration.
:::

The password store must be set up prior to first use:

```sh
pass init <GPG_KEY_ID>
```

Replace `<GPG_KEY_ID>` with your GPG key ID. You can use your personal GPG key or an alternative
one you may want to use specifically to encrypt the password store.

### KDE Wallet Manager

The `kwallet` backend uses `KDE Wallet Manager`, which comes installed by default on the
GNU/Linux distributions that ships KDE as default desktop environment. Please refer to
[KWallet Handbook](https://docs.kde.org/stable5/en/kdeutils/kwallet5/index.html) for more
information.

### Testing

The `test` backend is a password-less variation of the `file` backend. Keys are stored
**unencrypted** on disk. This keyring is provided for <u>testing purposes only</u>. Use at your own risk!

::: danger
🚨 **DANGER**: <u>Never</u> create your mainnet validator keys using a `test` keying backend. Doing so might result in a loss of funds by making your funds remotely accessible via the `eth_sendTransaction` JSON-RPC endpoint.

Ref: [Security Advisory: Insecurely configured geth can make funds remotely accessible](https://blog.ethereum.org/2015/08/29/security-alert-insecurely-configured-geth-can-make-funds-remotely-accessible/)
:::

### In Memory

The `memory` backend stores keys in memory. The keys are immediately deleted after the program has exited.

:::danger
**IMPORTANT**: Provided for testing purposes only. The `memory` backend is **not** recommended for use in production environments. Use at your own risk!
:::
