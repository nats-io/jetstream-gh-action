package main

import (
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
	return os.WriteFile(filepath.Join("/workspace/output", k, "output.txt"), []byte(v), 0644)
}

func (s stdout) Fatalf(msg string, args ...any) {
	log.Fatalf(msg, args...)
}

func (s stdout) Debugf(msg string, args ...any) {
	log.Printf(msg, args...)
}

func (s stdout) Printf(msg string, args ...any) {
	log.Printf(msg, args...)
}
