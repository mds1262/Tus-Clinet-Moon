package controller

import (
	"catenoid-company/tus-client/dto"
	"catenoid-company/tus-client/lib"
	"catenoid-company/tus-client/upload"
	"github.com/eventials/go-tus"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"log"
	"net/http"
)

type Handlers struct {
	*redis.Client
}

/**
2020.11.06
Writer: Deuksoo Moon
Content: Tus의 파일 생성 및 이어받기
*/

func (h *Handlers) UploadContinuousHandle(c *gin.Context) {
	//Tus Upload와 관련된 함수와 구조체 모음
	tusUtils := uploads.TusUploads{TusUtils: &uploads.TusUtils{Ctx: c}}

	// 최종결과값을 처리하는 구조체
	resResult := &dto.ResponseDto{
			Status: http.StatusOK,
			ResultMessage:  "Complete to continuous file",
	}

	// uploadKey가 실제 존재하는지 확인
	if hKeys,_ := h.HGetAll(c.Query(lib.UPLOADQUERYKEYFILED)).Result(); len(hKeys) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest,"Not find to Upload-Key")
		return
	}

	// Tus의 Redis에 저장하기위해 설정
	tusStore := lib.NewTusStore(h.Client)

	//Tus Uploader 설정
	uploader := tusUtils.RunTus(tusStore)

	// Upload 완료 후 확인용
	IsComplete := make(chan bool)

	// 취소시 사용
	go tusUtils.TusProcessAbort(uploader, IsComplete)

	// Tus 진행상황에 사용
	progressChan := make(chan tus.Upload)

	// Tus 진행상황
	uploader.NotifyUploadProgress(progressChan)

	// Tus 진행상황 처리
	go tusUtils.TusProcessBar(&progressChan, resResult, IsComplete)

	// 최종 종료
	defer tusUtils.TusCloseUpload(progressChan, resResult, tusStore, IsComplete)

	// 업로드 시작
	err := uploader.Upload()

	if err != nil {
		resResult = &dto.ResponseDto{
			Status: http.StatusBadRequest,
			ResultMessage:    "Fail to continuous file",
		}
		IsComplete <- false
	}
	// "X-Request-ID"
	//c.JSON(http.StatusOK, msg)
}

func (h *Handlers) DeleteHandle(c *gin.Context) {
	resResult := &dto.ResponseDto{
		Status: http.StatusOK,
		ResultMessage:    "Complete to continuous file",
	}

	//var msg = map[string]interface{}{
	//	"status": "Success",
	//	"msg":    "Delete to upload file",
	//}

	tu := uploads.TusUploads{TusUtils: &uploads.TusUtils{Ctx: c}}

	defer tu.TusUtils.TusCloseUpload(nil, resResult, nil, nil)

	_, err := tu.DeleteContinuousFile()

	if err != nil {
		resResult = &dto.ResponseDto{
			Status: http.StatusBadRequest,
			ResultMessage:    "The headers are not correct",
		}
		//msg["status"] = "Fail"
		//msg["msg"] = "The headers are not correct"
		//c.AbortWithStatusJSON(http.StatusBadRequest, msg)
	}

	//c.JSON(http.StatusOK, msg)
}

func (h *Handlers) GetContinueUploadFile(c *gin.Context) {
	resResult := &dto.ResponseDto{
		Status: http.StatusOK,
		ResultMessage:  "Complete to continuous file",
	}

	var err error

	tu := uploads.TusUploads{TusUtils: &uploads.TusUtils{Ctx: c, Err: err}}

	defer tu.TusUtils.TusCloseUpload(nil, resResult, nil, nil)

	tu.TusFileCopy()

	if tu.Err != nil {
		log.Print(tu.Err)
		resResult = &dto.ResponseDto{
			Status: http.StatusBadRequest,
			ResultMessage:   "Fail to file copy",
		}
		//c.AbortWithStatusJSON(http.StatusBadRequest, msg)
	}
	//c.JSON(http.StatusOK, msg)
}
