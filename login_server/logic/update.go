package logic

import (
	"xfx/login_server/define"
	"xfx/login_server/internal/middleware"
	"xfx/login_server/model/dto"
	"xfx/login_server/model/entity"
	"xfx/pkg/log"

	"github.com/gin-gonic/gin"
)

// ForceUpdate 更新
func ForceUpdate(c *gin.Context) {
	var forceUpdate dto.ForceUpdate
	if err := c.ShouldBindJSON(&forceUpdate); err != nil {
		middleware.RetGame(c, define.ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug(" version %v, chanel %v", forceUpdate.Version, forceUpdate.Channel)

	//取库
	item := entity.HotUpdateItem{}
	has, err := AccountEngine.Table(define.TableHotUpdate).Where("channel=?", forceUpdate.Channel).Get(&item)
	if err != nil {
		log.Error("获取线上版本错误: %s", err)
		middleware.RetGame(c, define.ERR_DB, "params err1")
		return
	}

	if !has {
		log.Error("获取线上版本错误: %s", err)
		middleware.RetGame(c, define.ERR_ACCOUNT_NOT_FOUND, "params err1")
		return
	}

	//获取线上版本
	//url := fmt.Sprintf("http://localhost:10001/hotupdate/%d/version/version.txt", forceUpdate.Channel)
	//resp, err := http.Get(url)
	//if err != nil {
	//	log.Error("获取线上版本错误: %s", err)
	//	middleware.RetGame(c, define.SUCCESS, "success",
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
	//	middleware.RetGame(c, define.SUCCESS, "success",
	//		map[string]any{
	//			"Status": 1,
	//			"Url":    "",
	//		})
	//	return
	//}
	version := item.Version
	if version == forceUpdate.Version {
		middleware.RetGame(c, define.SUCCESS, "success",
			map[string]any{
				"Status": 2,
				"Url":    "",
			})
		return
	}

	middleware.RetGame(c, define.SUCCESS, "success",
		map[string]any{
			"Status":  0,
			"Url":     "",
			"Version": item.Version,
		})
}
