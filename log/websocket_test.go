package log
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import(
	"os"
	"fmt"
	"io/ioutil"
	"time"
	"testing"
	"github.com/chelion/gws/fasthttp"
)

var TestHtml []byte = nil

func requestHandler(ctx *fasthttp.RequestCtx) {
	var err error
	ctx.SetContentType("text/html")
	if TestHtml != nil{
		fmt.Fprint(ctx, string(TestHtml))
	}else{
		TestHtml,err= readAll("./web/index.html")
		if nil == err{
			fmt.Fprint(ctx, string(TestHtml))
		}
	}
}

func readAll(filePth string) ([]byte, error) {
	f, err := os.Open(filePth)
	if err != nil {
	 return nil, err
	}
	return ioutil.ReadAll(f)
}

func TestWebSocketLog(t *testing.T){
	fmt.Println("start test websocket log,press Ctrl+C Stop test")
	go func(){
		logger,err := NewWebSocketLog("127.0.0.1:8089",true)//websocket 端口8089
		if nil != err{
			t.Errorf("get log console fail\n")
		}
		logger.Init()
		defer logger.DeInit()
		logger.SetLevel(DEBUG)
		for{
			time.Sleep(time.Duration(5)*time.Second)
			logger.Printf("printf,logger\n")
			logger.Println("println,logger")
			logger.Debug("debug")
			logger.Info("Info")
			logger.Warn("warn")
			logger.Error("error")
		}
	}()
	if err := fasthttp.ListenAndServe("127.0.0.1:8080", requestHandler); err != nil {//web server 端口8080
		t.Errorf("Error in ListenAndServe\n")
	}
}