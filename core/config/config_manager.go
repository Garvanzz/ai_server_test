package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"xfx/core/config/conf"
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
		//AllJson:       make(map[string]any),
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

	// 首次加载配置
	CfgMgr.Reload()

	if CfgMgr.reloadOnChange {
		go CfgMgr.watchConfigChange()
	}
}

// 提取出的 Reload 方法示例
func (m *Manager) Reload() {
	// 1. 加载所有最新文件
	data := m.loadFile()

	// 2. 创建一个全新的 Map 准备接收数据
	newAllJson := make(map[string]any)

	// 3. 调用 Parse 进行解析和校验
	err := m.parserFunc(data, newAllJson)
	if err != nil {
		// 【热更失败处理】
		// 打印错误日志，保留旧的 configData 不做替换，直接 return
		// log.Error("Hot-reload failed! error: %v", err)
		fmt.Printf("Hot-reload failed! error: %v\n", err)
		return
	}

	// 4. 解析和校验全部成功，执行无锁原子替换
	m.configData.Store(newAllJson)
	// log.Info("Configurations hot-reloaded successfully!")
	fmt.Println("Configurations hot-reloaded successfully!")
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
			// 过滤掉 chmod 等无关事件，只监听写入或创建
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				// 注意：在生产环境中建议在这里加一个防抖(Debounce)逻辑，
				// 防止短时间内多次修改触发多次 Reload
				m.Reload()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Error("watcher err: %v", err)
		}
	}
}

// 获取当前的配置 Map（包内使用）
func (m *Manager) getAll() map[string]any {
	data := m.configData.Load()
	if data == nil {
		return make(map[string]any)
	}
	return data.(map[string]any)
}

func (m *Manager) AllJson() map[string]any {
	return m.getAll()
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
	return m.getAll()["Global"].(conf.Global)
}
