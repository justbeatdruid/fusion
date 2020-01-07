package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/chinamobile/nlpt/apiserver/cmd/controller-manager/serviceunit"

	"k8s.io/component-base/logs"
)

func main() {
	// Important!!!
	rand.Seed(time.Now().UnixNano())

	logs.InitLogs()
	defer logs.FlushLogs()

	if err := serviceunit.Run(); err != nil {
		os.Exit(1)
	}
}
