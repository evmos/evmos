<!--
order: 3
-->

# Tendermint KMS + Ledger

Set up Tendermint KMS with the Tendermint Ledger app {synopsis}

## Pre-requisites

- [Ledger device](https://shop.ledger.com/) {prereq}
- [Install Ledger Live](https://www.ledger.com/ledger-live) {prereq}

## Checklist

::: warning
🚧  The following instructions are a brief walkthrough and not a comprehensive guideline. You should consider and research more about the [security implications](../validators/security.md) of activating an external KMS.
:::

- ✅ Ledger [Nano X](https://shop.ledger.com/pages/ledger-nano-x) or [Nano S](https://shop.ledger.com/products/ledger-nano-s) device (compare [here](https://shop.ledger.com/pages/hardware-wallets-comparison))
- ✅ [Ledger Live](https://www.ledger.com/ledger-live) installed
- ✅ Tendermint app installed (only in `Developer Mode`)
- ✅ Latest Versions (Firmware and Tendermint app)

## Tendermint Validator app (for Ledger devices)

::: danger
🚨**IMPORTANT**: KMS and Ledger Tendermint app are currently work in progress. Details may vary. Use under **your own risk**
:::

You should be able to find the Tendermint app in Ledger Live.

::: tip
You will need to enable `Developer Mode` in Ledger Live `Settings` in order to find the app.
:::

## KMS configuration

In this section, we will configure a KMS to use a Ledger device running the Tendermint Validator App.

### Config file

You can find other configuration examples [here](https://github.com/iqlusioninc/tmkms/blob/master/tmkms.toml.example)

- Create a `~/.tmkms/tmkms.toml` file with the following content (use an adequate `chain_id`)

  ```toml
  # Example KMS configuration file
  [[validator]]
  addr = "tcp://localhost:26658"                  # or "unix:///path/to/socket"
  chain_id = "evmos_9001-1"
  reconnect = true                                # true is the default
  secret_key = "~/.tmkms/secret_connection.key"

  [[providers.ledger]]
  chain_ids = ["evmos_9001-1"]
  ```

- Edit `addr` to point to your `evmosd` instance.
- Adjust `chain-id` to match your `.evmosd/config/config.toml` settings.
- `provider.ledger` has not additional parameters at the moment, however, it is important that you keep that header to enable the feature.

*Plug your Ledger device and open the Tendermint validator app.*

### Generate secret key

Now you need to generate a `secret_key`:

```bash
tmkms keygen ~/.tmkms/secret_connection.key
```

### Retrieve validator key

The last step is to retrieve the validator key that you will use in `evmosd`.

Start the KMS:

```bash
tmkms start -c ~/.tmkms/tmkms.toml
```

The output should look similar to:

```text
07:28:24 [INFO] tmkms 0.11.0 starting up...
07:28:24 [INFO] [keyring:ledger:ledger] added validator key evmosvalconspub1zcjduepqy53m39prgp9dz3nz96kaav3el5e0th8ltwcf8cpavqdvpxgr5slsd6wz6f
07:28:24 [INFO] KMS node ID: 1BC12314E2E1C29015B66017A397F170C6ECDE4A
```

The KMS may complain that it cannot connect to `evmosd`. That is fine, we will fix it in the next section.
This output indicates the validator key linked to this particular device is: `evmosvalconspub1zcjduepqy53m39prgp9dz3nz96kaav3el5e0th8ltwcf8cpavqdvpxgr5slsd6wz6f`
Take note of the validator pubkey that appears in your screen. *We will use it in the next section.*

## Evmos configuration

You need to enable KMS access by editing `.evmosd/config/config.toml`. In this file, modify `priv_validator_laddr` to create a listening address/port or a unix socket in `evmosd`.

For example:

```toml
...
# TCP or UNIX socket address for Tendermint to listen on for
# connections from an external PrivValidator process
priv_validator_laddr = "tcp://127.0.0.1:26658"
...
```

Let's assume that you have set up your validator account and called it `kmsval`. You can tell evmosd the key that we've got in the previous section.

```bash
evmosd gentx --name kmsval --pubkey <pub_key>
```

Now start `evmosd`. You should see that the KMS connects and receives a signature request.

Once the Ledger device receives the first message, it will ask for confirmation that the values are adequate.

![Tendermint Ledger app "Init Validation"](./../../img/kms_tm_ledger_01.jpg)

Click the right button, if the height and round are correct.

After that, you will see that the KMS will start forwarding all signature requests to the Ledger app:

![Tendermint Ledger app "Proposal"](./../../img/kms_tm_ledger_02.jpg)

::: warning
The word `TEST` in the second picture, second line appears because they were taken on a pre-release version.
Once the app as been released in Ledger's app store, this word should NOT appear.
:::
