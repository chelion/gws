package utils
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

import(
	"fmt"
	"time"
	"testing"
)
func TestTimingWheel(t *testing.T){
	timeCnt := 0
	timerWheel := NewTimingWheel(uint32(GetNowUnixSec()))
	timerWheel.AddTimer(NewTimer(5,2,func(args interface{}){
		fmt.Println(args.(string))
	},"timer1"))
	timerWheel.AddTimer(NewTimer(10,1,func(args interface{}){
		fmt.Println(args.(string))
	},"timer2"))
	timerWheel.AddTimer(NewTimer(89,2,func(args interface{}){
		fmt.Println(args.(string))
	},"timer3"))
	timerWheel.Start()
	for i:=0;i<100000;i++{
		showStr := fmt.Sprintf("%d\n",i)
		timerWheel.AddTimer(NewTimer(1,0,func(args interface{}){
			fmt.Print(args.(string))
		},showStr))
	}
	for{
		timeCnt ++
		if 100 == timeCnt{
			timerWheel.Stop()
			return
		}
		time.Sleep(1*time.Second)
	}
}
