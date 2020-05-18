package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type stdout struct{}

func (s stdout) GetInput(k string) string {
	return os.Getenv(strings.ToUpper(k))
}

func (s stdout) SetOutput(k string, v string) error {
	return ioutil.WriteFile(filepath.Join("/workspace/output", k, "output.txt"), []byte(v), 0644)
}

func (s stdout) Fatalf(msg string, args ...interface{}) {
	log.Fatalf(msg, args...)
}

func (s stdout) Debugf(msg string, args ...interface{}) {
	log.Printf(msg, args...)
}

func (s stdout) Printf(msg string, args ...interface{}) {
	log.Printf(msg, args...)
}
