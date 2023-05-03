package loader

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestMergeSuccess(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		override string
		expected map[string]any
	}{
		{
			"empty source & override",
			"",
			"",
			map[string]any{},
		},
		{
			"empty source",
			"",
			`
x: 123
y: [123]
z:
  nested: x
`,
			map[string]any{},
		},
		{
			"empty override",
			`
x: 123
y: [123]
z:
  nested: x
`,
			"",
			map[string]any{
				"x": int64(123),
				"y": []any{int64(123)},
				"z": map[string]any{
					"nested": "x",
				},
			},
		},
		{
			"override single value",
			`
x: 123
y: [123]
z:
  nested: x
`,
			`
x: 321
`,
			map[string]any{
				"x": int64(321),
				"y": []any{int64(123)},
				"z": map[string]any{
					"nested": "x",
				},
			},
		},
		{
			"override nested value",
			`
x: 123
y: [123]
z:
  nested: x
`,
			`
z:
  nested: z
`,
			map[string]any{
				"x": int64(123),
				"y": []any{int64(123)},
				"z": map[string]any{
					"nested": "z",
				},
			},
		},
		{
			"attribute added by override is ignored",
			`
x: 123
y: [123]
z:
  nested: x
`,
			`
z:
  nested2: z
`,
			map[string]any{
				"x": int64(123),
				"y": []any{int64(123)},
				"z": map[string]any{
					"nested": "x",
				},
			},
		},
		{
			"complex scenario",
			`
a: 123
b: [1, 2, 3, 4]
c:
  c1: x
  c2: 1
  c3:
    c31: 1
    c32: 1
  c4:
    c41: true
    c42: 123.321
    c43: [1.1, 2.2]
`,
			`
a: 123
b: [1, 2, 3, 4]
c:
  c1: z
  c2: 1
  c3:
    c31: 2
  c4:
    c41: false
    c42: 321.123
    c43: [3.3]
    cxx: 1
`,
			map[string]any{
				"a": int64(123),
				"b": []any{int64(1), int64(2), int64(3), int64(4)},
				"c": map[string]any{
					"c1": "z",
					"c2": int64(1),
					"c3": map[string]any{
						"c31": int64(2),
						"c32": int64(1),
					},
					"c4": map[string]any{
						"c41": false,
						"c42": 321.123,
						"c43": []any{3.3},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualYaml, err := mergeYamls([]byte(test.source), []byte(test.override))
			var actual map[string]any
			errUnmarshal := yaml.Unmarshal(actualYaml, &actual)
			require.NoError(t, errUnmarshal)

			if assert.NoError(t, err) {
				assert.EqualValues(t, test.expected, actual)
			}
		})
	}
}
