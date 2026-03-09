package huaguoshan

import (
	"encoding/json"
	"fmt"
	"time"
	"xfx/core/config"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/pkg/module"
	"xfx/pkg/module/modules"
)

var Module = func() module.Module {
	return &Manager{
		invites:         make(map[int64]*model.PartnerInvite),
		receiverInvites: make(map[int64][]int64),
	}
}

type Manager struct {
	modules.BaseModule
	inviteId        int64                          // 邀请ID自增
	invites         map[int64]*model.PartnerInvite // 所有邀请记录 key=inviteId
	receiverInvites map[int64][]int64              // 接收者的邀请列表 key=receiverId, value=[]inviteId
	lastSaveTime    int64                          // 上次保存时间
}

func (m *Manager) OnInit(app module.App) {
	m.BaseModule.OnInit(app)
	m.loadData()

	// 注册RPC方法
	m.Register("CreateInvite", m.CreateInvite)
	m.Register("GetInvite", m.GetInvite)
	m.Register("GetReceiverInvites", m.GetReceiverInvites)
	m.Register("ProcessInvite", m.ProcessInvite)
}

func (m *Manager) loadData() {

	reply, err := db.RedisExec("GET", define.HuaguoshanPartnerInvite)
	if err != nil {
		log.Error("load huaguoshan partner invite error: %v", err)
		return
	}

	if reply != nil {
		data := make(map[string]interface{})
		err = json.Unmarshal(reply.([]byte), &data)
		if err != nil {
			log.Error("unmarshal huaguoshan data error: %v", err)
			return
		}

		if id, ok := data["inviteId"]; ok {
			m.inviteId = int64(id.(float64))
		}
		if invites, ok := data["invites"]; ok {
			invitesData, _ := json.Marshal(invites)
			json.Unmarshal(invitesData, &m.invites)
		}
		if receiverInvites, ok := data["receiverInvites"]; ok {
			receiverInvitesData, _ := json.Marshal(receiverInvites)
			json.Unmarshal(receiverInvitesData, &m.receiverInvites)
		}
	}

	log.Debug("huaguoshan loadData success, inviteId: %d, invites: %d", m.inviteId, len(m.invites))
}

func (m *Manager) GetType() string { return define.ModuleHuaguoshan }

func (m *Manager) OnTick(delta time.Duration) {
	now := utils.Now().Unix()
	if now-m.lastSaveTime >= 60 {
		m.saveToRedis()
		m.lastSaveTime = now
	}
}

func (m *Manager) OnDestroy() {
	m.saveToRedis()
}

func (m *Manager) saveToRedis() {

	data := map[string]interface{}{
		"inviteId":        m.inviteId,
		"invites":         m.invites,
		"receiverInvites": m.receiverInvites,
	}

	b, err := json.Marshal(data)
	if err != nil {
		log.Error("marshal huaguoshan data error: %v", err)
		return
	}

	db.RedisExec("SET", define.HuaguoshanPartnerInvite, string(b))
}

func (m *Manager) OnMessage(msg interface{}) interface{} {
	return nil
}

// CreateInvite 创建邀请
func (m *Manager) CreateInvite(senderId int64, senderName string, receiverId int64) *model.PartnerInvite {
	m.inviteId++
	now := utils.Now().Unix()
	effectTime := config.Global.Get().PaternerInviteEffectTime
	expireTime := now + int64(effectTime*24*3600)

	invite := &model.PartnerInvite{
		Id:         m.inviteId,
		SenderId:   senderId,
		SenderName: senderName,
		ReceiverId: receiverId,
		Status:     define.PartnerInviteStatusPending,
		CreateTime: now,
		ExpireTime: expireTime,
	}

	m.invites[invite.Id] = invite
	m.receiverInvites[receiverId] = append(m.receiverInvites[receiverId], invite.Id)

	log.Debug("CreateInvite success, inviteId: %d, sender: %d, receiver: %d", invite.Id, senderId, receiverId)
	return invite
}

// GetInvite 获取邀请
func (m *Manager) GetInvite(inviteId int64) *model.PartnerInvite {
	return m.invites[inviteId]
}

// GetReceiverInvites 获取接收者的邀请列表
func (m *Manager) GetReceiverInvites(receiverId int64) []*model.PartnerInvite {
	inviteIds := m.receiverInvites[receiverId]
	if len(inviteIds) == 0 {
		return []*model.PartnerInvite{}
	}

	invites := make([]*model.PartnerInvite, 0)
	for _, id := range inviteIds {
		if invite := m.invites[id]; invite != nil {
			invites = append(invites, invite)
		}
	}

	return invites
}

// ProcessInvite 处理邀请
func (m *Manager) ProcessInvite(inviteId int64, accept bool) (*model.PartnerInvite, error) {
	invite := m.invites[inviteId]
	if invite == nil {
		return nil, fmt.Errorf("邀请不存在")
	}

	// 校验邀请状态
	if invite.Status != define.PartnerInviteStatusPending {
		return nil, fmt.Errorf("邀请已被处理")
	}

	// 校验是否过期
	now := utils.Now().Unix()
	if now > invite.ExpireTime {
		return nil, fmt.Errorf("邀请已过期")
	}

	// 更新状态
	if accept {
		invite.Status = define.PartnerInviteStatusAccepted
	} else {
		invite.Status = define.PartnerInviteStatusRejected
	}

	log.Debug("ProcessInvite success, inviteId: %d, accept: %v", inviteId, accept)
	return invite, nil
}
