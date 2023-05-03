package decodehooks

import (
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
)

// TODO szczygi

// StringToSetHookFunc returns a DecodeHookFunc that converts
// string to map[string]struct{} by splitting on the given separator.
func StringToSetHookFunc(separator string) mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {

		if f.Kind() != reflect.String {
			return data, nil
		}

		if t != reflect.TypeOf(map[string]struct{}{}) {
			return data, nil
		}

		raw := data.(string)
		if raw == "" {
			return map[string]struct{}{}, nil
		}

		split := strings.Split(raw, separator)

		out := make(map[string]struct{}, len(split))
		for _, v := range split {
			if key := strings.TrimSpace(v); len(key) > 0 {
				out[key] = struct{}{}
			}
		}

		return out, nil
	}
}
