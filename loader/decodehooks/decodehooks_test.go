package decodehooks_test

import (
	"os"
	"strings"
	"testing"

	"github.com/form3tech-oss/go-config/loader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type StringSet struct {
	x map[string]struct{}
}

func (s *StringSet) UnmarshalYAML(value *yaml.Node) error {
	if value.Value == "" {
		s.x = map[string]struct{}{}
		return nil
	}
	s.x = map[string]struct{}{}

	split := strings.Split(value.Value, ",")

	for _, v := range split {
		if key := strings.TrimSpace(v); len(key) > 0 {
			s.x[key] = struct{}{}
		}
	}
	return nil
}

func TestStringToSetHookFunc(t *testing.T) {
	require.Nil(t, os.Setenv("COMMA_SEPARATED_STRING", "v1, v2,v3,,,"))

	type TestStruct struct {
		Value StringSet
	}

	configYaml := `value: {{ env "COMMA_SEPARATED_STRING" }}`

	cl := loader.NewConfigLoader()
	err := cl.AppendConfig(configYaml)
	require.Nil(t, err)

	result := TestStruct{}
	err = cl.Unmarshal(&result)
	require.Nil(t, err)

	assert.Len(t, result.Value.x, 3)
	assert.Contains(t, result.Value.x, "v1")
	assert.Contains(t, result.Value.x, "v2")
	assert.Contains(t, result.Value.x, "v3")
}
