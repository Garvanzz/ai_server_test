package agent

import (
	"strings"
	"sync"

	"github.com/asynkron/protoactor-go/actor"
)

type PID *actor.PID

func NewPID(address, id string) PID { return actor.NewPID(address, id) }

// ParsePID is parse string to PID, like NewPID, ParsePID will new a PID
func Parse(value string) (PID, bool) {
	idx := strings.Index(value, "/")
	if idx == -1 {
		return nil, false
	}
	address, id := value[:idx], value[idx+1:]
	return NewPID(address, id), true
}

// 将pid转换成字符串
func Address(pid PID) string {
	if pid == nil {
		return "nil"
	}
	return pid.Address + "/" + pid.Id
}

var (
	NilPID   = PID(nil)
	name2pid map[string]PID
	addr2pid map[string]PID
	pidLock  sync.RWMutex
)

func init() {
	name2pid = make(map[string]PID)
	addr2pid = make(map[string]PID)
}

// Lookup is lookup PID by agent name or PID's address
func Lookup(name string) (PID, bool) {
	pidLock.RLock()
	defer pidLock.RUnlock()
	if strings.ContainsRune(name, ':') {
		pid, ok := addr2pid[name]
		return pid, ok
	}
	pid, ok := name2pid[name]
	return pid, ok
}

func _Store(name string, pid PID) {
	pidLock.Lock()
	name2pid[name] = pid
	addr2pid[Address(pid)] = pid
	pidLock.Unlock()
}

func _Delete(name string, pid PID) {
	pidLock.Lock()
	delete(name2pid, name)
	delete(addr2pid, Address(pid))
	pidLock.Unlock()
}
