package main

import (
	"logger"
	"runtime"
	"stream_server"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	//指定是否控制台打印，默认为true
	logger.SetConsole(true)
	//指定日志文件备份方式为日期的方式
	//第一个参数为日志文件存放目录
	//第二个参数为日志文件命名
	logger.SetRollingDaily("logtest", "test.log")
	//指定日志级别  ALL，DEBUG，INFO，WARN，ERROR，FATAL，OFF 级别由低到高
	//一般习惯是测试阶段为debug，生成环境为info以上
	logger.SetLevel(logger.DEBUG)

	server := stream_server.NewStreamServer()
	server.Run()
}
