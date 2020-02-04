package main

import(
	"io"
	"fmt"
	"io/ioutil"
	"errors"
	"net"
	"bytes"
	"bufio"
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
	//byteBufferPool *utils.ByteBufferPool = utils.NewByteBufferPool(BYTEBUFF_SIZE,1024)
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
	var err error
	var key []byte
	var value []byte
	var byteBuff bytes.Buffer
	bufioReader := bufio.NewReaderSize(conn,4096)
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
				line,err := bufioReader.ReadBytes('\n')
				if nil != err{
					fmt.Println("cmd",err)
					if err == io.EOF || bytes.Contains(utils.String2Bytes(err.Error()),[]byte("forcibly closed")){
						fmt.Println("close")
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
				line,err := ioutil.ReadAll(io.LimitReader(bufioReader,11))
				if nil != err{
					fmt.Println("keyvaluelen",err)
					if err == io.EOF || bytes.Contains(utils.String2Bytes(err.Error()),[]byte("forcibly closed")){
						fmt.Println("close")
						return
					}
				}
				if 11 == len(line) && '$' == line[0]{
					keyLen = int(utils.BytesToInt32(line[1:5]))
					valueLen = int(utils.BytesToInt32(line[5:9]))
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
				key, err = ioutil.ReadAll(io.LimitReader(bufioReader,int64(keyLen+2)))
				if nil != err || !bytes.HasSuffix(key, []byte("\r\n")) {
					fmt.Println("key",err)
					conn.Write([]byte("-Error\r\n"))
					if err == io.EOF || bytes.Contains(utils.String2Bytes(err.Error()),[]byte("forcibly closed")){
						fmt.Println("close")
						return
					}
				}
				if cmdType == 'S'{
					status = STATUS_VALUE
				}else{
					byteBuff.Reset()
					switch cmdType{
						case 'G':{
							value = fcc.cache.Get(nil,key[0:keyLen])
							byteBuff.Write([]byte("+"))
							if nil != value{
								byteBuff.Write(utils.Int32ToBytes(int32(len(value))))
								byteBuff.Write([]byte("\r\n"))
								byteBuff.Write(value)
							}else{
								byteBuff.Write(utils.Int32ToBytes(0))
							}
							byteBuff.Write([]byte("\r\n"))
							_,err = conn.Write(byteBuff.Bytes())
							if nil != err{
								return
							}
						}
						case 'E':{
							if fcc.cache.Has(key[0:keyLen]){
								byteBuff.Write([]byte(":1\r\n"))
							}else{
								byteBuff.Write([]byte(":0\r\n"))
							}
							_,err = conn.Write(byteBuff.Bytes())
							if nil != err{
								return
							}
						}
						case 'D':{
							if fcc.cache.Has(key[0:keyLen]){
								fcc.cache.Del(key[0:keyLen])
								byteBuff.Write([]byte(":1\r\n"))
							}else{
								byteBuff.Write([]byte(":0\r\n"))
							}
							_,err = conn.Write(byteBuff.Bytes())
							if nil != err{
								return
							}
						}
					}
					status = STATUS_CMD
				}
			}
			case STATUS_VALUE:{
				value, err = ioutil.ReadAll(io.LimitReader(bufioReader,int64(valueLen+2)))
				if nil != err || !bytes.HasSuffix(value, []byte("\r\n")){
					fmt.Println("value",err)
					conn.Write([]byte("-Error\r\n"))
					if err == io.EOF || bytes.Contains(utils.String2Bytes(err.Error()),[]byte("forcibly closed")){
						fmt.Println("close")
						return
					}
				}
				fcc.cache.Set(key[0:keyLen],value[0:valueLen])
				_,err = conn.Write([]byte("+OK\r\n"))
				if nil != err{
					return
				}
				status = STATUS_CMD
			}
		}
	}
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