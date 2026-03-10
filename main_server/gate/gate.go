package mgate

import (
	"fmt"
	"xfx/pkg/gate"
	"xfx/pkg/gate/tcpgate"
	"xfx/pkg/log"
	"xfx/pkg/module"
)

var Module = func() module.Module {
	gate := new(Gate)
	gate.SetCreateAgent(gate.CreateAgent)
	return gate
}

type Gate struct {
	tcpgate.Gate
}

func (gt *Gate) OnInit(app module.App) {
	gt.Gate.OnInit(app) // 模块初始化
}

func (gt *Gate) GetType() string { return "Gate" }

func (gt *Gate) CreateAgent(gate gate.Gate, session gate.Session) (gate.Agent, error) {
	agent := NewAgent(gt)
	agent.OnInit(gate, session)

	_, err := gt.Context.Create(fmt.Sprintf("session#%d", session.ID()), agent)
	if err != nil {
		log.Error("* gate CreateAgent session:%v failed: %v", session.ID(), err)
		return nil, err
	}

	return agent, nil
}
