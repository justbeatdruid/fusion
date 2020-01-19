package server

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/chinamobile/nlpt/apiserver/handler"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/options"

	"github.com/coreos/go-systemd/daemon"

	"k8s.io/apiserver/pkg/server/healthz"
	"k8s.io/klog"
)

type GenericServer struct {
	ListenAddress string

	APIHandler *handler.Handler

	stopped chan struct{}
}

func NewGenericServer(serverRunOptions *options.ServerRunOptions, k8sconfig *config.Config) (*GenericServer, error) {
	server := &GenericServer{
		stopped: make(chan struct{}),
	}
	err := server.setupOptions(serverRunOptions, k8sconfig)
	return server, err
}

func (s *GenericServer) setupOptions(serverRunOptions *options.ServerRunOptions, k8sconfig *config.Config) error {
	if len(serverRunOptions.ListenAddress) == 0 {
		return fmt.Errorf("server listen address is null")
	}
	s.ListenAddress = serverRunOptions.ListenAddress
	s.APIHandler = handler.NewHandler(k8sconfig)
	return nil
}

func (s *GenericServer) Run(stopCh <-chan struct{}, checks ...healthz.HealthChecker) {
	serverStop := make(chan struct{})
	serverIdle := make(chan struct{})
	go s.StartServer(serverStop, serverIdle, checks...)
	daemon.SdNotify(false, "READY=1")
	select {
	case <-stopCh:
		close(serverStop)
	}
	<-serverIdle
	klog.Infof("all goroutines are safely exited, ready to shutdown")
	close(s.stopped)
	os.Exit(0)
}

func (s *GenericServer) StartServer(stopCh <-chan struct{}, idleConnsClosed chan<- struct{}, checks ...healthz.HealthChecker) {
	wsContainer, err := s.APIHandler.CreateHTTPAPIHandler(checks...)
	if err != nil {
		klog.Fatalf("create http handler error: %+v", err)
	}
	server := &http.Server{Addr: s.ListenAddress, Handler: wsContainer}
	go func() {
		select {
		case <-stopCh:
			klog.Infof("rest sever will be stopped because signal received")
		}
		if err := server.Shutdown(context.Background()); err != nil {
			klog.Errorf("rest server shutdown error: %+v", err)
		}
		close(idleConnsClosed)
	}()

	klog.Infof("server is ready to start. listening on %s", s.ListenAddress)
	if err := server.ListenAndServe(); err != nil {
		klog.Errorf("server run error: %+v", err)
	}
}

func (s *GenericServer) WaitStop() {
	<-s.stopped
}
