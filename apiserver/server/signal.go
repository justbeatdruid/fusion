package server

import (
	"syscall"

	"os"
	"os/signal"

	"k8s.io/klog"
)

var onlyOneSignalHandler = make(chan struct{})
var shutdownHandler chan os.Signal
var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}

func SetupSignalHandler() <-chan struct{} {
	close(onlyOneSignalHandler) // panics when called twice

	shutdownHandler = make(chan os.Signal, 2)

	stop := make(chan struct{})
	signal.Notify(shutdownHandler, shutdownSignals...)
	go func() {
		sig := <-shutdownHandler
		klog.Infof("receive shutdown signal fisrt time: %s", sig.String())
		close(stop)
		sig = <-shutdownHandler
		klog.Infof("receive shutdown signal second time: %s, will exit directly", sig.String())
		os.Exit(1) // second signal. Exit directly.
	}()

	return stop
}

func RequestShutdown() bool {
	if shutdownHandler != nil {
		select {
		case shutdownHandler <- shutdownSignals[0]:
			return true
		default:
		}
	}

	return false
}
