# NATS JetStream Consumer Configuration Validator

A [GitHub Action](https://github.com/features/actions) to validate NATS [JetStream](https://github.com/nats-io/jetstream#readme) Consumer configuration files.

These files are suitable for input to the `nats consumer add --config <config file>` command used to create consumers.

## Usage

```yaml
on: push
name: consumer
jobs:
  consumer_validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@master
      - name: consumer
        uses: nats-io/jetstream-gh-action/validate/consumer@master
        with:
          config: consumer.json
```

## Inputs

|Input|Description|
|-----|-----------|
|`config`|The path to the configuration file that should be tested (required)|
