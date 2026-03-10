package config

import (
	"fmt"
)

func Parse(data map[string]any, dest map[string]any) error {
	for jsonName, raw := range data {
		parser, ok := registry[jsonName]
		if !ok {
			//log.Warn("config: JSON %q has no registered parser, skipped", jsonName)
			continue
		}
		dest[jsonName] = parser(raw, jsonName)
	}

	missing := make([]string, 0)
	for name := range registry {
		// 【关键改动】检查 dest 中是否有缺失
		if _, ok := dest[name]; !ok {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("config: registered configs not found in JSON directory: %v", missing)
	}

	return nil
}
