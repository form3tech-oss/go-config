package decodehooks_test

import (
	"os"
	"testing"

	"github.com/form3tech-oss/go-config/loader"
	"github.com/form3tech-oss/go-config/loader/decodehooks"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringToSetHookFunc(t *testing.T) {
	require.Nil(t, os.Setenv("COMMA_SEPARATED_STRING", "v1, v2,v3,,,"))

	type TestStruct struct {
		Value map[string]struct{}
	}

	configYaml := `value: {{ env "COMMA_SEPARATED_STRING" }}`

	cl := loader.NewConfigLoader()
	err := cl.AppendConfig(configYaml, "yaml")
	require.Nil(t, err)

	result := TestStruct{}
	err = cl.Unmarshal(&result, viper.DecodeHook(decodehooks.StringToSetHookFunc(",")))
	require.Nil(t, err)

	assert.Len(t, result.Value, 3)
	assert.Contains(t, result.Value, "v1")
	assert.Contains(t, result.Value, "v2")
	assert.Contains(t, result.Value, "v3")
}
