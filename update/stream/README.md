# NATS JetStream Stream Configuration Update

A [GitHub Action](https://github.com/features/actions) to update the configuration of a NATS [JetStream](https://github.com/nats-io/jetstream#readme) Stream.

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
        uses: nats-io/jetstream-gh-action/update/stream@master
        with:
          stream: ORDERS
          server: nats.example.net:4222
          config: ORDERS.json
```

## Inputs

|Input|Description|
|-----|-----------|
|`stream`|The Stream to delete (required)|
|`config`|The Configuration update to apply to the stream (required)|
|`server`|Comma separated list of NATS Server URLs (required)|
|`username`|Username or Token to connect with|
|`password`|Password to connect with|
|`credentials`|Path to a file holding NATS credentials|
