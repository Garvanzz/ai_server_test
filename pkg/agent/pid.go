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
	namelock *sync.RWMutex
	addrlock *sync.RWMutex
)

func init() {
	name2pid = make(map[string]PID)
	addr2pid = make(map[string]PID)
	namelock = new(sync.RWMutex)
	addrlock = new(sync.RWMutex)
}

// Lookup is lookup PID by agent name or PID's address
func Lookup(name string) (PID, bool) {
	if strings.ContainsRune(name, ':') {
		addrlock.RLock()
		pid, ok := addr2pid[name]
		addrlock.RUnlock()
		return pid, ok
	} else {
		namelock.RLock()
		pid, ok := name2pid[name]
		namelock.RUnlock()
		return pid, ok
	}
}

func _Store(name string, pid PID) {
	namelock.Lock()
	name2pid[name] = pid
	namelock.Unlock()

	addrlock.Lock()
	addr2pid[Address(pid)] = pid
	addrlock.Unlock()
}

func _Delete(name string, pid PID) {
	namelock.Lock()
	delete(name2pid, name)
	namelock.Unlock()

	addrlock.Lock()
	delete(addr2pid, Address(pid))
	addrlock.Unlock()
}
