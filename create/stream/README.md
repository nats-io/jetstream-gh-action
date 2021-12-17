# NATS JetStream Stream Creation

A [GitHub Action](https://github.com/features/actions) to create NATS [JetStream](https://github.com/nats-io/jetstream#readme) Streams.

## Usage

```yaml
on: push
name: consumer
jobs:
  create_orders:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@main
      - name: orders_stream
        uses: nats-io/jetstream-gh-action/create/stream@main
        with:
          config: ORDERS.json
          server: nats.example.net:4222
      - name: orders_new_consumer
        uses: nats-io/jetstream-gh-action/create/consumer@main
        with:
          config: ORDERS_NEW.json
          stream: ORDERS
          server: nats.example.net:4222
```

## Inputs

|Input|Description|
|-----|-----------|
|`config`|The path to the configuration file to use (required)|
|`server`|Comma separated list of NATS Server URLs (required)|
|`username`|Username or Token to connect with|
|`password`|Password to connect with|
|`credentials`|Path to a file holding NATS credentials|

## Outputs

|Output|Description|
|------|-----------|
|`config`|The effective configuration that was created|
