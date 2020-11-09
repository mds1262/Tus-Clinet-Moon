package main

import (
	"github.com/gin-gonic/gin"
	"github.com/mds1262/Tus-Clinet-Moon/controller"
	"github.com/mds1262/Tus-Clinet-Moon/lib"

	"log"
	"runtime"
)

/**
2020.11.06
Writer: Deuksoo Moon
Content: Tus을 이용한 파일 어이받기 서비스
*/

func main() {
	// Log 설정

	//logger := &lumberjack.Logger{
	//	Filename: lib.LOGPATH,
	//	MaxSize: 100,
	//	MaxBackups: 3,
	//	MaxAge: 28,
	//}
	//
	//defer logger.Close()
	//
	//gin.DefaultWriter = logger
	//
	//log.SetOutput(logger)

	app := gin.Default()

	redisClient := lib.SetRedisConn()

	defer redisClient.Close()

	handlers := &controller.Handlers{Client: redisClient}

	runtime.GOMAXPROCS(runtime.NumCPU())

	tgr := app.Group("/tus")
	{
		tgr.POST("/continuousFile", handlers.UploadContinuousHandle)
		tgr.POST("/moveToUploadFile", handlers.GetContinueUploadFile)
		tgr.POST("/deleteToUploadFile", handlers.DeleteHandle)
	}

	err := app.Run(":8082")

	if err != nil {
		log.Print(err)
	}

}
