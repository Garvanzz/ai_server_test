package invoke

import (
	"xfx/core/config/conf"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
)

type MailModClient struct {
	invoke Invoker
	Type   string
}

func MailClient(invoker Invoker) MailModClient {
	return MailModClient{
		invoke: invoker,
		Type:   define.ModuleMail,
	}
}

func (m MailModClient) SendMail(mailType int, CnTitle, CnContent, EnTitle, EnContent, sendName string,
	items []conf.ItemE, receiverIds []int64, expireTime int64, cfgId int32, attachment bool, params []string) bool {
	result, _ := Bool(m.invoke.Invoke(m.Type, "SendMail", mailType, CnTitle, CnContent, EnTitle, EnContent, sendName, items, receiverIds, expireTime, cfgId, params))
	return result
}

func (m MailModClient) SendDelayMails(id int64, delay int64, mType int, CnTitle, CnContent, EnTitle, EnContent string,
	items []conf.ItemE, expireTime int64, cfgId int32, params []string, receiverIds []int64) bool {
	result, _ := Bool(m.invoke.Invoke(m.Type, "SendDelayMails", id, delay, mType, CnTitle, CnContent, EnTitle, EnTitle, EnContent, items, expireTime, cfgId, params, receiverIds))
	return result
}

func (m MailModClient) DeleteDelayMails(id int64) bool {
	result, _ := Bool(m.invoke.Invoke(m.Type, "DeleteDelayMails", id))
	return result
}

func (m MailModClient) GetSystemMailById(id int64) *model.SysMailInfo {
	result, err := m.invoke.Invoke(m.Type, "GetSystemMailById", id)
	if err != nil {
		log.Error("GetSystemMailById err:%v", err)
		return nil
	}
	if result == nil {
		return nil
	}

	return result.(*model.SysMailInfo)
}

func (m MailModClient) GetMaxSystemMailId() int64 {
	result, _ := Int64(m.invoke.Invoke(m.Type, "GetMaxSystemMailId"))
	return result
}
