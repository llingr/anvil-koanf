// SPDX-FileCopyrightText: Copyright (c) 2026 The anvil Authors
// SPDX-License-Identifier: Apache-2.0

package conf

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/v2"
	"github.com/llingr/anvil"
)

// appPath is the files' top-level key the provider maps, and the APP_ prefix of the variables
const appPath = "app"

type provider[C any] struct {
	files  fs.FS
	loaded func(ctx context.Context, config C) // nil without OnLoaded
}

// NewProvider maps the app key of files, overlaid by APP_* environment variables through Load,
// onto a C; a key C has no field for fails the load. It logs nothing itself: OnLoaded hands the
// result back to the application.
func NewProvider[C any](files fs.FS, opts ...ProviderOption[C]) anvil.ConfigProvider[C] {
	p := &provider[C]{
		files: files,
	}
	for _, option := range opts {
		option(p)
	}
	return p
}

func (p *provider[C]) Load(ctx context.Context) (C, error) {
	var none C
	kfg, err := Load(p.files)
	if err != nil {
		return none, fmt.Errorf("read config files and environment - %w", err)
	}
	var config C
	err = kfg.UnmarshalWithConf(appPath, &config, koanf.UnmarshalConf{DecoderConfig: strictDecoding()})
	if err != nil {
		return none, fmt.Errorf("map %s - %s", appPath, strings.Join(problems(err), "; "))
	}
	if p.loaded != nil {
		p.loaded(ctx, config)
	}
	return config, nil
}

// strictDecoding is koanf's default decoding, except that a key C has no field for is an error
func strictDecoding() *mapstructure.DecoderConfig {
	return &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.TextUnmarshallerHookFunc()),
		WeaklyTypedInput: true,
		ErrorUnused:      true,
	}
}

// problems lists each mapping problem, which the decoder joins into one multi-line error
func problems(err error) []string {
	var joined interface{ Unwrap() []error }
	if !errors.As(err, &joined) {
		return []string{err.Error()}
	}
	var each []string
	for _, problem := range joined.Unwrap() {
		each = append(each, problems(problem)...)
	}
	return each
}

// ProviderOption adjusts NewProvider
type ProviderOption[C any] func(*provider[C])

// OnLoaded calls loaded inline in Load once the configuration has mapped, so the application decides
// what of it to log, through the logger loaded captures. Given twice, or given nil, it panics.
func OnLoaded[C any](loaded func(ctx context.Context, config C)) ProviderOption[C] {
	if loaded == nil {
		panic("conf: OnLoaded given a nil callback")
	}
	return func(p *provider[C]) {
		if p.loaded != nil {
			panic("conf: OnLoaded given twice")
		}
		p.loaded = loaded
	}
}
