# go-config

A wrapper around Viper and Go template for declarative configuration. 

## Usage

If you have a default config file and want to override some values with another config file like so:

```yaml
# config/default.yaml
log:
  level: debug
  format: console
```

```yaml
# config/env/default.yaml
log:
  format: json # override log format here
```

Load both files and merge them as follows, the declaration order will be the merging order:

```go
type Config struct {
	Level  string
	Format string
}

cl := NewConfigLoader()
if err := cl.LoadConfigFiles(
	"config/default.yaml", 
	"config/env/dev.yaml",
); err := nil {
	panic(err)
}

var cfg Config
if err := cl.Unmarshal(&cfg); err != nil {
	panic(err)
}
```

### Load values from environment variables

If you need to load secrets from environment variables, use `env` function inside a Go template delimiter.

The default env function is registered as `env` and has the following signature:

```go
func env(envName string, defaultValue ...string) (string, error) {}
```

Example:

```yaml
toto:
  foo: {{ env "FOO" }}
  bar: {{ env "BAR" "default_value" }}
```

### Load values from Hashicorp Vault

If you need to load secrets from Vault, you will need a pre-configured Vault client and
pass it to `WithVaultClient` option.

The default Vault function is registered as `vault` and has the following signature:

```go
func vault(path string, key string, defaultValue ...string) (string, error) {}
```

Example:

```yaml
toto:
  foo: {{ vault "secrets/path" "foo_key" }}
  bar: {{ vault "secrets/path" "bar_key" "default_value" }}
```

```go
cl := NewConfigLoader(WithVaultClient(vaultClient))
```

### Change Go template delimiter

In some case, you don't want to use `{{ }}` as delimiters, for example when using Helm. You can change
the delimiters with `WithDelimiters`.

Example:

```yaml
toto:
  foo: [[ env "FOO_VAL" ]]
```

```go
cl := NewConfigLoader(WithDelimiters("[[", "]]"))
```

### Use custom Go template function

If you need custom template function, you can register one with `WithCustomTemplateFunc`. The function
must return a string and optionally an error.

Example:

```yaml
toto:
  foo: {{ toUpper "bar" }}
```

```go
cl := NewConfigLoader(WithCustomTemplateFunc("toUpper", func(val string) string {
	return strings.ToUpper(val)
}))
```

### Viper

This package is a wrapper around a [Viper](https://github.com/spf13/viper) instance, so if you need extra viper config
or action, you can always get the viper instance:

```go
cl := NewConfigLoader()
viper := cl.Viper()
```

### Caveats

- Because Go template renders values as-is, so it is up to user to handle non-trivial cases such as multi-line string,
  unescaped characters, or indenting. It is just Go template under the hood, so you are free to customise your config
  file with any Go template feature.
- The `vault` function does not cache, so if there are many secrets from the same path, the data will be read multiple
  times.
