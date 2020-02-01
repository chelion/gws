package utils
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import(
	"sync"
	"errors"
	"sync/atomic"
)

var(
	WORKERPOOL_PARAM_ERR = errors.New("worker pool param is error")
	WORKERPOOL_DESTORY_ERR = errors.New("worker pool is destroy")
)

type Job interface{
	Run()
}

type Worker struct{
	jobChan chan Job
	workerPool *WorkerPool
}

type WorkerPool struct{
	idleNum int64
	maxNum int64
	curNum int64
	lock 	*sync.RWMutex
	workerChans chan *Worker
	isDestroy bool
}

func NewWorkerPool(maxNum int64)(*WorkerPool,error){
	if maxNum <= 0{
		return nil,WORKERPOOL_PARAM_ERR
	}
	workerPool := &WorkerPool{maxNum:maxNum,curNum:0,idleNum:0,
		isDestroy:false,lock:new(sync.RWMutex),workerChans:make(chan *Worker,maxNum)}
	newWorker := &Worker{jobChan:make(chan Job),workerPool:workerPool}
	go newWorker.do()
	atomic.StoreInt64(&workerPool.curNum,1)
	atomic.StoreInt64(&workerPool.idleNum,1)
	workerPool.workerChans <- newWorker
	return workerPool,nil
}

func (worker *Worker)do(){
	for{
		job,ok :=<- worker.jobChan
		if !ok {
			return
		}
		if nil != job{
			job.Run()
			if atomic.LoadInt64(&worker.workerPool.curNum) <= worker.workerPool.maxNum{
				worker.workerPool.lock.Lock()
				if worker.workerPool.isDestroy{
					worker.workerPool.lock.Unlock()
					close(worker.jobChan)
					return
				}
				worker.workerPool.workerChans <- worker
				atomic.AddInt64(&worker.workerPool.idleNum,1)
				worker.workerPool.lock.Unlock()
			}
		}else{
			close(worker.jobChan)
			return
		}
	}

}

func (workerPool *WorkerPool)Destroy()(err error){
	workerPool.lock.Lock()
	if workerPool.isDestroy{
		workerPool.lock.Unlock()
		return WORKERPOOL_DESTORY_ERR
	}
	for{
		select{
			case worker := <- workerPool.workerChans:{
				worker.jobChan <- nil
			}
			default:{
				close(workerPool.workerChans)
				workerPool.isDestroy = true
				workerPool.lock.Unlock()
				return nil
			}
		}
	}
}

func (workerPool *WorkerPool)Work(job Job)(err error){
	worker,ok := <- workerPool.workerChans
	if !ok{
		return WORKERPOOL_DESTORY_ERR
	}
	if 0 >= atomic.AddInt64(&workerPool.idleNum,-1){
		if atomic.LoadInt64(&workerPool.curNum) < workerPool.maxNum{
			atomic.AddInt64(&workerPool.curNum,1)
			workerPool.lock.RLock()
			if workerPool.isDestroy{
				workerPool.lock.RUnlock()
				return WORKERPOOL_DESTORY_ERR
			}
			newWorker := &Worker{jobChan:make(chan Job),workerPool:workerPool}
			go newWorker.do()
			workerPool.workerChans <- newWorker
			atomic.AddInt64(&workerPool.idleNum,1)
			workerPool.lock.RUnlock()
		}
	}
	worker.jobChan <- job
	return nil
}
