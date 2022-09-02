package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

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

func (g github) Fatalf(msg string, args ...any) {
	gha.Fatalf(msg, args...)
}

func (g github) Debugf(msg string, args ...any) {
	gha.Debugf(msg, args...)
}

func (g github) Printf(msg string, args ...any) {
	f := bufio.NewWriter(os.Stdout)
	defer f.Flush()

	fmt.Fprintf(f, time.Now().Format("15:04:05.000000")+": "+msg+"\n", args...)
}
