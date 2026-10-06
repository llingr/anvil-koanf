// SPDX-FileCopyrightText: Copyright (c) 2026 The anvil Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"log"
	"os"

	"github.com/llingr/anvil"
)

type stdLogging struct {
	logger *log.Logger
}

// newStdLogging is a logger provider on the standard library's log package, UTC timestamps to stderr
func newStdLogging() anvil.LoggerProvider[*log.Logger] {
	return &stdLogging{
		logger: log.New(os.Stderr, "", log.LstdFlags|log.LUTC),
	}
}

func (s *stdLogging) Logger() *log.Logger {
	return s.logger
}

func (s *stdLogging) LifecycleInfo(_ context.Context, msg string) {
	s.logger.Print(msg)
}

func (s *stdLogging) LifecycleError(_ context.Context, msg string, err error) {
	s.logger.Printf("%s: %v", msg, err)
}

// Flush has nothing to write out: log writes each line straight through
func (s *stdLogging) Flush() {
}
