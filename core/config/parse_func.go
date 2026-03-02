package config

import (
	"fmt"
	"xfx/pkg/log"
)

// Parse 使用 registry 中的注册信息解析所有 JSON 数据。
// JSON 文件无对应注册 → 打印警告（可能是废弃文件）。
// 已注册的配置无对应 JSON → panic（开服前发现问题）。
func Parse(data map[string]any) {
	for jsonName, raw := range data {
		parser, ok := registry[jsonName]
		if !ok {
			log.Warn("config: JSON %q has no registered parser, skipped", jsonName)
			continue
		}
		CfgMgr.AllJson[jsonName] = parser(raw, jsonName)
	}

	missing := make([]string, 0)
	for name := range registry {
		if _, ok := CfgMgr.AllJson[name]; !ok {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		panic(fmt.Sprintf("config: registered configs not found in JSON directory: %v", missing))
	}
}
