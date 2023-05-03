package loader

import (
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

type ErrUnsupportedType struct {
	Type reflect.Type
}

func (e ErrUnsupportedType) Error() string {
	return fmt.Sprintf("unsupported type: %s", e.Type.String())
}

func mergeYamls(sourceYaml, overrideYaml []byte) ([]byte, error) {
	var source map[string]any
	var override map[string]any
	if err := yaml.Unmarshal(sourceYaml, &source); err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(overrideYaml, &override); err != nil {
		return nil, err
	}

	result, err := mergeYamlMaps(source, override)
	if err != nil {
		return nil, err
	}

	res, err := yaml.Marshal(result)
	if err != nil {
		return nil, err
	}

	return res, nil
}
func mergeYamlMaps(source, override map[string]any) (map[string]any, error) {

	result := make(map[string]any)

	for k, v := range source {

		if mv, ok := v.(map[string]any); ok {
			var ov map[string]any
			ovv, ok := override[k].(map[string]any)
			if ok {
				ov = ovv
			}
			merged, err := mergeYamlMaps(mv, ov)
			if err != nil {
				return nil, err
			}
			result[k] = merged
		} else {
			var vof any
			ov, ok := override[k]
			if ok {
				vof = ov
			} else {
				vof = v
			}

			result[k] = vof
		}
	}

	return result, nil
}
