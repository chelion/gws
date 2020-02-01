package log
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import(
	"fmt"
	"log"
	"sync"
	"time"
	"net/http"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)


type WebSocketLog struct{
	LogHandler
	websocket *WebSocket
}

func NewWebSocketLog(addr string,date bool)(websocketlog *WebSocketLog,err error){
	websocketlog = &WebSocketLog{}
	websocket := &WebSocket{addr:addr,clients:make(map[string]*WebSocketClient),close:make(chan *WebSocketClient),hmutex:new(sync.RWMutex),exit:make(chan bool,1)}
	log := log.New(websocket, "",0)
	websocketlog.LogHandler = LogHandler{log:log,level:DEBUG,date:date}
	websocketlog.websocket = websocket
	return websocketlog,nil
}

func (websocketlog *WebSocketLog)Init()(err error){
	if nil == websocketlog.log{
		return ILOGCLIENT_NIL
	}
	go websocketlog.websocket.Start()
	return nil
}

func (websocketlog *WebSocketLog)DeInit()(err error){
	if nil != websocketlog.log{
		if nil != websocketlog.websocket{
			websocketlog.websocket.Stop()
			websocketlog.websocket = nil
		}
		websocketlog.log = nil
		return nil
	}
	return ILOGCLIENT_NIL
}


func (websocketlog *WebSocketLog)Debug(v ...interface{}){
	if websocketlog.level <= DEBUG && nil != websocketlog.log{
		websocketlog.log.Output(2,websocketlog.getDate())
		websocketlog.log.Output(2, fmt.Sprintln("<font color='lime'>Debug-></font>", v))
	}
}

func (websocketlog *WebSocketLog)Info(v ...interface{}){
	if websocketlog.level <= INFO && nil != websocketlog.log{
		websocketlog.log.Output(2,websocketlog.getDate())
		websocketlog.log.Output(2, fmt.Sprintln("<font color='aqua'>Info-></font>", v))
	}
}

func (websocketlog *WebSocketLog)Warn(v ...interface{}){
	if websocketlog.level <= WARN && nil != websocketlog.log{
		websocketlog.log.Output(2,websocketlog.getDate())
		websocketlog.log.Output(2, fmt.Sprintln("<font color='yellow'>Warn-></font>", v))
	}
}

func (websocketlog *WebSocketLog)Error(v ...interface{}){
	if websocketlog.level <= ERROR && nil != websocketlog.log{
		websocketlog.log.Output(2,websocketlog.getDate())
		websocketlog.log.Output(2, fmt.Sprintln("<font color='red'>Error-></font>", v))
	}
}


const (
	// Time allowed to write a message to the peer.
	WriteWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	PongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	PingPeriod = (PongWait * 9) / 10

	// Maximum message size allowed from peer.
	MaxMessageSize = 4096
)

var (
	WEBSOCKETCLIENT_NIL = errors.New("web socket client nil") 
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

type WebSocketClient struct {
	id string
	wscon *websocket.Conn
	send 	chan []byte
	ticker	chan bool
	timeout int64
	once *sync.Once
}

type WebSocket struct {
	addr string
	clients map[string]*WebSocketClient
	close  chan *WebSocketClient
	hmutex  *sync.RWMutex
	exit chan bool
}

func (websocket *WebSocket)ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		conn, err := upgrader.Upgrade(w, r, nil)
		if nil != err {
			return
		}
		uuid, err := uuid.NewUUID()
		if err != nil {
			return
		}
		uuidstr := uuid.String()
		wscc := &WebSocketClient{once:new(sync.Once),id:uuidstr,wscon: conn, send:make(chan []byte,8),ticker:make(chan bool,1),timeout:0}
		websocket.hmutex.Lock()
		websocket.clients[uuidstr] = wscc
		websocket.hmutex.Unlock()
		go wscc.readPump(websocket)
		go wscc.writePump(websocket)
	}
}

func (websocket * WebSocket)Start() (err error){
	go func(){
		var ticktimeout int64
		ticker := time.NewTicker(PingPeriod)
		ticktimeout = 30
		defer ticker.Stop()
		for{
			select {
				case <-ticker.C:{
					nowtime := time.Now().Unix()
					if  nil == websocket.clients{
						return
					}
					websocket.hmutex.RLock()
					for _,wscc := range websocket.clients{
						if nowtime - wscc.timeout >= ticktimeout{
							wscc.ticker <- true
						}
					}
					websocket.hmutex.RUnlock()
				}
				case <-websocket.exit:{
					return
				}
			}
		}
	}()
	err = http.ListenAndServe(websocket.addr,websocket)
	if err != nil {
		return
	}
	return nil
}

func (websocket * WebSocket)Stop(){
	websocket.hmutex.Lock()
	for _,wscc := range websocket.clients{
		wscc.close(websocket)
	}
	websocket.hmutex.Unlock()
	websocket.clients = nil
	websocket.exit <- true
}

func (websocket * WebSocket)Write(p []byte) (n int, err error){
	if nil != websocket.clients{
		var data []byte = make([]byte,len(p))
		copy(data,p)
		websocket.hmutex.RLock()
		for _,wscc := range websocket.clients{
			wscc.send<-data
		}
		websocket.hmutex.RUnlock()
		return len(data),nil
 	}
	return 0,WEBSOCKETCLIENT_NIL
}


func (wscc *WebSocketClient) close(ws *WebSocket) {
	wscc.once.Do(
		func(){
			wscc.wscon.Close()
			ws.hmutex.Lock()
			delete(ws.clients,wscc.id)
			ws.hmutex.Unlock()
		})
}

func (wscc *WebSocketClient) writePump(ws *WebSocket) {
	defer wscc.close(ws)
	for {
		select {
			case message,ok := <-wscc.send:{
				wscc.wscon.SetWriteDeadline(time.Now().Add(WriteWait))
				if !ok {
					wscc.wscon.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				err := wscc.wscon.WriteMessage(websocket.TextMessage,message)
				if nil != err{
					return
				}
			}
			case <-wscc.ticker:
				wscc.wscon.SetWriteDeadline(time.Now().Add(WriteWait))
				if err := wscc.wscon.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
				wscc.timeout = time.Now().Unix()
		}
	}
}


func (wscc *WebSocketClient) readPump(ws *WebSocket) {
	defer wscc.close(ws)
	wscc.wscon.SetReadLimit(MaxMessageSize)
	wscc.wscon.SetReadDeadline(time.Now().Add(PongWait))
	wscc.wscon.SetPongHandler(func(string) error { wscc.wscon.SetReadDeadline(time.Now().Add(PongWait)); return nil })
	for {
		_, message, err := wscc.wscon.ReadMessage()
		if nil != err{
			return
		}else{
			wscc.send <- message
		}
	}
}
