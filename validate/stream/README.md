# NATS JetStream Stream Configuration Validator

A [GitHub Action](https://github.com/features/actions) to validate NATS [JetStream](https://github.com/nats-io/jetstream#readme) Stream configuration files.

These files are suitable for input to the `nats stream add --config <config file>` command used to create streams.

## Usage

```yaml
on: push
name: stream
jobs:
  stream_validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@master
      - name: stream
        uses: nats-io/jetstream-gh-action/validate/stream@master
        with:
          config: stream.json
```

## Inputs

|Input|Description|
|-----|-----------|
|`config`|The path to the configuration file that should be tested (required)|
