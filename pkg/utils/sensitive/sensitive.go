package sensitive

import "xfx/pkg/log"

var Filter *Manager

func Init() {
	var err error
	Filter, err = NewFilter(
		StoreOption{Type: StoreMemory},
		FilterOption{Type: FilterDfa},
	)
	if err != nil {
		log.Error("敏感词服务启动失败, err:%v", err)
		return
	}

	// 加载敏感词库
	//err = Filter.Store.LoadDictPath("F:/project/xiuxian/server/pkg/utils/sensitive/text/dict2.txt")
	//err = Filter.Store.LoadDictPath("./sensitive/dict2.txt")
	//if err != nil {
	//	log.Error("加载词库发生了错误, err:%v", err)
	//	return
	//}
	//log.Debug("加载了敏感词库")
	// 动态自定义敏感词
	//err = Filter.Store.AddWord("测试1", "测试2", "成小王")
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
}
