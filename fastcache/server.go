package main

import(
	"fmt"
	"errors"
	"net"
	"net/rpc"
	"github.com/chelion/gws/log"
	"github.com/chelion/gws/configure"
	"github.com/VictoriaMetrics/fastcache"
)

const(
	CACHEDEFAULTSIZE = 1024 * 1024 * 256
	CACHEITEMMAXSIZE = 64 * 1024 - 16 - 4 - 1
)

var(
	CACHECLIENT_NIL  = errors.New("client is nil")
	CACHEPARAM_ERROR = errors.New("param is error")
	CACHEMAXSIZE_OVER = errors.New("cache over max size")
)

type FastCache struct{
	cache *fastcache.Cache
	cachesize int64
}

type Args struct{
	Key []byte
	Data []byte
	Expire int
}

func NewFastCache(cachesize int64)(fcc *FastCache,err error){
	fcc = &FastCache{cachesize:cachesize}
	fcc.cache = fastcache.New(int(cachesize))
	return fcc,nil
}

func (fcc *FastCache)Get(args *Args,replay *[]byte)(err error){
	value := fcc.cache.Get(nil,args.Key)
	*replay = value
	return nil
}

func (fcc *FastCache)Set(args *Args,sta *bool)(err error){
	if nil == args.Data{
		*sta = false
		return CACHEPARAM_ERROR
	}
	if len(args.Data) > CACHEITEMMAXSIZE{
		*sta = false
		return CACHEMAXSIZE_OVER
	}
	fcc.cache.Set(args.Key,args.Data)
	*sta = true
	return nil
}

func (fcc *FastCache)Delete(args *Args,sta *bool)(err error){
	fcc.cache.Del(args.Key)
	*sta = true
	return nil
}

func (fcc *FastCache)Exists(args *Args,sta *bool)(err error){
	*sta = fcc.cache.Has(args.Key)
	return nil
}

func (fcc *FastCache)Ping(args *Args,sta *bool)(err error){
	*sta = true
	return nil
}

func main(){
	var logger log.Log
	config,err := configure.NewIniConfigure("./server.ini")
	if nil != err{
		fmt.Println(err)
		return
	}
	if nil == config.Init(){
		defer config.DeInit()
	}else{
		fmt.Println(err)
		return
	}
	config.SetSection("Log")
	logType,err := config.GetString("logType")
	switch(logType){
		case "file":{
			logger,err = log.NewFileLog("http",true,true)
		}
		case "socket":{
			serverAddr,err := config.GetString("serverAddr")
			if nil != err{
				fmt.Println(err)
				panic("use socket log need configure socket server addr\n")
			}
			logger,err = log.NewSocketLog(serverAddr,true)
		}
		case "websocket":{
			serverAddr,err := config.GetString("serverAddr")
			if nil != err{
				fmt.Println(err)
				panic("use websocket log need configure websocket server addr\n")
			}
			logger,err = log.NewWebSocketLog(serverAddr,true)
		}
		default:{
			logger,err = log.NewConsoleLog(true)
		}
	}
	if nil != err{
		fmt.Println(err)
	}
	logLevel,err := config.GetInt("level",0)
	if nil != err{
		fmt.Println(err)
	}
	logger.SetLevel(log.LevelEnum(logLevel))
	config.SetSection("FastCacheServer")
	addr,err := config.GetString("addr")
	if nil != err{
		fmt.Println(err)
		return
	}
	netWork,err := config.GetString("netWork")
	if nil != err{
		fmt.Println(err)
		return
	}
	cacheSize,err := config.GetInt64("cacheSize",CACHEDEFAULTSIZE)
	if nil != err{
		fmt.Println(err)
		return
	}
	if cacheSize <=0 {
		cacheSize = CACHEDEFAULTSIZE
	}
	fcc,err := NewFastCache(cacheSize)
	if nil != err{
		logger.Fatalln(err)
		return
	}
	rpc.Register(fcc)
	tcpaddr, err := net.ResolveTCPAddr(netWork,addr)
    if err != nil {
		logger.Fatalln(err)
		return
    }
    listen, err := net.ListenTCP(netWork,tcpaddr)
    if err != nil {
		logger.Fatalln(err)
		return
	}
	logger.Println("fast cache addr",addr)
	logger.Println("fast cache size",int(cacheSize/1024/1024),"MB")
	logger.Println("start server ok,working......!")
    for {
        conn, err := listen.Accept()
        if err != nil {
            continue
        }
        go func(conn net.Conn) {
            rpc.ServeConn(conn)
        }(conn)
    }
}