package utils
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

import(
	"fmt"
	"time"
	"testing"
)

type TestJob struct{
 	i int
}
func (job *TestJob)Run(){
	time.Sleep(10*time.Millisecond)
	fmt.Println("do job",job.i)
}

func BenchmarkWorkerPool(b *testing.B){
	workerPool,_ := NewWorkerPool(1000)
	fmt.Println("----start test worker pool----")
	s := time.Now()
	for i:=0;i<200000;i++{
		if nil != workerPool.Work(Job(&TestJob{i:i})){
			fmt.Println("worker destroy")
		}
	}
	b.N = 200000
	workerPool.Destroy()
	elapsed := time.Since(s)
    fmt.Println("start test worker pool elapsed:", elapsed)
	fmt.Println("----end test worker pool----")
}
