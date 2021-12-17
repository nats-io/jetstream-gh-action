# NATS JetStream Consumer State Evaluator

A [GitHub Action](https://github.com/features/actions) to evaluate the state of a NATS [JetStream](https://github.com/nats-io/jetstream#readme) Consumer.

Use this action to confirm that after a Consumer was created that it matches the desired state.

The expression language used in the `expression` language is the same [as used in several Hashicorp products](https://github.com/hashicorp/go-bexpr).
The evaluation is against the `ConsumerInfo` structure:

```go
type ConsumerInfo struct {
	Stream         string         `json:"stream_name"`
	Name           string         `json:"name"`
	Config         ConsumerConfig `json:"config"`
	Created        time.Time      `json:"created"`
	Delivered      SequencePair   `json:"delivered"`
	AckFloor       SequencePair   `json:"ack_floor"`
	NumPending     int            `json:"num_pending"`
	NumRedelivered int            `json:"num_redelivered"`
}

type ConsumerConfig struct {
	Durable         string        `json:"durable_name,omitempty"`
	DeliverSubject  string        `json:"deliver_subject,omitempty"`
	DeliverPolicy   DeliverPolicy `json:"deliver_policy"`
	OptStartSeq     uint64        `json:"opt_start_seq,omitempty"`
	OptStartTime    *time.Time    `json:"opt_start_time,omitempty"`
	AckPolicy       AckPolicy     `json:"ack_policy"`
	AckWait         time.Duration `json:"ack_wait,omitempty"`
	MaxDeliver      int           `json:"max_deliver,omitempty"`
	FilterSubject   string        `json:"filter_subject,omitempty"`
	ReplayPolicy    ReplayPolicy  `json:"replay_policy"`
	SampleFrequency string        `json:"sample_freq,omitempty"`
}

type SequencePair struct {
	ConsumerSeq uint64 `json:"consumer_seq"`
	StreamSeq   uint64 `json:"stream_seq"`
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
          consumer: NEW
          expression: Delivered.StreamSeq == 100
          server: nats.example.net:4222
```

## Inputs

|Input|Description|
|-----|-----------|
|`stream`|The Stream that the consumer belongs to (required)|
|`consumer`|The Consumer to evaluate (required)|
|`expression`|The expression to apply to the Stream state (required)|
|`server`|Comma separated list of NATS Server URLs (required)|
|`username`|Username or Token to connect with|
|`password`|Password to connect with|
|`credentials`|Path to a file holding NATS credentials|
