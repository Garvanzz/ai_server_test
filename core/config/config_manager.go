package config

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"xfx/core/config/conf"
)

var CfgMgr *Manager
var excelMap map[string]int

var numPrefixRe = regexp.MustCompile(`^\d+`)

type Manager struct {
	AllJson        map[string]any
	filePath       string
	parserFunc     ParseFunc
	c              chan any
	reloadOnChange bool // TODO: 热更配置
}

type ParseFunc func(map[string]any)

func NewManager(filePath string, reloadOnChange bool, parser ParseFunc) *Manager {
	s := &Manager{
		AllJson:        make(map[string]any),
		filePath:       filePath,
		reloadOnChange: reloadOnChange,
		parserFunc:     parser,
		c:              make(chan interface{}),
	}
	return s
}

func InitConfig(confPath string) {
	CfgMgr = NewManager(confPath, false, Parse)

	if CfgMgr.parserFunc == nil {
		panic("config parser func is nil")
	}

	data := CfgMgr.loadFile()
	CfgMgr.parserFunc(data)
}

func (m *Manager) loadFile() map[string]any {
	files, err := os.ReadDir(m.filePath)
	if err != nil {
		panic(err)
	}

	excelMap = make(map[string]int)
	jsonData := make(map[string]any)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if filepath.Ext(file.Name()) == ".json" {
			data, err := os.ReadFile(filepath.Join(m.filePath, file.Name()))
			if err != nil {
				panic(err)
			}

			excelId := 0
			match := numPrefixRe.FindString(file.Name())
			if match != "" {
				num, _ := strconv.Atoi(match)
				excelId = num
			}

			fileName := strings.TrimSuffix(strings.TrimLeft(file.Name(), "0123456789"), ".json")
			jsonData[fileName] = data
			excelMap[fileName] = excelId
		}
	}
	return jsonData
}

func (m *Manager) GetGlobal() conf.Global {
	return m.AllJson["Global"].(conf.Global)
}
