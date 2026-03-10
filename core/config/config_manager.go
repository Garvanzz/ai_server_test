package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"xfx/core/define"
	"xfx/core/event"
	"xfx/pkg/log"

	"github.com/fsnotify/fsnotify"
)

var CfgMgr *Manager
var excelMap map[string]int

var numPrefixRe = regexp.MustCompile(`^\d+`)

type Manager struct {
	configData     atomic.Value
	filePath       string
	parserFunc     ParseFunc
	reloadOnChange bool
}

type ParseFunc func(data map[string]any, dest map[string]any) error

func NewManager(filePath string, reloadOnChange bool, parser ParseFunc) *Manager {
	s := &Manager{
		filePath:       filePath,
		reloadOnChange: reloadOnChange,
		parserFunc:     parser,
	}

	s.configData.Store(make(map[string]any))
	return s
}

func InitConfig(confPath string) {
	CfgMgr = NewManager(confPath, true, Parse)
	if CfgMgr.parserFunc == nil {
		panic("config parser func is nil")
	}

	CfgMgr.reload(true)

	if CfgMgr.reloadOnChange {
		go CfgMgr.watchConfigChange()
	}
}

func (m *Manager) reload(flag bool) {
	data := m.loadFile()

	newAllJson := make(map[string]any)

	err := m.parserFunc(data, newAllJson)
	if err != nil {
		if flag {
			panic(err)
		}

		fmt.Printf("Hot-reload failed! error: %v\n", err)
		return
	}

	m.configData.Store(newAllJson)
	fmt.Println("Configurations hot-reloaded successfully!")

	// 通知各模块配置已热更
	event.DoEvent(define.EventTypeConfigReload, nil)
}

// 监听文件变化
func (m *Manager) watchConfigChange() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("fsnotify err: %v", err)
	}
	defer watcher.Close()

	watcher.Add(m.filePath)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				m.reload(false)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Error("watcher err: %v", err)
		}
	}
}

// 获取当前的配置 Map
func (m *Manager) getAll() map[string]any {
	data := m.configData.Load()
	if data == nil {
		return make(map[string]any)
	}
	return data.(map[string]any)
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
