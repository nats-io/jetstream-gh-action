package main

import (
	"io/ioutil"
	"path/filepath"

	gha "github.com/sethvargo/go-githubactions"
)

type github struct{}

func (g github) GetInput(k string) string {
	return gha.GetInput(k)
}

func (g github) SetOutput(k string, v string) error {
	return ioutil.WriteFile(filepath.Join("/worksspace/output", k, "output.txt"), []byte(v), 0644)
}

func (g github) Fatalf(msg string, args ...interface{}) {
	gha.Fatalf(msg, args...)
}

func (g github) Debugf(msg string, args ...interface{}) {
	gha.Debugf(msg, args...)
}
