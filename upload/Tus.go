package uploads

import (
	"catenoid-company/tus-client/dto"
	"catenoid-company/tus-client/lib"
	"fmt"
	"github.com/eventials/go-tus"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type TusUploads struct {
	*TusUtils
}

func (t *TusUploads) RunTus(store tus.Store) *tus.Uploader {
	url := lib.HOST+lib.PATH
	tusConfig := &tus.Config{}

	uploadKey := t.Ctx.Query(lib.UPLOADQUERYKEYFILED)

	if  path, isPath := store.Get(uploadKey); path != "" && isPath{
		tusConfig.Store = store
		tusConfig.Resume = true
		tusConfig.OverridePatchMethod = false
		tusConfig.ChunkSize = lib.CHUNKSIZE
		url= path
	}else {
		tusConfig = nil
	}

	client, _ := tus.NewClient(url, tusConfig)
	// create an upload from a file.
	upload, err := t.getTusUpload()
	if err != nil {
		log.Println(err)
	}

	// create the uploader.
	uploader, _ := client.CreateOrResumeUpload(upload)

	//uploadProcess := make(chan tus.Upload)
	//
	//uploader.NotifyUploadProgress(uploadProcess)

	if tusConfig == nil {
		store.Set(uploadKey, uploader.Url())
	}

	// start the uploading process.
	return uploader
}

type TusUtilsInterface interface {
	getTusUpload() (*tus.Upload, error)
	TusFileCopy()
	DeleteContinuousFile(resResult *dto.ResponseDto) (*http.Response, error)
}

type TusUtils struct {
	Ctx    *gin.Context
	//Result *dto.ResponseDto
	Err    error
}

func (t *TusUtils) getTusUpload() (*tus.Upload, error) {
	cu := &CustomUtils{&CustomTusUtils{C: t.Ctx}}

	upload, err := cu.NewUploadFromFile()

	return upload, err
}

func (t *TusUtils) TusFileCopy() {
	var resp *http.Response
	var err error
	msg := make(map[string]interface{})

	cu := &CustomUtils{&CustomTusUtils{C: t.Ctx}}

	url := t.Ctx.PostForm(lib.FILEDOWNLOADNAME)

	if url == "" {
		msg["status"] = "Fail"
		msg["msg"] = "The parameters are not correct"
		//t.Result = msg
		t.Err = err
	}

	setHttpInfo := map[string]string{
		lib.URI:    url,
		lib.METHOD: "GET",
		lib.PARAMS: "",
	}

	cu.SendHttpInfo = setHttpInfo

	resp, err = cu.SendToHttp(nil)
	defer resp.Body.Close()

	if err != nil {
		log.Print(err)
		msg["status"] = "Fail"
		msg["msg"] = "Make sure the URL you requested is correct"
		//t.Result = msg
		t.Err = err
	}

	msg, err = cu.CreateAndCopyFromResFile(resp)

	if err != nil {
		//t.Result = msg
		t.Err = err
	}
}

func (t *TusUtils) DeleteContinuousFile() (*http.Response, error) {
	cu := &CustomUtils{&CustomTusUtils{C: t.Ctx}}

	setSendToFileInfo := map[string]string{
		lib.URI:    lib.HOST + "findFileInfo",
		lib.METHOD: "GET",
	}

	cu.SendHttpInfo = setSendToFileInfo

	resp, err := cu.SendToHttp(nil)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Println(err)
	}

	fileInfoDto := &dto.FileInfoDto{}
	err = lib.ConvertToUnMarshalJson(b, fileInfoDto)

	if err != nil {
		log.Println(err)
	}

	var m = map[string]string{
		lib.TUSRESUMEABLE:    lib.TUSRESUMEALBEVERSION,
		lib.TUSCONTENTLENGTH: strconv.Itoa(fileInfoDto.Size),
	}

	deleteFileName := cu.C.PostForm(lib.FILEDELETENAME)

	setSendToDelete := map[string]string{
		lib.URI:    lib.HOST + lib.PATH + deleteFileName,
		lib.METHOD: "DELETE",
		lib.PARAMS: "",
	}

	cu.SendHttpInfo = setSendToDelete

	return cu.SendToHttp(m)
}

func (t *TusUtils) TusProcessAbort(uploader *tus.Uploader, isComplete chan bool) {
	isDisconnect := <- t.Ctx.Writer.CloseNotify()
	if isDisconnect {
		uploader.Abort()
	}

	if uploader.IsAborted() {
		isComplete <- false
		log.Print("Tus DisConnected")
	}
}

func (t *TusUtils) TusProcessBar(TusProcessChan *chan tus.Upload, resResult *dto.ResponseDto, isComplete chan bool) {
	fmt.Print("progress")
	var op int64 = 0

	for {
		startingTime := time.Now().UTC()
		up, ok := <-*TusProcessChan
		if !ok {
			fmt.Print("chan closed\n")
			break
		}

		endingTime := time.Now().UTC()
		duration  := endingTime.Sub(startingTime)
		elapsedSec := duration.Seconds()
		speed := (float64)(lib.CHUNKSIZE) / 1024 / 1024 / elapsedSec

		p := up.Progress()
		if p == 100 || up.Finished() {
			processStr := "...100%,Done"
			fmt.Println(processStr)
			resResult.ProcessStatus = processStr
			isComplete <- true
			return
		}
		if p != op {
			op = p
			processStr := fmt.Sprint("...(", fmt.Sprintf("%.3f", speed), "MB/s)", p, "%")
			fmt.Println(processStr)
			resResult.ProcessStatus = processStr

		}
	}
}

func (t *TusUtils) TusCloseUpload(processStr chan tus.Upload, resResult *dto.ResponseDto, store tus.Store, isComplete chan bool) {
	if processStr == nil{
		if resResult.Status == http.StatusBadRequest{
			t.Ctx.AbortWithStatusJSON(http.StatusBadRequest, resResult)
			return
		}
	}

	defer close(processStr)

	//bMsg := lib.ConvertToMarshalJson(msg)
	//lib.ConvertToUnMarshalJson(bMsg,resResult)

	if resResult.ProcessStatus == http.StatusBadRequest{
		t.Ctx.AbortWithStatusJSON(http.StatusBadRequest, resResult)
		return
	}

	if isProcessCompete := <- isComplete; !isProcessCompete {
		resResult.ResultMessage =  "Not Complete to continuous file"
		t.Ctx.JSON(http.StatusOK,resResult)
		return
	}

	store.Delete(t.Ctx.Query(lib.UPLOADQUERYKEYFILED))
	t.Ctx.JSON(http.StatusOK,resResult)
}
