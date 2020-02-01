package log
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import(
	"fmt"
	"testing"
)

func TestConsoleLog(t *testing.T){
	fmt.Println("start test console log")
	logger,err := NewConsoleLog(true)
	if nil != err{
		t.Errorf("get log console fail\n")
	}
	logger.Init()
	defer logger.DeInit()
	logger.SetLevel(DEBUG)
	logger.Printf("printf,logger\n")
	logger.Println("println,logger")
	logger.SetLevel(INFO)
	logger.Debug("debug")
	logger.Info("Info")
	logger.Error("error")
	logger.SetLevel(OFF)
	logger.Debug("off")
	fmt.Println("end test console log")
}