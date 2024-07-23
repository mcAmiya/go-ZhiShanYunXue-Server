package main

import (
	"ZhiShanYunXue/router"
	"ZhiShanYunXue/util"
	"fmt"
	"github.com/gin-gonic/gin"
)

func init() {

	//初始化完成输出Logo
	info := `
 __  /  |     _)   ___|   |                 \ \   /             \ \  /
    /   __ \   | \___ \   __ \    _` + "`" + ` |  __ \ \   /  |   |  __ \  \  /   |   |   _ \
   /    | | |  |       |  | | |  (   |  |   |   |   |   |  |   |    \   |   |   __/
 ____| _| |_| _| _____/  _| |_| \__,_| _|  _|  _|  \__,_| _|  _| _/\_\ \__,_| \___|

##############################################################################################`
	fmt.Println(info)

	logger, _ := util.NewLogger()
	logger.Info("程序初始化")

	// 非调试模式
	gin.SetMode(gin.ReleaseMode)

	// SQLite DataBase初始化
	util.InitSqlite()
}

func main() {
	// 日志
	logger, _ := util.NewLogger()
	logger.Info("Main函数运行中")

	logger.Infof("现已改为默认监听全部地址，下面的是本机地址")
	logger.Infof("IpAddress: http://localhost:24748/")

	r := router.InitRouter()
	err := r.Run(":24748")
	if err != nil {
		logger.Fatal("运行服务器时出错: ", err)
		return
	}
}
