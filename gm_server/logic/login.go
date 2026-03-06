package logic

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"xfx/gm_server/db"
	"xfx/gm_server/define"
	"xfx/gm_server/gm_model"
	"xfx/pkg/log"
	"xfx/pkg/utils"
)

// 登录账号
func GmLogin(c *gin.Context) {
	var loginUser gm_model.GMLogin
	if err := c.ShouldBindJSON(&loginUser); err != nil {
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug(" account %v, version %v", loginUser.UserName, loginUser.Password)

	//账号不能为空
	if len(loginUser.UserName) <= 0 {
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "account not found")
		return
	}

	player := new(gm_model.GmAccount)
	player.UserName = loginUser.UserName
	player.Password = loginUser.Password

	has, err := db.AccountDb.Table(define.Admin).Get(player)
	if err != nil {
		log.Error("account failed, err : %v", err)
		httpRetGame(c, ERR_DB, "db err")
		return
	}

	if !has {
		httpRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	//生成登录token
	loginToken := fmt.Sprintf("%x", utils.RandomNumeric(15))
	player.Token = loginToken
	db.AccountDb.Table(define.Admin).Where("user_name = ?", player.UserName).MustCols("token").Update(player)

	httpRetGame(c, SUCCESS, "success",
		map[string]any{
			"xiaoxiaoxiyou": loginToken,
		})
}

// 退出登录
func GmLoginout(c *gin.Context) {
	httpRetGame(c, SUCCESS, "success")
}

// GM用户
func GmAdminUserInfo(c *gin.Context) {
	rawData, _ := c.GetRawData()

	var gmUserInfo gm_model.GMUserInfo
	err := json.Unmarshal(rawData, &gmUserInfo)
	if err != nil {
		log.Fatal("解析失败:", err)
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug(" gmUserInfo %v", gmUserInfo)
	//验证码不能为空
	if len(gmUserInfo.Xiaoxiaoxiyou) <= 0 {
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "accessToken not found")
		return
	}

	player := new(gm_model.GmAccount)
	player.UserName = gmUserInfo.UserName

	has, err := db.AccountDb.Table(define.Admin).Get(player)
	if err != nil {
		log.Error("account failed, err : %v", err)
		httpRetGame(c, ERR_DB, "db err")
		return
	}
	if !has {
		httpRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	if player.Token != gmUserInfo.Xiaoxiaoxiyou {
		httpRetGame(c, ERR_ACCOUNT_NOT_FOUND, "accessToken is mistake")
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

	httpRetGame(c, SUCCESS, "success",
		map[string]any{
			"username":    player.Name,
			"permissions": permiss,
			"avatar": []string{
				"https://gcore.jsdelivr.net/gh/zxwk1998/image/avatar/avatar_1.png",
				"https://gcore.jsdelivr.net/gh/zxwk1998/image/avatar/avatar_2.png",
				"https://gcore.jsdelivr.net/gh/zxwk1998/image/avatar/avatar_3.png",
				"https://gcore.jsdelivr.net/gh/zxwk1998/image/avatar/avatar_4.png",
				"https://gcore.jsdelivr.net/gh/zxwk1998/image/avatar/avatar_5.png",
			},
		})
}
