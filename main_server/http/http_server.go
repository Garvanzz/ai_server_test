package http

import (
	"github.com/gin-gonic/gin"
	"xfx/core/define"
	"xfx/pkg/log"
	"xfx/pkg/module"
	"xfx/pkg/module/modules"
)

var Module = func() module.Module {
	httpModule := new(HttpModule)
	return httpModule
}

type HttpModule struct {
	modules.BaseModule
	router *gin.Engine
}

func (m *HttpModule) Version() string { return "1.0.0" }

func (m *HttpModule) GetType() string { return define.ModuleHttp }

func (m *HttpModule) OnInit(app module.App) {
	m.BaseModule.OnInit(app)
	//正式服直接改成发布模式
	gin.SetMode(gin.ReleaseMode)
	m.router = gin.New()
	m.router.Use(gin.Recovery())

	m.startHttpServer(app)
}

func (m *HttpModule) startHttpServer(app module.App) {
	m.register()

	go func() {
		err := m.router.Run(app.GetEnv().HttpUrl)
		if err != nil {
			log.Error("gin server run error:%v", err)
			return
		}
	}()
}
