package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/chinamobile/nlpt/cmd/apiserver/app"

	"k8s.io/component-base/logs"
)

func main() {
	command := app.NewServerCommand()

	// Important!!!
	rand.Seed(time.Now().UnixNano())

	logs.InitLogs()
	defer logs.FlushLogs()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
