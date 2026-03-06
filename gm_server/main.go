package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"xfx/core/config"
	"xfx/gm_server/conf"
	"xfx/gm_server/logic"
	"xfx/pkg/log"
)

func main() {
	//日志处理
	log.Init(conf.Server.Log)

	//加载配置
	config.InitConfig("./json")

	//正式服直接改成发布模式
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())

	//无页面处理
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "PAGE_NOT_FOUND",
			"message": "Page not found",
		})
	})

	//GM
	r.POST("gmlogin", logic.GmLogin) //Gm登录
	r.POST("userPlayerInfo", logic.GmAdminUserInfo)
	r.POST("gmlogout", logic.GmLoginout)                               //GM退出
	r.POST("gmgetServerList", logic.GmGetServerList)                   //GM服务器列表
	r.POST("getPlayerInfo", logic.GmGetPlayerInfo)                     //GM玩家信息
	r.POST("getPlayerGameInfo", logic.GmGetPlayerGameInfo)             //GM玩家游戏信息
	r.POST("gmItem", logic.GmItem)                                     //GM玩家获取道具
	r.POST("gmAddItem", logic.GmAddItem)                               //GM玩家添加道具
	r.POST("gmDeleteItem", logic.GmDeleteItem)                         //GM玩家删除道具
	r.POST("gmEquip", logic.GmEquip)                                   //GM玩家获取装备
	r.POST("gmDeleteEquip", logic.GmDeleteEquip)                       //GM玩家删除装备
	r.POST("gmHero", logic.GmHero)                                     //GM玩家获取角色
	r.POST("gmEditHero", logic.GmEditHero)                             //GM玩家编辑角色
	r.POST("gmStartServer", logic.GmStartServer)                       //GM启动服务器
	r.POST("gmReStartServer", logic.GmReStartServer)                   //GM重启服务器
	r.POST("gmStopServer", logic.GmStopServer)                         //GM停止服务器
	r.POST("gmGetGameServerList", logic.GmGetGameServerList)           //GM获取游戏服务器列表
	r.POST("gmGetOrderList", logic.GmGetOrderList)                     //GM获取充值列表
	r.POST("gmGetOrderCacheList", logic.GmGetCacheOrderList)           //GM获取缓存充值列表
	r.POST("gmGetHotUpdate", logic.GmGetHotUpdate)                     //GM获取热更
	r.POST("gmEditHotUpdateVersion", logic.GmEditHotUpdateVersion)     //GM更改热更
	r.POST("gmCreateHotUpdateVersion", logic.GmCreateHotUpdateVersion) //GM创建热更
	r.POST("gmDeleteHotUpdateVersion", logic.GmDeleteHotUpdateVersion) //GM删除热更
	r.POST("gmCreateHotUpdatePath", logic.GmCreateHotUpdatePath)       //GM创建热更路径
	r.POST("upload", logic.Gmupload)                                   //GM上传文件
	r.POST("CreateAdminMail", logic.GmCreateAdminMail)                 //GM创建邮件
	r.POST("SetServerTime", logic.GmSetServerTime)                     //GM创建邮件
	r.POST("gmGetGuidList", logic.GmGetGuidList)                       //GM获取帮会列表
	r.POST("gmSendNotice", logic.GmSendNotice)                         //GM发送公告
	r.POST("gmUpdateConfig", logic.GmUpdateConfig)                     //GM更新配置表
	r.POST("gmBuildServer", logic.GmBuildServer)                       //GM编译服务器
	r.POST("gmOneKeyAddItem", logic.GmOneKeyAddItem)                   //GM一键添加道具
	r.POST("gmSendHorse", logic.GmSendHorse)                           //GM发送跑马灯
	r.POST("gmGameStartServer", logic.GmStartGameServer)               //GM启动游戏服务器
	r.POST("gmGameReStartServer", logic.GmReStartGameServer)           //GM重启游戏服务器
	r.POST("gmGameStopServer", logic.GmStopGameServer)                 //GM停止游戏服务器
	r.POST("gmGameGetGameServerList", logic.GmGetGameGameServerList)   //GM获取游戏游戏服务器列表
	r.POST("gmGameUpdateConfig", logic.GmGameUpdateConfig)             //GM更新配置表
	r.POST("gmGameBuildServer", logic.GmGameBuildServer)               //GM编译服务器
	r.POST("gmGetStageInfo", logic.GmGetStageInfo)                     //获取关卡列表信息
	r.POST("gmSetStageInfo", logic.GmSetStageInfo)                     //设置关卡信息
	r.POST("gmAddStageInfo", logic.GmAddStageInfo)                     //添加关卡信息
	r.POST("gmDeleteStageInfo", logic.GmDeleteStageInfo)               //删除关卡信息

	log.Debug("http service listen at %v", conf.Server.HttpPort)
	if err := http.ListenAndServe(conf.Server.HttpPort, r); err != nil {
		log.Fatal("ListenAndServe err : ", err)
	}
}
