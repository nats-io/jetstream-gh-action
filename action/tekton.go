package main

import (
	"log"
	"os"
	"strings"

	gha "github.com/sethvargo/go-githubactions"
)

type tekton struct{}

func (t tekton) GetInput(k string) string {
	return os.Getenv(strings.ToUpper(k))
}

func (t tekton) SetOutput(k string, v string) error {
	gha.SetOutput(k, v)
	return nil
}

func (t tekton) Fatalf(msg string, args ...interface{}) {
	log.Fatalf(msg, args...)
}

func (t tekton) Debugf(msg string, args ...interface{}) {
	log.Printf(msg, args...)
}
