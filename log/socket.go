package log
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import(
	"log"
	"net"
	"time"
	"sync"
	"errors"
)

var(
	CONISNIL_ERR = errors.New("connect is nil")
)

type SocketLogConn struct{
	serverAddr string
	conn net.Conn
	lock *sync.RWMutex
}

type SocketLog struct{
	LogHandler
	conn *SocketLogConn
	exit chan bool
}

func (socketLogConn *SocketLogConn)Read(b []byte) (n int, err error){
	socketLogConn.lock.RLock()
	if nil != socketLogConn.conn{
		socketLogConn.lock.RUnlock()
		return socketLogConn.conn.Read(b)
	}
	socketLogConn.lock.RUnlock()
	return 0,CONISNIL_ERR
}

func (socketLogConn *SocketLogConn)Write(b []byte) (n int, err error){
	socketLogConn.lock.RLock()
	if nil != socketLogConn.conn{
		socketLogConn.lock.RUnlock()
		n,err = socketLogConn.conn.Write(b)
		if nil != err{
			if CONISNIL_ERR != socketLogConn.Close(){
				go socketLogConn.reconnect()
			}
		}
		return n,err
	}
	socketLogConn.lock.RUnlock()
	return 0,CONISNIL_ERR
}

func (socketLogConn *SocketLogConn)Close() error{
	socketLogConn.lock.RLock()
	if nil != socketLogConn.conn{
		socketLogConn.lock.RUnlock()
		err := socketLogConn.conn.Close()
		socketLogConn.lock.Lock()
		socketLogConn.conn = nil
		socketLogConn.lock.Unlock()
		return err
	}
	socketLogConn.lock.RUnlock()
	return CONISNIL_ERR
}

func (socketLogConn *SocketLogConn)LocalAddr() net.Addr{
	socketLogConn.lock.RLock()
	if nil != socketLogConn.conn{
		socketLogConn.lock.RUnlock()
		return socketLogConn.conn.LocalAddr()
	}
	socketLogConn.lock.RUnlock()
	return nil
}

func (socketLogConn *SocketLogConn)RemoteAddr() net.Addr{
	socketLogConn.lock.RLock()
	if nil != socketLogConn.conn{
		socketLogConn.lock.RUnlock()
		return socketLogConn.conn.RemoteAddr()
	}
	socketLogConn.lock.RUnlock()
	return nil
}

func (socketLogConn *SocketLogConn)SetDeadline(t time.Time) error{
	socketLogConn.lock.RLock()
	if nil != socketLogConn.conn{
		socketLogConn.lock.RUnlock()
		return socketLogConn.conn.SetDeadline(t)
	}
	socketLogConn.lock.RUnlock()
	return CONISNIL_ERR
}

func (socketLogConn *SocketLogConn)SetReadDeadline(t time.Time) error{
	socketLogConn.lock.RLock()
	if nil != socketLogConn.conn{
		socketLogConn.lock.RUnlock()
		return socketLogConn.conn.SetReadDeadline(t)
	}
	socketLogConn.lock.RUnlock()
	return CONISNIL_ERR
}

func (socketLogConn *SocketLogConn)SetWriteDeadline(t time.Time) error{
	socketLogConn.lock.RLock()
	if nil != socketLogConn.conn{
		socketLogConn.lock.RUnlock()
		return socketLogConn.conn.SetWriteDeadline(t)
	}
	socketLogConn.lock.RUnlock()
	return CONISNIL_ERR
}

func (socketLogConn *SocketLogConn)reconnect(){
	for{
		tcpAddr, err := net.ResolveTCPAddr("tcp",socketLogConn.serverAddr)  //TCP连接地址
		if err != nil{
			return
		}
		conn, err := net.DialTCP("tcp", nil, tcpAddr) 
		if nil != err{
			time.Sleep(time.Duration(2)*time.Second)//2秒钟继续连接服务器
			continue
		}
		socketLogConn.lock.Lock()
		socketLogConn.conn = conn
		socketLogConn.lock.Unlock()
		return
	}
}

func (socketLog *SocketLog)socketTick(){
	ticker := time.NewTicker(time.Duration(30)*time.Second)
	defer ticker.Stop()
	for{
		select{
			case <-ticker.C:{
				if nil != socketLog.conn{
					_,err := socketLog.conn.Write([]byte("\r\nTick\r\n"))
					if nil != err && err != CONISNIL_ERR{
						if CONISNIL_ERR != socketLog.conn.Close(){
							go socketLog.conn.reconnect()
						}
					}
				}
			}
			case <-socketLog.exit:{
				socketLog.conn.Close()
				return
			}
		}
	}
}

func NewSocketLog(serverAddr string,date bool)(socketLog *SocketLog,err error){
	socketLog = &SocketLog{conn:&SocketLogConn{serverAddr:serverAddr,conn:nil,lock:new(sync.RWMutex)},exit:make(chan bool,1)}
	log := log.New(socketLog.conn, "",0)
	socketLog.LogHandler = LogHandler{log:log,level:DEBUG,date:date}
	go socketLog.conn.reconnect()
	go socketLog.socketTick()
	return socketLog,nil
}

func (socketLog *SocketLog)Init()(err error){
	if nil == socketLog.log{
		return ILOGCLIENT_NIL
	}
	return nil
}

func (socketLog *SocketLog)DeInit()(err error){
	if nil != socketLog.log{
		socketLog.exit <- true
	}
	if nil != socketLog.log{
		socketLog.log = nil
		return nil
	}
	return ILOGCLIENT_NIL
}