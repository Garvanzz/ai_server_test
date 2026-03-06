package logic

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"xfx/core/model"
	"xfx/gm_server/db"
	"xfx/gm_server/define"
	"xfx/gm_server/gm_model"
	"xfx/pkg/log"
)

const (
	SUCCESS                              = 0
	ERR_SERVER_INTERNAL                  = 1    // 服务器内部错误
	ERR_PAY_ORDER_NOT_FOUND              = 2    // 订单不存在
	ERR_PAY_SIGN                         = 3    // 签名不正确
	ERR_INVITE_CODE_FAIL                 = 29   // 邀请码验证不通过
	ERR_DB                               = 1801 // 数据库错误
	ERR_ACCOUNT_EXISTS                   = 1802 // 账号已存在
	ERR_ACCOUNT_PASSWORD_FAILED          = 1803 // 账号密码错误
	ERR_ACCOUNT_TYPE_UNKNOWN             = 1804 // 账号类型错误
	ERR_ACCOUNT_NOT_FOUND                = 1805 // 账号不存在
	ERR_ACCOUNT_VERIFY_CODE_INCORRECT    = 1806 // 验证码不正确
	ERR_ACCOUNT_GET_VERIFY_CODE_FAILED   = 1807 // 获取验证码失败
	ERR_ACCOUNT_REGISTER_CLOSED          = 1808 // 注册服务关闭
	ERR_ACCOUNT_LOGIN_SERVER_MAINTAIN    = 1809 // 服务器维护 其实只有白名单账号可以进
	ERR_ACCOUNT_BANNED                   = 1810 // 账号被ban中
	ERR_ACCOUNT_PARAMS_ERROR             = 1811 // 参数错误
	ERR_ACCOUNT_CLIENT_VERSION_UNMATCHED = 1812 // 客户端版本不匹配
	ERR_ACCOUNT_SDK_TOKEN_AUTH_FAILED    = 1813 // 登录SDK Token效验失败
	ERR_ACCOUNT_SDK_TOKEN_EXPIRED        = 1814 // 登录SDK Token过期
	ERR_ACCOUNT_HAS_NO_NFT_HERO          = 1815 // 帐号没有nft英雄
	ERR_ACCOUNT_FORCED_OFFLINE           = 1816 // 帐号强制下线中
	ERR_GIT_ERROR                        = 1817 // git错误
)

// 获取服务器列表
func GmGetServerList(c *gin.Context) {
	p := new(gm_model.ServerItem)
	if has, _ := db.AccountDb.IsTableExist(define.ServerGroup); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
		}
	}

	items := make([]gm_model.ServerItem, 0)
	err := db.AccountDb.Table(define.ServerGroup).Find(&items)
	if err != nil {
		log.Error("getserverlist find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	log.Debug("获取服务器列表:%s", items)
	servers := make([]*gm_model.GMGetServerList, 0)
	for i := 0; i < len(items); i++ {
		//这里要获取每个服务器的时间，后面去处理
		//暂时设置本服
		timestr := time.Now().Format("2006-01-02 15:04:05")
		servers = append(servers, &gm_model.GMGetServerList{
			Name: items[i].ServerName,
			Id:   items[i].Id,
			Time: timestr,
		})
	}

	js, _ := json.Marshal(servers)
	httpRetGame(c, SUCCESS, "success", map[string]any{
		"list": string(js),
	})
}

// 启动服务器
func GmStartServer(c *gin.Context) {
	p := new(gm_model.ServerItem)
	if has, _ := db.AccountDb.IsTableExist(define.ServerGroup); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
			httpRetGame(c, ERR_DB, err.Error())
			return
		}
	}

	rawData, _ := c.GetRawData()
	var result map[string]interface{}
	if err := json.Unmarshal(rawData, &result); err != nil {
		log.Error("GmStartServer find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	serverId, ok := result["serverId"].(float64)
	if !ok {
		log.Error("GmStartServer find serverName err")
		httpRetGame(c, ERR_DB, "GmStartServer find serverName err")
		return
	}

	log.Debug("请求服务中心数据:%v", serverId)

	serverItem := new(gm_model.ServerItem)
	serverItem.Id = int64(serverId)

	has, err := db.AccountDb.Table(define.ServerGroup).Get(serverItem)
	if err != nil {
		log.Error("getserverlist2 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	if !has {
		log.Error("getserverlist2 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	//查询是否运行
	cmd := exec.Command("pgrep", "-x", serverItem.ExeName)
	_, err = cmd.Output()
	if err == nil {
		log.Error("GmStartServer is run")
		httpRetGame(c, ERR_SERVER_INTERNAL, "GmStartServer is run")
		return
	} else {
		cmd := exec.Command(serverItem.ExePath)
		cmd.Dir = "/usr/local/games/xiyou/server"
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		err = cmd.Start()
		if err != nil {
			log.Error("Start failed: %v, Stderr: %s", err, stderr.String())
			httpRetGame(c, ERR_DB, fmt.Sprintf("start failed: %v", err))
			return
		}

		// 短暂检查进程是否启动成功
		time.Sleep(1 * time.Second)
		if cmd.Process == nil {
			httpRetGame(c, ERR_SERVER_INTERNAL, "process failed to start")
			return
		}

		httpRetGame(c, SUCCESS, "success")

		// 异步回收进程
		go func() {
			if err := cmd.Wait(); err != nil {
				log.Error("Process crashed: %v, Stderr: %s", err, stderr.String())
				// 可选：通过回调或日志服务通知崩溃事件
			}
		}()
	}
}

// 停止服务器
func GmStopServer(c *gin.Context) {
	p := new(gm_model.ServerItem)
	if has, _ := db.AccountDb.IsTableExist(define.ServerGroup); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
			httpRetGame(c, ERR_DB, err.Error())
			return
		}
	}

	rawData, _ := c.GetRawData()
	var result map[string]interface{}
	if err := json.Unmarshal(rawData, &result); err != nil {
		log.Error("GmStartServer find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	serverId, ok := result["serverId"].(float64)
	if !ok {
		log.Error("GmStartServer find serverName err")
		httpRetGame(c, ERR_DB, "GmStartServer find serverName err")
		return
	}

	log.Debug("请求服务中心数据:%v", serverId)

	serverItem := new(gm_model.ServerItem)
	serverItem.Id = int64(serverId)

	has, err := db.AccountDb.Table(define.ServerGroup).Get(serverItem)
	if err != nil {
		log.Error("getserverlist2 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	if !has {
		log.Error("getserverlist2 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	//查询是否运行
	cmd := exec.Command("pgrep", "-x", serverItem.ExeName)
	output, err := cmd.Output()
	if err == nil {
		pidStr := strings.TrimSpace(string(output))
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			log.Error("pid find has err :%v", err.Error())
			httpRetGame(c, ERR_DB, err.Error())
			return
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			log.Error("Failed to find process: %v\n", err)
			httpRetGame(c, ERR_DB, err.Error())
			return
		}

		err = process.Signal(syscall.SIGTERM)
		if err != nil {
			log.Error("Failed to find process: %v\n", err)
			httpRetGame(c, ERR_DB, err.Error())
			return
		}
	} else {
		log.Error("GmStopServer find  pid has err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	httpRetGame(c, SUCCESS, "success")
}

// 重启服务器
func GmReStartServer(c *gin.Context) {
	p := new(gm_model.ServerItem)
	if has, _ := db.AccountDb.IsTableExist(define.ServerGroup); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
			httpRetGame(c, ERR_DB, err.Error())
			return
		}
	}

	rawData, _ := c.GetRawData()
	var result map[string]interface{}
	if err := json.Unmarshal(rawData, &result); err != nil {
		log.Error("GmStartServer find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	serverId, ok := result["serverId"].(float64)
	if !ok {
		log.Error("GmStartServer find serverName err")
		httpRetGame(c, ERR_DB, "GmStartServer find serverName err")
		return
	}

	log.Debug("请求服务中心数据:%v", serverId)

	serverItem := new(gm_model.ServerItem)
	serverItem.Id = int64(serverId)

	has, err := db.AccountDb.Table(define.ServerGroup).Get(serverItem)
	if err != nil {
		log.Error("getserverlist2 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	if !has {
		log.Error("getserverlist2 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	//查询是否运行
	cmd := exec.Command("pgrep", "-x", serverItem.ExeName)
	output, err := cmd.Output()
	if err == nil {
		pidStr := strings.TrimSpace(string(output))
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			log.Error("pid find has err :%v", err.Error())
			httpRetGame(c, ERR_DB, err.Error())
			return
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			log.Error("Failed to find process: %v\n", err)
			httpRetGame(c, ERR_DB, err.Error())
			return
		}

		err = process.Signal(syscall.SIGTERM)
		if err != nil {
			log.Error("Failed to find process: %v\n", err)
			httpRetGame(c, ERR_DB, err.Error())
			return
		}
	}

	cmd = exec.Command(serverItem.ExePath)
	err = cmd.Start() // 异步启动（不阻塞）
	if err != nil {
		log.Error("GmGetServer find Start err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	httpRetGame(c, SUCCESS, "success")

	go func() {
		if err := cmd.Wait(); err != nil { // 必须调用 Wait() 回收子进程
			log.Error("Command failed:", err)
		}
	}()
}

// 获取游戏服务器列表
func GmGetGameServerList(c *gin.Context) {
	p := new(gm_model.ServerItem)
	if has, _ := db.AccountDb.IsTableExist(define.ServerGroup); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
		}
	}

	log.Debug("请求游戏服列表")

	var serverItem []gm_model.ServerItem
	err := db.AccountDb.Table(define.ServerGroup).Find(&serverItem)
	if err != nil {
		log.Error("getserverlist2 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}
	items := make([]*gm_model.GMRespServerItem, 0)
	for i := 0; i < len(serverItem); i++ {
		t := time.Unix(serverItem[i].OpenServerTime, 8)
		openTime := t.Format("2006-01-02 15:04:05")

		t1 := time.Unix(serverItem[i].StopServerTime, 8)
		closeTime := t1.Format("2006-01-02 15:04:05")

		State := "正常"
		if serverItem[i].ServerState == 0 {
			State = "正常"
		} else if serverItem[i].ServerState == 1 {
			State = "拥挤"
		} else if serverItem[i].ServerState == 2 {
			State = "爆满"
		} else if serverItem[i].ServerState == 3 {
			State = "维护"
		} else if serverItem[i].ServerState == 4 {
			State = "未开服"
		} else if serverItem[i].ServerState == 5 {
			State = "停服"
		}

		//查询是否运行
		cmd := exec.Command("pgrep", "-x", serverItem[i].ExeName)
		err := cmd.Run()
		state := ""
		if err == nil {
			state = "运行中"
		} else {
			state = "离线"
		}

		var gameServerItem gm_model.GameServerItem
		_, err = db.AccountDb.Table(define.GameServer).Where("id = ? ", serverItem[i].GameServer).Get(&gameServerItem)
		if err != nil {
			log.Error("getserverlist2 find err :%v", err.Error())
			httpRetGame(c, ERR_DB, err.Error())
			continue
		}

		items = append(items, &gm_model.GMRespServerItem{
			Id:             serverItem[i].Id,
			Ip:             serverItem[i].Ip,
			Port:           serverItem[i].Port,
			Channel:        serverItem[i].Channel,
			Group:          serverItem[i].ServerGroup,
			ServerName:     serverItem[i].ServerName,
			RedisPort:      serverItem[i].RedisPort,
			MysqlAddr:      serverItem[i].MysqlAddr,
			LoginServerUrl: serverItem[i].LoginServerUrl,
			OpenServerTime: openTime,
			StopServerTime: closeTime,
			ServerState:    State,
			RunState:       state,
			GameServer:     gameServerItem.ServerName,
			GameServerId:   int(gameServerItem.Id),
		})
	}

	js, _ := json.Marshal(items)

	httpRetGame(c, SUCCESS, "success", map[string]any{
		"data":       string(js),
		"totalCount": len(items),
	})
}

// 获取热更列表
func GmGetHotUpdate(c *gin.Context) {
	p := new(gm_model.HotUpdateItem)
	if has, _ := db.AccountDb.IsTableExist(define.HotUpdate); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
		}
	}

	log.Debug("请求热更列表")

	var hotItem []gm_model.HotUpdateItem
	err := db.AccountDb.Table(define.HotUpdate).Find(&hotItem)
	if err != nil {
		log.Error("getserverlist2 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	items := make([]*gm_model.GMRespHotUpdateItem, 0)
	for i := 0; i < len(hotItem); i++ {
		items = append(items, &gm_model.GMRespHotUpdateItem{
			Id:          hotItem[i].Id,
			Channel:     hotItem[i].Channel,
			ChannelName: hotItem[i].ChannelName,
			Version:     hotItem[i].Version,
		})
	}
	js, _ := json.Marshal(items)

	httpRetGame(c, SUCCESS, "success", map[string]any{
		"data":       string(js),
		"totalCount": len(items),
	})
}

// 编辑热更
func GmEditHotUpdateVersion(c *gin.Context) {
	p := new(gm_model.HotUpdateItem)
	if has, _ := db.AccountDb.IsTableExist(define.HotUpdate); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
		}
	}

	rawData, _ := c.GetRawData()
	var result map[string]interface{}
	if err := json.Unmarshal(rawData, &result); err != nil {
		log.Error("GmStartServer find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	log.Debug("请求热更编辑:%v", result)
	channel, ok := result["channel"].(string)
	if !ok {
		log.Error("GmStartServer find serverName err")
		httpRetGame(c, ERR_DB, "GmStartServer find serverName err")
		return
	}

	version, ok := result["version"].(string)
	if !ok {
		log.Error("GmStartServer find serverName err")
		httpRetGame(c, ERR_DB, "GmStartServer find serverName err")
		return
	}

	hotItem := new(gm_model.HotUpdateItem)
	hotItem.Channel = channel
	has, err := db.AccountDb.Table(define.HotUpdate).Get(hotItem)
	if err != nil {
		log.Error("getserverlist2 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}
	if !has {
		log.Error("getserverlist2 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	hotItem.Version = version

	//更新数据库
	db.AccountDb.Table(define.HotUpdate).Where("id=?", hotItem.Id).Cols("version").Update(hotItem)

	httpRetGame(c, SUCCESS, "success")
}

// 创建热更
func GmCreateHotUpdateVersion(c *gin.Context) {
	p := new(gm_model.HotUpdateItem)
	if has, _ := db.AccountDb.IsTableExist(define.HotUpdate); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
		}
	}

	log.Debug("请求热更创建")
	rawData, _ := c.GetRawData()
	var result map[string]interface{}
	if err := json.Unmarshal(rawData, &result); err != nil {
		log.Error("GmStartServer find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	channel, ok := result["channel"].(string)
	if !ok {
		log.Error("GmStartServer find serverName err")
		httpRetGame(c, ERR_DB, "GmStartServer find serverName err")
		return
	}

	channelName, ok := result["channelName"].(string)
	if !ok {
		log.Error("GmStartServer find serverName err")
		httpRetGame(c, ERR_DB, "GmStartServer find serverName err")
		return
	}

	version, ok := result["version"].(string)
	if !ok {
		log.Error("GmStartServer find serverName err")
		httpRetGame(c, ERR_DB, "GmStartServer find serverName err")
		return
	}

	//先查询是否存在
	has, err := db.AccountDb.Table(define.HotUpdate).Where("channel=?", channel).Exist()
	if err != nil {
		log.Error("getserverlist2 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}
	if has {
		log.Error("getserverlist3 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	hotItem := new(gm_model.HotUpdateItem)
	hotItem.Channel = channel
	hotItem.ChannelName = channelName
	hotItem.Version = version

	//插入数据库
	db.AccountDb.Table(define.HotUpdate).Insert(hotItem)

	httpRetGame(c, SUCCESS, "success")
}

// 删除热更
func GmDeleteHotUpdateVersion(c *gin.Context) {
	p := new(gm_model.HotUpdateItem)
	if has, _ := db.AccountDb.IsTableExist(define.HotUpdate); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
		}
	}

	rawData, _ := c.GetRawData()
	var result map[string]interface{}
	if err := json.Unmarshal(rawData, &result); err != nil {
		log.Error("GmStartServer find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}
	log.Debug("请求热更删除, %v", result)
	channels, ok := result["channel"].([]interface{})
	if !ok {
		log.Error("GmStartServer find serverName err")
		httpRetGame(c, ERR_DB, "GmStartServer find serverName err")
		return
	}

	for i := 0; i < len(channels); i++ {
		channel := channels[i].(string)
		//先查询是否存在
		has, err := db.AccountDb.Table(define.HotUpdate).Where("channel=?", channel).Exist()
		if err != nil {
			log.Error("getserverlist2 find err :%v", err.Error())
			httpRetGame(c, ERR_DB, err.Error())
			continue
		}
		if !has {
			log.Error("getserverlist3 find err :%v", err.Error())
			httpRetGame(c, ERR_DB, err.Error())
			continue
		}

		db.AccountDb.Table(define.HotUpdate).Where("channel=?", channel).Delete()
	}

	httpRetGame(c, SUCCESS, "success")
}

// 创建热更路径
func GmCreateHotUpdatePath(c *gin.Context) {
	p := new(gm_model.HotUpdateItem)
	if has, _ := db.AccountDb.IsTableExist(define.HotUpdate); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
		}
	}

	rawData, _ := c.GetRawData()
	var result map[string]interface{}
	if err := json.Unmarshal(rawData, &result); err != nil {
		log.Error("GmStartServer find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}
	log.Debug("请求创建热更路径, %v", result)
	channel, ok := result["channel"].(string)
	if !ok {
		log.Error("GmStartServer find serverName err")
		httpRetGame(c, ERR_DB, "GmStartServer find serverName err")
		return
	}

	version, ok := result["version"].(string)
	if !ok {
		log.Error("GmStartServer find serverName err")
		httpRetGame(c, ERR_DB, "GmStartServer find serverName err")
		return
	}

	//获取路径文件夹是否存在，不存在就创建，： /usr/local/games/xiyou/hotupdate
	hotUpdatePath := fmt.Sprintf("/usr/local/games/xiyou/hotupdate/%s/%s", channel, version)

	// 1. 首先检查路径是否已存在
	if _, err := os.Stat(hotUpdatePath); err == nil {
		// 目录已存在
		log.Error("热更目录已存在: %s", hotUpdatePath)
		httpRetGame(c, SUCCESS, "success", map[string]any{
			"state": true,
		})
		return
	} else if !os.IsNotExist(err) {
		// 不是"不存在"错误，是其他错误
		log.Error("检查目录失败: %v", err)
		httpRetGame(c, ERR_DB, "检查热更目录失败")
		return
	}

	// 2. 创建目录（包括所有必要父目录）
	if err := os.MkdirAll(hotUpdatePath, 0755); err != nil {
		log.Error("创建目录失败: %v", err)
		httpRetGame(c, ERR_DB, "无法创建热更目录")
		return
	}

	// 3. 验证目录是否创建成功
	if _, err := os.Stat(hotUpdatePath); err != nil {
		log.Error("验证目录创建失败: %v", err)
		httpRetGame(c, ERR_DB, "热更目录验证失败")
		return
	}

	// 4. 设置合适的目录权限（如果需要）
	if err := os.Chmod(hotUpdatePath, 0755); err != nil {
		log.Error("设置目录权限失败: %v", err)
		// 这里不返回错误，因为目录创建成功了
	}

	httpRetGame(c, SUCCESS, "success", map[string]any{
		"state": true,
	})
}

// GM上传文件
func Gmupload(c *gin.Context) {
	// 单文件
	file, err := c.FormFile("file")
	if err != nil {
		httpRetGame(c, ERR_DB, "热更目录验证失败")
		return
	}

	Channel := c.PostForm("Channel")
	Version := c.PostForm("Version")

	// 生成保存路径
	filename := filepath.Base(file.Filename)
	uploadDir := filepath.Join("/usr/local/games/xiyou/hotupdate/", Channel, Version)
	dst := filepath.Join(uploadDir, filename)

	// 确保目录存在并设置权限
	if err := os.MkdirAll(uploadDir, 0777); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建目录失败", "details": err.Error()})
		return
	}

	// 保存上传的 ZIP 文件
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败", "details": err.Error()})
		return
	}

	// 修改文件的拥有者为 nginx 用户
	if err := os.Chown(dst, 1001, 1001); err != nil { // 1001 是 nginx 用户和组的 ID
		c.JSON(http.StatusInternalServerError, gin.H{"error": "修改文件拥有者失败", "details": err.Error()})
		return
	}

	// 解压 ZIP 文件
	if err := unzip(dst, uploadDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解压文件失败", "details": err.Error()})
		return
	}

	// 删除原始 ZIP 文件（可选）
	if err := os.Remove(dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除 ZIP 文件失败", "details": err.Error()})
		return
	}

	httpRetGame(c, SUCCESS, "success", map[string]any{
		"filename": filename,
	})
}

// unzip 解压 ZIP 文件到目标目录
func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// 防止 ZIP 路径遍历攻击（Zip Slip）
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("非法文件路径: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			// 如果是目录，创建目录
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// 确保父目录存在
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		// 创建目标文件
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
		if err != nil {
			return err
		}
		defer outFile.Close()

		// 显式设置权限
		if err := outFile.Chmod(0777); err != nil {
			return err
		}

		// 打开 ZIP 文件
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		// 复制内容
		_, err = io.Copy(outFile, rc)
		if err != nil {
			return err
		}
	}

	return nil
}

// 设置服务器时间
func GmSetServerTime(c *gin.Context) {
	var Info gm_model.GmSetServerTime
	if err := c.ShouldBindJSON(&Info); err != nil {
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug(" Info %v", Info)
	//判断时间
	t, err := time.ParseInLocation("2006-01-02 15:04:05", Info.SetTime, time.Local)
	if err != nil {
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	if t.Unix() <= time.Now().Unix() {
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "fail")
		return
	}

	// 执行date命令修改系统时间
	timeStr := t.Format("2006-01-02 15:04:05")
	cmd := exec.Command("date", "-s", timeStr)
	if err := cmd.Run(); err != nil {
		fmt.Errorf("设置时间失败: %v", err)
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	httpRetGame(c, SUCCESS, "success")
}

// 启动游戏服务器
func GmStartGameServer(c *gin.Context) {
	p := new(model.ServerItem)
	if has, _ := db.AccountDb.IsTableExist(define.GameServer); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
			httpRetGame(c, ERR_DB, err.Error())
			return
		}
	}

	rawData, _ := c.GetRawData()
	var result map[string]interface{}
	if err := json.Unmarshal(rawData, &result); err != nil {
		log.Error("GmStartGameServer find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	serverId, ok := result["serverId"].(float64)
	if !ok {
		log.Error("GmStartGameServer find serverName err")
		httpRetGame(c, ERR_DB, "GmStartGameServer find serverName err")
		return
	}

	log.Debug("请求服务中心数据:%v", serverId)

	serverItem := new(gm_model.GameServerItem)
	serverItem.Id = int64(serverId)

	has, err := db.AccountDb.Table(define.GameServer).Get(serverItem)
	if err != nil {
		log.Error("getserverlist2 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	if !has {
		log.Error("getserverlist2 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	//查询是否运行
	cmd := exec.Command("pgrep", "-x", serverItem.ExeName)
	_, err = cmd.Output()
	if err == nil {
		log.Error("GmStartServer is run")
		httpRetGame(c, ERR_SERVER_INTERNAL, "GmStartServer is run")
		return
	} else {
		cmd := exec.Command(serverItem.ExePath)
		cmd.Dir = "/usr/local/games/xiyou/server"
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		err = cmd.Start()
		if err != nil {
			log.Error("Start failed: %v, Stderr: %s", err, stderr.String())
			httpRetGame(c, ERR_DB, fmt.Sprintf("start failed: %v", err))
			return
		}

		// 短暂检查进程是否启动成功
		time.Sleep(1 * time.Second)
		if cmd.Process == nil {
			httpRetGame(c, ERR_SERVER_INTERNAL, "process failed to start")
			return
		}

		httpRetGame(c, SUCCESS, "success")

		// 异步回收进程
		go func() {
			if err := cmd.Wait(); err != nil {
				log.Error("Process crashed: %v, Stderr: %s", err, stderr.String())
				// 可选：通过回调或日志服务通知崩溃事件
			}
		}()
	}
}

// 停止游戏服务器
func GmStopGameServer(c *gin.Context) {
	p := new(gm_model.GameServerItem)
	if has, _ := db.AccountDb.IsTableExist(define.GameServer); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
			httpRetGame(c, ERR_DB, err.Error())
			return
		}
	}

	rawData, _ := c.GetRawData()
	var result map[string]interface{}
	if err := json.Unmarshal(rawData, &result); err != nil {
		log.Error("GmStartGameServer find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	serverId, ok := result["serverId"].(float64)
	if !ok {
		log.Error("GmStartGameServer find serverName err")
		httpRetGame(c, ERR_DB, "GmStartGameServer find serverName err")
		return
	}

	log.Debug("请求服务中心数据:%v", serverId)

	serverItem := new(gm_model.GameServerItem)
	serverItem.Id = int64(serverId)

	has, err := db.AccountDb.Table(define.GameServer).Get(serverItem)
	if err != nil {
		log.Error("getserverlist2 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	if !has {
		log.Error("getserverlist2 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	//查询是否运行
	cmd := exec.Command("pgrep", "-x", serverItem.ExeName)
	output, err := cmd.Output()
	if err == nil {
		pidStr := strings.TrimSpace(string(output))
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			log.Error("pid find has err :%v", err.Error())
			httpRetGame(c, ERR_DB, err.Error())
			return
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			log.Error("Failed to find process: %v\n", err)
			httpRetGame(c, ERR_DB, err.Error())
			return
		}

		err = process.Signal(syscall.SIGTERM)
		if err != nil {
			log.Error("Failed to find process: %v\n", err)
			httpRetGame(c, ERR_DB, err.Error())
			return
		}
	} else {
		log.Error("GmStopServer find  pid has err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	httpRetGame(c, SUCCESS, "success")
}

// 重启游戏服务器
func GmReStartGameServer(c *gin.Context) {
	p := new(gm_model.GameServerItem)
	if has, _ := db.AccountDb.IsTableExist(define.GameServer); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
			httpRetGame(c, ERR_DB, err.Error())
			return
		}
	}

	rawData, _ := c.GetRawData()
	var result map[string]interface{}
	if err := json.Unmarshal(rawData, &result); err != nil {
		log.Error("GmStartServer find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	serverId, ok := result["serverId"].(float64)
	if !ok {
		log.Error("GmStartServer find serverName err")
		httpRetGame(c, ERR_DB, "GmStartServer find serverName err")
		return
	}

	log.Debug("请求服务中心数据:%v", serverId)

	serverItem := new(gm_model.GameServerItem)
	serverItem.Id = int64(serverId)

	has, err := db.AccountDb.Table(define.GameServer).Get(serverItem)
	if err != nil {
		log.Error("getserverlist2 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	if !has {
		log.Error("getserverlist2 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	//查询是否运行
	cmd := exec.Command("pgrep", "-x", serverItem.ExeName)
	output, err := cmd.Output()
	if err == nil {
		pidStr := strings.TrimSpace(string(output))
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			log.Error("pid find has err :%v", err.Error())
			httpRetGame(c, ERR_DB, err.Error())
			return
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			log.Error("Failed to find process: %v\n", err)
			httpRetGame(c, ERR_DB, err.Error())
			return
		}

		err = process.Signal(syscall.SIGTERM)
		if err != nil {
			log.Error("Failed to find process: %v\n", err)
			httpRetGame(c, ERR_DB, err.Error())
			return
		}
	}

	cmd = exec.Command(serverItem.ExePath)
	err = cmd.Start() // 异步启动（不阻塞）
	if err != nil {
		log.Error("GmGetServer find Start err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	httpRetGame(c, SUCCESS, "success")

	go func() {
		if err := cmd.Wait(); err != nil { // 必须调用 Wait() 回收子进程
			log.Error("Command failed:", err)
		}
	}()
}

// 获取游戏游戏服务器列表
func GmGetGameGameServerList(c *gin.Context) {
	p := new(gm_model.GameServerItem)
	if has, _ := db.AccountDb.IsTableExist(define.GameServer); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
		}
	}

	log.Debug("请求游戏服列表")

	var serverItem []gm_model.GameServerItem
	err := db.AccountDb.Table(define.GameServer).Find(&serverItem)
	if err != nil {
		log.Error("getserverlist2 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}
	items := make([]*gm_model.GMGameRespServerItem, 0)
	for i := 0; i < len(serverItem); i++ {
		//查询是否运行
		cmd := exec.Command("pgrep", "-x", serverItem[i].ExeName)
		err := cmd.Run()
		state := ""
		if err == nil {
			state = "运行中"
		} else {
			state = "离线"
		}

		items = append(items, &gm_model.GMGameRespServerItem{
			Id:         serverItem[i].Id,
			Ip:         serverItem[i].Ip,
			Port:       serverItem[i].Port,
			ServerName: serverItem[i].ServerName,
			RedisPort:  serverItem[i].RedisPort,
			MysqlAddr:  serverItem[i].MysqlAddr,
			RunState:   state,
		})
	}

	js, _ := json.Marshal(items)

	httpRetGame(c, SUCCESS, "success", map[string]any{
		"data":       string(js),
		"totalCount": len(items),
	})
}
