package logic

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"xfx/core/define"
	"xfx/gm_server/db"
	"xfx/gm_server/dto"
	"xfx/pkg/log"
	"xfx/pkg/utils"
)

// GmListAdminAccounts 列出所有 GM 账号（脱敏：不返回 token/password）
func GmListAdminAccounts(c *gin.Context) {
	var list []dto.GmAccount
	err := db.AccountDb.Table(define.AdminTable).Find(&list)
	if err != nil {
		log.Error("GmListAdminAccounts err: %v", err)
		HTTPRetGame(c, ERR_DB, "db err")
		return
	}
	result := make([]map[string]any, 0, len(list))
	for _, a := range list {
		result = append(result, map[string]any{
			"id":         a.Id,
			"username":   a.UserName,
			"name":       a.Name,
			"permission": a.Permission,
		})
	}
	HTTPRetGame(c, SUCCESS, "success", map[string]any{"data": result})
}

// GmCreateAdminAccount 创建 GM 账号
func GmCreateAdminAccount(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req struct {
		UserName   string `json:"username"`
		Password   string `json:"password"`
		Name       string `json:"name"`
		Permission int    `json:"permission"`
	}
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if req.UserName == "" || req.Password == "" {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "username and password required")
		return
	}
	if req.Permission < 1 || req.Permission > 3 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "permission must be 1, 2 or 3")
		return
	}

	// 检查是否已存在
	exist := new(dto.GmAccount)
	exist.UserName = req.UserName
	has, err := db.AccountDb.Table(define.AdminTable).Get(exist)
	if err != nil {
		HTTPRetGame(c, ERR_DB, "db err")
		return
	}
	if has {
		HTTPRetGame(c, ERR_ACCOUNT_EXISTS, "account already exists")
		return
	}

	hashedPwd := utils.MD5(req.Password)
	acc := &dto.GmAccount{
		UserName:   req.UserName,
		Password:   hashedPwd,
		Name:       req.Name,
		Permission: req.Permission,
	}
	if _, err := db.AccountDb.Table(define.AdminTable).Insert(acc); err != nil {
		log.Error("GmCreateAdminAccount insert err: %v", err)
		HTTPRetGame(c, ERR_DB, "db err")
		return
	}
	HTTPRetGame(c, SUCCESS, "success")
}

// GmUpdateAdminAccount 修改 GM 账号信息（name、permission、password 均可选更新）
func GmUpdateAdminAccount(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req struct {
		Id         int64  `json:"id"`
		Name       string `json:"name"`
		Password   string `json:"password"`
		Permission int    `json:"permission"`
	}
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if req.Id <= 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "id required")
		return
	}

	acc := new(dto.GmAccount)
	acc.Id = req.Id
	has, err := db.AccountDb.Table(define.AdminTable).Get(acc)
	if err != nil || !has {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	cols := []string{}
	if req.Name != "" {
		acc.Name = req.Name
		cols = append(cols, "name")
	}
	if req.Permission >= 1 && req.Permission <= 3 {
		acc.Permission = req.Permission
		cols = append(cols, "permission")
	}
	if req.Password != "" {
		acc.Password = utils.MD5(req.Password)
		cols = append(cols, "password")
	}
	if len(cols) == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "nothing to update")
		return
	}

	session := db.AccountDb.Table(define.AdminTable).Where("id = ?", req.Id)
	for _, col := range cols {
		session = session.MustCols(col)
	}
	if _, err := session.Update(acc); err != nil {
		log.Error("GmUpdateAdminAccount err: %v", err)
		HTTPRetGame(c, ERR_DB, "db err")
		return
	}
	HTTPRetGame(c, SUCCESS, "success")
}

// GmDeleteAdminAccount 删除 GM 账号
func GmDeleteAdminAccount(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req struct {
		Id int64 `json:"id"`
	}
	if err := json.Unmarshal(rawData, &req); err != nil || req.Id <= 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "id required")
		return
	}
	// 不允许删除自身（通过 token 取当前操作者）
	currentUser, _ := c.Get(ContextKeyGmUser)
	if cu, ok := currentUser.(*dto.GmAccount); ok && cu.Id == req.Id {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "cannot delete yourself")
		return
	}

	if _, err := db.AccountDb.Table(define.AdminTable).Where("id = ?", req.Id).Delete(&dto.GmAccount{}); err != nil {
		log.Error("GmDeleteAdminAccount err: %v", err)
		HTTPRetGame(c, ERR_DB, "db err")
		return
	}
	HTTPRetGame(c, SUCCESS, "success")
}
