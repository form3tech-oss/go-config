package loader_test

import (
	"os"
	"strings"
	"testing"

	"github.com/form3tech-oss/go-config/loader"
	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/require"
)

type CType struct {
	Val string
}

type Test struct {
	A string
	B string
	C CType
}

type CleanupFn func()

func initVault(t *testing.T) (*api.Client) {
	t.Helper()

	cfg := api.DefaultConfig()
	cfg.Address = "http://localhost:8200"

	client, err := api.NewClient(cfg)
	require.Nil(t, err)

	client.SetToken("dev-token")

	return client
}

func TestConfigLoader_AppendConfig(t *testing.T) {
	defaultCfgJson := `
{
  "a": "default_a",
  "b": "default_b",
  "c": {
    "val": "default_c"
  }
}
`

	tests := []struct {
		name     string
		inputs   [][]string
		setupFn  func(*testing.T) (*loader.ConfigLoader, CleanupFn)
		expected Test
	}{
		{
			"two simple overrides",
			[][]string{
				// JSON config
				{defaultCfgJson, "json"},

				// Yaml config that overrides b
				{
					`
b: override_b
`, "yaml",
				},

				// Yaml config that overrides c
				{
					`
c:
  val: override_c
`, "yaml",
				},
			},
			func(*testing.T) (*loader.ConfigLoader, CleanupFn) {
				return loader.NewConfigLoader(), func() {}
			},
			Test{
				A: "default_a",
				B: "override_b",
				C: CType{Val: "override_c"},
			},
		},
		{
			"overrides with env",
			[][]string{
				// default json config
				{defaultCfgJson, "json"},

				// yaml config that overrides a with env
				{
					`
a: {{ env "A_VAL" }}
`, "yaml",
				},

				// yaml config that overrides b with env, but takes default value
				{
					`
b: {{ env "B_VAL" "override_b_env_default" }}
`, "yaml",
				},
			},
			func(*testing.T) (*loader.ConfigLoader, CleanupFn) {
				require.Nil(t, os.Setenv("A_VAL", "override_a_env"))
				return loader.NewConfigLoader(), func() {
					require.Nil(t, os.Unsetenv("A_VAL"))
				}
			},
			Test{
				A: "override_a_env",
				B: "override_b_env_default",
				C: CType{Val: "default_c"},
			},
		},
		{
			"overrides with vault",
			[][]string{
				// default json config
				{defaultCfgJson, "json"},

				// yaml config that overrides a with env
				{
					`
a: {{ vault "secret/test" "A_VAL" }}
`, "yaml",
				},

				// yaml config that overrides b with env, but takes default value
				{
					`
b: {{ vault "secret/test" "B_VAL" "override_b_vault_default" }}
`, "yaml",
				},
			},
			func(*testing.T) (*loader.ConfigLoader, CleanupFn) {
				vc := initVault(t)
				_, err := vc.Logical().Write("secret/test", map[string]interface{}{
					"A_VAL": "override_a_vault",
				})
				require.Nil(t, err)

				return loader.NewConfigLoader(loader.WithVaultClient(vc)), func() {}
			},
			Test{
				A: "override_a_vault",
				B: "override_b_vault_default",
				C: CType{Val: "default_c"},
			},
		},
		{
			"overrides with env custom delimiters",
			[][]string{
				// default json config
				{defaultCfgJson, "json"},

				// yaml config that overrides a with env, use [[ ]] as delimiters
				{
					`
a: [[ env "A_VAL" ]]
`, "yaml",
				},
			},
			func(*testing.T) (*loader.ConfigLoader, CleanupFn) {
				require.Nil(t, os.Setenv("A_VAL", "override_a_env"))
				return loader.NewConfigLoader(loader.WithDelimiters("[[", "]]")), func() {
					require.Nil(t, os.Unsetenv("A_VAL"))
				}
			},
			Test{
				A: "override_a_env",
				B: "default_b",
				C: CType{Val: "default_c"},
			},
		},
		{
			"overrides with custom template func",
			[][]string{
				// default json config
				{defaultCfgJson, "json"},

				// yaml config that overrides a with env, register strings.ToUpper as uppercase func in template
				{
					`
a: {{ uppercase "override_a" }}
`, "yaml",
				},
			},
			func(*testing.T) (*loader.ConfigLoader, CleanupFn) {
				return loader.NewConfigLoader(loader.WithCustomTemplateFunc("uppercase", strings.ToUpper)), func() {}
			},
			Test{
				A: "OVERRIDE_A",
				B: "default_b",
				C: CType{Val: "default_c"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create a ConfigLoader, setup env var and vault secret if needed
			cl, cleanupFn := test.setupFn(t)
			defer cleanupFn()

			// Append input one by one
			for _, input := range test.inputs {
				err := cl.AppendConfig(input[0], input[1])
				require.Nil(t, err)
			}

			// Unmarshal into concrete type
			var actual Test
			err := cl.Unmarshal(&actual)
			require.Nil(t, err)

			// Compare with expected value
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestConfigLoader_LoadConfigFiles(t *testing.T) {
	require.Nil(t, os.Setenv("B_VAL", "override_b"))

	cl := loader.NewConfigLoader()
	err := cl.LoadConfigFiles(
		"tests/default.yaml",
		"tests/override.json",
	)
	require.Nil(t, err)

	var actual Test
	err = cl.Unmarshal(&actual)
	require.Nil(t, err)

	expected := Test{
		A: "default_a",
		B: "override_b",
		C: CType{
			Val: "override_c",
		},
	}
	require.Equal(t, expected, actual)
}
