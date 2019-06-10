# PRS3R - Parser the Minter in DB:ClickHouse/Redis

* [About](#about)
* [Installing DB](#installing-in-ubuntu)
* [Commands](#commands)
* [Key benefits from PostgreSQL](#key-benefits-from-postgresql)
* [Control](#control)

## About

Actual for [Minter Blockchain](https://minter.network) version 1.0.x.

## Installing in Ubuntu

#### Yandex ClickHouse

To install official packages add the Yandex repository in `/etc/apt/sources.list` or in a separate `/etc/apt/sources.list.d/clickhouse.list` file:

```ini
deb http://repo.yandex.ru/clickhouse/deb/stable/ main/
```

If you want to use the most recent version, replace `stable` with `testing` (this is not recommended for production environments).

Then run these commands to actually install packages:

```bash
sudo apt-get install dirmngr    # optional
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv E0C56BD4    # optional
sudo apt-get update
sudo apt-get install clickhouse-client clickhouse-server
```

Instructions from the [official site](https://clickhouse.yandex/docs/en/getting_started/#installation).

#### Redis

```bash
sudo apt-get install redis-server
```

## Commands

Creating and/or cleaning tables in databases:
```bash
prs3rd del
```

Downloading preliminary data (if needed) from the `start.json` file:
```bash
prs3rd json
```

Running the daemon:
```bash
prs3rd
```

## Key benefits from PostgreSQL:

* Size of base on disk
* Request execution speed

## Control

* `localhost:8018/status` - parser status
* `localhost:8018/start` - starting parser
* `localhost:8018/stop` - stoping parser
* `localhost:8018/exit` - exit