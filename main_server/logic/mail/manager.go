package mail

import (
	"encoding/json"
	"time"
	"xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
	"xfx/pkg/module"
	"xfx/pkg/module/modules"
	"xfx/pkg/utils"
)

// 游戏内使用
type delayMailInfo struct {
	startTime   int64
	mType       int
	CnTitle     string
	CnContent   string
	EnTitle     string
	EnContent   string
	items       []conf.ItemE
	receiverIds []int64
	expireTime  int64
	cfgId       int32
	params      []string
	SenderName  string
}

// DbMailRecord db用
type DbMailRecord struct {
	Id          int64
	Type        int
	PlayerIds   []string
	DbIds       []int64
	CnTitle     string
	CnContent   string
	EnTitle     string
	EnContent   string
	Items       []mailItem
	EffectTime  time.Time
	CreateTime  time.Time
	CreatorName string
	Status      int
	SenderName  string
}

type mailItem struct {
	Id   int32 `json:"id"`
	Num  int32 `json:"num"`
	Type int32 `json:"type"`
}

var Module = func() module.Module {
	return &Manager{
		Mails:      make(map[int64]*model.SysMailInfo),
		DailyMails: make(map[int64]bool),
		DelayMails: make(map[int64]*delayMailInfo),
	}
}

type Manager struct {
	modules.BaseModule
	SystemMailId   int64                        // 当前最大系统邮件id
	Mails          map[int64]*model.SysMailInfo // 系统邮件列表
	DailyMails     map[int64]bool               // 日常邮件列表
	LastUpdateTime time.Time                    // 上一次tick时间
	DelayMails     map[int64]*delayMailInfo     // 延迟发送邮件
}

func (m *Manager) OnInit(app module.App) {
	m.BaseModule.OnInit(app)

	m.Register("SendMail", m.SendMail)
	m.Register("SendDelayMails", m.SendDelayMails)
	m.Register("DeleteDelayMails", m.DeleteDelayMails)
	m.Register("GetSystemMailById", m.GetSystemMailById)
	m.Register("GetMaxSystemMailId", m.GetMaxSystemMailId)

	rdb, _ := db.GetEngine(m.App.GetEnv().ID)
	sysMails := make([]*model.SysMailInfo, 0)
	err := rdb.Mysql.Find(&sysMails)
	if err != nil {
		log.Error("load sysMails from redis err:", err)
		return
	}

	// 加载系统邮件列表
	now := time.Now().Unix()
	for _, mail := range sysMails {
		if mail.ExpireTime > now {
			m.Mails[mail.Id] = mail
		}
	}

	reply, err := rdb.RedisExec("get", "dailyMail")
	if err != nil {
		log.Error("mail manager load daily mail err:%v", err)
		return
	}

	// 加载日常邮件
	if reply != nil {
		err = json.Unmarshal(reply.([]byte), &m.DailyMails)
		if err != nil {
			log.Error("mail manager load daily mail err1:%v", err)
			return
		}
	}

	reply, err = rdb.RedisExec("get", "systemMailId")
	if err != nil {
		log.Error("mail manager load system mail Id err:%v", err)
		return
	}

	// 加载邮件id
	if reply != nil {
		err = json.Unmarshal(reply.([]byte), &m.SystemMailId)
		log.Debug("load systemMailId ：%+v", m.SystemMailId)
		if err != nil {
			log.Error("mail manager load system mail Id err:%v", err)
			return
		}
	}

	records := make([]*DbMailRecord, 0)
	if err := rdb.Mysql.Table(define.AdminMailTable).Where("status = ?", 1).Find(&records); err != nil {
		log.Error("mail manager load delay mail form db err:%v", err)
		return
	}

	for _, record := range records {

		info := &delayMailInfo{
			startTime: record.EffectTime.Unix(),
			mType:     record.Type,
			CnTitle:   record.CnTitle,
			CnContent: record.CnContent,
			EnTitle:   record.EnTitle,
			EnContent: record.EnContent,
			//items:         _items,
			receiverIds: record.DbIds,
		}

		if record.Items != nil && len(record.Items) > 0 {

			_items := make([]conf.ItemE, 0)
			for _, v := range record.Items {
				_items = append(_items, conf.ItemE{
					ItemId:   v.Id,
					ItemType: v.Type,
					ItemNum:  v.Num,
				})
			}

			info.items = _items
		}

		m.DelayMails[record.Id] = info
	}

}

func (m *Manager) GetType() string { return define.ModuleMail }

func (m *Manager) OnTick(delta time.Duration) {
	now := time.Now()
	m.LastUpdateTime = now

	rdb, _ := db.GetEngine(m.App.GetEnv().ID)

	for id, info := range m.DelayMails {
		if info.startTime <= now.Unix() {

			record := new(DbMailRecord)
			_, err := rdb.Mysql.Table(define.AdminMailTable).Where("id = ?", id).Get(record)
			if err != nil {
				log.Error("get mail record form DB error:%v", err)
			} else {
				if record.Status == 1 {

					record.Status = 2 // 修改为已发送状态
					if _, err := rdb.Mysql.Table(define.AdminMailTable).Where("id = ?", id).Cols("status").Update(record); err != nil {
						log.Error("mail manager update mail record err:%v", err)
					} else {

						//默认结束时间
						if info.expireTime == 0 {
							//当天结束时间+ (过期日期-1)*24小时*3600分钟
							info.expireTime = utils.GetTodayEndUnix() + ((define.MailExpiration - 1) * 3600 * 24)
						} else {
							info.expireTime = info.expireTime + now.Unix()
						}

						if info.mType == define.SystemMail {
							m.sendSysMail(info.items, info.CnTitle, info.CnContent, info.EnTitle, info.EnContent, info.SenderName, info.expireTime, info.cfgId, info.params)
						} else {
							m.sendPlayerMail(info.CnTitle, info.CnContent, info.EnTitle, info.EnContent, info.SenderName, info.items, info.expireTime, info.cfgId, info.params, info.receiverIds)
						}
					}
				}
			}

			delete(m.DelayMails, id)
		}
	}
}

func (m *Manager) OnMessage(msg interface{}) interface{} {
	log.Debug("* room message %v", msg)
	return nil
}

func (m *Manager) OnDestroy() {
	m.OnSave()
}

func (m *Manager) OnSave() {
	data, err := json.Marshal(m.SystemMailId)
	if err != nil {
		log.Error("mailManager stop marshal systemMailId data error:", err)
		return
	}

	rdb, _ := db.GetEngine(m.App.GetEnv().ID)
	_, err = rdb.RedisExec("set", "systemMailId", data)
	if err != nil {
		log.Error("mailManager stop set systemMailid error: %v", err)
		return
	}
}

// SendDelayMails 发送延时邮件
func (m *Manager) SendDelayMails(id int64, delay int64, mType int, CnTitle, CnContent, EnTitle, EnContent string, items []conf.ItemE, expireTime int64, cfgId int32, params []string, receiverIds []int64) bool {
	delayInfo := &delayMailInfo{
		startTime:   delay,
		mType:       mType,
		CnTitle:     CnTitle,
		CnContent:   CnContent,
		EnTitle:     EnTitle,
		EnContent:   EnContent,
		items:       items,
		receiverIds: receiverIds,
		expireTime:  expireTime,
		cfgId:       cfgId,
		params:      params,
	}

	m.DelayMails[id] = delayInfo
	return true
}

// SendMail 发送邮件 1是系统 2是玩家
func (m *Manager) SendMail(mailType int, CnTitle, CnContent, EnTitle, EnContent, sendName string, items []conf.ItemE, receiverIds []int64, expireTime int64, cfgId int32, attachment bool, params []string) bool {
	//默认结束时间
	if expireTime == 0 {
		//当天结束时间+ (过期日期-1)*24小时*3600分钟
		expireTime = utils.GetTodayEndUnix() + ((define.MailExpiration - 1) * 3600 * 24)
	} else {
		expireTime = time.Now().Unix() + expireTime
	}

	log.Debug("发送邮件 类型 :%v", mailType)

	if mailType == define.SystemMail {
		// 发送系统邮件
		return m.sendSysMail(items, CnTitle, CnContent, EnTitle, EnContent, sendName, expireTime, cfgId, params)
	} else {
		// 发送个人邮件
		return m.sendPlayerMail(CnTitle, CnContent, EnTitle, EnContent, sendName, items, expireTime, cfgId, params, receiverIds)
	}
}

// 发送个人邮件
func (m *Manager) sendPlayerMail(CnTitle, CnContent, EnTitle, EnContent, senderName string, items []conf.ItemE, expireTime int64, cfgId int32, params []string, receiverIds []int64) bool {
	gotItem := true
	if len(items) > 0 {
		gotItem = false
	}

	//log.Debug("sendPlayerMail receiverIds:%+v", receiverIds)
	mailType := define.MailTypeNormal
	//mailConf, ok := cfg.Configm.GetCfg("ConfMail").(map[int64]global.ConfMailElement)[int64(cfgId)]
	//if ok && mailConf.Type == 2 {
	//	mailType = global.MailTypeGuild
	//}
	for _, receiverId := range receiverIds {
		log.Info("Ids:%v", receiverId)
		newMail := &model.PlayerMailInfo{
			SysId: 0, //个人邮件的话系统id为0
			MailInfos: map[string]model.MailInfo{
				define.LanguageChinese:            {CnTitle, CnContent},
				define.LanguageEnglish:            {EnTitle, EnContent},
				define.LanguageChineseTraditional: {CnTitle, CnContent},
			},
			CreateTime: time.Now().Unix(),
			Items:      items,
			ExpireTime: define.MailExpiration*86400 + time.Now().Unix(), // 根据配置设置过期时间 ,
			CfgId:      cfgId,
			Params:     params,
			DbId:       receiverId,
			GotItem:    gotItem,
			Type:       mailType,
			SenderName: senderName,
			//AccountId:  receiverId,
		}
		log.Info("newMail:%v", newMail)
		rdb, _ := db.GetEngine(m.App.GetEnv().ID)

		num, err := rdb.Mysql.Insert(newMail)
		if err != nil {
			log.Error("send player mail insert error:%v", err)
			return false
		}
		if num == 0 {
			log.Error("send player private mail insert db failed , num is 0, player account_id %v", receiverId)
			return false
		}
	}
	return true
}

// 发送系统邮件
func (m *Manager) sendSysMail(items []conf.ItemE, CnTitle, CnContent, EnTitle, EnContent, sendName string, expireTime int64, cfgId int32, params []string) bool {
	newSysMail := &model.SysMailInfo{
		Items: items,
		MailInfos: map[string]model.MailInfo{
			define.LanguageChinese:            {CnTitle, CnContent},
			define.LanguageChineseTraditional: {CnTitle, CnContent},
			define.LanguageEnglish:            {EnTitle, EnContent},
		},
		ExpireTime: expireTime,
		CreateTime: time.Now().Unix(),
		CfgId:      cfgId,
		Params:     params,
		SenderName: sendName,
	}

	rdb, _ := db.GetEngine(m.App.GetEnv().ID)
	num, err := rdb.Mysql.Insert(newSysMail)
	if err != nil {
		log.Error("send sysMail insert mysql error:%s", err)
		return false
	}
	if num == 0 {
		log.Error("send system mail insert db failed , num is 0")
		return false
	}

	m.Mails[newSysMail.Id] = newSysMail
	m.SystemMailId = newSysMail.Id
	m.OnSave() // systemMailId修改及时落库

	return true
}

// GetSystemMailById 根据id获取系统邮件
func (m *Manager) GetSystemMailById(id int64) *model.SysMailInfo {
	mail, ok := m.Mails[id]

	if !ok {
		return nil
	}

	sysMail := new(model.SysMailInfo)
	*sysMail = *mail
	return sysMail
}

// GetMaxSystemMailId 获取当前最大的系统邮件id
func (m *Manager) GetMaxSystemMailId() int64 {
	return m.SystemMailId
}

// DeleteDelayMails 删除延时邮件
func (m *Manager) DeleteDelayMails(id int64) bool {
	_, ok := m.DelayMails[id]
	if !ok {
		return false
	}
	delete(m.DelayMails, id)

	return true
}

// 删除db邮件
func (m *Manager) deleteDBMail(id int64) bool {
	rdb, _ := db.GetEngine(m.App.GetEnv().ID)

	num, err := rdb.Mysql.Where("id = ?", id).Delete(&model.PlayerMailInfo{})
	if err != nil {
		log.Error("delete DB mail error:%v", err)
		return false
	}

	if num == 0 {
		log.Error("delete DB mail num is 0,%v", id)
		return false
	}

	return true
}
