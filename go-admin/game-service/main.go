package main

import (
	_ "game-service/models"
	_ "game-service/routers"
	_ "game-service/yalla/models"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	// 初始化日志
	logs.SetLogger(logs.AdapterFile, `{"filename":"logs/game-service.log","level":7,"maxlines":0,"maxsize":0,"daily":true,"maxdays":10}`)
}

func main() {
	// 设置静态文件路径
	web.SetStaticPath("/static", "static")

	// 启动Web服务
	web.Run()
}
