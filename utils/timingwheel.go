package utils
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import(
	"sync"
	"time"
	"sync/atomic"
	"github.com/google/uuid"
	"github.com/cespare/xxhash"
)

const(
	WHEEL_NUM = 4//4个轮，每个轮256个槽,平行轮，都是1Tick为刻度，1024 Tick一个轮回
	SLOT_NUM = 256
	SLOT_2POW = 8
	WHEEL_COROUTINE_MAXNUM = WHEEL_NUM*SLOT_NUM
)

type Timer struct{
	interval 	uint32
	tick		uint32
	loop 		uint32
	callBack func(interface{})
	args interface{}
	tuuid uint64
	wheelID uint8
	slotID uint8
}

type TimingWheel struct{
	stop chan struct{}
	isStart bool
	lock *sync.RWMutex
	wheelSlotLock [WHEEL_NUM][SLOT_NUM]*sync.RWMutex//槽锁
	currentTickCnt uint32
	workerPool *WorkerPool
	slotTimerMap [WHEEL_NUM][SLOT_NUM]map[uint64]*Timer
}

type UserCallBack struct{
	callBack func(interface{})
	args interface{}
}

func (userCallBack *UserCallBack)Run(){
	if nil != userCallBack.callBack{
		userCallBack.callBack(userCallBack.args)
	}
}

func NewTimer(interval uint32,loop uint32,callBack func(interface{}),args interface{})(timer *Timer){
	uuidv, err := uuid.NewUUID()
	if nil != err{
		panic(err)
	}
	return &Timer{
		interval:interval,
		tick:0,
		loop:loop,
		callBack:callBack,
		args:args,
		tuuid:xxhash.Sum64(uuidv[0:]),
	}
}

func NewTimingWheel(baseTick uint32)(timingWheel * TimingWheel){
	timingWheel =  &TimingWheel{
		lock:new(sync.RWMutex),
		isStart:false,
		stop:make(chan struct{}),
	}
	timingWheel.workerPool,_ = NewWorkerPool(WHEEL_COROUTINE_MAXNUM)
	atomic.StoreUint32(&timingWheel.currentTickCnt,baseTick)
	for i:=0;i<WHEEL_NUM;i++{
		for j:=0;j<SLOT_NUM;j++{
			timingWheel.wheelSlotLock[i][j] =  new(sync.RWMutex)
		}
	}
	return
}

func (timingWheel *TimingWheel)getTickWheelSlotId(tick uint32)(slotID,wheelID uint8){
	wheelID = 0
	tmp := tick
	for i:=0;i<WHEEL_NUM;i++{
		tmp >>= SLOT_2POW
		if 0 == tmp{
			break
		}else{
			wheelID ++
		}
	}
	slotID = uint8(uint64(tick)%SLOT_NUM)
	return
}

func (timingWheel *TimingWheel)Start(){
	if false == timingWheel.isStart{
		go func(timingWheel *TimingWheel){
			timingWheel.isStart = true
			var slotID uint8 = 0 
			var wheelID uint8 = 0 
			ticker := time.NewTicker(1*time.Second)
			defer ticker.Stop()
			for{
				select{
					case <-timingWheel.stop:{
						timingWheel.lock.Lock()
						if true == timingWheel.isStart{
							timingWheel.isStart = false
							for i:=0;i<WHEEL_NUM;i++{
								timingWheel.wheelSlotLock[wheelID][slotID].Lock()
								for j:=0;j<SLOT_NUM;j++{
									timingWheel.slotTimerMap[i][j] = nil
								}
								timingWheel.wheelSlotLock[wheelID][slotID].Unlock()
							}
							timingWheel.workerPool.Destroy()
						}
						timingWheel.lock.Unlock()
						return
					}
					case <-ticker.C:{
						atomic.AddUint32(&timingWheel.currentTickCnt, 1)
						currentTickCnt := atomic.LoadUint32(&timingWheel.currentTickCnt)
						slotID,wheelID = timingWheel.getTickWheelSlotId(currentTickCnt)
						timingWheel.wheelSlotLock[wheelID][slotID].RLock()
						timerMap := timingWheel.slotTimerMap[wheelID][slotID]
						timingWheel.wheelSlotLock[wheelID][slotID].RUnlock()
						if nil != timerMap{
							for _,t := range timerMap{
								if t.tick == currentTickCnt{
									timingWheel.workerPool.Work(Job(&UserCallBack{callBack:t.callBack,args:t.args}))
									timingWheel.wheelSlotLock[wheelID][slotID].Lock()
									delete(timingWheel.slotTimerMap[wheelID][slotID],t.tuuid)
									timingWheel.wheelSlotLock[wheelID][slotID].Unlock()
									if t.loop == 0  || t.loop > 1{
										if t.loop != 0{
											t.loop = t.loop - 1
										}
										t.tick = t.interval + currentTickCnt
										slotIDNew,wheelIDNew := timingWheel.getTickWheelSlotId(t.tick)
										t.slotID = slotIDNew
										t.wheelID = wheelIDNew
										timingWheel.wheelSlotLock[wheelIDNew][slotIDNew].Lock()
										if nil == timingWheel.slotTimerMap[wheelIDNew][slotIDNew]{
											timingWheel.slotTimerMap[wheelIDNew][slotIDNew] = make(map[uint64]*Timer)
										}
										timingWheel.slotTimerMap[wheelIDNew][slotIDNew][t.tuuid] = t
										timingWheel.wheelSlotLock[wheelIDNew][slotIDNew].Unlock()
									}
								}
							}
							if 0 == len(timingWheel.slotTimerMap[wheelID][slotID]){
								timingWheel.wheelSlotLock[wheelID][slotID].Lock()
								timingWheel.slotTimerMap[wheelID][slotID] = nil
								timingWheel.wheelSlotLock[wheelID][slotID].Unlock()
							}
						}
					}
				}
			}
		}(timingWheel)
	}
}

func (timingWheel *TimingWheel)Stop(){
	timingWheel.lock.RLock()
	if true == timingWheel.isStart{
		timingWheel.lock.RUnlock()
		timingWheel.stop <- struct{}{}
		return
	}
	timingWheel.lock.RUnlock()
}

func (timingWheel *TimingWheel)AddTimer(timer *Timer){
	currentTickCnt := atomic.LoadUint32(&timingWheel.currentTickCnt)
	timer.tick = timer.interval+currentTickCnt
	slotID,wheelID := timingWheel.getTickWheelSlotId(timer.tick)
	timer.slotID = slotID
	timer.wheelID = wheelID
	timingWheel.wheelSlotLock[wheelID][slotID].Lock()
	if nil == timingWheel.slotTimerMap[wheelID][slotID]{
		timingWheel.slotTimerMap[wheelID][slotID] = make(map[uint64]*Timer)
	}
	timingWheel.slotTimerMap[wheelID][slotID][timer.tuuid] = timer
	timingWheel.wheelSlotLock[wheelID][slotID].Unlock()
}

func (timingWheel * TimingWheel)RemoveTimer(timer * Timer){
	timingWheel.wheelSlotLock[timer.wheelID][timer.slotID].Lock()
	if nil != timingWheel.slotTimerMap[timer.wheelID][timer.slotID]{
		delete(timingWheel.slotTimerMap[timer.wheelID][timer.slotID],timer.tuuid)
	}
	if 0 == len(timingWheel.slotTimerMap[timer.wheelID][timer.slotID]){
		timingWheel.slotTimerMap[timer.wheelID][timer.slotID] = nil
	}
	timingWheel.wheelSlotLock[timer.wheelID][timer.slotID].Unlock()
}