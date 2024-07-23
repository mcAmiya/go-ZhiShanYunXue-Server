package router

import (
	"ZhiShanYunXue/api/middleware"
	v1 "ZhiShanYunXue/api/v1"
	"ZhiShanYunXue/setting"
	_ "embed"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
)

func setupStaticRoutes(r *gin.Engine) {
	// 自定义静态文件中间件，确保 MIME 类型正确设置
	staticHandler := static.Serve("/", static.LocalFile("./front", false))
	r.Use(staticHandler)

	// 当请求的路径未匹配到任何静态资源时，返回前端应用的入口HTML文件
	r.NoRoute(func(c *gin.Context) {
		indexPath := filepath.Join("./front", "index.html")
		http.ServeFile(c.Writer, c.Request, indexPath)
	})
}

func InitRouter() *gin.Engine {

	r := gin.Default()

	// 配置 CORS 中间件
	r.Use(middleware.Cors())

	// 配置静态资源路由
	setupStaticRoutes(r)

	apiBaseUrl := "/zsyx/api/" + setting.ApiVersion

	api := r.Group(apiBaseUrl)
	{

		// 任务 任务管理类
		task := api.Group("/tasks")
		{
			task.POST("/new_task", v1.NewTask)
			task.GET("/get_info", v1.GetInfo)
			task.GET("/get_task_data", v1.GetTaskData)
			task.GET("/get_report", v1.GetReport)
			task.GET("/get_status", v1.GetStatusReportData)
			task.POST("/push_answer", v1.PushAnswer)
		}
	}

	return r
}
