<!--
order: 2
-->

# Faucet

Check how to obtain testnet tokens from the Evmos faucet website {synopsis}

## Obtain address

### Keyring

1. You can generate an Evmos key by using the [Keyring](./../guides/keys-wallets/keyring.md):

```bash
evmosd keys add testnet-key
```

You can obtain your key [Bech32](./../basics/accounts.md#addresses-and-public-keys) address by typing:

```bash
evmosd keys show testnet-key
```

### Metamask

1. Add the Testnet to the [Metamask](./../guides/keys-wallets/metamask.md) Networks settings.
2. Copy your Hex address and use the [`debug`](./../basics/accounts.md#addresses-conversion) command to obtain the Bech32 address

  ```bash
  evmosd debug addr 0x...
  ```

::: tip
Следуйте руководству [Инструкция по Metamask](./../guides/keys-wallets/metamask.md) для получения дополнительной информации о том, как настроить счета кошельков
:::

## Запросить токены

<!-- TODO: update to support Hex format -->
Вы можете запросить токены для тестнета с помощью [крана] Evmos (https://faucet.evmos.org).
Просто заполните свой адрес в поле ввода в формате Bech32 (`evmos1...`).

::: warning
Если вы используете свой адрес Bech32, убедитесь, что вы ввели [адрес учетной записи](./../basics/accounts.md#addresses-and-public-keys) (`evmos1...`), а **НЕ** адрес оператора валидатора (`evmosvaloper1...`).
:::

![faucet site](../.././img/faucet_web_page.png)

## Rate limits

To prevent the faucet account from draining the available funds, the Evmos testnet faucet
imposes a maximum number of request for a period of time. By default the faucet service accepts 1
request per day per address. All addresses **must** be authenticated using
[Auth0](https://auth0.com/) before requesting tokens.

<!-- TODO: add screenshots of authentication window -->

## Amount

For each request, the faucet transfers 1 {{ $themeConfig.project.testnet_denom }} to the given address.

## Faucet Addresses

The public faucet addresses for the testnet are:

- **Hex**: [`0x1549d29D1d51A694Cd5bbC89bF2c5F86ea5cE151`](https://evm.evmos.org/address/0x1549d29D1d51A694Cd5bbC89bF2c5F86ea5cE151/transactions)
- **Bech32**: [`evmos1z4ya98ga2xnffn2mhjym7tzlsm49ec23890sze`](https://explorer.evmos.org/accounts/evmos1z4ya98ga2xnffn2mhjym7tzlsm49ec23890sze)
