<!--
layout: home
title: Evmos документация на русском
description: Evmos - это масштабируемый и совместимый с Ethereum блокчейн, построенный на основе Proof-of-Stake с быстрой финализацией.
sections:
  - title: Введение
    desc: Прочитайте обзор Evmos и его архитектуры.
    url: ./intro
    icon: ethereum-intro
  - title: Основы
    desc: Начните с основных понятий Evmos, таких как счета и транзакции.
    url: ./basics
    icon: basics
  - title: Основные концепции
    desc: Прочитайте об основных понятиях, таких как encoding и events.
    url: ./core
    icon: core
stack:
  - title: Cosmos SDK
    desc: Cosmos SDK - это самая популярная в мире платформа для создания блокчейн, ориентированных на конкретные приложения.
    color: "#5064FB"
    label: sdk
    url: http://docs.cosmos.network
  - title: Ethereum
    desc: Ethereum - это глобальная платформа с открытым исходным кодом для децентрализованных приложений.
    color: "#1A1F36"
    label: ethereum-black
    url: https://eth.wiki
  - title: Tendermint Core
    desc: Ведущий движок BFT для создания блокчейн, на котором работает Evmos.
    color: "#00BB00"
    label: core
    url: http://docs.tendermint.com
footer:
  newsletter: false
aside: false
-->

# Evmos документация на русском

## Начало работы

- **[Введение](./intro/overview.md)**: Обзор Evmos.

## Справочник

- **[Основы](./basics/)**: Документация по основным концепциям Evmos, таким как стандартная анатомия приложения, жизненный цикл транзакций и управление счетами.
- **[Ядро](./core/)**: Документация по основным концепциям Evmos, таким как `encoding` и `events`.
- **[Разработка модулей](./building-modules/)**: Важные понятия для разработчиков модулей, такие как `message`, `keeper`, `handler` и `querier`.
- **[Интерфейсы](./interfaces/)**: Документация по созданию интерфейсов для приложений Evmos.

## Другие источники

- **[Каталог модулей](../x/)**: Реализации модулей и соответствующая документация к ним.
- **[Справочник по API Ethermint](https://godoc.org/github.com/tharsis/ethermint)**: Godocs по API Ethermint.
- **[REST API spec](https://cosmos.network/rpc/)**:  Список эндпоинтов REST для взаимодействия с узлом через REST.

## Contribute

Посмотрите [этот файл](https://github.com/tharsis/evmos/blob/main/docs/DOCS_README.md) для получения подробной информации о процессе сборки и соображениях при внесении изменений.
