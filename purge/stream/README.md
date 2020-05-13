# NATS JetStream Stream Data Purge

A [GitHub Action](https://github.com/features/actions) to purge all data from a NATS [JetStream](https://github.com/nats-io/jetstream#readme) Stream.

## Usage

```yaml
on: push
name: consumer
jobs:
  clean_orders:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@master
      - name: purge_orders
        uses: nats-io/jetstream-gh-action/purge/stream@master
        with:
          stream: ORDERS
          server: nats.example.net:4222
```

## Inputs

|Input|Description|
|-----|-----------|
|`stream`|The Stream to delete (required)|
|`server`|Comma separated list of NATS Server URLs (required)|
|`username`|Username or Token to connect with|
|`password`|Password to connect with|
|`credentials`|Path to a file holding NATS credentials|
