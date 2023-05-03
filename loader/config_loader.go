// Package loader provides an easy way to load and merge configuration files and decode them into one single
// configuration struct.
package loader

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/hashicorp/vault/api"
	"gopkg.in/yaml.v3"
)

type ConfigLoader struct {
	funcMap      template.FuncMap
	vaultClient  *api.Client
	delimiters   []string
	mergedConfig []byte
}

// NewConfigLoader creates a new ConfigLoader instances `env` function pre-loaded.
func NewConfigLoader(opts ...OptFunc) *ConfigLoader {
	cl := &ConfigLoader{}
	cl.funcMap = template.FuncMap{
		"env": cl.env,
	}
	for _, opt := range opts {
		opt(cl)
	}
	return cl
}

type OptFunc func(*ConfigLoader)

// WithVaultClient enables vault function to fetch secrets from Hashicorp Vault.
// Example: {{ vault "secrets/path" "bar_key" "default_value" }}.
func WithVaultClient(vaultClient *api.Client) OptFunc {
	return func(cl *ConfigLoader) {
		cl.vaultClient = vaultClient
		cl.funcMap["vault"] = cl.vault
	}
}

// WithDelimiters changes the default template delimiters {{ }}.
func WithDelimiters(left, right string) OptFunc {
	return func(cl *ConfigLoader) {
		cl.delimiters = []string{left, right}
	}
}

// WithCustomTemplateFunc registers a custom function to use with the template, similar to `env` and `vault`.
func WithCustomTemplateFunc(name string, fn interface{}) OptFunc {
	return func(cl *ConfigLoader) {
		cl.funcMap[name] = fn
	}
}

// LoadConfigFiles loads and parses one or more config files. The latter files will merge with previous ones in order.
// Should be followed by Unmarshal to create a struct out of the config.
func (cl *ConfigLoader) LoadConfigFiles(fileNames ...string) error {
	for _, fileName := range fileNames {
		rawFile, err := os.ReadFile(fileName)
		if err != nil {
			return fmt.Errorf("failed to read config file '%s': %w", fileName, err)
		}

		ext := filepath.Ext(fileName)
		ext = strings.TrimPrefix(ext, ".")

		if err := cl.AppendConfig(string(rawFile)); err != nil {
			return fmt.Errorf("failed to append config file '%s': %w", fileName, err)
		}
	}

	return nil
}

// AppendConfig appends a plain string config (similar to a config file but in string), parses and merges it with
// current config.
func (cl *ConfigLoader) AppendConfig(config string) error {

	tmpl := template.New("").Funcs(cl.funcMap)
	if cl.delimiters != nil && len(cl.delimiters) == 2 {
		tmpl = tmpl.Delims(cl.delimiters[0], cl.delimiters[1])
	}

	tmpl, err := tmpl.Parse(config)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, nil); err != nil {
		return fmt.Errorf("failed to render config: %w", err)
	}

	b := buf.Bytes()
	//TODO szczygi, probably should validate yaml
	if cl.mergedConfig == nil {
		cl.mergedConfig = b
	} else {
		res, err := mergeYamls(cl.mergedConfig, b)
		if err != nil {
			return err
		}
		cl.mergedConfig = res
	}

	return nil
}

// Unmarshal unmarshals current config into a struct.
func (cl *ConfigLoader) Unmarshal(v interface{}) error {
	if err := yaml.Unmarshal(cl.mergedConfig, v); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

func (cl *ConfigLoader) env(envName string, defaultVal ...string) string {
	val, isSet := os.LookupEnv(envName)
	if !isSet && len(defaultVal) > 0 {
		val = defaultVal[0]
	}
	return val
}

func (cl *ConfigLoader) vault(path string, key string, defaultVal ...string) (string, error) {
	secrets, err := cl.vaultClient.Logical().Read(path)
	if err != nil {
		return "", fmt.Errorf("failed to read secrets from vault path '%s': %w", path, err)
	}

	if secrets != nil {
		val, ok := secrets.Data[key]
		if ok {
			return fmt.Sprintf("%v", val), nil
		}
	}

	if len(defaultVal) > 0 {
		return defaultVal[0], nil
	}
	return "", fmt.Errorf("vault: key '%s' does not exist in '%s' and no default value has been provided", key, path)
}
