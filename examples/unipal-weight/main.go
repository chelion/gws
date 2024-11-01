/*
 * @Author: your name
 * @Date: 2022-04-20 10:55:57
 * @LastEditTime: 2022-04-20 11:03:40
 * @LastEditors: Please set LastEditors
 * @Description: 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 * @FilePath: \gws\examples\unipal-weight\main.go
 */
package main

import (
	"fmt"
	"github.com/chelion/gws"
	"github.com/chelion/gws/configure"
	"github.com/chelion/gws/log"
	"runtime"
	"strconv"
	"time"
)

func NeedUpWeightData(ctx *gws.Context) {
	fmt.Println("NeedUpWeightData here 1")
	DeviceSN := ctx.Get("DeviceSN")
	fmt.Println(string(DeviceSN))
	FileSize := ctx.Get("FileSize")
	fmt.Println(string(FileSize))
	FileSizeInt, _ := strconv.Atoi(string(FileSize))
	if FileSizeInt >= 1*1024*1024 {
		ctx.String(200, "{\"Status\":1}")
	} else {
		ctx.String(200, "{\"Status\":0}")
	}

}

func UploadWeightData(ctx *gws.Context) {
	fmt.Println("here 1")
	FileName := ctx.Get("FileName")
	fmt.Println(string(FileName))
	fileHeader, err := ctx.FormFile(string(FileName))
	if nil != err {
		ctx.JSON(200, gws.M{
			"code": 1,
			"info": err.Error(),
		})
		return
	}
	fmt.Println("here 2")
	dst := "/usr/local/" + strconv.FormatInt(time.Now().UnixNano(), 10) + "__" + fileHeader.Filename
	fmt.Println("UploadWeightData", dst)
	if err := ctx.SaveUploadedFile(fileHeader, dst); err != nil {
		ctx.JSON(200, gws.M{
			"code": 1,
			"info": err.Error(),
		})
		return
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	config, err := configure.NewIniConfigure("./httpserver.ini")
	if nil != err {
		fmt.Println(err)
		return
	}
	config.Init()
	defer config.DeInit()
	logger, err := log.NewConsoleLog(true)
	if nil != err {
		fmt.Println(err)
		return
	}
	logger.Init()
	defer logger.DeInit()
	g, err := gws.New("http", "HttpServer", config, logger)
	if nil != err {
		fmt.Println(err)
		return
	}
	g.UseDefaultSession()
	g.GET("/", func(ctx *gws.Context) {
		ctx.Rctx.Write([]byte("gws is ok!"))
	})
	g.POST("/UploadWeightData", UploadWeightData)
	g.POST("/NeedUpWeightData", NeedUpWeightData)

	g.Start()
	defer g.Stop()
}
