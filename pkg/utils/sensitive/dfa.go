package sensitive

import (
	"bufio"
	"io"
	"os"
	"strings"
	"sync"
)

// dfaNode DFA 节点。
type dfaNode struct {
	children map[rune]*dfaNode
	isEnd    bool
}

func newDFANode() *dfaNode {
	return &dfaNode{
		children: make(map[rune]*dfaNode),
		isEnd:    false,
	}
}

// DFA DFA 敏感词过滤器实现。
type DFA struct {
	root *dfaNode
	mu   sync.RWMutex
}

// NewDFA 创建新的 DFA 敏感词过滤器。
func NewDFA() *DFA {
	return &DFA{
		root: newDFANode(),
	}
}

// LoadFromFile 从文件加载敏感词。
func (d *DFA) LoadFromFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return d.LoadFromReader(f)
}

// LoadFromReader 从 io.Reader 加载敏感词。
func (d *DFA) LoadFromReader(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" && !strings.HasPrefix(word, "#") {
			d.AddWord(word)
		}
	}
	return scanner.Err()
}

// AddWord 添加单个敏感词。
func (d *DFA) AddWord(word string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	node := d.root
	for _, r := range word {
		if next, ok := node.children[r]; ok {
			node = next
		} else {
			next = newDFANode()
			node.children[r] = next
			node = next
		}
	}
	node.isEnd = true
}

// AddWords 添加多个敏感词。
func (d *DFA) AddWords(words ...string) {
	for _, word := range words {
		d.AddWord(word)
	}
}

// RemoveWord 删除单个敏感词。
func (d *DFA) RemoveWord(word string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// DFA 删除比较复杂，这里简单处理：标记为非结束节点
	// 如果需要完全删除，需要重建整个树
	node := d.root
	for _, r := range word {
		if next, ok := node.children[r]; ok {
			node = next
		} else {
			return // 词不存在
		}
	}
	node.isEnd = false
}

// FindAll 找到文本中所有敏感词。
func (d *DFA) FindAll(text string) []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]string, 0)
	seen := make(map[string]bool)
	runes := []rune(text)
	length := len(runes)

	for i := 0; i < length; i++ {
		node := d.root
		for j := i; j < length; j++ {
			next, ok := node.children[runes[j]]
			if !ok {
				break
			}
			node = next
			if node.isEnd {
				word := string(runes[i : j+1])
				if !seen[word] {
					seen[word] = true
					result = append(result, word)
				}
			}
		}
	}

	return result
}

// FindAllCount 找到所有敏感词及出现次数。
func (d *DFA) FindAllCount(text string) map[string]int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make(map[string]int)
	runes := []rune(text)
	length := len(runes)

	for i := 0; i < length; i++ {
		node := d.root
		for j := i; j < length; j++ {
			next, ok := node.children[runes[j]]
			if !ok {
				break
			}
			node = next
			if node.isEnd {
				word := string(runes[i : j+1])
				result[word]++
			}
		}
	}

	return result
}

// FindOne 找到文本中第一个敏感词。
func (d *DFA) FindOne(text string) string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	runes := []rune(text)
	length := len(runes)

	for i := 0; i < length; i++ {
		node := d.root
		for j := i; j < length; j++ {
			next, ok := node.children[runes[j]]
			if !ok {
				break
			}
			node = next
			if node.isEnd {
				return string(runes[i : j+1])
			}
		}
	}

	return ""
}

// IsSensitive 检查是否包含敏感词。
func (d *DFA) IsSensitive(text string) bool {
	return d.FindOne(text) != ""
}

// Replace 替换敏感词为指定字符。
func (d *DFA) Replace(text string, repl rune) string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	runes := []rune(text)
	length := len(runes)
	result := []rune(text)

	for i := 0; i < length; i++ {
		node := d.root
		for j := i; j < length; j++ {
			next, ok := node.children[runes[j]]
			if !ok {
				break
			}
			node = next
			if node.isEnd {
				// 替换敏感词
				for k := i; k <= j; k++ {
					result[k] = repl
				}
				// 跳过已匹配的敏感词
				i = j
				break
			}
		}
	}

	return string(result)
}

// Remove 删除所有敏感词。
func (d *DFA) Remove(text string) string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	runes := []rune(text)
	length := len(runes)
	result := make([]rune, 0, length)

	for i := 0; i < length; {
		node := d.root
		matchEnd := -1

		for j := i; j < length; j++ {
			next, ok := node.children[runes[j]]
			if !ok {
				break
			}
			node = next
			if node.isEnd {
				matchEnd = j
			}
		}

		if matchEnd >= 0 {
			// 跳过敏感词
			i = matchEnd + 1
		} else {
			// 保留当前字符
			result = append(result, runes[i])
			i++
		}
	}

	return string(result)
}

// WordCount 返回敏感词库中的词数。
func (d *DFA) WordCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.countWords(d.root)
}

func (d *DFA) countWords(node *dfaNode) int {
	count := 0
	if node.isEnd {
		count++
	}
	for _, child := range node.children {
		count += d.countWords(child)
	}
	return count
}

// Clear 清空敏感词库。
func (d *DFA) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.root = newDFANode()
}

// Reader 接口定义（用于上面的函数）。
type Reader interface {
	Read(p []byte) (n int, err error)
}

// Writer 接口定义。
type Writer interface {
	Write(p []byte) (n int, err error)
}

// Closer 接口定义。
type Closer interface {
	Close() error
}

// ReadWriter 接口定义。
type ReadWriter interface {
	Reader
	Writer
}

// ReadCloser 接口定义。
type ReadCloser interface {
	Reader
	Closer
}

// WriteCloser 接口定义。
type WriteCloser interface {
	Writer
	Closer
}

// ReadWriteCloser 接口定义。
type ReadWriteCloser interface {
	Reader
	Writer
	Closer
}

// Seeker 接口定义。
type Seeker interface {
	Seek(offset int64, whence int) (int64, error)
}
