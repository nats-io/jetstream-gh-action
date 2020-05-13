package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/nats-io/jsm.go"
	"github.com/nats-io/jsm.go/api"
	"github.com/nats-io/nats.go"
	gha "github.com/sethvargo/go-githubactions"
)

type validatable interface {
	Validate() (bool, []string)
}

func main() {
	var err error

	command := gha.GetInput("COMMAND")
	switch command {
	case "VALIDATE_STREAM_CONFIG":
		err = handleValidateStreamConfig()

	case "VALIDATE_CONSUMER_CONFIG":
		err = handleValidateConsumerConfig()

	case "CREATE_STREAM":
		err = handleCreateStream()

	case "CREATE_CONSUMER":
		err = handleCreateConsumer()

	default:
		err = fmt.Errorf("invalid command '%s'", command)
	}

	if err != nil {
		gha.Fatalf("JetStream Action failed: %s", err)
	}
}

func handleCreateStream() error {
	gha.Group("JetStream Connection")
	nc, err := connect()
	if err != nil {
		return err
	}
	log.Printf("Connected to %s", nc.ConnectedUrl())
	gha.EndGroup()

	gha.Group("Create Stream")
	cfile := gha.GetInput("CONFIG")
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

	stream, err := jsm.NewStreamFromDefault(cfg.Name, cfg, jsm.StreamConnection(jsm.WithConnection(nc)))
	if err != nil {
		return err
	}

	cj, err = json.MarshalIndent(stream.Configuration(), "", "  ")
	if err != nil {
		return err
	}
	gha.SetOutput("config", string(cj))
	gha.Debugf("Created Stream: \n%s", string(cj))

	gha.EndGroup()

	return nil
}

func handleCreateConsumer() error {
	gha.Group("JetStream Connection")
	nc, err := connect()
	if err != nil {
		return err
	}
	log.Printf("Connected to %s", nc.ConnectedUrl())
	gha.EndGroup()

	gha.Group("Create Consumer")
	cfile := gha.GetInput("CONFIG")
	if cfile == "" {
		return fmt.Errorf("CONFIG is required")
	}

	stream := gha.GetInput("STREAM")
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

	consumer, err := jsm.NewConsumerFromDefault(stream, cfg, jsm.ConsumerConnection(jsm.WithConnection(nc)))
	if err != nil {
		return err
	}

	cj, err = json.MarshalIndent(consumer.Configuration(), "", "  ")
	if err != nil {
		return err
	}
	gha.SetOutput("config", string(cj))
	gha.Debugf("Created Consumer: \n%s", string(cj))

	gha.EndGroup()

	return nil
}

func connect() (*nats.Conn, error) {
	creds := gha.GetInput("CREDENTIALS")
	user := gha.GetInput("USERNAME")
	pass := gha.GetInput("PASSWORD")

	server := gha.GetInput("SERVER")
	if server == "" {
		return nil, fmt.Errorf("SERVER is required")
	}

	opts := []nats.Option{
		nats.ErrorHandler(func(_ *nats.Conn, _ *nats.Subscription, err error) {
			gha.Fatalf("NATS Error: %s", err)
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
	cfile := gha.GetInput(input)
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
