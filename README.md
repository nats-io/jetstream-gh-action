# NATS JetStream GitHub Action Pack

This is a collection of [GitHub Actions](https://github.com/features/actions) to interact with NATS [JetStream](https://github.com/nats-io/jetstream#readme).

|Action|Description|
|------|-----------|
|[nats-io/jetstream-gh-action/validate/stream@master](https://github.com/nats-io/jetstream-gh-action/tree/master/validate/stream)|Validates Stream Configuration|
|[nats-io/jetstream-gh-action/validate/consumer@master](https://github.com/nats-io/jetstream-gh-action/tree/master/validate/consumer)|Validates Consumer Configuration|
|[nats-io/jetstream-gh-action/create/stream@master](https://github.com/nats-io/jetstream-gh-action/tree/master/create/strean)|Creates Streams|
|[nats-io/jetstream-gh-action/create/consumer@master](https://github.com/nats-io/jetstream-gh-action/tree/master/create/consumer)|Creates Consumers|
|[nats-io/jetstream-gh-action/update/stream@master](https://github.com/nats-io/jetstream-gh-action/tree/master/update/strean)|Updates Streams|
|[nats-io/jetstream-gh-action/delete/stream@master](https://github.com/nats-io/jetstream-gh-action/tree/master/delete/strean)|Deletes Streams|
|[nats-io/jetstream-gh-action/delete/consumer@master](https://github.com/nats-io/jetstream-gh-action/tree/master/delete/consumer)|Deletes Consumers|
|[nats-io/jetstream-gh-action/purge/stream@master](https://github.com/nats-io/jetstream-gh-action/tree/master/purge/stream)|Purge all data from a Stream|
|[nats-io/jetstream-gh-action@master](https://github.com/nats-io/jetstream-gh-action/)|Publish to a JetStream Stream|

See individual action directory for detailed usage instructions.

## Publishing Messages

Messages can be published to a Stream - or any NATS subject - using the base action.

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
          message: Deployed versoin xxx
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
