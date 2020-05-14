package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nats-io/jsm.go"
	"github.com/nats-io/jsm.go/api"
	"github.com/nats-io/nats.go"
)

type handler func() error

type validatable interface {
	Validate() (bool, []string)
}

type environment interface {
	GetInput(string) string
	SetOutput(string, string) error
	Fatalf(string, ...interface{})
	Debugf(string, ...interface{})
}

var (
	commands = make(map[string]handler)
	env      environment
)

func main() {
	switch {
	case os.Getenv("GITHUB_ACTIONS") == "true":
		env = github{}

	case os.Getenv("TEKTON") == "true":
		env = tekton{}

	default:
		panic("Cannot determine execution environment")

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

	runAction()
}

func handlePurgeStream() error {
	stream := env.GetInput("STREAM")
	if stream == "" {
		return fmt.Errorf("STREAM is required")
	}

	nc, err := connect()
	if err != nil {
		return err
	}
	log.Printf("Connected to %s", nc.ConnectedUrl())

	str, err := jsm.LoadStream(stream, jsm.WithConnection(nc))
	if err != nil {
		return err
	}

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

	nc, err := connect()
	if err != nil {
		return err
	}

	log.Printf("Connected to %s", nc.ConnectedUrl())

	if shouldAck {
		resp, err := nc.Request(subj, []byte(msg), 5*time.Second)
		if err != nil {
			env.SetOutput("response", err.Error())
			return fmt.Errorf("publish Request failed: %s", err)
		}

		if !jsm.IsOKResponse(resp) {
			env.SetOutput("response", string(resp.Data))
			return fmt.Errorf("publish failed: %s", string(resp.Data))
		}

		log.Println(string(resp.Data))

		env.SetOutput("response", string(resp.Data))

		return nil
	}

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

	nc, err := connect()
	if err != nil {
		return err
	}
	log.Printf("Connected to %s", nc.ConnectedUrl())

	str, err := jsm.LoadStream(stream, jsm.WithConnection(nc))
	if err != nil {
		return err
	}

	err = str.UpdateConfiguration(cfg)
	if err != nil {
		return err
	}

	str.Reset()

	cj, err = json.Marshal(str.Configuration())
	if err != nil {
		return err
	}
	env.SetOutput("config", string(cj))

	log.Printf("Configuration: %s", string(cj))

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

	nc, err := connect()
	if err != nil {
		return err
	}
	log.Printf("Connected to %s", nc.ConnectedUrl())

	known, err := jsm.IsKnownStream(stream, jsm.WithConnection(nc))
	if err != nil {
		return err
	}

	if missingok && !known {
		log.Printf("Stream %s was not present", stream)
		return nil
	}

	if !known {
		return fmt.Errorf("stream %s does not exist", stream)
	}

	known, err = jsm.IsKnownConsumer(stream, consumer, jsm.WithConnection(nc))
	if err != nil {
		return err
	}

	if missingok && !known {
		log.Printf("Consumer %s > %s was not present", stream, consumer)
		return nil
	}

	if !known {
		return fmt.Errorf("consumer %s > %s does not exist", stream, consumer)
	}

	cons, err := jsm.LoadConsumer(stream, consumer, jsm.WithConnection(nc))
	if err != nil {
		return err
	}

	return cons.Delete()
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

	nc, err := connect()
	if err != nil {
		return err
	}
	log.Printf("Connected to %s", nc.ConnectedUrl())

	known, err := jsm.IsKnownStream(stream, jsm.WithConnection(nc))
	if err != nil {
		return err
	}

	if missingok && !known {
		log.Printf("Stream %s was not present", stream)
		return nil
	}

	if !known {
		return fmt.Errorf("stream %s does not exist, cannot delete it", stream)
	}

	str, err := jsm.LoadStream(stream, jsm.WithConnection(nc))
	if err != nil {
		return err
	}

	return str.Delete()
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

	nc, err := connect()
	if err != nil {
		return err
	}
	log.Printf("Connected to %s", nc.ConnectedUrl())

	stream, err := jsm.NewStreamFromDefault(cfg.Name, cfg, jsm.StreamConnection(jsm.WithConnection(nc)))
	if err != nil {
		return err
	}

	cj, err = json.Marshal(stream.Configuration())
	if err != nil {
		return err
	}
	env.SetOutput("config", string(cj))

	log.Printf("Configuration: %s", string(cj))

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

	nc, err := connect()
	if err != nil {
		return err
	}
	log.Printf("Connected to %s", nc.ConnectedUrl())

	consumer, err := jsm.NewConsumerFromDefault(stream, cfg, jsm.ConsumerConnection(jsm.WithConnection(nc)))
	if err != nil {
		return err
	}

	cj, err = json.Marshal(consumer.Configuration())
	if err != nil {
		return err
	}
	env.SetOutput("config", string(cj))

	log.Printf("Configuration: %s", string(cj))

	return nil
}

func handleValidateStreamConfig() error {
	cfg := &api.StreamConfig{}
	cfile, valid, errs, err := validateHelper("CONFIG", cfg)
	if err != nil {
		return err
	}

	if valid {
		log.Printf("%s is a valid Stream Configuration", cfile)
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
		log.Printf("%s is a valid Consumer Configuration", cfile)
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

func connect() (*nats.Conn, error) {
	creds := env.GetInput("CREDENTIALS")
	user := env.GetInput("USERNAME")
	pass := env.GetInput("PASSWORD")

	server := env.GetInput("SERVER")
	if server == "" {
		return nil, fmt.Errorf("SERVER is required")
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

	return nats.Connect(server, opts...)
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

func runAction() {
	command := env.GetInput("COMMAND")
	env.Debugf("Running command: %s", command)

	cmd, ok := commands[command]
	if !ok {
		env.Fatalf("Unknown command %s", command)
	}

	err := cmd()
	if err != nil {
		env.Fatalf("Could not run command %s: %s", command, err)
	}
}
