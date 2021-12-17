# NATS JetStream Consumer Deletion

A [GitHub Action](https://github.com/features/actions) to delete NATS [JetStream](https://github.com/nats-io/jetstream#readme) Consumers.

## Usage

```yaml
on: push
name: consumer
jobs:
  clean_orders:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@main
      - name: orders_new_consumer
        uses: nats-io/jetstream-gh-action/delete/consumer@main
        with:
          missing_ok: 1
          stream: ORDERS
          consumer: NEW
          server: nats.example.net:4222
```

## Inputs

|Input|Description|
|-----|-----------|
|`stream`|The Stream to delete the Consumer from (required)|
|`consumer`|The Consumer to delete (required)|
|`missing_ok`|If the task should complete successfully if the Consumer does not exist already|
|`server`|Comma separated list of NATS Server URLs (required)|
|`username`|Username or Token to connect with|
|`password`|Password to connect with|
|`credentials`|Path to a file holding NATS credentials|
