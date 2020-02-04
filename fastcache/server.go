package main

import(
	"io"
	"fmt"
	"errors"
	"net"
	"bytes"
	"github.com/chelion/gws/log"
	"github.com/chelion/gws/utils"
	"github.com/chelion/gws/configure"
	"github.com/VictoriaMetrics/fastcache"
)

const(
	BYTEBUFF_SIZE = 4096
	CACHEDEFAULTSIZE = 1024 * 1024 * 256
	CACHEITEMMAXSIZE = 64 * 1024 - 16 - 4 - 1
)

var(
	logger log.Log
	CACHECLIENT_NIL  = errors.New("client is nil")
	CACHEPARAM_ERROR = errors.New("param is error")
	CACHEMAXSIZE_OVER = errors.New("cache over max size")
)

type FastCache struct{
	cache *fastcache.Cache
	cachesize int64
}

func NewFastCache(cachesize int64)(fcc *FastCache,err error){
	fcc = &FastCache{cachesize:cachesize}
	fcc.cache = fastcache.New(int(cachesize))
	return fcc,nil
}
/*
func (fcc *FastCache)Get(key []byte,replay *[]byte)([]byte){
	return fcc.cache.Get(nil,key)
}

func (fcc *FastCache)Set(key []byte,value []byte)(err error){
	fcc.cache.Set(args.Key,args.Data)
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
*/

/*
CMD
PING

*PNG\r\n


SET
*SET\r\n
$keylen valuelen\r\n 1+4+4+2 = 11字节
key \r\n
value \r\n

*SET
$b0011 b0010
god
good


GET
*GET\r\n
$keylen spacelen\r\n
key \r\n

*GET
$b0011 b0000
god


DEL
*DEL\r\n
$keylen spacelen\r\n
key \r\n

*DEL
$b0011 b0000
god


EXISTS
*EXS\r\n
$keylen spacelen\r\n
key \r\n

*EXS
$b0011 b0000
god

go tool pprof http://localhost:8899/debug/pprof/profile?seconds=60

*/

const(
	STATUS_CMD = 0x01
	STATUS_KEYVALUELEN = 0x02
	STATUS_KEY = 0x03
	STATUS_VALUE = 0x04
)

func serverConn(fcc *FastCache,conn net.Conn){
	//不用bufio提高速度
	var key []byte
	var value []byte
	var keyvalueLenBytes = make([]byte,11)
	var cacheBytes = make([]byte,4096)
	var byteReadBuff bytes.Buffer
	var byteWriteBuff bytes.Buffer
	status := STATUS_CMD
	cmdType := byte(' ')
	keyLen := 0
	valueLen := 0
	defer conn.Close()
	for{
		switch status{
			case STATUS_CMD:{
				cmdType = byte(' ')
				keyLen = 0
				valueLen = 0
				cacheBytes = cacheBytes[0:]
				rlen,err := conn.Read(cacheBytes)
				if nil != err{
					logger.Println(err)
					return
				}
				byteReadBuff.Write(cacheBytes[0:rlen])
				line,err := byteReadBuff.ReadBytes('\n')
				if nil != err{
					if err == io.EOF{
						logger.Println(err)
						return
					}
				}
				if 6 == len(line) && '*' == line[0]{
					switch line[1]{
						case 'S':
							fallthrough
						case 'G':
							fallthrough
						case 'D':
							fallthrough
						case 'E':{
							status = STATUS_KEYVALUELEN
						}
						case 'P':{
							_,err = conn.Write([]byte("+PNG\r\n"))
							if nil != err{
								logger.Println(err)
								return
							}
							status = STATUS_CMD
						}
						default:{
							conn.Write([]byte("-Error\r\n"))
							return
						}
					}
					cmdType = line[1]
				}else{
					conn.Write([]byte("-Error\r\n"))
					return
				}
			}
			case STATUS_KEYVALUELEN:{
				if byteReadBuff.Len() < 11{
					cacheBytes = cacheBytes[0:]
					rlen,err := conn.Read(cacheBytes)
					if nil != err{
						logger.Println(err)
						return
					}
					byteReadBuff.Write(cacheBytes[0:rlen])
				}
				byteReadBuff.Read(keyvalueLenBytes[0:])
				if '$' == keyvalueLenBytes[0]{
					keyLen = int(utils.BytesToInt32(keyvalueLenBytes[1:5]))
					valueLen = int(utils.BytesToInt32(keyvalueLenBytes[5:9]))
					if keyLen > CACHEITEMMAXSIZE || valueLen > CACHEITEMMAXSIZE ||
					(keyLen+valueLen) > CACHEITEMMAXSIZE{
						conn.Write([]byte("-Error\r\n"))
						return
					}
					status = STATUS_KEY
				}else{
					conn.Write([]byte("-Error\r\n"))
					return
				}
			}
			case STATUS_KEY:{
				for{
					if byteReadBuff.Len() < keyLen+2{
						cacheBytes = cacheBytes[0:]
						rlen,err := conn.Read(cacheBytes)
						if nil != err{
							logger.Println(err)
							return
						}
						byteReadBuff.Write(cacheBytes[0:rlen])
					}else{
						break
					}
				}
				key = make([]byte,keyLen+2)
				byteReadBuff.Read(key[0:])
				if !bytes.HasSuffix(key, []byte("\r\n")) {
					conn.Write([]byte("-Error\r\n"))
					return
				}
				if cmdType == 'S'{
					status = STATUS_VALUE
				}else{
					byteWriteBuff.Reset()
					switch cmdType{
						case 'G':{
							value = fcc.cache.Get(nil,key[0:keyLen])
							byteWriteBuff.Write([]byte("+"))
							if nil != value{
								byteWriteBuff.Write(utils.Int32ToBytes(int32(len(value))))
								byteWriteBuff.Write([]byte("\r\n"))
								byteWriteBuff.Write(value)
							}else{
								byteWriteBuff.Write(utils.Int32ToBytes(0))
							}
							byteWriteBuff.Write([]byte("\r\n"))
							_,err := conn.Write(byteWriteBuff.Bytes())
							if nil != err{
								logger.Println(err)
								return
							}
						}
						case 'E':{
							if fcc.cache.Has(key[0:keyLen]){
								byteWriteBuff.Write([]byte(":1\r\n"))
							}else{
								byteWriteBuff.Write([]byte(":0\r\n"))
							}
							_,err := conn.Write(byteWriteBuff.Bytes())
							if nil != err{
								logger.Println(err)
								return
							}
						}
						case 'D':{
							if fcc.cache.Has(key[0:keyLen]){
								fcc.cache.Del(key[0:keyLen])
								byteWriteBuff.Write([]byte(":1\r\n"))
							}else{
								byteWriteBuff.Write([]byte(":0\r\n"))
							}
							_,err := conn.Write(byteWriteBuff.Bytes())
							if nil != err{
								logger.Println(err)
								return
							}
						}
					}
					status = STATUS_CMD
				}
			}
			case STATUS_VALUE:{
				for{
					if byteReadBuff.Len() < valueLen+2{
						cacheBytes = cacheBytes[0:]
						rlen,err := conn.Read(cacheBytes)
						if nil != err{
							logger.Println(err)
							return
						}
						byteReadBuff.Write(cacheBytes[0:rlen])
					}else{
						break
					}
				}
				value = make([]byte,valueLen+2)
				byteReadBuff.Read(value[0:])
				if !bytes.HasSuffix(value, []byte("\r\n")){
					conn.Write([]byte("-Error\r\n"))
					return
				}
				fcc.cache.Set(key[0:keyLen],value[0:valueLen])
				_,err := conn.Write([]byte("+OK\r\n"))
				if nil != err{
					logger.Println(err)
					return
				}
				status = STATUS_CMD
			}
		}
	}
}

func main(){
	
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
        go serverConn(fcc,conn)
    }
}