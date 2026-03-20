// Package sensitive 提供敏感词过滤功能，使用 DFA（确定有限自动机）算法。
//
// 使用方法：
//
//	// 创建过滤器
//	filter := sensitive.New()
//
//	// 加载词库
//	filter.LoadFromFiles("dict1.txt", "dict2.txt")
//	filter.LoadFromStrings("敏感词1", "敏感词2")
//
//	// 检测文本
//	if filter.IsSensitive("这是一条包含敏感词的文本") {
//	    // 处理敏感词
//	}
//
//	// 替换敏感词
//	clean := filter.Replace("这是一段文本", '*')
//
//	// 删除敏感词
//	clean := filter.Remove("这是一段文本")
package sensitive

// Filter 敏感词过滤器接口。
type Filter interface {
	// FindAll 找到文本中所有敏感词。
	FindAll(text string) []string
	// FindAllCount 找到所有敏感词及出现次数。
	FindAllCount(text string) map[string]int
	// FindOne 找到一个敏感词。
	FindOne(text string) string
	// IsSensitive 检查是否包含敏感词。
	IsSensitive(text string) bool
	// Replace 替换敏感词为指定字符。
	Replace(text string, repl rune) string
	// Remove 删除所有敏感词。
	Remove(text string) string
}

// New 创建新的敏感词过滤器（使用 DFA 算法）。
func New() Filter {
	return NewDFA()
}

// LoadFromFiles 从文件加载敏感词并创建过滤器。
func LoadFromFiles(paths ...string) (Filter, error) {
	filter := NewDFA()
	for _, path := range paths {
		if err := filter.LoadFromFile(path); err != nil {
			return nil, err
		}
	}
	return filter, nil
}

// LoadFromStrings 从字符串数组创建过滤器。
func LoadFromStrings(words ...string) Filter {
	filter := NewDFA()
	filter.AddWords(words...)
	return filter
}
