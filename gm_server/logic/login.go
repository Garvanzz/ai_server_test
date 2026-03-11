package logic

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"xfx/core/define"
	"xfx/gm_server/db"
	"xfx/gm_server/dto"
	"xfx/pkg/log"
	"xfx/pkg/utils"
)

// 登录账号
func GmLogin(c *gin.Context) {
	var loginUser dto.GMLogin
	if err := c.ShouldBindJSON(&loginUser); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug(" account %v, version %v", loginUser.UserName, loginUser.Password)

	//账号不能为空
	if len(loginUser.UserName) <= 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "account not found")
		return
	}

	player := new(dto.GmAccount)
	player.UserName = loginUser.UserName
	player.Password = loginUser.Password

	has, err := db.AccountDb.Table(define.AdminTable).Get(player)
	if err != nil {
		log.Error("account failed, err : %v", err)
		HTTPRetGame(c, ERR_DB, "db err")
		return
	}

	if !has {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	//生成登录token
	loginToken := fmt.Sprintf("%x", utils.RandomNumeric(15))
	player.Token = loginToken
	db.AccountDb.Table(define.AdminTable).Where("user_name = ?", player.UserName).MustCols("token").Update(player)

	HTTPRetGame(c, SUCCESS, "success",
		map[string]any{
			"xiaoxiaoxiyou": loginToken,
		})
}

// GmLogout GM 登出，清除服务端 token，使当前 token 立即失效
func GmLogout(c *gin.Context) {
	token := c.GetHeader(TokenHeaderName)
	if token != "" {
		player := new(dto.GmAccount)
		player.Token = token
		has, err := db.AccountDb.Table(define.AdminTable).Where("token = ?", token).Get(player)
		if err == nil && has {
			player.Token = ""
			_, _ = db.AccountDb.Table(define.AdminTable).Where("user_name = ?", player.UserName).MustCols("token").Update(player)
		}
	}
	HTTPRetGame(c, SUCCESS, "success")
}

// GM用户
func GmAdminUserInfo(c *gin.Context) {
	rawData, _ := c.GetRawData()

	var gmUserInfo dto.GMUserInfo
	err := json.Unmarshal(rawData, &gmUserInfo)
	if err != nil {
		log.Error("解析失败: %v", err)
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug(" gmUserInfo %v", gmUserInfo)
	//验证码不能为空
	if len(gmUserInfo.Xiaoxiaoxiyou) <= 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "accessToken not found")
		return
	}

	player := new(dto.GmAccount)
	player.UserName = gmUserInfo.UserName

	has, err := db.AccountDb.Table(define.AdminTable).Get(player)
	if err != nil {
		log.Error("account failed, err : %v", err)
		HTTPRetGame(c, ERR_DB, "db err")
		return
	}
	if !has {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	if player.Token != gmUserInfo.Xiaoxiaoxiyou {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "accessToken is mistake")
		return
	}

	permiss := make([]string, 0)
	if player.Permission == 1 {
		permiss = []string{"admin", "editor"}
	} else if player.Permission == 2 {
		permiss = []string{"admin"}
	} else if player.Permission == 3 {
		permiss = []string{"editor"}
	}

	HTTPRetGame(c, SUCCESS, "success",
		map[string]any{
			"username":    player.Name,
			"permissions": permiss,
		})
}
