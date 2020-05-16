# NATS JetStream GitHub Action Pack

This is a collection of [GitHub Actions](https://github.com/features/actions) to interact with NATS [JetStream](https://github.com/nats-io/jetstream#readme).

|Action|Description|
|------|-----------|
|[nats-io/jetstream-gh-action/validate/stream@master](https://github.com/nats-io/jetstream-gh-action/tree/master/validate/stream)|Validates Stream Configuration|
|[nats-io/jetstream-gh-action/validate/consumer@master](https://github.com/nats-io/jetstream-gh-action/tree/master/validate/consumer)|Validates Consumer Configuration|
|[nats-io/jetstream-gh-action/create/stream@master](https://github.com/nats-io/jetstream-gh-action/tree/master/create/stream)|Creates Streams|
|[nats-io/jetstream-gh-action/create/consumer@master](https://github.com/nats-io/jetstream-gh-action/tree/master/create/consumer)|Creates Consumers|
|[nats-io/jetstream-gh-action/update/stream@master](https://github.com/nats-io/jetstream-gh-action/tree/master/update/stream)|Updates Streams|
|[nats-io/jetstream-gh-action/delete/stream@master](https://github.com/nats-io/jetstream-gh-action/tree/master/delete/stream)|Deletes Streams|
|[nats-io/jetstream-gh-action/delete/consumer@master](https://github.com/nats-io/jetstream-gh-action/tree/master/delete/consumer)|Deletes Consumers|
|[nats-io/jetstream-gh-action/eval/stream@master](https://github.com/nats-io/jetstream-gh-action/tree/master/eval/stream)|Evaluate Stream state|
|[nats-io/jetstream-gh-action/eval/consumer@master](https://github.com/nats-io/jetstream-gh-action/tree/master/eval/consumer)|Evaluate Consumer state|
|[nats-io/jetstream-gh-action/purge/stream@master](https://github.com/nats-io/jetstream-gh-action/tree/master/purge/stream)|Purge all data from a Stream|
|[nats-io/jetstream-gh-action@master](https://github.com/nats-io/jetstream-gh-action/)|Publish to a JetStream Stream|

See individual action directory for detailed usage instructions.

## JetStream Service In Workflow

JetStream can be run within the workflow job as a local service, here's an example starting the Service and creating a Stream in it.

```yaml
on: push
name: orders
jobs:
  orders:
    runs-on: ubuntu-latest
    services:
      # starts a JetStream service locally known as "jetstream" on the network
      jetstream:
        image: synadia/jsm:latest
        options: >-
          -e JSM_MODE=server

      # creates a stream on the "jetstream:4222" server started above
      - name: orders_stream
        uses: nats-io/jetstream-gh-action/create/stream@master
        with:
          config: ORDERS.json
          server: jetstream:4222
```

This server is available to all steps in the job that hosts it.

## Publishing Messages

Messages can be published to a Stream, or any NATS subject, using the base action.

### Usage

```yaml
on: push
name: consumer
jobs:
  consumer_validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@master
      - name: hello
        uses: nats-io/jetstream-gh-action@master
        with:
          subject: ORDERS.deploy
          message: Published new deployment via "${{ github.event_name }}" in "${{ github.repository }}"
          should_ack: 1
          server: nats.example.net:4222
```

### Inputs

|Input|Description|
|-----|-----------|
|`subject`|The subject to publish to (required)|
|`message`|The message payload (required)|
|`should_ack`|If a positive response from JetStream is required for success|
|`server`|Comma separated list of NATS Server URLs (required)|
|`username`|Username or Token to connect with|
|`password`|Password to connect with|
|`credentials`|Path to a file holding NATS credentials|

### Outputs

|Output|Description|
|------|-----------|
|`response`|Response received or error body|
