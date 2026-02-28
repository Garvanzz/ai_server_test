package modules

import (
	"xfx/pkg/agent"
	"xfx/pkg/module"
)

type DefaultModule struct {
	mi  module.Module
	pid agent.PID
}

// const (
// 	LenStackBuf = 3
// )

// func run(m *DefaultModule) {
// 	defer func() {
// 		if r := recover(); r != nil {
// 			if LenStackBuf > 0 {
// 				buf := make([]byte, LenStackBuf)
// 				l := runtime.Stack(buf, false)
// 				log.Error("%v: %s", r, buf[:l])
// 			} else {
// 				log.Error("%v", r)
// 			}
// 		}
// 	}()
// 	// m.mi.Run(m.closeSig)
// 	m.wg.Done()
// }

// func destroy(m *DefaultModule) {
// 	defer func() {
// 		if r := recover(); r != nil {
// 			if LenStackBuf > 0 {
// 				buf := make([]byte, LenStackBuf)
// 				l := runtime.Stack(buf, false)
// 				log.Error("%v: %s", r, buf[:l])
// 			} else {
// 				log.Error("%v", r)
// 			}
// 		}
// 	}()
// 	m.mi.OnDestroy()
// }
