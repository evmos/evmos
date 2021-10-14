<!--
order: 1
-->

# Обзор

Узнайте о Evmos и его основных особенностях. {synopsis}

## Что такое Evmos

Evmos - это масштабируемый, высокопроизводительный блокчейн Proof-of-Stake, который полностью совместимс Ethereum. Он создан с использованием [Cosmos SDK](https://github.com/cosmos/cosmos-sdk/), который работает на базе движка консенсуса [Tendermint Core](https://github.com/tendermint/tendermint).

Evmos позволяет запускать классический Ethereum в качестве блокчейна для конкретного приложения [Cosmos](https://cosmos.network/). Это позволяет разработчикам иметь все желаемые функции Ethereum и в то же время пользоваться преимуществами реализации PoS в Tendermint. Кроме того, поскольку Evmos
построен на базе Cosmos SDK, он сможет обмениваться данными с остальной Cosmos
экосистемой через протокол межблокчейновой связи (IBC).

### Особенности

Вот некоторые ключевые особенности Evmos:

* Web3 и EVM совместимость
* Высокая пропускная способность благодаря [Tendermint Core](https://github.com/tendermint/tendermint)
* Горизонтальное масштабирование посредством [IBC](https://cosmos.network/ibc)
* Быстрое завершение транзакций

Evmos обеспечивает эти ключевые возможности благодаря:

* Реализация прикладного блокчейн-интерфейса Tendermint Core ([ABCI](https://docs.tendermint.com/master/spec/abci/)) для управления блокчейном.
* Использование [модулей](https://docs.cosmos.network/master/building-modules/intro.html) и других механизмов, реализованных в [Cosmos SDK](https://docs.cosmos.network/).
* Использование [`geth`](https://github.com/ethereum/go-ethereum) в качестве библиотеки для избежания повторного использования кода и улучшения сопровождаемости.
* Предоставление полностью совместимого с Web3 [JSON-RPC](./../basic/json_rpc.md) слоя для взаимодействия с существующими клиентами Ethereum и инструментарием ([Metamask](./../guides/keys-wallets/metamask.md), [Remix](./../guides/tools/remix.md), [Truffle](./../guides/tools/truffle.md) и т.п.).

Совокупность этих возможностей позволяет разработчикам использовать существующие инструменты и программное обеспечение экосистемы Ethereum для беспрепятственного развертывания смарт-контрактов, которые взаимодействуют с остальной частью [экосистемы Cosmos](https://cosmos.network/ecosystem)!

## Дальше {hide}

Узнайте об [архитектуре](./architecture.md) Evmos {hide}
