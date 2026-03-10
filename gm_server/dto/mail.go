package dto

// 邮件信息
type GmMailInfo struct {
	Delaytime       string `json:"delaytime"`
	Immediatelysend bool   `json:"immediatelysend"`
	Content         string `json:"content"`
	Title           string `json:"title"`
	Uid             string `json:"uid"` // player_id 列表，用 | 分隔，如 "123|456"
	Server          int32  `json:"server"`
	Fullserversend  bool   `json:"fullserversend"`
	Name            string `json:"name"`
	Type            string `json:"mailType"` // 邮件类型：system / person
	SenderName      string `json:"sendname"` //发送者名字
	Itmes           string `json:"reward"`
}
