package main

import (
	"context"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"monitor/config"
	"monitor/internal/api"
	"monitor/internal/service/scene"
	"monitor/internal/service/task"
	"time"
)

func main() {

	if err := config.InitConfig(); err != nil {
		log.Fatalf("无法加载配置文件: %v", err)
	}

	svc := api.NewServiceContext()
	engine := gin.Default()
	grafanaConf := config.Grafana
	iGrafanaService := scene.NewGrafanaService(&grafanaConf)
	sc := api.NewScene(iGrafanaService)
	ms := api.NewModelReq(iGrafanaService)
	lg := api.NewLedger(context.Background())
	// 配置CORS中间件

	//启动任务队列
	scheduler := task.NewDelayedTaskScheduler()
	scheduler.Start(3) // 单工作线程保证顺序

	engine.Use(cors.New(cors.Config{
		// 允许的域名列表
		AllowOriginFunc: func(origin string) bool {
			return true // 无限制接受任何来源
		},

		// 允许的HTTP方法
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},

		// 允许的请求头
		AllowHeaders: []string{
			"Authorization",
			"Content-Type",
			"Origin",
			"Accept",
			"X-Requested-With",
			"Accept",
			"Referer", // 显式添加 Referer
			"User-Agent",
			// 添加正则表达式捕获所有请求头
			".*",
		},
		// 暴露所有响应头
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
			"Authorization",
			"Set-Cookie",    // 需要特殊处理的头
			"Location",      // 重定向时需要
			"Access-Token",  // 如有自定义token
			"Refresh-Token", // 如有刷新token
			// 正则表达式匹配所有响应头
			".*",
		},
		// 是否允许携带凭证
		AllowCredentials: true,

		// 预检请求缓存时间
		MaxAge: 12 * time.Hour,

		AllowWildcard: true,

		// 允许WebSockets
		AllowWebSockets: true,

		// 允许文件上传
		AllowFiles: false,
	}))
	{
		monitor := engine.Group("/apis/gpu.monitor.io/monitor/list")
		monitor.Use(api.MakeToken())
		monitor.GET("/clusterinfo", svc.ListCluster)
		monitor.GET("/clustername", svc.ListClusterName)
		monitor.GET("/nodesinfo", svc.ListNodes)
		monitor.GET("/clusterdetail", svc.ClusterDetail)
		monitor.GET("/nodedetail", svc.NodeDetail)

		scene := engine.Group("/apis/gpu.monitor.io/scene")
		scene.Use(api.MakeToken())
		scene.GET("/list", sc.CountScenes)
		scene.GET("/models", sc.CountModels)
		scene.GET("/details", sc.CountModelDetail)

		model := engine.Group("/apis/gpu.monitor.io/model")
		model.Use(api.MakeToken())
		model.GET("/list", ms.ModelCards)         //外部模型列表
		model.GET("/logs", ms.ModelLogs)          //模型对应场景详情中的Logs
		model.GET("/timerecord", ms.ModelReqTime) //模型对应场景调用次数趋势
		model.GET("/details", ms.ModelDetail)     //模型对应场景

		ledger := engine.Group("/apis/gpu.monitor.io/ledger")
		ledger.Use(api.MakeToken())
		model.GET("/tasklist", lg.LedgerTasksList)   //任务列表
		model.GET("/Preview", lg.LedgerAllInfo)      //台账预览
		model.GET("/download", lg.DownloadLedger)    //下载台账
		model.POST("/saveledger", lg.GenerateLedger) //生成任务，生成台账
		model.POST("/savetask", lg.GenerateTask)     //生成任务，生成台账
	}
	addr := fmt.Sprintf(":%d", config.GetServerConfig().Port)

	engine.Run(addr)

}
