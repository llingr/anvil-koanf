// SPDX-FileCopyrightText: Copyright (c) 2026 The anvil Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"embed"
	"time"
)

//go:embed config.yaml
var configFiles embed.FS

// Config is every setting, by config.yaml section under app
type Config struct {
	Server Server `koanf:"server"`
}

// Server path: `app.server.`
type Server struct {
	Port              int           `koanf:"port"`
	ReadHeaderTimeout time.Duration `koanf:"readHeaderTimeout"`
}
