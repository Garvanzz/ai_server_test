package gate

type Gate interface {
	NewAgent(session Session) (Agent, error)
}

type Session interface {
	ID() uint64
	IsClosed() bool
	Close() error
	CloseWithFlush(timeout time.Duration) error  // 新增
	Receive() (any, error)
	Send(msg any) error
	Set(key string, value any)
	Get(key string) (any, bool)
	Delete(key string)
}

type Agent interface {
	OnInit(gate Gate, session Session)
	Send(msg interface{})
	OnRecv(msg interface{})
	Close() error
	Set(key string, value any)
	Get(key string) (value any, ok bool)
	GetSession() Session
}
