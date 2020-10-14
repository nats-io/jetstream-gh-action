package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-bexpr"
	"github.com/nats-io/jsm.go"
	"github.com/nats-io/jsm.go/api"
	"github.com/nats-io/nats.go"
)

type handler func() error

type validatable interface {
	Validate(v ...api.StructValidator) (bool, []string)
}

type environment interface {
	GetInput(string) string
	SetOutput(string, string) error
	Fatalf(string, ...interface{})
	Debugf(string, ...interface{})
	Printf(string, ...interface{})
}

var (
	commands = make(map[string]handler)
	env      environment
)

func main() {
	switch {
	case os.Getenv("GITHUB_ACTIONS") == "true":
		env = github{}

	default:
		env = stdout{}
	}

	mustRegister("VALIDATE_STREAM_CONFIG", handleValidateStreamConfig)
	mustRegister("VALIDATE_CONSUMER_CONFIG", handleValidateConsumerConfig)

	mustRegister("DELETE_STREAM", handleDeleteStream)
	mustRegister("DELETE_CONSUMER", handleDeleteConsumer)

	mustRegister("CREATE_STREAM", handleCreateStream)
	mustRegister("CREATE_CONSUMER", handleCreateConsumer)

	mustRegister("UPDATE_STREAM", handleUpdateStream)
	mustRegister("PURGE_STREAM", handlePurgeStream)
	mustRegister("PUBLISH", handlePublish)

	mustRegister("EVAL_STREAM", handleEvalStream)
	mustRegister("EVAL_CONSUMER", handleEvalConsumer)

	err := runAction()
	if err != nil {
		env.Fatalf("Command failed: %s", err)
	}
}

func handleEvalStream() error {
	stream := env.GetInput("STREAM")
	if stream == "" {
		return fmt.Errorf("STREAM is required")
	}

	expr := env.GetInput("EXPRESSION")
	if expr == "" {
		return fmt.Errorf("EXPRESSOIN is required")
	}

	_, mgr, err := connect()
	if err != nil {
		return err
	}

	str, err := mgr.LoadStream(stream)
	if err != nil {
		return err
	}

	info, err := str.Information()
	if err != nil {
		return err
	}

	env.Printf("Stream State: %#v", info)
	var failures []string

	for _, e := range strings.Split(expr, "\n") {
		eval, err := bexpr.CreateEvaluatorForType(e, nil, info)
		if err != nil {
			return err
		}

		result, err := eval.Evaluate(info)
		if err != nil {
			return err
		}

		if result {
			env.Printf("stream %q state matched %q", stream, e)
			continue
		}

		failures = append(failures, e)
		return fmt.Errorf("stream %q state did not match %q", stream, e)
	}

	if len(failures) > 0 {
		return fmt.Errorf("stream state did not match %d expressions:\n\t%s", len(failures), strings.Join(failures, "\n\t"))
	}

	return nil
}

func handleEvalConsumer() error {
	stream := env.GetInput("STREAM")
	if stream == "" {
		return fmt.Errorf("STREAM is required")
	}

	consumer := env.GetInput("CONSUMER")
	if consumer == "" {
		return fmt.Errorf("CONSUMER is required")
	}

	expr := env.GetInput("EXPRESSION")
	if expr == "" {
		return fmt.Errorf("EXPRESSOIN is required")
	}

	_, mgr, err := connect()
	if err != nil {
		return err
	}

	str, err := mgr.LoadStream(stream)
	if err != nil {
		return err
	}

	cons, err := str.LoadConsumer(consumer)
	if err != nil {
		return err
	}

	info, err := cons.State()
	if err != nil {
		return err
	}

	env.Printf("Consumer State: %#v", info)
	var failures []string

	for _, e := range strings.Split(expr, "\n") {
		eval, err := bexpr.CreateEvaluatorForType(e, nil, info)
		if err != nil {
			return err
		}

		result, err := eval.Evaluate(info)
		if err != nil {
			return err
		}

		if result {
			env.Printf("consumer %q > %q state matched %q", stream, consumer, e)
			continue
		}

		failures = append(failures, e)
		return fmt.Errorf("consumer %q > %q state did not match %q", stream, consumer, e)
	}

	if len(failures) > 0 {
		return fmt.Errorf("consumer state did not match %d expressions:\n\t%s", len(failures), strings.Join(failures, "\n\t"))
	}

	return nil
}

func handlePurgeStream() error {
	stream := env.GetInput("STREAM")
	if stream == "" {
		return fmt.Errorf("STREAM is required")
	}

	_, mgr, err := connect()
	if err != nil {
		return err
	}

	env.Printf("Purging Stream %q", stream)
	str, err := mgr.LoadStream(stream)
	if err != nil {
		return err
	}
	env.Printf("Purged Stream %q", stream)

	return str.Purge()
}

func handlePublish() error {
	subj := env.GetInput("SUBJECT")
	if subj == "" {
		return fmt.Errorf("SUBJECT is required")
	}

	msg := env.GetInput("MESSAGE")
	if msg == "" {
		return fmt.Errorf("MESSAGE is required")
	}

	shouldAck, err := strconv.ParseBool(env.GetInput("SHOULD_ACK"))
	if err != nil {
		shouldAck = true
	}

	nc, _, err := connect()
	if err != nil {
		return err
	}

	if shouldAck {
		env.Printf("Publishing %d bytes to %q and waiting for acknowledgement", len(msg), subj)
		resp, err := nc.Request(subj, []byte(msg), 5*time.Second)
		if err != nil {
			env.SetOutput("response", err.Error())
			return fmt.Errorf("publish Request failed: %s", err)
		}

		if !jsm.IsOKResponse(resp) {
			env.SetOutput("response", string(resp.Data))
			return fmt.Errorf("publish failed: %s", string(resp.Data))
		}

		env.Printf(string(resp.Data))

		env.SetOutput("response", string(resp.Data))

		return nil
	}

	env.Printf("Publishing %d bytes to %q", len(msg), subj)
	err = nc.Publish(subj, []byte(msg))
	if err != nil {
		env.SetOutput("response", err.Error())
		return err
	}

	err = nc.Flush()
	if err != nil {
		env.SetOutput("response", err.Error())
		return err
	}

	env.SetOutput("response", "published without requesting Ack")

	return nil
}

func handleUpdateStream() error {
	stream := env.GetInput("STREAM")
	if stream == "" {
		return fmt.Errorf("STREAM is required")
	}

	cfile := env.GetInput("CONFIG")
	if cfile == "" {
		return fmt.Errorf("CONFIG is required")
	}

	cj, err := ioutil.ReadFile(cfile)
	if err != nil {
		return err
	}

	var cfg api.StreamConfig
	err = json.Unmarshal(cj, &cfg)
	if err != nil {
		return err
	}

	_, mgr, err := connect()
	if err != nil {
		return err
	}

	str, err := mgr.LoadStream(stream)
	if err != nil {
		return err
	}

	// sorts strings to subject lists that only differ in ordering is considered equal
	sorter := cmp.Transformer("Sort", func(in []string) []string {
		out := append([]string(nil), in...)
		sort.Strings(out)
		return out
	})

	diff := cmp.Diff(str.Configuration(), cfg, sorter)
	if diff != "" {
		env.Printf("Differences (-old +new):\n%s", diff)
	} else {
		env.Printf("No difference")
		return nil
	}

	err = str.UpdateConfiguration(cfg)
	if err != nil {
		return err
	}

	err = str.Reset()
	if err != nil {
		return err
	}

	cj, err = json.Marshal(str.Configuration())
	if err != nil {
		return err
	}
	env.SetOutput("config", string(cj))

	env.Printf("Updated Stream %q with configuration: %s", stream, string(cj))

	return nil
}

func handleDeleteConsumer() error {
	stream := env.GetInput("STREAM")
	if stream == "" {
		return fmt.Errorf("STREAM is required")
	}

	consumer := env.GetInput("CONSUMER")
	if consumer == "" {
		return fmt.Errorf("CONSUMER is required")
	}

	missingok, err := strconv.ParseBool(env.GetInput("MISSING_OK"))
	if err != nil {
		missingok = false
	}

	_, mgr, err := connect()
	if err != nil {
		return err
	}

	known, err := mgr.IsKnownStream(stream)
	if err != nil {
		return err
	}

	if missingok && !known {
		env.Printf("Stream %s was not present", stream)
		return nil
	}

	if !known {
		return fmt.Errorf("stream %s does not exist", stream)
	}

	known, err = mgr.IsKnownConsumer(stream, consumer)
	if err != nil {
		return err
	}

	if missingok && !known {
		env.Printf("Consumer %s > %s was not present", stream, consumer)
		return nil
	}

	if !known {
		return fmt.Errorf("consumer %s > %s does not exist", stream, consumer)
	}

	cons, err := mgr.LoadConsumer(stream, consumer)
	if err != nil {
		return err
	}

	err = cons.Delete()
	if err != nil {
		return err
	}

	env.Printf("Deleted consumer %s > %s", stream, consumer)

	return nil
}

func handleDeleteStream() error {
	stream := env.GetInput("STREAM")
	if stream == "" {
		return fmt.Errorf("STREAM is required")
	}

	missingok, err := strconv.ParseBool(env.GetInput("MISSING_OK"))
	if err != nil {
		missingok = false
	}

	_, mgr, err := connect()
	if err != nil {
		return err
	}

	known, err := mgr.IsKnownStream(stream)
	if err != nil {
		return err
	}

	if missingok && !known {
		env.Printf("Stream %s was not present", stream)
		return nil
	}

	if !known {
		return fmt.Errorf("stream %s does not exist, cannot delete it", stream)
	}

	str, err := mgr.LoadStream(stream)
	if err != nil {
		return err
	}

	err = str.Delete()
	if err != nil {
		return err
	}

	env.Printf("Deleted stream %q", stream)

	return nil
}

func handleCreateStream() error {
	cfile := env.GetInput("CONFIG")
	if cfile == "" {
		return fmt.Errorf("CONFIG is required")
	}

	cj, err := ioutil.ReadFile(cfile)
	if err != nil {
		return err
	}

	var cfg api.StreamConfig
	err = json.Unmarshal(cj, &cfg)
	if err != nil {
		return err
	}

	_, mgr, err := connect()
	if err != nil {
		return err
	}

	stream, err := mgr.NewStreamFromDefault(cfg.Name, cfg)
	if err != nil {
		return err
	}

	cj, err = json.Marshal(stream.Configuration())
	if err != nil {
		return err
	}
	env.SetOutput("config", string(cj))

	env.Printf("Created stream %q using %q\n%s", cfg.Name, cfile, string(cj))

	return nil
}

func handleCreateConsumer() error {
	cfile := env.GetInput("CONFIG")
	if cfile == "" {
		return fmt.Errorf("CONFIG is required")
	}

	stream := env.GetInput("STREAM")
	if stream == "" {
		return fmt.Errorf("STREAM is required")
	}

	cj, err := ioutil.ReadFile(cfile)
	if err != nil {
		return err
	}

	var cfg api.ConsumerConfig
	err = json.Unmarshal(cj, &cfg)
	if err != nil {
		return err
	}

	_, mgr, err := connect()
	if err != nil {
		return err
	}

	consumer, err := mgr.NewConsumerFromDefault(stream, cfg)
	if err != nil {
		return err
	}

	cj, err = json.Marshal(consumer.Configuration())
	if err != nil {
		return err
	}
	env.SetOutput("config", string(cj))

	env.Printf("Created consumer %q > %q using %q\n%s", stream, consumer.Name(), cfile, string(cj))

	return nil
}

func handleValidateStreamConfig() error {
	cfg := &api.StreamConfig{}
	cfile, valid, errs, err := validateHelper("CONFIG", cfg)
	if err != nil {
		return err
	}

	if valid {
		env.Printf("%s is a valid Stream Configuration", cfile)
		return nil
	}

	return fmt.Errorf("%s is an invalid JetStream Stream configuration:\n\t%s\n", cfile, strings.Join(errs, "\n\t"))
}

func handleValidateConsumerConfig() error {
	cfg := &api.ConsumerConfig{}
	cfile, valid, errs, err := validateHelper("CONFIG", cfg)
	if err != nil {
		return err
	}

	if valid {
		env.Printf("%s is a valid Consumer Configuration", cfile)
		return nil
	}

	return fmt.Errorf("%s is an invalid JetStream Consumer configuration:\n\t%s\n", cfile, strings.Join(errs, "\n\t"))
}

func validateHelper(input string, cfg validatable) (string, bool, []string, error) {
	cfile := env.GetInput(input)
	if cfile == "" {
		return "", false, nil, fmt.Errorf("%s is required", input)
	}

	cb, err := ioutil.ReadFile(cfile)
	if err != nil {
		return cfile, false, nil, err
	}

	err = json.Unmarshal(cb, cfg)
	if err != nil {
		return cfile, false, nil, fmt.Errorf("could not parse '%s': %s", cfile, err)
	}

	ok, errs := cfg.Validate()

	return cfile, ok, errs, nil
}

func connect() (*nats.Conn, *jsm.Manager, error) {
	creds := env.GetInput("CREDENTIALS")
	user := env.GetInput("USERNAME")
	pass := env.GetInput("PASSWORD")

	server := env.GetInput("SERVER")
	if server == "" {
		return nil, nil, fmt.Errorf("SERVER is required")
	}

	opts := []nats.Option{
		nats.Name("jetstream-gh-action"),
		nats.ErrorHandler(func(_ *nats.Conn, _ *nats.Subscription, err error) {
			env.Fatalf("NATS Error: %s", err)
		}),
	}

	if user != "" {
		opts = append(opts, nats.UserInfo(user, pass))
	}

	if creds != "" {
		opts = append(opts, nats.UserCredentials(creds))
	}

	env.Debugf("Attempting to connect to %q", server)
	nc, err := nats.Connect(server, opts...)
	if err == nil {
		env.Printf("Connected to %q", nc.ConnectedUrl())
	}

	mgr, err := jsm.New(nc, jsm.WithAPIValidation(new(SchemaValidator)))
	if err != nil {
		return nil, nil, err
	}

	return nc, mgr, err
}

func register(command string, h handler) error {
	_, ok := commands[command]
	if ok {
		return fmt.Errorf("already registered")
	}

	commands[command] = h

	return nil
}

func mustRegister(command string, h handler) {
	err := register(command, h)
	if err != nil {
		env.Fatalf("Could not register '%s': %s", command, err)
	}
}

func runAction() error {
	start := time.Now()
	command := env.GetInput("COMMAND")

	defer func() {
		elapsed := time.Now().Sub(start)
		env.Printf("Ran %q in %v", command, elapsed)
	}()

	env.Printf("Starting NATS JetStream Action Pack command %q", command)

	cmd, ok := commands[command]
	if !ok {
		return fmt.Errorf("unknown command %s", command)
	}

	err := cmd()
	if err != nil {
		return err
	}

	return nil
}
