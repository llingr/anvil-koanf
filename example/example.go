// SPDX-FileCopyrightText: Copyright (c) 2026 The anvil Authors
// SPDX-License-Identifier: Apache-2.0

// An HTTP server on anvil with its configuration from conf.NewProvider: config.yaml embedded, overlaid
// by APP_* environment variables, with an OnLoaded callback logging the one setting it chooses. Its
// logging is the standard library's log package, standing in for a real logger provider such as
// anvil-zap's.
//
//	go run ./example
//	APP_SERVER_PORT=9090 go run ./example
package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/llingr/anvil"
	"github.com/llingr/anvil-koanf/conf"
	"github.com/llingr/anvil/shutdown"
)

// Shell keeps the two type parameters out of every function that takes the shell
type Shell = anvil.Shell[Config, *log.Logger]

func main() {
	loggerProvider := newStdLogging()
	logger := loggerProvider.Logger()
	configProvider := conf.NewProvider(configFiles,
		conf.OnLoaded(func(_ context.Context, config Config) {
			logger.Printf("configuration: port %d", config.Server.Port)
		}))
	exitCode := anvil.Run(context.Background(), "anvil-koanf-example", configProvider, loggerProvider, wire)
	os.Exit(exitCode)
}

// wire serves HTTP on the configured port, shut down in Ingress
func wire(ctx context.Context, shell Shell) error {
	logger := shell.Logger()
	config := shell.Config()
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(config.Server.Port)))
	if err != nil {
		return err
	}
	server := &http.Server{
		Handler: http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			_, _ = writer.Write([]byte("hello from anvil\n"))
		}),
		ReadHeaderTimeout: config.Server.ReadHeaderTimeout,
	}
	shell.Go(shutdown.Ingress, "http server", func(context.Context) error {
		if serveErr := server.Serve(listener); !errors.Is(serveErr, http.ErrServerClosed) {
			return serveErr
		}
		return nil
	})
	logger.Print("serving on ", listener.Addr())
	shell.RegisterShutdownHandler(shutdown.Ingress, "http shutdown", server.Shutdown)
	return nil
}
