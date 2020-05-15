package main

import (
	"log"

	gha "github.com/sethvargo/go-githubactions"
)

type github struct{}

func (g github) GetInput(k string) string {
	return gha.GetInput(k)
}

func (g github) SetOutput(k string, v string) error {
	gha.SetOutput(k, v)
	return nil
}

func (g github) Fatalf(msg string, args ...interface{}) {
	gha.Fatalf(msg, args...)
}

func (g github) Debugf(msg string, args ...interface{}) {
	gha.Debugf(msg, args...)
}

func (g github) Printf(msg string, args ...interface{}) {
	log.Printf(msg, args...)
}
