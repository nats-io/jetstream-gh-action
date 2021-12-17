# NATS JetStream Stream State Evaluator

A [GitHub Action](https://github.com/features/actions) to evaluate the state of a NATS [JetStream](https://github.com/nats-io/jetstream#readme) Stream.

Use this action to confirm that after a Stream was created/purged/updated that it matches the desired state.

The expression language used in the `expression` language is the same [as used in several Hashicorp products](https://github.com/hashicorp/go-bexpr).
The evaluation is against the `StreamInfo` structure:

```go
type StreamInfo struct {
	Config StreamConfig `json:"config"`
	State  StreamState  `json:"state"`
}

type StreamConfig struct {
	Name         string          `json:"name"`
	Subjects     []string        `json:"subjects,omitempty"`
	Retention    RetentionPolicy `json:"retention"`
	MaxConsumers int             `json:"max_consumers"`
	MaxMsgs      int64           `json:"max_msgs"`
	MaxBytes     int64           `json:"max_bytes"`
	MaxAge       time.Duration   `json:"max_age"`
	MaxMsgSize   int32           `json:"max_msg_size,omitempty"`
	Storage      StorageType     `json:"storage"`
	Discard      DiscardPolicy   `json:"discard"`
	Replicas     int             `json:"num_replicas"`
	NoAck        bool            `json:"no_ack,omitempty"`
	Template     string          `json:"template_owner,omitempty"`
}

type StreamState struct {
	Msgs      uint64 `json:"messages"`
	Bytes     uint64 `json:"bytes"`
	FirstSeq  uint64 `json:"first_seq"`
	LastSeq   uint64 `json:"last_seq"`
	Consumers int    `json:"consumer_count"`
}
```

## Usage

```yaml
on: push
name: consumer
jobs:
  clean_orders:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@main
      - name: purge_orders
        uses: nats-io/jetstream-gh-action/update/stream@main
        with:
          stream: ORDERS
          expression: |
            State.Msgs == 1
            Config.Storage == 1
          server: nats.example.net:4222
```

## Inputs

|Input|Description|
|-----|-----------|
|`stream`|The Stream to evaluate (required)|
|`expression`|The expression to apply to the Stream state (required)|
|`server`|Comma separated list of NATS Server URLs (required)|
|`username`|Username or Token to connect with|
|`password`|Password to connect with|
|`credentials`|Path to a file holding NATS credentials|
