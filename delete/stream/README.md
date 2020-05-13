# NATS JetStream Stream Deletion

A [GitHub Action](https://github.com/features/actions) to delete NATS [JetStream](https://github.com/nats-io/jetstream#readme) Streams.

## Usage

```yaml
on: push
name: consumer
jobs:
  clean_orders:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@master
      - name: orders_stream
        uses: nats-io/jetstream-gh-action/delete/stream@master
        with:
          missing_ok: 1
          stream: ORDERS
          server: nats.example.net:4222
```

## Inputs

|Input|Description|
|-----|-----------|
|`stream`|The Stream to delete (required)|
|`missing_ok`|If the task should complete successfully if the Stream does not exist already|
|`server`|Comma separated list of NATS Server URLs (required)|
|`username`|Username or Token to connect with|
|`password`|Password to connect with|
|`credentials`|Path to a file holding NATS credentials|
