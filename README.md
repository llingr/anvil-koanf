# anvil-koanf

An [anvil](https://github.com/llingr/anvil) config provider for [koanf](https://github.com/knadh/koanf):
YAML files under an `app` key, overlaid by `APP_*` environment variables, mapped onto the service's
config type.

## Getting Started

```sh
go get github.com/llingr/anvil-koanf
```

```yaml
app:
  server:
    port: 8080
    readHeaderTimeout: 5s
```

```go
type Config struct {
    Server struct {
        Port              int           `koanf:"port"`
        ReadHeaderTimeout time.Duration `koanf:"readHeaderTimeout"`
    } `koanf:"server"`
}

//go:embed *.yaml
var configFiles embed.FS

configProvider := conf.NewProvider[Config](configFiles)
```

## Features

- **Strict mapping.** A key with no matching field, such as a misspelling, fails the load and is
  named, rather than leaving the field at its zero value.
- **Environment overrides.** `APP_SERVER_PORT` overrides `app.server.port`. A variable naming no key
  is ignored, so the files stay the complete list of settings.
- **Embedded files.** The `*.yaml` files at the root of an `fs.FS`: `go:embed` chooses the files,
  `fs.Sub` the directory.
- **Safe config logging.** `conf.OnLoaded` hands the loaded config to a callback, such as
  [anvil-zap](https://github.com/llingr/anvil-zap)'s `zaplog.LogConfig`, which logs only the fields
  the config's `MarshalLogObject` names.
- **Example.** `go run ./example` serves HTTP on the port in its `config.yaml`;
  `APP_SERVER_PORT=9090 go run ./example` overrides it.
