package agent

import (
	"fmt"
	"time"
	"xfx/pkg/serialize"

	"github.com/gogo/protobuf/proto"
)

// Control message: create agent
type createMessage struct {
	name  string
	agent Agent
	opts  []Option
}

// Control message: tick agent
type tickMessage time.Duration

type LocalMessage struct {
	msg      interface{}
	response bool
	sender   PID
}

func wrapMessage(fromPid, toPid PID, msg interface{}, response bool) (interface{}, error) {
	remote := fromPid.Address != toPid.Address
	if remote {
		switch msg.(type) {
		case proto.Message:
			name, data, err := serialize.MarshalType(msg)
			if err != nil {
				panic(err)
			}
			return &RemoteMessage{
				TypeName:    name,
				MessageData: data,
				Response:    response,
				Sender:      Address(fromPid),
			}, nil
		default:
			return nil, fmt.Errorf("agent message: expect proto message")
		}
	} else {
		switch msg.(type) {
		case tickMessage:
			return msg, nil
		default:
			return &LocalMessage{
				msg:      msg,
				response: response,
				sender:   fromPid,
			}, nil
		}
	}
}

func unwrapMessage(msg interface{}) (interface{}, PID, string, bool) {
	switch m := msg.(type) {
	case tickMessage:
		return m, nil, "", false
	case *RemoteMessage:
		p, err := serialize.UnmarshalType(m.TypeName, m.MessageData)
		if err != nil {
			panic(err)
		}
		return p, nil, m.Sender, m.Response
	case *LocalMessage:
		return m.msg, m.sender, "", m.response
	default:
		return m, nil, "", false
	}
}
