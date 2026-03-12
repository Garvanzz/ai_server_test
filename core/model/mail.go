package model

import (
	"encoding/json"
	"time"
	"xfx/core/config/conf"
	"xfx/pkg/utils"
	"xfx/proto/proto_mail"
	"xfx/proto/proto_public"
)

// SysMailInfo 系统邮件信息
type SysMailInfo struct {
	Id         int64               `json:"id"`
	MailInfos  map[string]MailInfo `json:"mail_infos"` // 邮件内容
	Items      []conf.ItemE        `json:"items"`      //道具
	CreateTime int64               `json:"createTime"` //创建时间
	ExpireTime int64               `json:"expireTime"` //过期时间
	CfgId      int32               `json:"cfgId"`      //配置id 默认0
	Params     []string            `json:"params"`     //参数 为空就是无
	SenderName string              `json:"senderName"` //发送者名字
}

// GM消息结构体
type GMMailInfo struct {
	CreatorName string
	CreateTime  time.Time
	EffectTime  time.Time
	CnContent   string
	CnTitle     string
	PlayerIds   []int64
	Status      int
	Type        int
	SenderName  string
	Items       []MailItem
}

type MailItem struct {
	Id   int32 `json:"id"`
	Num  int32 `json:"num"`
	Type int32 `json:"type"`
}

// PlayerMailInfo 玩家邮件信息
type PlayerMailInfo struct {
	Id              int64               // mysql自增id
	MailInfos       map[string]MailInfo `json:"mail_infos"` // 邮件内容
	OpenTime        int64               // 开启时间
	CreateTime      int64               // 创建时间
	Items           []conf.ItemE        // 附件
	GotItem         bool                // 是否领取奖励
	CfgId           int32               // 配置id
	Params          []string            // 参数
	ExpireTime      int64               // 过期时间
	SysId           int64               // 系统邮件id
	AccountId       string              // 玩家account_id
	DbId            int64               // 玩家id
	Type            int32               // 邮件类型 0默认普通邮件 1联盟邮件
	SenderName      string              //发送者名字
	IsHasAttachment bool                //是否有附件
}

type MailInfo struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (mail *PlayerMailInfo) ToProto() *proto_mail.Mail {
	m := make(map[string]*proto_mail.MailInfo, 0)
	for k, v := range mail.MailInfos {
		m[k] = &proto_mail.MailInfo{
			Title:   v.Title,
			Content: v.Content,
		}
	}

	reward := make([]*proto_public.Item, 0)
	for i := 0; i < len(mail.Items); i++ {
		reward = append(reward, &proto_public.Item{
			ItemId:   mail.Items[i].ItemId,
			ItemType: mail.Items[i].ItemType,
			ItemNum:  mail.Items[i].ItemNum,
		})
	}

	var attachmentInfo *proto_public.AttachmentOption
	if mail.IsHasAttachment && len(mail.Params) >= 2 {
		// 从 Content[1] 反序列化 AttachmentData
		var attachmentData AttachmentData
		orderId := utils.MustParseInt64(mail.Params[0])
		err := json.Unmarshal([]byte(mail.Params[1]), &attachmentData)
		if err == nil {
			// 转换为 AttachmentOption 协议结构
			attachmentInfo = &proto_public.AttachmentOption{
				Id:    orderId,
				Type:  attachmentData.Type,
				Value: attachmentData.ItemId,
				Level: attachmentData.Level,
				Stage: attachmentData.Stage,
				Star:  attachmentData.Star,
			}
		}
	}

	return &proto_mail.Mail{
		Id:              mail.Id,
		Infos:           m,
		OpenTime:        mail.OpenTime,
		CreateTime:      mail.CreateTime,
		Item:            reward,
		GotItem:         mail.GotItem,
		CfgId:           mail.CfgId,
		Params:          mail.Params,
		ExpireTime:      mail.ExpireTime,
		SenderName:      mail.SenderName,
		IsHasAttachment: mail.IsHasAttachment,
		AttachmentInfo:  attachmentInfo,
	}
}
