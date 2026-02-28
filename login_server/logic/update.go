package logic

import (
	"github.com/gin-gonic/gin"
	"xfx/login_server/define"
	"xfx/login_server/model"
	"xfx/pkg/log"
)

// 更新
func Forceupdate(c *gin.Context) {
	var forceUpdate model.ForceUpdate
	if err := c.ShouldBindJSON(&forceUpdate); err != nil {
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug(" version %v, chanel %v", forceUpdate.Version, forceUpdate.Channel)

	//取库
	item := model.HotUpdateItem{}
	has, err := AccountEngine.Table(define.HotUpdate).Where("channel=?", forceUpdate.Channel).Get(&item)
	if err != nil {
		log.Error("获取线上版本错误: %s", err)
		httpRetGame(c, ERR_DB, "params err1")
		return
	}

	if !has {
		log.Error("获取线上版本错误: %s", err)
		httpRetGame(c, ERR_ACCOUNT_NOT_FOUND, "params err1")
		return
	}

	//获取线上版本
	//url := fmt.Sprintf("http://localhost:10001/hotupdate/%d/version/version.txt", forceUpdate.Channel)
	//resp, err := http.Get(url)
	//if err != nil {
	//	log.Error("获取线上版本错误: %s", err)
	//	httpRetGame(c, SUCCESS, "success",
	//		map[string]any{
	//			"Status": 1,
	//			"Url":    "",
	//		})
	//	return
	//}
	//defer resp.Body.Close()

	// Read the response body
	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	httpRetGame(c, SUCCESS, "success",
	//		map[string]any{
	//			"Status": 1,
	//			"Url":    "",
	//		})
	//	return
	//}
	version := string(item.Version)
	if version == forceUpdate.Version {
		httpRetGame(c, SUCCESS, "success",
			map[string]any{
				"Status": 2,
				"Url":    "",
			})
		return
	}

	httpRetGame(c, SUCCESS, "success",
		map[string]any{
			"Status":  0,
			"Url":     "",
			"Version": item.Version,
		})
}
