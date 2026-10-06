// SPDX-FileCopyrightText: Copyright (c) 2026 The anvil Authors
// SPDX-License-Identifier: Apache-2.0

package conf_test

import (
	"context"
	"log/slog"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/llingr/anvil-koanf/conf"
)

type providerConfig struct {
	Server struct {
		Port int `koanf:"port"`
	} `koanf:"server"`
}

func providerFiles(document string) fstest.MapFS {
	return fstest.MapFS{
		"config.yaml": &fstest.MapFile{Data: []byte(document)},
	}
}

// The app key maps onto C, and the environment overlays it
func TestProviderMapsAppKey(t *testing.T) {
	t.Setenv("APP_SERVER_PORT", "9090")
	provider := conf.NewProvider[providerConfig](providerFiles("app:\n  server:\n    port: 8080\n"))
	config, err := provider.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if config.Server.Port != 9090 {
		t.Fatalf("port %d, want the variable's 9090", config.Server.Port)
	}
}

// OnLoaded receives the mapped config, and Load's ctx, once the mapping has succeeded
func TestProviderOnLoaded(t *testing.T) {
	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "load ctx")
	var gotCtx any
	var gotConfig providerConfig
	provider := conf.NewProvider(providerFiles("app:\n  server:\n    port: 8080\n"),
		conf.OnLoaded(func(ctx context.Context, config providerConfig) {
			gotCtx = ctx.Value(ctxKey{})
			gotConfig = config
		}))
	config, err := provider.Load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if gotCtx != "load ctx" || !reflect.DeepEqual(gotConfig, config) || gotConfig.Server.Port != 8080 {
		t.Fatalf("OnLoaded got ctx value %v and config %+v", gotCtx, gotConfig)
	}
}

// Files that cannot be read, here none matching *.yaml, fail the load before anything maps
func TestProviderReadError(t *testing.T) {
	provider := conf.NewProvider[providerConfig](fstest.MapFS{})
	config, err := provider.Load(context.Background())
	if err == nil || !strings.Contains(err.Error(), "read config files and environment") {
		t.Fatalf("error %v, want the read failure", err)
	}
	if !reflect.DeepEqual(config, providerConfig{}) {
		t.Fatalf("config %+v, want the zero value", config)
	}
}

// A config the files cannot map onto returns the zero C, names the key, and skips OnLoaded
func TestProviderMappingError(t *testing.T) {
	called := false
	provider := conf.NewProvider(providerFiles("app:\n  server: eighty\n"),
		conf.OnLoaded(func(context.Context, providerConfig) {
			called = true
		}))
	config, err := provider.Load(context.Background())
	if err == nil || !strings.Contains(err.Error(), "map app") {
		t.Fatalf("error %v, want one naming the app key", err)
	}
	if !reflect.DeepEqual(config, providerConfig{}) || called {
		t.Fatalf("config %+v, OnLoaded called %v, want the zero value and no call", config, called)
	}
}

// A key C has no field for, such as a misspelling, fails the load and is named, rather than leaving
// the field at its zero value
func TestProviderUnknownKeyFails(t *testing.T) {
	provider := conf.NewProvider[providerConfig](providerFiles("app:\n  server:\n    port: 8080\n    prot: 9090\n"))
	config, err := provider.Load(context.Background())
	if err == nil || !strings.Contains(err.Error(), "prot") {
		t.Fatalf("error %v, want one naming the unknown key prot", err)
	}
	if !reflect.DeepEqual(config, providerConfig{}) {
		t.Fatalf("config %+v, want the zero value", config)
	}
}

// Strict decoding keeps koanf's conversions: a duration from its text, and a TextUnmarshaler
func TestProviderDecodesDurationAndText(t *testing.T) {
	type textConfig struct {
		Timeout time.Duration `koanf:"timeout"`
		Level   slog.Level    `koanf:"level"`
	}
	provider := conf.NewProvider[textConfig](providerFiles("app:\n  timeout: 5s\n  level: warn\n"))
	config, err := provider.Load(context.Background())
	if err != nil || config.Timeout != 5*time.Second || config.Level != slog.LevelWarn {
		t.Fatalf("config %+v, error %v, want 5s and WARN", config, err)
	}
}

// Every mapping problem is reported, on one line, so a log line stays whole
func TestProviderMappingErrorsOnOneLine(t *testing.T) {
	provider := conf.NewProvider[providerConfig](providerFiles("app:\n  server:\n    port: eighty\n    prot: 9090\n"))
	_, err := provider.Load(context.Background())
	if err == nil || strings.Contains(err.Error(), "\n") || !strings.Contains(err.Error(), "prot") || !strings.Contains(err.Error(), "port") {
		t.Fatalf("error %q, want both problems on one line", err)
	}
}

// OnLoaded given nil or twice panics at construction
func TestProviderOnLoadedMisuse(t *testing.T) {
	expectPanic := func(name string, do func()) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Errorf("%s did not panic", name)
			}
		}()
		do()
	}
	loaded := func(context.Context, providerConfig) {
	}
	expectPanic("nil", func() {
		conf.OnLoaded[providerConfig](nil)
	})
	expectPanic("twice", func() {
		conf.NewProvider(providerFiles(""), conf.OnLoaded(loaded), conf.OnLoaded(loaded))
	})
}
