package cache

import(
	"github.com/chelion/gws/utils"
	"github.com/VictoriaMetrics/fastcache"
)

const(
	CACHEITEMMAXSIZE = 64 * 1024 - 16 - 4 - 1
)

type LocalCacheConfig struct{
	Cachesize int64
} 

type LocalCache struct{
	cacheSize int64
	client *fastcache.Cache
}

func NewLocalCache(config *LocalCacheConfig)(lcc *LocalCache,err error){
	lcc = &LocalCache{cacheSize:config.Cachesize}
	lcc.client = fastcache.New(int(lcc.cacheSize))
	return lcc,nil
}

func (lcc *LocalCache)Start()(err error){
	if nil != lcc.client{
		lcc.client.Reset()
		return nil
	}
	return CACHECLIENT_NIL
}

func (lcc *LocalCache)Stop()(err error){
	if nil != lcc.client{
		return nil
	}
	return CACHECLIENT_NIL
}

func (lcc *LocalCache)Get(key string)(value []byte,err error){
	if nil != lcc.client{
		value = lcc.client.Get(nil,utils.String2Bytes(key))
		return value,nil
	}
	return nil,CACHECLIENT_NIL
}

func (lcc *LocalCache)Set(key string,value []byte,expire uint32)(err error){
	if len(value) > CACHEITEMMAXSIZE{
		return CACHEMAXSIZE_OVER
	}
	if nil != lcc.client{
		lcc.client.Set(utils.String2Bytes(key),value)
		return nil
	}
	return CACHECLIENT_NIL
}

func (lcc *LocalCache)Delete(key string)(sta bool,err error){
	if nil != lcc.client{
		lcc.client.Del(utils.String2Bytes(key))
		return true,nil
	}
	return false,CACHECLIENT_NIL
}

func (lcc *LocalCache)Exists(key string)(sta bool,err error){
	if nil != lcc.client{
		sta = lcc.client.Has(utils.String2Bytes(key))
		return sta,nil
	}
	return false,CACHECLIENT_NIL
}